package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"

	"app/logger"
	"app/simulator/processor"
	"app/simulator/processor/config"
	"app/simulator/processor/consts"
	"app/simulator/translator"
)

func main() {
	app := cli.NewApp()
	app.Name = "Superscalar Processor Simulator"
	app.Usage = "University of Bristol\n   Repository: https://github.com/felipegs01/processor_simulator"
	app.HelpName = "simulator"
	app.Author = "Felipe Galindo Sanchez"
	app.Email = "felipegs01@gmail.com"
	app.Version = "1.0"
	app.Commands = []cli.Command{
		{
			Name:        "run",
			Usage:       "run <assembly-filename>",
			Description: "translate assembly file and run all instructions of an specified assembly program",
			Action:      runCommand,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "s, step-by-step",
					Usage: "Run interactively step by step",
				},
				cli.BoolFlag{
					Name:  "v, verbose",
					Usage: "Verbose on debug mode",
				},
				cli.StringFlag{
					Name:  "o, output-folder",
					Value: "",
					Usage: "Output folder where to store debug and memory files. if not provided output-folder will be on the same directory where the assembly-filename is",
				},
				cli.StringFlag{
					Name:  "c, config-filename",
					Value: "",
					Usage: "Processor config filename, if not provided a deafult config will be provided",
				},
				cli.IntFlag{
					Name:  "max-cycles",
					Value: 3000,
					Usage: "Maximum number of cycles to execute",
				},
			},
		},
	}

	app.Run(os.Args)
}

func runCommand(c *cli.Context) {

	printHeader()
	logger.SetVerboseDebug(c.Bool("verbose"))

	if len(c.Args()) != 1 {
		logger.Error("Expecting <assembly-filename> and got %d parameters", len(c.Args()))
		os.Exit(1)
	}

	assemblyFilename, _ := filepath.Abs(c.Args()[0])
	if _, err := os.Stat(assemblyFilename); os.IsNotExist(err) {
		logger.Error("File %s does not exists", assemblyFilename)
		os.Exit(1)
	}

	outputFolder, _ := filepath.Abs(c.String("output-folder"))
	if outputFolder == "" {
		outputFolder = filepath.Join(filepath.Dir(assemblyFilename), getFileName(assemblyFilename))
	}

	configFilename, _ := filepath.Abs(c.String("config-filename"))
	if configFilename == "" {
		logger.Error("Configuration file not provided, please provide a valid configuration file")
		os.Exit(1)
	}

	cfg, err := config.Load(configFilename)
	if err != nil {
		logger.Error("Failed loading config. %s", err.Error())
		os.Exit(1)
	}
	logger.Print(" => Configuration file: %s", configFilename)

	err = runProgram(assemblyFilename, c.Bool("step-by-step"), outputFolder, cfg, uint32(c.Int("max-cycles")))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runProgram(assemblyFilename string, interactive bool, outputFolder string, config *config.Config, maxCycles uint32) error {

	err := os.MkdirAll(outputFolder, 0777)
	if err != nil {
		return err
	}

	// Translate assembly file to hex file
	hexFilename, err := translator.TranslateFromFile(assemblyFilename, filepath.Join(outputFolder, "assembly.hex"))
	if err != nil {
		return err
	}

	// Instanciate processor
	p, err := processor.New(hexFilename, config)
	if err != nil {
		return err
	}

	// Start simulation
	p.Start()

	// Run as many instructions as they are
	result := consts.PROGRAM_RUNNING
	for result == consts.PROGRAM_RUNNING {
		p.PauseClock()
		if interactive {
			for runInteractiveStep(p) {
			}
		}
		// Unpause and execute next cycle
		p.ContinueClock()
		result = p.NextCycle()
		// If max cycles option selected
		if maxCycles > 0 && p.Cycles() >= maxCycles {
			p.PauseClock()
			break
		}
	}

	logger.Print(p.Stats())
	return p.SaveOutputFiles(outputFolder)
}

func getFileName(filename string) string {
	filename = filepath.Base(filename)
	extension := filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

func runInteractiveStep(p *processor.Processor) bool {
	// Small sleep to allow pipeline stages messages to be printed first in console
	time.Sleep(consts.MENU_DELAY)

	// Display menu
	fmt.Println("Press the desired key and then hit [ENTER]...")
	fmt.Println(" - (R) to see registers memory")
	fmt.Println(" - (D) to see data memory")
	fmt.Println(" - (E) to exit and quit")
	fmt.Println(" - (*) Any other key to continue...")
	fmt.Println("Option: ")

	var option string
	fmt.Scan(&option)

	switch option {
	case "R", "r":
		fmt.Println(p.RegistersMemory().ToString())
		fmt.Println("--------------------------------------------")
	case "D", "d":
		fmt.Println(p.DataMemory().ToString())
		fmt.Println("--------------------------------------------")
	case "E", "e":
		os.Exit(0)
	default:
		return false
	}
	return true
}

func printHeader() {
	logger.Print("")
	logger.Print("################################################")
	logger.Print("#     Superscalar Processor Simulator v1.0     #")
	logger.Print("################################################")
	logger.Print("")
}
