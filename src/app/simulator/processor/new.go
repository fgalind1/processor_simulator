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

			branchHistoryTable:    map[uint32]uint32{},
			conditionalBranches:   0,
			unconditionalBranches: 0,
			mispredictedBranches:  0,
			noTakenBranches:       0,
			branchPredictorBits:   0,
			speculativeJumps:      0,

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
		return p.processor.done && p.InstructionsFetchedCounter() == p.InstructionsCompletedCounter() && p.SpeculativeJumps() == 0
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

		// Reach pre-filled memory data lines
		if strings.Contains(line, "@0x") {
			parts := strings.Split(strings.Replace(line, "@0x", "", -1), ":")
			err = this.processPreFilledDataMemoryLine(parts[0], strings.Split(parts[1], " "))
			if err != nil {
				return errors.New(fmt.Sprintf("Failed parsing memory macro. %s", err.Error()))
			}
			continue
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

func (this *Processor) processPreFilledDataMemoryLine(startAddress string, values []string) error {

	bytesStartAddress, err := hex.DecodeString(startAddress)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed parsing memory macro (hex) value: %s. %s", startAddress, err.Error()))
	}
	address := uint32(0)
	for i := 0; i < len(bytesStartAddress); i++ {
		address += uint32(bytesStartAddress[len(bytesStartAddress)-1-i]) << (uint32(i) * 8)
	}

	for i := 0; i < len(values); i++ {
		if strings.TrimSpace(values[i]) == "" {
			continue
		}
		bytes, err := hex.DecodeString(values[i])
		if err != nil {
			return errors.New(fmt.Sprintf("Failed parsing memory macro (hex) value: %s. %s", values[i], err.Error()))
		}
		value := uint32(0)
		for i := 0; i < len(bytes); i++ {
			value += uint32(bytes[len(bytes)-1-i]) << (uint32(i) * 8)
		}
		this.DataMemory().StoreUint32(address, value)
		address += consts.BYTES_PER_WORD
	}
	return nil
}
