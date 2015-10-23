package alu

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/components/storagebus"
	"app/simulator/consts"
	"app/simulator/models/instruction/data"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
	"app/simulator/models/instruction/set"
)

type Alu struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *Alu {
	return &Alu{bus: bus}
}

func (this *Alu) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *Alu) Process(instruction *instruction.Instruction) error {
	result, outputAddress, err := this.compute(instruction.Info, instruction.Data)
	if err != nil {
		return err
	}
	logger.Print(" => [E]: [R%d(%#02X) = %#08X]", outputAddress, outputAddress*consts.BYTES_PER_WORD, result)
	this.Bus().StoreRegister(outputAddress, result)
	return nil
}

func (this *Alu) getOperands(info *info.Info, operands interface{}) (uint32, uint32, uint32, error) {

	var op1, op2, outputAddr uint32
	switch info.Type {
	case data.TypeR:
		op1 = operands.(*data.DataR).RegisterS.ToUint32()
		op2 = operands.(*data.DataR).RegisterT.ToUint32()
		outputAddr = operands.(*data.DataR).RegisterD.ToUint32()
	case data.TypeI:
		op1 = operands.(*data.DataI).RegisterS.ToUint32()
		op2 = operands.(*data.DataI).Immediate.ToUint32()
		outputAddr = operands.(*data.DataI).RegisterD.ToUint32()
	default:
		return 0, 0, 0, errors.New(fmt.Sprintf("Invalid data type to process by Alu unit. Type: %d", info.Type))
	}
	return op1, op2, outputAddr, nil
}

func (this *Alu) compute(info *info.Info, operands interface{}) (uint32, uint32, error) {

	op1, op2, outputAddr, err := this.getOperands(info, operands)
	if err != nil {
		return 0, 0, err
	}

	switch info.Opcode {
	case set.OP_ADD:
		return this.Bus().LoadRegister(op1) + this.Bus().LoadRegister(op2), outputAddr, nil
	case set.OP_ADDU:
		return this.Bus().LoadRegister(op1) + this.Bus().LoadRegister(op2), outputAddr, nil
	case set.OP_SUB:
		return this.Bus().LoadRegister(op1) - this.Bus().LoadRegister(op2), outputAddr, nil
	case set.OP_SUBU:
		return this.Bus().LoadRegister(op1) - this.Bus().LoadRegister(op2), outputAddr, nil
	case set.OP_ADDI:
		return this.Bus().LoadRegister(op1) + op2, outputAddr, nil
	case set.OP_ADDIU:
		return this.Bus().LoadRegister(op1) + op2, outputAddr, nil
	case set.OP_CMP:
		val1, val2 := this.Bus().LoadRegister(op1), this.Bus().LoadRegister(op2)
		if val1 < val2 {
			return 1, outputAddr, nil
		} else if val1 == val2 {
			return 2, outputAddr, nil
		} else {
			return 4, outputAddr, nil
		}
	default:
		return 0, 0, errors.New(fmt.Sprintf("Invalid operation to process by Alu unit. Opcode: %d", info.Opcode))
	}
}
