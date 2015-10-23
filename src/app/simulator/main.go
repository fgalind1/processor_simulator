package main

import (
	"os"

	"github.com/codegangsta/cli"

	"app/logger"
	"app/simulator/components/processor"
	"app/simulator/components/translator"
	"app/simulator/consts"
)

const (
	RESGITERS_MEMORY_SIZE    = 32 * consts.BYTES_PER_WORD // 32 words
	DATA_MEMORY_SIZE         = 10 * 1024                  // 10KB
	INSTRUCTIONS_MEMORY_SIZE = 10 * 1024                  // 10KB
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
	}

	app.Run(os.Args)
}

func runAll(c *cli.Context) {

	if len(c.Args()) != 1 {
		logger.Error("Expecting <assembly-filename> and got %d parameters", len(c.Args()))
		os.Exit(1)
	}

	// Translate assembly file to hex file
	filename, err := translator.TranslateFromFile(c.Args()[0])
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Instanciate processor
	p, err := processor.New(filename, uint32(INSTRUCTIONS_MEMORY_SIZE), uint32(DATA_MEMORY_SIZE), uint32(RESGITERS_MEMORY_SIZE))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	for {
		result := p.RunNext()
		if result == consts.INSTRUCTION_FAIL {
			os.Exit(1)
		} else if result == consts.INSTRUCTION_REACHED_END {
			break
		}
	}
}
