package executor

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/pipeline/executor/alu"
	"app/simulator/processor/components/pipeline/executor/branch"
	"app/simulator/processor/components/pipeline/executor/loadstore"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/operation"
)

type Executor struct {
	*executor
}

type executor struct {
	index     uint32
	processor iprocessor.IProcessor
	category  info.CategoryEnum
	bus       *storagebus.StorageBus
	isActive  bool
}

func New(index uint32, processor iprocessor.IProcessor, bus *storagebus.StorageBus, category info.CategoryEnum) *Executor {
	return &Executor{
		&executor{
			index:     index,
			processor: processor,
			bus:       bus,
			category:  category,
			isActive:  true,
		},
	}
}

func (this *Executor) Index() uint32 {
	return this.executor.index
}

func (this *Executor) Category() info.CategoryEnum {
	return this.executor.category
}

func (this *Executor) Processor() iprocessor.IProcessor {
	return this.executor.processor
}

func (this *Executor) Bus() *storagebus.StorageBus {
	return this.executor.bus
}

func (this *Executor) IsActive() bool {
	return this.executor.isActive
}

func (this *Executor) Close() {
	this.executor.isActive = false
}

func (this *Executor) Run(input map[info.CategoryEnum]channel.Channel, commonDataBus channel.Channel) {
	// Launch each unit as a goroutine
	unit, event := this.getUnitFromCategory(this.Category())
	logger.Print(" => Initializing execution unit (%s) %d", this.Category(), this.Index())
	go func() {
		for {
			value, running := <-input[this.Category()].Channel()
			if !running || !this.IsActive() {
				logger.Print(" => Flushing execution unit (%s) %d", this.Category(), this.Index())
				return
			}
			op := operation.Cast(value)
			// Iterate instructions received via the channel
			this.executeOperation(unit, event, op)
			// Send data to common bus for reservation station feedback
			commonDataBus.Add(op)
			// Release one item from input Channel
			input[this.Category()].Release()
		}
	}()
}

func (this *Executor) getUnitFromCategory(category info.CategoryEnum) (IExecutor, string) {
	switch category {
	case info.Aritmetic:
		return alu.New(this.Bus()), consts.ALU_EVENT
	case info.LoadStore:
		return loadstore.New(this.Bus()), consts.LOAD_STORE_EVENT
	case info.Control:
		return branch.New(this.Bus()), consts.BRANCH_EVENT
	}
	return nil, ""
}

func (this *Executor) executeOperation(unit IExecutor, event string, op *operation.Operation) error {
	startCycles := this.Processor().Cycles()

	// Do decode once a data instruction is received
	err := unit.Process(op)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed executing instruction. %s]", err.Error()))
	}
	// Wait cycles of a execution stage
	logger.Collect(" => [%s%d][%03d]: Executing %s, %s", event, this.Index(), op.Id(), op.Instruction().Info.ToString(), op.Instruction().Data.ToString())
	for i := uint8(0); i < op.Instruction().Info.Cycles; i++ {
		this.Processor().Wait(1)
	}
	// Log completion
	if this.IsActive() {
		this.Processor().LogEvent(event, this.Index(), op.Id(), startCycles)
	}
	return nil
}
