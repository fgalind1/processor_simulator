package branch

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/operation"
	"app/simulator/processor/models/set"
)

type Branch struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *Branch {
	return &Branch{bus: bus}
}

func (this *Branch) Bus() *storagebus.StorageBus {
	return this.bus
}

func ComputeOffsetTypeI(operands data.Data) int32 {
	immediate := operands.(*data.DataI).Immediate.ToUint32()
	// Cast as signed int32 so it gets performed Two's complement (if negative)
	offsetAddress := int32(immediate<<16) >> 16
	// Transform offset from N instructions domain to N bytes domain (2 bytes per instruction)
	return offsetAddress << 2
}

func ComputeAddressTypeJ(operands data.Data) uint32 {
	// Transform address from N instructions domain to N bytes domain (2 bytes per instruction)
	return operands.(*data.DataJ).Address.ToUint32() << 2
}

func (this *Branch) Process(operation *operation.Operation) error {
	instruction := operation.Instruction()
	info := instruction.Info
	operands := instruction.Data

	switch info.Type {
	case data.TypeI:
		registerD := operands.(*data.DataI).RegisterD.ToUint32()
		op1 := this.Bus().LoadRegister(registerD)
		registerS := operands.(*data.DataI).RegisterS.ToUint32()
		op2 := this.Bus().LoadRegister(registerS)
		logger.Collect(" => [E][%03d]: [R%d(%#02X) = %#08X ? R%d(%#02X) = %#08X]",
			operation.Id(), registerD, registerD*consts.BYTES_PER_WORD, op1, registerS, registerS*consts.BYTES_PER_WORD, op2)

		taken, err := processOperation(op1, op2, info.Opcode)
		if err != nil {
			return err
		}
		this.Bus().SetBranchResult(operation.Address(), taken)

		if taken {
			offsetAddress := ComputeOffsetTypeI(operands)
			this.Bus().IncrementProgramCounter(operation, offsetAddress)
			logger.Collect(" => [E][%03d]: [PC(offset) = 0x%06X", operation.Id(), offsetAddress)
		} else {
			// Notify ROB
			this.Bus().IncrementProgramCounter(operation, 0)
		}
	case data.TypeJ:
		address := ComputeAddressTypeJ(operands)
		this.Bus().SetProgramCounter(operation, uint32(address-consts.BYTES_PER_WORD))
		logger.Collect(" => [E][%03d]: [Address = %06X]", operation.Id(), address)
	default:
		return errors.New(fmt.Sprintf("Invalid data type to process by Branch unit. Type: %d", info.Type))
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
		return false, errors.New(fmt.Sprintf("Invalid operation to process by Branch unit. Opcode: %d", opcode))
	}
}
