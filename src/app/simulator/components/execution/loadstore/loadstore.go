package loadstore

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

type LoadStore struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *LoadStore {
	return &LoadStore{bus: bus}
}

func (this *LoadStore) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *LoadStore) Process(instruction *instruction.Instruction) error {

	rdAddress := instruction.Data.(*data.DataI).RegisterD.ToUint32()
	rsAddress := instruction.Data.(*data.DataI).RegisterS.ToUint32()
	immediate := instruction.Data.(*data.DataI).Immediate.ToUint32()

	switch instruction.Info.Opcode {
	case set.OP_LW:
		rsValue := this.Bus().LoadRegister(rsAddress)
		value := this.Bus().LoadData(rsValue + immediate)
		this.Bus().StoreRegister(rdAddress, value)
		logger.Collect(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, value)
	case set.OP_SW:
		rdValue := this.Bus().LoadRegister(rdAddress)
		rsValue := this.Bus().LoadRegister(rsAddress)
		this.Bus().StoreData(rsValue+immediate, rdValue)
		logger.Collect(" => [E]: [MEM(%#02X) = %#08X]", rsValue+immediate, rdValue)
	case set.OP_LLI:
		this.Bus().StoreRegister(rdAddress, immediate)
		logger.Collect(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate)
	case set.OP_SLI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(rdValue, immediate)
		logger.Collect(" => [E]: [MEM(%#02X) = %#08X]", rdValue, immediate)
	case set.OP_LUI:
		this.Bus().StoreRegister(rdAddress, immediate<<16)
		logger.Collect(" => [E]: [R%d(%#02X) = %#08X]", rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate<<16)
	case set.OP_SUI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(rdValue, immediate<<16)
		logger.Collect(" => [E]: [MEM(%#02X) = %#08X]", rdValue, immediate<<16)
	default:
		return errors.New(fmt.Sprintf("Invalid operation to process by Data unit. Opcode: %d", instruction.Info.Opcode))
	}
	return nil
}
