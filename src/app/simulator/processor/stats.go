package processor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"app/logger"
	"app/simulator/processor/config"
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
	logger.SetVerboseQuiet(false)
	stats := "\n Program Stats:\n\n"
	stats += fmt.Sprintf(" => Instructions found: %d\n", len(this.InstructionsMap()))
	stats += fmt.Sprintf(" => Instructions executed: %d\n", this.InstructionsCompletedCounter())
	stats += fmt.Sprintf(" => Cycles performed: %d\n", this.Cycles())
	stats += fmt.Sprintf(" => Cycles per instruction: %3.2f cycles\n", float32(this.Cycles())/float32(this.InstructionsCompletedCounter()))
	stats += fmt.Sprintf(" => Simulation duration: %d ms\n", this.DurationMs())
	stats += fmt.Sprintf("\n")
	totalBranches := this.processor.conditionalBranches + this.processor.unconditionalBranches
	stats += fmt.Sprintf(" => Total Branches: %d\n", totalBranches)
	stats += fmt.Sprintf(" => Conditional Branches: %d\n", this.processor.conditionalBranches)
	stats += fmt.Sprintf(" => Unconditional Branches: %d\n", this.processor.unconditionalBranches)
	if this.Config().BranchPredictorType() != config.StallPredictor {
		stats += fmt.Sprintf(" => Mispredicted Branches: %d\n", this.processor.mispredictedBranches)
		stats += fmt.Sprintf(" => Misprediction Percentage: %3.2f\n", 100*float32(this.processor.mispredictedBranches)/float32(totalBranches))
	}
	return stats
}

func (this *Processor) PipelineFlow() string {

	instructions := this.InstructionsFetchedCounter()
	str := "\n------------- Pipeline Flow ---------------\n"

	// Create matrix
	str += fmt.Sprintf("   |%-30s", " - Instruction -")
	pipeline := make([][]string, instructions)
	for c := uint32(0); c < this.Cycles(); c++ {
		for i := range pipeline {
			pipeline[i] = append(pipeline[i], "   ")
		}
		str += fmt.Sprintf("|%03d", c)
	}
	str += "\n"

	// Iterate each pipeline stage
	for instructionId, events := range this.processor.dataLog {
		for _, event := range events {
			for c := event.Start; c < event.End; c++ {
				if uint32(len(pipeline)) > instructionId {
					pipeline[instructionId][c] = event.Id
				}
			}
		}
	}

	// Construct string matrix
	for i := range pipeline {
		instr := strings.Replace(this.InstructionsFetched()[i], "\t", " ", -1)
		for strings.Contains(instr, "  ") {
			instr = strings.Replace(instr, "  ", " ", -1)
		}
		str += fmt.Sprintf("%03d|%-30s|%s\n", i, instr, strings.Join(pipeline[i], "|"))
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

	// Save pipeline flow
	filename = filepath.Join(outputFolder, "pipeline.dat")
	err = ioutil.WriteFile(filename, []byte(this.PipelineFlow()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Pipeline flow saved at %s", filename)

	// Save stats
	filename = filepath.Join(outputFolder, "output.log")
	err = ioutil.WriteFile(filename, []byte(this.Config().ToString()+this.Stats()), 0644)
	if err != nil {
		return err
	}
	logger.Print(" => Output saved at %s", filename)

	return nil
}
