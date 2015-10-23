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
			instructionsMap: map[uint32]string{},
			instructionsSet: set.Init(),

			programCounter:    0,
			registerMemory:    memory.New(registersSize),
			instructionMemory: memory.New(instructionsSize),
			dataMemory:        memory.New(dataSize),
		},
	}

	bus := &storagebus.StorageBus{
		LoadRegister: func(address uint32) uint32 {
			return p.RegistersMemory().LoadUint32(address * consts.BYTES_PER_WORD)
		},
		StoreRegister: func(address, value uint32) {
			p.RegistersMemory().StoreUint32(address*consts.BYTES_PER_WORD, value)
		},

		LoadData:  p.DataMemory().LoadUint32,
		StoreData: p.DataMemory().StoreUint32,
	}

	p.processor.aluUnit = alu.New(bus)
	p.processor.dataUnit = data.New(bus)

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
		parts := strings.Split(line, ";")

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

func (this *Processor) IncrementProgramCounter() {
	this.processor.programCounter += consts.BYTES_PER_WORD
}
