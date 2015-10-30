package control

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/components/storagebus"
	"app/simulator/consts"
	"app/simulator/models/instruction/data"
	"app/simulator/models/instruction/instruction"
	"app/simulator/models/instruction/set"
)

type Control struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *Control {
	return &Control{bus: bus}
}

func (this *Control) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *Control) Process(instruction *instruction.Instruction) error {
	info := instruction.Info
	operands := instruction.Data

	switch info.Type {
	case data.TypeI:
		registerD := operands.(*data.DataI).RegisterD.ToUint32()
		op1 := this.Bus().LoadRegister(registerD)
		registerS := operands.(*data.DataI).RegisterS.ToUint32()
		op2 := this.Bus().LoadRegister(registerS)
		immediate := operands.(*data.DataI).Immediate.ToUint32()
		logger.Print(" => [E]: [R%d(%#02X) = %#08X ? R%d(%#02X) = %#08X]",
			registerD, registerD*consts.BYTES_PER_WORD, op1, registerS, registerS*consts.BYTES_PER_WORD, op2)

		doBranch, err := processOperation(op1, op2, info.Opcode)
		if err != nil {
			return err
		}

		if doBranch {
			// Cast as signed int32 so it gets performed Two's complement (if negative)
			offsetAddress := int32(immediate<<16) >> 16
			// Transform offset from N instructions domain to N bytes domain (2 bytes per instruction)
			offsetAddress = offsetAddress << 2
			// Increment program counter
			this.Bus().IncrementProgramCounter(offsetAddress)
			logger.Print(" => [E]: [PC(offset) = %d]", offsetAddress)
		}
	case data.TypeJ:
		// Transform address from N instructions domain to N bytes domain (2 bytes per instruction)
		address := operands.(*data.DataJ).Address.ToUint32() << 2
		// Decrement 4 bytes so the automatic +4 process set the desired address
		this.Bus().SetProgramCounter(address - 4)
		logger.Print(" => [E]: [Address = %06X]", address)
	default:
		return errors.New(fmt.Sprintf("Invalid data type to process by Control unit. Type: %d", info.Type))
	}
	return nil
}

func processOperation(registerD uint32, registerS uint32, opcode uint8) (bool, error) {
	switch opcode {
	case set.OP_BEQ:
		return registerD == registerS, nil
	case set.OP_BNE:
		return registerD != registerS, nil
	case set.OP_BLT:
		return registerD < registerS, nil
	case set.OP_BGT:
		return registerD > registerS, nil
	default:
		return false, errors.New(fmt.Sprintf("Invalid operation to process by Control unit. Opcode: %d", opcode))
	}
}
