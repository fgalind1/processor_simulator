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
	alu *alu
}

type alu struct {
	Result uint32
	Status uint32
}

func New(bus *storagebus.StorageBus) *Alu {
	return &Alu{
		bus: bus, alu: &alu{}}
}

func (this *Alu) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *Alu) SetResult(result uint32) {
	this.alu.Result = result
}

func (this *Alu) CleanStatus() {
	this.alu.Status = 0x00
}

func (this *Alu) SetStatusFlag(value bool, flag uint8) {
	if value {
		this.alu.Status += 0x01 << flag
	}
}

func (this *Alu) Result() uint32 {
	return this.alu.Result
}

func (this *Alu) Status() uint32 {
	return this.alu.Status
}

func (this *Alu) Process(instruction *instruction.Instruction) error {

	// Clean Status
	this.CleanStatus()

	outputAddress, err := this.compute(instruction.Info, instruction.Data)
	if err != nil {
		return err
	}
	logger.Print(" => [E]: [R%d(%#02X) = %#08X]", outputAddress, outputAddress*consts.BYTES_PER_WORD, this.Result())

	// Set status flags
	this.SetStatusFlag(this.Result()%2 == 0, consts.FLAG_PARITY)
	this.SetStatusFlag(this.Result() == 0, consts.FLAG_ZERO)
	this.SetStatusFlag(getSign(this.Result()), consts.FLAG_SIGN)

	// Persist status register and output data
	this.Bus().StoreRegister(outputAddress, this.Result())
	this.Bus().StoreRegister(consts.STATUS_REGISTER, this.Status())
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

func (this *Alu) compute(info *info.Info, operands interface{}) (uint32, error) {

	op1, op2, outputAddr, err := this.getOperands(info, operands)
	if err != nil {
		return 0, err
	}

	switch info.Opcode {
	case set.OP_ADD:
		value1 := this.Bus().LoadRegister(op1)
		value2 := this.Bus().LoadRegister(op2)
		this.SetResult(value1 + value2)
		this.SetStatusFlag(getSign(value1) == getSign(value2) && getSign(value1) != getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_ADDU:
		this.SetResult(this.Bus().LoadRegister(op1) + this.Bus().LoadRegister(op2))
	case set.OP_SUB:
		value1 := this.Bus().LoadRegister(op1)
		value2 := this.Bus().LoadRegister(op2)
		this.SetResult(value1 - value2)
		this.SetStatusFlag(getSign(value1) != getSign(value2) && getSign(value2) == getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_SUBU:
		this.SetResult(this.Bus().LoadRegister(op1) - this.Bus().LoadRegister(op2))
	case set.OP_ADDI:
		value1 := this.Bus().LoadRegister(op1)
		this.SetResult(value1 + op2)
		this.SetStatusFlag(getSign(value1) == getSign(op2) && getSign(value1) != getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_ADDIU:
		this.SetResult(this.Bus().LoadRegister(op1) + op2)
	case set.OP_CMP:
		val1, val2 := this.Bus().LoadRegister(op1), this.Bus().LoadRegister(op2)
		if val1 < val2 {
			this.SetResult(1)
		} else if val1 == val2 {
			this.SetResult(2)
		} else {
			this.SetResult(4)
		}
	default:
		return 0, errors.New(fmt.Sprintf("Invalid operation to process by Alu unit. Opcode: %d", info.Opcode))
	}
	return outputAddr, nil
}

func getSign(value uint32) bool {
	return (value >> 31) == 1
}
