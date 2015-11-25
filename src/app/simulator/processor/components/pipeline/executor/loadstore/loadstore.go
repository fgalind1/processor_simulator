package loadstore

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

type LoadStore struct {
	bus *storagebus.StorageBus
}

func New(bus *storagebus.StorageBus) *LoadStore {
	return &LoadStore{bus: bus}
}

func (this *LoadStore) Bus() *storagebus.StorageBus {
	return this.bus
}

func (this *LoadStore) Process(operation *operation.Operation) error {

	instruction := operation.Instruction()
	rdAddress := instruction.Data.(*data.DataI).RegisterD.ToUint32()
	rsAddress := instruction.Data.(*data.DataI).RegisterS.ToUint32()
	immediate := instruction.Data.(*data.DataI).Immediate.ToUint32()

	switch instruction.Info.Opcode {
	case set.OP_LW:
		rsValue := this.Bus().LoadRegister(rsAddress)
		value := this.Bus().LoadData(rsValue + immediate)
		this.Bus().StoreRegister(operation, rdAddress, value)
		logger.Collect(" => [E][%03d]: [R%d(%#02X) = %#08X]", operation.Id(), rdAddress, rdAddress*consts.BYTES_PER_WORD, value)
	case set.OP_SW:
		rdValue := this.Bus().LoadRegister(rdAddress)
		rsValue := this.Bus().LoadRegister(rsAddress)
		this.Bus().StoreData(operation, rdValue+immediate, rsValue)
		logger.Collect(" => [E][%03d]: [MEM(%#02X) = %#08X]", operation.Id(), rdValue+immediate, rsValue)
	case set.OP_LLI:
		this.Bus().StoreRegister(operation, rdAddress, immediate)
		logger.Collect(" => [E][%03d]: [R%d(%#02X) = %#08X]", operation.Id(), rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate)
	case set.OP_SLI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(operation, rdValue, immediate)
		logger.Collect(" => [E][%03d]: [MEM(%#02X) = %#08X]", operation.Id(), rdValue, immediate)
	case set.OP_LUI:
		this.Bus().StoreRegister(operation, rdAddress, immediate<<16)
		logger.Collect(" => [E][%03d]: [R%d(%#02X) = %#08X]", operation.Id(), rdAddress, rdAddress*consts.BYTES_PER_WORD, immediate<<16)
	case set.OP_SUI:
		rdValue := this.Bus().LoadRegister(rdAddress)
		this.Bus().StoreData(operation, rdValue, immediate<<16)
		logger.Collect(" => [E][%03d]: [MEM(%#02X) = %#08X]", operation.Id(), rdValue, immediate<<16)
	default:
		return errors.New(fmt.Sprintf("Invalid operation to process by Data unit. Opcode: %d", instruction.Info.Opcode))
	}
	return nil
}
