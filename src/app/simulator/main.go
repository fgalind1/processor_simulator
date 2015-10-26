package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"

	"app/logger"
	"app/simulator/components/processor"
	"app/simulator/components/translator"
	"app/simulator/consts"
)

const (
	REGISTERS_MEMORY_SIZE    = 32 * consts.BYTES_PER_WORD // 32 words
	INSTRUCTIONS_MEMORY_SIZE = 1 * 1024                   // 1KB (254 words)
	DATA_MEMORY_SIZE         = 1 * 1024                   // 1KB (254 words)
)

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
					Name:  "d, data-memory",
					Value: "",
					Usage: "The filename where to save the data memory once the program has finished",
				},
				cli.StringFlag{
					Name:  "r, registers-memory",
					Value: "",
					Usage: "The filename where to save the registers memory once the program has finished",
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

	err := runProgram(c.Args()[0], c.Bool("step-by-step"), c.String("data-memory"), c.String("registers-memory"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runProgram(assemblyFilename string, interactive bool, dataFile, registersFile string) error {

	// Translate assembly file to hex file
	hexFilename, err := translator.TranslateFromFile(assemblyFilename)
	if err != nil {
		return err
	}

	// Instanciate processor
	p, err := processor.New(hexFilename, uint32(REGISTERS_MEMORY_SIZE), uint32(INSTRUCTIONS_MEMORY_SIZE), uint32(DATA_MEMORY_SIZE))
	if err != nil {
		return err
	}

	// Run as many instructions as they are
	for {
		result := p.RunNext()
		if result == consts.INSTRUCTION_FAIL {
			return errors.New("Executing instruction failed")
		} else if result == consts.INSTRUCTION_REACHED_END {
			break
		}
		if interactive {
			for runInteractiveStep(p, true) {
			}
		}
	}

	// Ask at the end to see the last state of the memory (if desired)
	runInteractiveStep(p, false)

	// Save memory (if selected)
	if dataFile != "" {
		dataFile, _ = filepath.Abs(dataFile)
		err = ioutil.WriteFile(dataFile, []byte(p.DataMemory().ToString()), 0644)
		logger.Print(" => Data memory saved at %s", dataFile)
		if err != nil {
			return err
		}
	}

	if registersFile != "" {
		registersFile, _ = filepath.Abs(registersFile)
		err = ioutil.WriteFile(registersFile, []byte(p.RegistersMemory().ToString()), 0644)
		logger.Print(" => Registers memory saved at %s", dataFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func runInteractiveStep(p *processor.Processor, showExit bool) bool {
	logger.Print("Press the desired key and then hit [ENTER]...")
	logger.Print(" - (R) to see registers memory")
	logger.Print(" - (D) to see data memory")
	if showExit {
		logger.Print(" - (E) to exit and quit")
	}
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
		if showExit {
			os.Exit(0)
		} else {
			return false
		}
	default:
		return false
	}
	return true
}
