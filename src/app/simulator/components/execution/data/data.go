package data

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

type Data struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *Data {
	return &Data{bus: bus}
}

func (this *Data) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *Data) Process(instruction *instruction.Instruction) error {

	rdAddress := instruction.Data.(*data.DataI).RegisterD.ToUint32()
	rsAddress := instruction.Data.(*data.DataI).RegisterS.ToUint32()
	immediate := instruction.Data.(*data.DataI).Immediate.ToUint32()

	switch instruction.Info.Opcode {
	case set.OP_LW:
		rsValue := this.Bus().LoadRegister(rsAddress)
		value := this.Bus().LoadData(rsValue + immediate)
		this.Bus().StoreRegister(rdAddress, value)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, value)
	case set.OP_SW:
		rdValue := this.Bus().LoadRegister(rdAddress)
		rsValue := this.Bus().LoadRegister(rsAddress)
		this.Bus().StoreData(rsValue+immediate, rdValue)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rsValue+immediate, rdValue)
	case set.OP_LLI:
		this.Bus().StoreRegister(rdAddress, immediate)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate)
	case set.OP_SLI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(rdValue, immediate)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rdValue, immediate)
	case set.OP_LUI:
		this.Bus().StoreRegister(rdAddress, immediate<<16)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate<<16)
	case set.OP_SUI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(rdValue, immediate<<16)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rdValue, immediate<<16)
	default:
		return errors.New(fmt.Sprintf("Invalid operation to process by Data unit. Opcode: %d", instruction.Info.Opcode))
	}
	return nil
}
