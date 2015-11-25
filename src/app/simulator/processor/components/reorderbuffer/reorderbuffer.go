package reorderbuffer

import (
	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/operation"
)

type RobType string

const (
	MemoryType         RobType = "M"
	RegisterType       RobType = "R"
	ProgramCounterType RobType = "PC"

	AbsoluteType = 0
	OffsetType   = 1
)

type ReorderBuffer struct {
	*reorderBuffer
}

type reorderBuffer struct {
	index                       uint32
	processor                   iprocessor.IProcessor
	startOperationId            uint32
	buffer                      map[uint32]RobEntry
	robEntries                  uint32
	instructionsWrittenPerCycle uint32
}

type RobEntry struct {
	Operation   *operation.Operation
	Type        RobType
	Destination uint32
	Value       int32
	Cycle       uint32
}

func New(index uint32, processor iprocessor.IProcessor, startOperationId, robEntries uint32, instructionsWrittenPerCycle uint32) *ReorderBuffer {
	return &ReorderBuffer{
		&reorderBuffer{
			index:                       index,
			processor:                   processor,
			startOperationId:            startOperationId,
			buffer:                      map[uint32]RobEntry{},
			robEntries:                  robEntries,
			instructionsWrittenPerCycle: instructionsWrittenPerCycle,
		},
	}
}

func (this *ReorderBuffer) Index() uint32 {
	return this.reorderBuffer.index
}

func (this *ReorderBuffer) Processor() iprocessor.IProcessor {
	return this.reorderBuffer.processor
}

func (this *ReorderBuffer) StartOperationId() uint32 {
	return this.reorderBuffer.startOperationId
}

func (this *ReorderBuffer) Buffer() map[uint32]RobEntry {
	return this.reorderBuffer.buffer
}

func (this *ReorderBuffer) RobEntries() uint32 {
	return this.reorderBuffer.robEntries
}

func (this *ReorderBuffer) InstructionsWrittenPerCycle() uint32 {
	return this.reorderBuffer.instructionsWrittenPerCycle
}

func (this *ReorderBuffer) LoadRegister(index uint32) uint32 {
	robEntry, ok := this.getEntryByDestination(RegisterType, index)
	if ok {
		return uint32(robEntry.Value)
	}
	return this.Processor().RegistersMemory().LoadUint32(index * consts.BYTES_PER_WORD)
}

func (this *ReorderBuffer) StoreRegister(op *operation.Operation, index, value uint32) {
	this.Buffer()[op.Id()] = RobEntry{
		Operation:   op,
		Type:        RegisterType,
		Destination: index,
		Value:       int32(value),
		Cycle:       this.Processor().Cycles(),
	}
}

func (this *ReorderBuffer) LoadData(address uint32) uint32 {
	robEntry, ok := this.getEntryByDestination(MemoryType, address)
	if ok {
		return uint32(robEntry.Value)
	}
	return this.Processor().DataMemory().LoadUint32(address)
}

func (this *ReorderBuffer) StoreData(op *operation.Operation, address, value uint32) {
	this.Buffer()[op.Id()] = RobEntry{
		Operation:   op,
		Type:        MemoryType,
		Destination: address,
		Value:       int32(value),
		Cycle:       this.Processor().Cycles(),
	}
}

func (this *ReorderBuffer) IncrementProgramCounter(op *operation.Operation, value int32) {
	this.Buffer()[op.Id()] = RobEntry{
		Operation:   op,
		Type:        ProgramCounterType,
		Destination: OffsetType,
		Value:       value,
		Cycle:       this.Processor().Cycles(),
	}
}

func (this *ReorderBuffer) SetProgramCounter(op *operation.Operation, value uint32) {
	this.Buffer()[op.Id()] = RobEntry{
		Operation:   op,
		Type:        ProgramCounterType,
		Destination: AbsoluteType,
		Value:       int32(value),
		Cycle:       this.Processor().Cycles(),
	}
}

func (this *ReorderBuffer) Run(commonDataBus channel.Channel, recoveryBus channel.Channel) {
	// Launch unit as a goroutine
	logger.Print(" => Initializing re-order buffer unit %d", this.Index())
	opId := this.StartOperationId()
	misprediction := false

	go func() {
		for {
			_, running := <-commonDataBus.Channel()
			if !running || misprediction {
				logger.Print(" => Flushing re-order buffer unit %d", this.Index())
				return
			}
			commonDataBus.Release()

			// Commit in order, if missing an operation, wait for it
			computedAddress := uint32(0)
			robEntries := []RobEntry{}
			for robEntry, exists := this.Buffer()[opId]; exists; robEntry, exists = this.Buffer()[opId] {
				if uint32(len(robEntries)) >= this.InstructionsWrittenPerCycle() {
					commonDataBus.Add(true)
					break
				}
				// Ensure we can write results the next cycle result was written into ROB
				if this.Processor().Cycles() > robEntry.Cycle+1 {
					// Check for misprediction
					misprediction, computedAddress = this.checkForMisprediction(this.Buffer()[opId], robEntries)
					// Add to queue for commit
					robEntries = append(robEntries, robEntry)
					opId += 1
					// If misprediction, do not process more rob entries
					if misprediction {
						break
					}
				}
			}
			this.commitRobEntries(robEntries)
			if misprediction {
				this.Processor().Wait(consts.WRITEBACK_CYCLES)
				recoveryBus.Add(operation.New(opId, computedAddress))
			}
		}
	}()
}

