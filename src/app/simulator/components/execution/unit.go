package execution

import "app/simulator/models/instruction/instruction"

type Unit interface {
	Process(instruction *instruction.Instruction) error
}
