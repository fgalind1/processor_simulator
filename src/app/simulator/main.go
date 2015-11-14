package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"

	"app/logger"
	"app/simulator/components/processor"
	"app/simulator/components/translator"
	"app/simulator/config"
	"app/simulator/consts"
)

const (
	CYCLE_PERIOD_MS = 50

	REGISTERS_MEMORY_SIZE    = 32 * 4   // 32 words (4 bytes each)
	INSTRUCTIONS_MEMORY_SIZE = 1 * 1024 // 1KB (254 words)
	DATA_MEMORY_SIZE         = 1 * 1024 // 1KB (254 words)

	INSTRUCTIONS_BUFFER_SIZE = 2
	DECODER_UNITS            = 2
	BRANCH_UNITS             = 1
	LOAD_STORE_UNITS         = 1
	ALU_UNITS                = 1
)

func getConfig() *config.Config {
	return config.New(
		CYCLE_PERIOD_MS,
		REGISTERS_MEMORY_SIZE,
		INSTRUCTIONS_MEMORY_SIZE,
		DATA_MEMORY_SIZE,
		INSTRUCTIONS_BUFFER_SIZE,
		DECODER_UNITS,
		BRANCH_UNITS,
		LOAD_STORE_UNITS,
		ALU_UNITS)
}

func main() {
	app := cli.NewApp()
	app.Name = "Superscalar Processor Simulator"
	app.Usage = "University of Bristol"
	app.HelpName = "simulator"
	app.Author = "Felipe Galindo Sanchez"
	app.Email = "felipegs01@gmail.com"
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
			},
		},
	}

	app.Run(os.Args)
}

func runCommand(c *cli.Context) {

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

	cfg := getConfig()
	configFilename, _ := filepath.Abs(c.String("config-filename"))
	if configFilename != "" {
		var err error
		cfg, err = config.Load(configFilename)
		if err != nil {
			logger.Error("Failed loading config. %s", err.Error())
			os.Exit(1)
		}
	}

	err := runProgram(assemblyFilename, c.Bool("step-by-step"), outputFolder, cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runProgram(assemblyFilename string, interactive bool, outputFolder string, config *config.Config) error {

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
	logger.Print("Press the desired key and then hit [ENTER]...")
	logger.Print(" - (R) to see registers memory")
	logger.Print(" - (D) to see data memory")
	logger.Print(" - (E) to exit and quit")
	logger.Print(" - (*) Any other key to continue...")
	fmt.Print("Option: ")

	var option string
	fmt.Scan(&option)

	switch option {
	case "R", "r":
		logger.Print(p.RegistersMemory().ToString())
		logger.Print("--------------------------------------------")
	case "D", "d":
		logger.Print(p.DataMemory().ToString())
		logger.Print("--------------------------------------------")
	case "E", "e":
		os.Exit(0)
	default:
		return false
	}
	return true
}
