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
		value := this.Bus().LoadData(rsAddress + immediate)
		this.Bus().StoreRegister(rdAddress, value)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, value)
	case set.OP_SW:
		value := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(rsAddress+immediate, value)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rsAddress+immediate, value)
	case set.OP_LLI:
		this.Bus().StoreRegister(rdAddress, immediate)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate)
	case set.OP_SLI:
		this.Bus().StoreData(rdAddress, immediate)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rdAddress, immediate)
	case set.OP_LUI:
		this.Bus().StoreRegister(rdAddress, immediate<<16)
		logger.Print(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate<<16)
	case set.OP_SUI:
		this.Bus().StoreData(rdAddress, immediate<<16)
		logger.Print(" => [E]: [MEM(%#02X) = %#08X]", rdAddress, immediate<<16)
	default:
		return errors.New(fmt.Sprintf("Invalid operation to process by Data unit. Opcode: %d", instruction.Info.Opcode))
	}
	return nil
}
