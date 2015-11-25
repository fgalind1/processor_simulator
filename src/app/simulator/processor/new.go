package processor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"app/logger"
	"app/simulator/processor/components/clock"
	"app/simulator/processor/components/memory"
	"app/simulator/processor/config"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/set"
	"app/utils"
)

func New(assemblyFileName string, config *config.Config) (*Processor, error) {

	p := &Processor{
		&processor{
			done:                  false,
			instructionsFetched:   []string{},
			instructionsCompleted: []uint32{},
			dataLog:               map[uint32][]LogEvent{},

			branchHistoryTable:    map[uint32]bool{},
			conditionalBranches:   0,
			unconditionalBranches: 0,
			mispredictedBranches:  0,

			instructionsMap: map[uint32]string{},
			instructionsSet: set.Init(),
			config:          config,

			programCounter:    0,
			registerMemory:    memory.New(config.RegistersMemorySize()),
			instructionMemory: memory.New(config.InstructionsMemorySize()),
			dataMemory:        memory.New(config.DataMemorySize()),
		},
	}

	logger.Print(config.ToString())

	// Instanciate functional units
	instructionsFinished := func() bool {
		return p.processor.done && p.InstructionsFetchedCounter() == p.InstructionsCompletedCounter()
	}
	p.processor.clockUnit = clock.New(config.CyclePeriod(), instructionsFinished)

	err := p.loadInstructionsMemory(assemblyFileName)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (this *Processor) loadInstructionsMemory(assemblyFileName string) error {

	logger.Print(" => Reading hex file: %s", assemblyFileName)
	lines, err := utils.ReadLines(assemblyFileName)
	if err != nil {
		return err
	}

	address := uint32(0)
	for _, line := range lines {

		// Split hex value and humand readable comment
		parts := strings.Split(line, "//")

		// Save human readable for debugging purposes
		if len(parts) > 1 {
			this.InstructionsMap()[address] = strings.TrimSpace(parts[1])
		}

		// Save hex value into instructions memory
		bytes, err := hex.DecodeString(strings.TrimSpace(parts[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Failed parsing instruction (hex) value: %s. %s", parts[0], err.Error()))
		}
		this.InstructionsMemory().Store(address, bytes...)

		// Increment address
		address += uint32(len(bytes))
	}
	for i := address; i < this.InstructionsMemory().Size(); i++ {
		this.InstructionsMemory().Store(i, []byte{consts.ENDING_BYTE}...)
	}
	return nil
}
