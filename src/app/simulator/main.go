package main

import (
	"errors"
	"fmt"
	"os"

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
			Name:        "run-all",
			Usage:       "run-all <assembly-filename>",
			Description: "translate assembly file and run all instructions of an specified assembly program",
			Action:      runAll,
		},
		{
			Name:        "run-step",
			Usage:       "run-step <assembly-filename>",
			Description: "translate assembly file and run instructions interactively step by step",
			Action:      runStep,
		},
	}

	app.Run(os.Args)
}

func runAll(c *cli.Context) {

	if len(c.Args()) != 1 {
		logger.Error("Expecting <assembly-filename> and got %d parameters", len(c.Args()))
		os.Exit(1)
	}

	err := runProgram(c.Args()[0], false)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runStep(c *cli.Context) {

	if len(c.Args()) != 1 {
		logger.Error("Expecting <assembly-filename> and got %d parameters", len(c.Args()))
		os.Exit(1)
	}

	err := runProgram(c.Args()[0], true)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func runProgram(assemblyFilename string, interactive bool) error {

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

	for {
		result := p.RunNext()
		if result == consts.INSTRUCTION_FAIL {
			return errors.New("Executing instruction failed")
		} else if result == consts.INSTRUCTION_REACHED_END {
			break
		}
		if interactive {
			for runInteractiveStep(p) {
			}
		}
	}
	runInteractiveStep(p)
	return nil
}

func runInteractiveStep(p *processor.Processor) bool {
	fmt.Printf("Press the desired key and then hit [ENTER]...\n")
	fmt.Printf(" - (R) to see registers memory\n")
	fmt.Printf(" - (D) to see data memory\n")
	fmt.Printf(" - (E) to exit and quit\n")
	fmt.Printf(" - (*) Any other key to continue to the next step\n")
	fmt.Printf("Option: ")

	var option string
	fmt.Scan(&option)

	switch option {
	case "R", "r":
		fmt.Printf(p.RegistersMemory().ToString())
	case "D", "d":
		fmt.Printf(p.DataMemory().ToString())
	case "E", "e":
		os.Exit(0)
	default:
		return false
	}
	return true
}
