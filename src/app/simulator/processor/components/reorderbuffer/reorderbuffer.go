package reorderbuffer

import (
	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/registeraliastable"
	"app/simulator/processor/components/storagebus"
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
	bus                         *storagebus.StorageBus
	startOperationId            uint32
	buffer                      map[uint32]RobEntry
	robEntries                  uint32
	instructionsWrittenPerCycle uint32
	registerAliasTable          *registeraliastable.RegisterAliasTable
}

type RobEntry struct {
	Operation   *operation.Operation
	Type        RobType
	Destination uint32
	Value       int32
	Cycle       uint32
}

func New(index uint32, processor iprocessor.IProcessor, startOperationId, robEntries uint32,
	instructionsWrittenPerCycle uint32, rat *registeraliastable.RegisterAliasTable) *ReorderBuffer {
	rob := &ReorderBuffer{
		&reorderBuffer{
			index:                       index,
			processor:                   processor,
			startOperationId:            startOperationId,
			buffer:                      map[uint32]RobEntry{},
			robEntries:                  robEntries,
			instructionsWrittenPerCycle: instructionsWrittenPerCycle,
			registerAliasTable:          rat,
		},
	}
	rob.reorderBuffer.bus = rob.getStorageBus()
	return rob
}

func (this *ReorderBuffer) Index() uint32 {
	return this.reorderBuffer.index
}

func (this *ReorderBuffer) Processor() iprocessor.IProcessor {
	return this.reorderBuffer.processor
}

