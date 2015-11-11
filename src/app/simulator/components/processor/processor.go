package processor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"app/logger"
	"app/simulator/components/execution/alu"
	"app/simulator/components/execution/control"
	"app/simulator/components/execution/data"
	"app/simulator/components/memory"
	"app/simulator/components/storagebus"
	"app/simulator/consts"
	"app/simulator/models/instruction/set"
	"app/utils"
)

type Processor struct {
	processor *processor
}

type processor struct {
	// internals
	cycles               uint32
	instructionsExecuted uint32
	fetchState           bool
	decodeState          bool
	executeState         bool
	dataLog              map[string][]uint32

	// metadata
	instructionsMap map[uint32]string
	instructionsSet set.Set

	// execution units
	aluUnit     *alu.Alu
	dataUnit    *data.Data
	controlUnit *control.Control

	// data/memory
	programCounter    uint32
	registerMemory    *memory.Memory
	instructionMemory *memory.Memory
	dataMemory        *memory.Memory
}

func New(filename string, registersSize, instructionsSize, dataSize uint32) (*Processor, error) {

	p := &Processor{
		&processor{
			cycles:               0,
			instructionsExecuted: 0,
			fetchState:           consts.ACTIVE,
			decodeState:          consts.ACTIVE,
			executeState:         consts.ACTIVE,
			dataLog:              map[string][]uint32{},

			instructionsMap: map[uint32]string{},
			instructionsSet: set.Init(),

			programCounter:    0,
			registerMemory:    memory.New(registersSize),
			instructionMemory: memory.New(instructionsSize),
			dataMemory:        memory.New(dataSize),
		},
	}

	logger.Print(" => Bytes per word: %d", consts.BYTES_PER_WORD)
	logger.Print(" => Registers:      %d", registersSize/consts.BYTES_PER_WORD)
	logger.Print(" => Instr Memory:   %d Bytes", instructionsSize)
	logger.Print(" => Data Memory:    %d Bytes", dataSize)

	bus := &storagebus.StorageBus{
		LoadRegister: func(address uint32) uint32 {
			return p.RegistersMemory().LoadUint32(address * consts.BYTES_PER_WORD)
		},
		StoreRegister: func(address, value uint32) {
			p.RegistersMemory().StoreUint32(address*consts.BYTES_PER_WORD, value)
		},

		LoadData:  p.DataMemory().LoadUint32,
		StoreData: p.DataMemory().StoreUint32,

		SetProgramCounter:       p.SetProgramCounter,
		IncrementProgramCounter: p.IncrementProgramCounter,
	}

	p.processor.aluUnit = alu.New(bus)
	p.processor.dataUnit = data.New(bus)
	p.processor.controlUnit = control.New(bus)

	err := p.loadInstructionsMemory(filename)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (this *Processor) loadInstructionsMemory(filename string) error {

	logger.Print(" => Reading hex file: %s", filename)
	lines, err := utils.ReadLines(filename)
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
	this.InstructionsMemory().Store(address, []byte{consts.ENDING_BYTE, consts.ENDING_BYTE, consts.ENDING_BYTE, consts.ENDING_BYTE}...)
	return nil
}

///////////////////////////
//       Internals       //
///////////////////////////

func (this *Processor) Cycles() uint32 {
	return this.processor.cycles
}

func (this *Processor) InstructionsExecuted() uint32 {
	return this.processor.instructionsExecuted
}

func (this *Processor) IncrementCycles() {
	this.processor.cycles += 1
}

func (this *Processor) IncrementInstructionsExecuted() {
	this.processor.instructionsExecuted += 1
}

func (this *Processor) LogEvent(event string) {
	this.processor.dataLog[event] = append(this.processor.dataLog[event], this.Cycles())
}

///////////////////////////
//       Metadata        //
///////////////////////////

func (this *Processor) InstructionsMap() map[uint32]string {
	return this.processor.instructionsMap
}

func (this *Processor) InstructionsSet() set.Set {
	return this.processor.instructionsSet
}

///////////////////////////
//    Execution Units    //
///////////////////////////

func (this *Processor) AluUnit() *alu.Alu {
	return this.processor.aluUnit
}

func (this *Processor) DataUnit() *data.Data {
	return this.processor.dataUnit
}

func (this *Processor) ControlUnit() *control.Control {
	return this.processor.controlUnit
}

///////////////////////////
//      Data/Memory      //
///////////////////////////

func (this *Processor) DataMemory() *memory.Memory {
	return this.processor.dataMemory
}

func (this *Processor) InstructionsMemory() *memory.Memory {
	return this.processor.instructionMemory
}

func (this *Processor) RegistersMemory() *memory.Memory {
	return this.processor.registerMemory
}

func (this *Processor) ProgramCounter() uint32 {
	return this.processor.programCounter
}

func (this *Processor) SetProgramCounter(value uint32) {
	this.processor.programCounter = value
}

func (this *Processor) IncrementProgramCounter(offset int32) {
	if offset < 0 {
		this.processor.programCounter -= uint32(offset * -1)
	} else {
		this.processor.programCounter += uint32(offset)
	}
}

///////////////////////////
//         Stats         //
///////////////////////////

func (this *Processor) Stats() string {
	stats := "\n------------- Program Stats ---------------\n"
	stats += fmt.Sprintf(" => Instructions available: %d\n", len(this.InstructionsMap()))
	stats += fmt.Sprintf(" => Instructions executed: %d\n", this.InstructionsExecuted())
	stats += fmt.Sprintf(" => Cycles performed: %d\n", this.Cycles())
	stats += "-------------------------------------------\n"
	return stats
}

func (this *Processor) PipelineFlow() string {

	str := "\n------------- Pipeline Flow ---------------\n"

	// Create matrix
	str += "      "
	pipeline := make([][]string, this.InstructionsExecuted())
	for c := uint32(0); c < this.Cycles(); c++ {
		for i := range pipeline {
			pipeline[i] = append(pipeline[i], " ")
		}
		str += fmt.Sprintf("|%03d", c)
	}
	str += "\n"

	// Iterate each pipeline stage
	for stage, array := range this.processor.dataLog {
		for i := 0; i < len(array)/2; i += 1 {
			startCycle := array[i*2]
			endCycle := array[i*2+1]
			if endCycle > startCycle {
				for c := startCycle; c < endCycle; c++ {
					pipeline[i][c] = stage
				}
			}
		}
	}

	// Construct string matrix

	for i := range pipeline {
		str += fmt.Sprintf(" I%03d | %s\n", i, strings.Join(pipeline[i], " | "))
	}
	return str
}