func (this *ReorderBuffer) checkForMisprediction(targetEntry RobEntry, cachedEntries []RobEntry) (bool, uint32) {
	// If operation does not have a predicted address, then return
	if targetEntry.Operation.PredictedAddress() == -1 {
		if targetEntry.Operation.Instruction().Info.IsUnconditionalBranch() {
			this.Processor().LogBranchInstruction(false, false)
		} else if targetEntry.Operation.Instruction().Info.IsConditionalBranch() {
			this.Processor().LogBranchInstruction(true, false)
		}
		return false, 0
	}

	// If predicted address is equal to the computed address, then return
	computedAddress := this.getNextProgramCounter(targetEntry, this.getCachedProgramCounter(cachedEntries))
	failed := computedAddress != uint32(targetEntry.Operation.PredictedAddress())
	if failed {
		logger.Collect(" => [RB%d][%03d]: Misprediction found, it was predicted: %#04X and computed: %#04X",
			this.Index(), targetEntry.Operation.Id(), targetEntry.Operation.PredictedAddress(), computedAddress)
	}
	this.Processor().LogBranchInstruction(targetEntry.Operation.Instruction().Info.IsConditionalBranch(), failed)
	return failed, computedAddress
}

func (this *ReorderBuffer) getCachedProgramCounter(cachedEntries []RobEntry) uint32 {
	pc := this.Processor().ProgramCounter()
	for _, robEntry := range cachedEntries {
		if robEntry.Type == ProgramCounterType {
			pc = this.getNextProgramCounter(robEntry, pc)
		}
		// Increment program counter
		pc += consts.BYTES_PER_WORD
	}
	return pc + consts.BYTES_PER_WORD
}

func (this *ReorderBuffer) commitRobEntries(robEntries []RobEntry) {
	startCycles := this.Processor().Cycles()
	// Commit results in order
	opIds := []uint32{}
	for _, robEntry := range robEntries {
		opIds = append(opIds, robEntry.Operation.Id())
		this.commitRobEntry(robEntry, startCycles)
	}
	// Wait and log completion in a go routine
	go func(opIds []uint32, startCycles uint32) {
		this.Processor().Wait(consts.WRITEBACK_CYCLES)
		for _, opId := range opIds {
			this.Processor().LogEvent(consts.WRITEBACK_EVENT, this.Index(), opId, startCycles)
			this.Processor().LogInstructionCompleted(opId)
		}
	}(opIds, startCycles)
}

func (this *ReorderBuffer) commitRobEntry(robEntry RobEntry, startCycles uint32) {
	// Commit update
	opId := robEntry.Operation.Id()
	if robEntry.Type == RegisterType {
		logger.Collect(" => [RB%d][%03d]: Writing %#08X to %s%d...", this.Index(), opId, robEntry.Value, robEntry.Type, robEntry.Destination)
		this.Processor().RegistersMemory().StoreUint32(robEntry.Destination*consts.BYTES_PER_WORD, uint32(robEntry.Value))
	} else if robEntry.Type == MemoryType {
		logger.Collect(" => [RB%d][%03d]: Writing %#08X to %s[%#X]...", this.Index(), opId, robEntry.Value, robEntry.Type, robEntry.Destination)
		this.Processor().DataMemory().StoreUint32(robEntry.Destination, uint32(robEntry.Value))
	} else {
		this.Processor().SetProgramCounter(this.getNextProgramCounter(robEntry, this.Processor().ProgramCounter()))
	}

	// Increment program counter
	this.Processor().IncrementProgramCounter(consts.BYTES_PER_WORD)
	logger.Collect(" => [RB%d][%03d]: PC = %#04X", this.Index(), opId, this.Processor().ProgramCounter())

	// Release ROB entry
	delete(this.Buffer(), opId)
}

func (this *ReorderBuffer) getNextProgramCounter(robEntry RobEntry, programCounter uint32) uint32 {
	if robEntry.Destination == AbsoluteType {
		return uint32(robEntry.Value)
	} else {
		return uint32(int32(programCounter) + robEntry.Value)
	}
}

func (this *ReorderBuffer) getEntryByDestination(robType RobType, destination uint32) (RobEntry, bool) {
	maxOpId := int32(-1)
	for opId, value := range this.Buffer() {
		if value.Type == robType && value.Destination == destination && int32(opId) >= maxOpId {
			maxOpId = int32(opId)
		}
	}
	if maxOpId >= 0 {
		return this.Buffer()[uint32(maxOpId)], true
	}
	return RobEntry{}, false
}
