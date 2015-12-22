package fpu

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
	"app/simulator/standards/ieee754"
)

type Fpu struct {
	bus *storagebus.StorageBus
	fpu *fpu
}

type fpu struct {
	Result uint32
}

func New(bus *storagebus.StorageBus) *Fpu {
	return &Fpu{
		bus: bus, fpu: &fpu{}}
}

func (this *Fpu) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *Fpu) SetResult(result uint32) {
	this.fpu.Result = result
}

func (this *Fpu) Result() uint32 {
	return this.fpu.Result
}

func (this *Fpu) Process(operation *operation.Operation) (*operation.Operation, error) {
	instruction := operation.Instruction()
	outputAddress, err := this.compute(operation, instruction.Info, instruction.Data)
	if err != nil {
		return operation, err
	}
	logger.Collect(" => [FPU][%03d]: [R%d(%#02X) = %#08X]", operation.Id(), outputAddress, outputAddress*consts.BYTES_PER_WORD, this.Result())

	// Persist output data
	this.Bus().StoreRegister(operation, outputAddress, this.Result())
	return operation, nil
}

func (this *Fpu) getOperands(info *info.Info, operands interface{}) (uint32, uint32, uint32, error) {

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
		return 0, 0, 0, errors.New(fmt.Sprintf("Invalid data type to process by Fpu unit. Type: %d", info.Type))
	}
	return op1, op2, outputAddr, nil
}

func (this *Fpu) compute(op *operation.Operation, info *info.Info, operands interface{}) (uint32, error) {

	op1, op2, outputAddr, err := this.getOperands(info, operands)
	if err != nil {
		return 0, err
	}

	switch info.Opcode {
	case set.OP_FADD:
		val1 := this.Bus().LoadRegister(op, op1)
		val2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(ieee754.PackFloat754_32(ieee754.UnPackFloat754_32(val1) + ieee754.UnPackFloat754_32(val2)))
	case set.OP_FSUB:
		val1 := this.Bus().LoadRegister(op, op1)
		val2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(ieee754.PackFloat754_32(ieee754.UnPackFloat754_32(val1) - ieee754.UnPackFloat754_32(val2)))
	case set.OP_FMUL:
		val1 := this.Bus().LoadRegister(op, op1)
		val2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(ieee754.PackFloat754_32(ieee754.UnPackFloat754_32(val1) * ieee754.UnPackFloat754_32(val2)))
	case set.OP_FDIV:
		val1 := this.Bus().LoadRegister(op, op1)
		val2 := this.Bus().LoadRegister(op, op2)
		this.SetResult(ieee754.PackFloat754_32(ieee754.UnPackFloat754_32(val1) / ieee754.UnPackFloat754_32(val2)))
	default:
		return 0, errors.New(fmt.Sprintf("Invalid operation to process by FPU unit. Opcode: %d", info.Opcode))
	}
	return outputAddr, nil
}
