package alu

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/operation"
	"app/simulator/processor/models/set"
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

func (this *Alu) Process(operation *operation.Operation) (*operation.Operation, error) {
	instruction := operation.Instruction()

	// Clean Status
	this.CleanStatus()

	outputAddress, err := this.compute(operation, instruction.Info, instruction.Data)
	if err != nil {
		return operation, err
	}
	logger.Collect(" => [ALU][%03d]: [R%d(%#02X) = %#08X]", operation.Id(), outputAddress, outputAddress*consts.BYTES_PER_WORD, this.Result())

	// Set status flags
	this.SetStatusFlag(this.Result()%2 == 0, consts.FLAG_PARITY)
	this.SetStatusFlag(this.Result() == 0, consts.FLAG_ZERO)
	this.SetStatusFlag(getSign(this.Result()), consts.FLAG_SIGN)

	// Persist output data
	this.Bus().StoreRegister(operation, outputAddress, this.Result())
	return operation, nil
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

func (this *Alu) compute(op *operation.Operation, info *info.Info, operands interface{}) (uint32, error) {

	op1, op2, outputAddr, err := this.getOperands(info, operands)
	if err != nil {
		return 0, err
	}

	switch info.Opcode {
	case set.OP_ADD:
		// Arithmetic
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 + value2)
		this.SetStatusFlag(getSign(value1) == getSign(value2) && getSign(value1) != getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_ADDI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 + op2)
		this.SetStatusFlag(getSign(value1) == getSign(op2) && getSign(value1) != getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_ADDU:
		this.SetResult(this.Bus().LoadRegister(op, op1) + this.Bus().LoadRegister(op, op2))
	case set.OP_ADDIU:
		this.SetResult(this.Bus().LoadRegister(op, op1) + op2)
	case set.OP_SUB:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 - value2)
		this.SetStatusFlag(getSign(value1) != getSign(value2) && getSign(value2) == getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_SUBI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 - op2)
		this.SetStatusFlag(getSign(value1) != getSign(op2) && getSign(op2) == getSign(this.Result()), consts.FLAG_OVERFLOW)
	case set.OP_SUBU:
		this.SetResult(this.Bus().LoadRegister(op, op1) - this.Bus().LoadRegister(op, op2))
	case set.OP_MUL:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 * value2)
	// Bitwise Shifts
	case set.OP_SHL:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 << value2)
	case set.OP_SHLI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 << op2)
	case set.OP_SHR:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 >> value2)
	case set.OP_SHRI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 >> op2)
	// Logical
	case set.OP_CMP:
		val1, val2 := this.Bus().LoadRegister(op, op1), this.Bus().LoadRegister(op, op2)
		if val1 < val2 {
			this.SetResult(1)
		} else if val1 == val2 {
			this.SetResult(2)
		} else {
			this.SetResult(4)
		}
	case set.OP_AND:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 & value2)
	case set.OP_ANDI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 & op2)
	case set.OP_OR:
		value1 := this.Bus().LoadRegister(op, op1)
		value2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(value1 | value2)
	case set.OP_ORI:
		value1 := this.Bus().LoadRegister(op, op1)
		this.SetResult(value1 | op2)
	default:
		return 0, errors.New(fmt.Sprintf("Invalid operation to process by Alu unit. Opcode: %d", info.Opcode))
	}
	return outputAddr, nil
}

func getSign(value uint32) bool {
	return (value >> 31) == 1
}
