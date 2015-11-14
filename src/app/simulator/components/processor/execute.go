package processor

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/components/channel"
	"app/simulator/components/execution"
	"app/simulator/components/execution/alu"
	"app/simulator/components/execution/branch"
	"app/simulator/components/execution/loadstore"
	"app/simulator/consts"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
)

var (
	branchChannel    channel.Channel
	aluChannel       channel.Channel
	loadStoreChannel channel.Channel
)

func (this *Processor) LaunchExecuteUnits(aluUnits, loadStoreUnits, branchUnits uint32, input channel.Channel, channels map[info.CategoryEnum]channel.Channel) {

	go func() {
		for value := range input.Channel() {
			//Redirect input operations to the required execution unit channels
			instruction := value.(*instruction.Instruction)
			channels[instruction.Info.Category].Add(instruction)
			input.Release()
		}
	}()

	// Launch execution units
	this.launchExecutionUnitsByCategory(info.Aritmetic, aluUnits, channels)
	this.launchExecutionUnitsByCategory(info.LoadStore, loadStoreUnits, channels)
	this.launchExecutionUnitsByCategory(info.Control, branchUnits, channels)
}

func (this *Processor) launchExecutionUnitsByCategory(category info.CategoryEnum, units uint32, channels map[info.CategoryEnum]channel.Channel) {
	channels[category] = channel.New(units)

	for index := uint32(0); index < units; index++ {
		// Launch each unit as a goroutine
		go func(index uint32) {
			unit, event := this.getUnitFromCategory(category)
			logger.Print(" => Initializing execution unit (%s) %d", category, index)
			for value := range channels[category].Channel() {
				// Iterate instructions received via the channel
				this.executeOperation(unit, event, index, value.(*instruction.Instruction))
				// Release one item from input Channel
				channels[category].Release()
			}
		}(index)
	}
}

func (this *Processor) getUnitFromCategory(category info.CategoryEnum) (execution.Unit, string) {
	switch category {
	case info.Aritmetic:
		return alu.New(this.StorageBus()), consts.ALU_EVENT
	case info.LoadStore:
		return loadstore.New(this.StorageBus()), consts.LOAD_STORE_EVENT
	case info.Control:
		return branch.New(this.StorageBus()), consts.BRANCH_EVENT
	}
	return nil, ""
}

func (this *Processor) executeOperation(unit execution.Unit, event string, index uint32, instruction *instruction.Instruction) error {
	startCycles := this.Cycles()

	// Do decode once a data instruction is received
	err := unit.Process(instruction)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed executing instruction. %s]", err.Error()))
	}
	// Increment program counter
	this.IncrementProgramCounter(consts.BYTES_PER_WORD)
	// Wait cycles of a execution stage
	for i := uint8(0); i < instruction.Info.Cycles; i++ {
		logger.Collect(" => [E%s%d]: Executing %s, %s", event, index, instruction.Info.ToString(), instruction.Data.ToString())
		this.Wait(1)
	}
	// Log completion
	this.LogEvent(consts.EXECUTE_EVENT, event, index, startCycles)
	this.LogInstructionCompleted()
	return nil
}
