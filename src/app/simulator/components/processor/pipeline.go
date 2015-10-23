package processor

import (
	"errors"

	"app/simulator/consts"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
)

func (this *Processor) Fetch(address uint32) ([]byte, error) {
	return this.InstructionsMemory().Load(address, consts.BYTES_PER_WORD), nil
}

func (this *Processor) Decode(data []byte) (*instruction.Instruction, error) {
	return this.InstructionsSet().GetInstructionFromBytes(data)
}

func (this *Processor) Execute(instruction *instruction.Instruction) error {
	switch instruction.Info.Category {
	case info.Aritmetic:
		return this.AluUnit().Process(instruction)
	case info.Data:
		return this.DataUnit().Process(instruction)
	case info.Control:
		return this.ControlUnit().Process(instruction)
	default:
		return errors.New("Invalid instruction info category")
	}
}