func (this *ReorderBuffer) Bus() *storagebus.StorageBus {
	return this.reorderBuffer.bus
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

func (this *ReorderBuffer) RegisterAliasTable() *registeraliastable.RegisterAliasTable {
	return this.reorderBuffer.registerAliasTable
}

func (this *ReorderBuffer) LoadRegister(op *operation.Operation, index uint32) uint32 {
	lookupRegister := index
	// If renaming register enabled
	if len(this.RegisterAliasTable().Entries()) > 0 {
		ratEntry, ok := this.RegisterAliasTable().GetPhysicalRegister(op.Id()-1, index)
		if ok {
			// Proceed to search register in ROB with renamed dest from RAT
			lookupRegister = ratEntry
		} else {
			// Alias does not exist, value was already commited (search on memory)
			return this.Processor().RegistersMemory().LoadUint32(index * consts.BYTES_PER_WORD)
		}
	}

	// Search register on ROB
	robEntry, ok := this.getEntryByDestination(op.Id(), RegisterType, lookupRegister)
	if ok {
		return uint32(robEntry.Value)
	}
	return this.Processor().RegistersMemory().LoadUint32(index * consts.BYTES_PER_WORD)
}

func (this *ReorderBuffer) StoreRegister(op *operation.Operation, index, value uint32) {

	dest := index
	// If renaming register enabled
	if op.RenamedDestRegister() != -1 {
		dest = uint32(op.RenamedDestRegister())
	}
	this.Buffer()[op.Id()] = RobEntry{
		Operation:   op,
		Type:        RegisterType,
		Destination: dest,
		Value:       int32(value),
		Cycle:       this.Processor().Cycles(),
	}
}

func (this *ReorderBuffer) Allocate(op *operation.Operation) {
	this.waitStallOperationIfFull(op)
}

func (this *ReorderBuffer) LoadData(op *operation.Operation, address uint32) uint32 {
	robEntry, ok := this.getEntryByDestination(op.Id(), MemoryType, address)
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
	clockAllowed := this.Processor().Cycles()
	forceClose := false

	go func() {
		for {
			_, running := <-commonDataBus.Channel()
			if !running || misprediction {
				forceClose = true
				logger.Print(" => Flushing re-order buffer unit %d", this.Index())
				return
			}
			commonDataBus.Release()
		}
	}()

	go func() {
		for {
			if forceClose {
				return
			}

			if this.Processor().Cycles() < clockAllowed {
				this.Processor().Wait(1)
				continue
			}

			// Commit in order, if missing an operation, wait for it
			computedAddress := uint32(0)
			robEntries := []RobEntry{}
			for robEntry, exists := this.Buffer()[opId]; exists; robEntry, exists = this.Buffer()[opId] {
				if uint32(len(robEntries)) >= this.InstructionsWrittenPerCycle() {
					break
				}
				// Ensure we can write results the next cycle result was written into ROB
				if this.Processor().Cycles() > robEntry.Cycle+1 {
					// Check for misprediction
					misprediction, computedAddress = this.checkForMisprediction(this.Buffer()[opId], robEntries)
					// Decrement speculative jumps
					this.Processor().DecrementSpeculativeJump()
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
			clockAllowed = this.Processor().Cycles() + 1
		}
	}()
}

func (this *ReorderBuffer) waitStallOperationIfFull(op *operation.Operation) {
	for uint32(len(this.Buffer())) >= this.RobEntries() {
		lastCompletedOpId := this.Processor().LastOperationIdCompleted()
		if op.Id() == lastCompletedOpId+1 {
			logger.Collect(" => [RB%d][%03d]: Writing latest instruction.", this.Index(), op.Id())
			return
		}
		logger.Collect(" => [RB%d][%03d]: ROB is full, wait for free entries. Current: %d, Max: %d, LastOpId: %d...",
			this.Index(), op.Id(), len(this.Buffer()), this.RobEntries(), lastCompletedOpId)
		this.Processor().Wait(1)
	}
}

func (this *ReorderBuffer) checkForMisprediction(targetEntry RobEntry, cachedEntries []RobEntry) (bool, uint32) {
	op := targetEntry.Operation

	if !op.Instruction().Info.IsBranch() {
		return false, 0
	}

	// If operation does not have a predicted address, then return
	if op.PredictedAddress() == -1 {
		this.Processor().LogBranchInstruction(op.Address(), op.Instruction().Info.IsConditionalBranch(), false, op.Taken())
		return false, 0
	}

	// If predicted address is equal to the computed address, then return
	computedAddress := this.getNextProgramCounter(targetEntry, this.getCachedProgramCounter(cachedEntries))
	failed := computedAddress != uint32(op.PredictedAddress())
	if failed {
		logger.Collect(" => [RB%d][%03d]: Misprediction found, it was predicted: %#04X and computed: %#04X",
			this.Index(), targetEntry.Operation.Id(), op.PredictedAddress(), computedAddress)
	}
	this.Processor().LogBranchInstruction(op.Address(), op.Instruction().Info.IsConditionalBranch(), failed, op.Taken())
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
		logger.Collect(" => [RB%d][%03d]: Commiting operation %d...", this.Index(), robEntry.Operation.Id(), robEntry.Operation.Id())
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

		// Register renaming (RAT)
		dest := robEntry.Destination
		if robEntry.Operation.RenamedDestRegister() != -1 {
			dest = this.RegisterAliasTable().Entries()[robEntry.Destination].ArchRegister

			// Release entry from RAT
			this.RegisterAliasTable().Release(opId)
		}

		logger.Collect(" => [RB%d][%03d]: Writing %#08X to %s%d...", this.Index(), opId, robEntry.Value, robEntry.Type, dest)
		this.Processor().RegistersMemory().StoreUint32(dest*consts.BYTES_PER_WORD, uint32(robEntry.Value))
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

func (this *ReorderBuffer) getEntryByDestination(operationId uint32, robType RobType, destination uint32) (RobEntry, bool) {
	maxOpId := int32(-1)
	for opId, value := range this.Buffer() {
		if value.Type == robType && value.Destination == destination && int32(opId) >= maxOpId && opId <= operationId {
			maxOpId = int32(opId)
		}
	}
	if maxOpId >= 0 {
		return this.Buffer()[uint32(maxOpId)], true
	}
	return RobEntry{}, false
}

func (this *ReorderBuffer) getStorageBus() *storagebus.StorageBus {

	return &storagebus.StorageBus{

		// Registers handlers
		LoadRegister: func(op *operation.Operation, index uint32) uint32 {
			return this.LoadRegister(op, index)
		},
		StoreRegister: func(op *operation.Operation, index, value uint32) {
			this.StoreRegister(op, index, value)
		},

		// Data Memory handlers
		LoadData: func(op *operation.Operation, address uint32) uint32 {
			return this.LoadData(op, address)
		},
		StoreData: func(op *operation.Operation, address, value uint32) {
			this.StoreData(op, address, value)
		},

		// Program Counter handlers
		IncrementProgramCounter: func(op *operation.Operation, value int32) {
			this.IncrementProgramCounter(op, value)
		},
		SetProgramCounter: func(op *operation.Operation, value uint32) {
			this.SetProgramCounter(op, value)
		},
	}
}
