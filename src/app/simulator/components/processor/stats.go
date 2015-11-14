package processor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"app/logger"
)

type LogEvent struct {
	Id    string
	Start uint32
	End   uint32
}

func NewEvent(id string, start uint32, end uint32) LogEvent {
	return LogEvent{Id: id, Start: start, End: end}
}

func (this *Processor) Stats() string {
	stats := "\n Program Stats:\n\n"
	stats += fmt.Sprintf(" => Instructions found: %d\n", len(this.InstructionsMap()))
	stats += fmt.Sprintf(" => Instructions executed: %d\n", len(this.InstructionsCompleted()))
	stats += fmt.Sprintf(" => Cycles performed: %d\n", this.Cycles())
	stats += fmt.Sprintf(" => Simulation duration: %d ms\n", this.DurationMs())
	return stats
}

func (this *Processor) PipelineFlow() string {

	instructions := len(this.InstructionsCompleted())
	str := "\n------------- Pipeline Flow ---------------\n"

	// Create matrix
	str += fmt.Sprintf("%-30s", " - Instruction -")
	pipeline := make([][]string, instructions)
	for c := uint32(0); c < this.Cycles(); c++ {
		for i := range pipeline {
			pipeline[i] = append(pipeline[i], "   ")
		}
		str += fmt.Sprintf("|%03d", c)
	}
	str += "\n"

	// Iterate each pipeline stage
	for _, array := range this.processor.dataLog {
		for i := 0; i < instructions; i += 1 {
			startCycle := array[i].Start
			endCycle := array[i].End
			if endCycle > startCycle {
				for c := startCycle; c < endCycle; c++ {
					pipeline[i][c] = array[i].Id
				}
			}
		}
	}

	// Construct string matrix
	for i := range pipeline {
		instr := strings.Replace(this.InstructionsCompleted()[i], "\t", " ", -1)
		for strings.Contains(instr, "  ") {
			instr = strings.Replace(instr, "  ", " ", -1)
		}
		str += fmt.Sprintf("%-30s|%s\n", instr, strings.Join(pipeline[i], "|"))
	}
	return str
}

func (this *Processor) SaveOutputFiles(outputFolder string) error {

	logger.Print("\n Output Files:\n")

	// Save debug buffer
	filename := filepath.Join(outputFolder, "debug.log")
	err := logger.WriteBuffer(filename)
	if err != nil {
		return err
	}
	logger.Print(" => Debug buffer saved at %s", filename)

	// Save memory file
	filename = filepath.Join(outputFolder, "memory.dat")
	err = ioutil.WriteFile(filename, []byte(this.DataMemory().ToString()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Data memory saved at %s", filename)

	// Save registers file
	filename = filepath.Join(outputFolder, "registers.dat")
	err = ioutil.WriteFile(filename, []byte(this.RegistersMemory().ToString()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Registers memory saved at %s", filename)

	// Save stats
	filename = filepath.Join(outputFolder, "stats.log")
	err = ioutil.WriteFile(filename, []byte(this.Stats()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Stats saved at %s", filename)

	// Save pipeline flow
	filename = filepath.Join(outputFolder, "pipeline.dat")
	err = ioutil.WriteFile(filename, []byte(this.PipelineFlow()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Pipeline flow saved at %s", filename)

	return nil
}
