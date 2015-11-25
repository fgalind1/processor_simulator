package reservationstation

import (
	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/instruction"
	"app/simulator/processor/models/operation"
	"app/simulator/processor/models/set"
)

const (
	REGISTER_FREE = -1
	INVALID_INDEX = -1
)

type Register int32
type EntryIndex int32

type ReservationStation struct {
	*reservationStation
}

type reservationStation struct {
	index                       uint32
	processor                   iprocessor.IProcessor
	entries                     []RsEntry
	input                       channel.Channel
	output                      map[info.CategoryEnum]channel.Channel
	lock                        chan bool
	isActive                    bool
	instructionsDispatchedStack chan bool
}

type RsEntry struct {
	Operation    *operation.Operation
	Destination  Register
	Operands     []Register
	Dependencies []Register
	Free         bool
	Busy         bool
}

func New(index uint32, processor iprocessor.IProcessor, registers uint32, reservationStationEntries uint32,
	instructionsDispatchedPerCycle uint32) *ReservationStation {
	rs := &ReservationStation{
		&reservationStation{
			index:                       index,
			processor:                   processor,
			entries:                     make([]RsEntry, reservationStationEntries),
			input:                       channel.New(reservationStationEntries),
			lock:                        make(chan bool, 1),
			isActive:                    true,
			instructionsDispatchedStack: make(chan bool, instructionsDispatchedPerCycle),
		},
	}
	for entryIndex, _ := range rs.Entries() {
		rs.Entries()[entryIndex] = RsEntry{
			Free: true,
			Busy: false,
		}
	}
	return rs
}

func (this *ReservationStation) Index() uint32 {
	return this.reservationStation.index
}

func (this *ReservationStation) Processor() iprocessor.IProcessor {
	return this.reservationStation.processor
}

func (this *ReservationStation) Entries() []RsEntry {
	return this.reservationStation.entries
}

func (this *ReservationStation) Input() channel.Channel {
	return this.reservationStation.input
}

func (this *ReservationStation) Output() map[info.CategoryEnum]channel.Channel {
	return this.reservationStation.output
}

func (this *ReservationStation) SetOutput(output map[info.CategoryEnum]channel.Channel) {
	this.reservationStation.output = output
}

func (this *ReservationStation) Lock() chan bool {
	return this.reservationStation.lock
}

func (this *ReservationStation) InstructionsFetchedStack() chan bool {
	return this.reservationStation.instructionsDispatchedStack
}

func (this *ReservationStation) CleanInstructionsFetchedStack() {
	this.reservationStation.instructionsDispatchedStack = make(chan bool, cap(this.reservationStation.instructionsDispatchedStack))
}

func (this *ReservationStation) Schedule(op *operation.Operation) {
	// Wait until there is an entry free
	this.Input().Add(op)
}

func (this *ReservationStation) Run(commonDataBus channel.Channel, output map[info.CategoryEnum]channel.Channel) {
	// Launch unit as a goroutine
	logger.Print(" => Initializing reservation station unit %d", this.Index())
	this.SetOutput(output)

	go func() {
		for this.reservationStation.isActive {
			this.Processor().Wait(consts.DISPATCH_CYCLES)
			this.CleanInstructionsFetchedStack()
		}
	}()

	go this.runScheduler()
	go this.runCommonBusListener(commonDataBus)
}

func (this *ReservationStation) runScheduler() {
	// For each operation received to schedule, process it
	for {
		value, running := <-this.Input().Channel()
		if !running {
			this.reservationStation.isActive = false
			logger.Print(" => Flushing reservation station unit %d (scheduler)", this.Index())
			return
		}
		this.Lock() <- true
		op := operation.Cast(value)

		// Get next entry free
		entryIndex := this.getNextIndexFreeEntry()
		dest, operands := this.getComponentsFromInstruction(op.Instruction())
		dependencies := this.getDependencies(op.Id(), dest, operands)
		logger.Collect(" => [RS%d][%03d]: Adding op to entry %d [D: %d, O's: %v, V's: %v]...",
			this.Index(), op.Id(), entryIndex, dest, operands, dependencies)

		// Store entry depency into reservation station
		this.Entries()[entryIndex] = RsEntry{
			Operation:    op,
			Destination:  dest,
			Operands:     operands,
			Dependencies: dependencies,
			Free:         false,
			Busy:         false,
		}

		// If no waiting dependencies, release and execute
		if len(dependencies) == 0 {
			this.dispatchOperation(EntryIndex(entryIndex), op)
		}
		<-this.Lock()
	}
}

func (this *ReservationStation) runCommonBusListener(commonDataBus channel.Channel) {
	// For each operation executed, feed reservation station to release operands
	for {
		value, running := <-commonDataBus.Channel()
		if !running {
			this.reservationStation.isActive = false
			logger.Print(" => Flushing reservation station unit %d (CDB listener)", this.Index())
			return
		}
		this.Lock() <- true
		op := operation.Cast(value)

		dest, operands := this.getComponentsFromInstruction(op.Instruction())
		entryIndex := this.getEntryIndexFromOperationId(op.Id())
		logger.Collect(" => [RS%d][%03d]: Operation completed, releasing entry %d", this.Index(), op.Id(), entryIndex)
		if entryIndex != INVALID_INDEX {
			// Release entry
			this.Entries()[entryIndex].Busy = false
			this.Entries()[entryIndex].Free = true
			// Release entry from reservation station queue
			this.Input().Release()
		}
		// Release destination register
		if dest != INVALID_INDEX {
			logger.Collect(" => [RS%d][%03d]: Register %d resolved", this.Index(), op.Id(), dest)
			this.releaseOperation(Register(dest))
		}
		// Release operands registers
		for _, operand := range operands {
			if operand != INVALID_INDEX {
				logger.Collect(" => [RS%d][%03d]: Register %d resolved", this.Index(), op.Id(), operand)
				this.releaseOperation(Register(operand))
			}
		}
		commonDataBus.Release()
		<-this.Lock()
	}
}

func (this *ReservationStation) dispatchOperation(entryIndex EntryIndex, op *operation.Operation) {

	// Lock a slot in max quantity of instructions dispatcher per cycle
	this.InstructionsFetchedStack() <- true
	// Release reservation station entry
	this.Entries()[entryIndex].Busy = true

	go func() {
		// Release one entry and send operation to execution, when output channel (execution unit) is free
		logger.Collect(" => [RS%d][%03d]: Sending entry %d to %s queue...", this.Index(), op.Id(), entryIndex, op.Instruction().Info.Category)
		// Wait dispatch cycles
		startCycles := this.Processor().Cycles()
		this.Processor().Wait(consts.DISPATCH_CYCLES)
		// Log completion
		this.Processor().LogEvent(consts.DISPATCH_EVENT, this.Index(), op.Id(), startCycles)
		// Send data to execution unit in a go routine, when execution unit is available
		go this.Output()[op.Instruction().Info.Category].Add(op)
	}()
}

func (this *ReservationStation) releaseOperation(registerIndex Register) {

	// Remove dependencies from entries
	for entryIndex, entry := range this.Entries() {
		if !entry.Free && !entry.Busy {
			index := this.getIndexFromDependencies(Register(registerIndex), entry.Dependencies)
			if index != INVALID_INDEX {
				this.Entries()[entryIndex].Dependencies = this.getDependencies(entry.Operation.Id(), entry.Destination, entry.Operands)
			}
			// If no dependencies anymore, release and execute
			if len(this.Entries()[entryIndex].Dependencies) == 0 {
				this.dispatchOperation(EntryIndex(entryIndex), entry.Operation)
			}
		}
	}
}

func (this *ReservationStation) getDependencies(operationId uint32, destRegister Register, operandRegisters []Register) []Register {
	dependencies := []Register{}
	targetRegisters := append(operandRegisters, destRegister)
	for _, entry := range this.Entries() {
		if !entry.Free && entry.Operation.Id() < operationId {

			// Check entry destination and operands against target registers
			for _, entryRegister := range append(entry.Operands, entry.Destination) {
				for _, targetRegister := range targetRegisters {
					if entryRegister != INVALID_INDEX && entryRegister == targetRegister {
						dependencies = append(dependencies, Register(targetRegister))
					}
				}
			}
		}
	}
	return dependencies
}

func (this *ReservationStation) getEntryIndexFromOperationId(operationId uint32) EntryIndex {
	for entryIndex, entry := range this.Entries() {
		if entry.Operation.Id() == operationId {
			return EntryIndex(entryIndex)
		}
	}
	return EntryIndex(INVALID_INDEX)
}

func (this *ReservationStation) getNextIndexFreeEntry() EntryIndex {
	for entryIndex, entry := range this.Entries() {
		if entry.Free {
			return EntryIndex(entryIndex)
		}
	}
	return INVALID_INDEX
}

func (this *ReservationStation) getComponentsFromInstruction(instruction *instruction.Instruction) (Register, []Register) {
	if instruction.Info.Type == data.TypeI {
		data := instruction.Data.(*data.DataI)
		if instruction.Info.IsBranch() {
			return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32()), Register(data.RegisterS.ToUint32())}
		} else {
			switch instruction.Info.Opcode {
			case set.OP_LW:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}
			case set.OP_LLI, set.OP_LUI:
				return Register(data.RegisterD.ToUint32()), []Register{}
			case set.OP_SW:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32()), Register(data.RegisterS.ToUint32())}
			case set.OP_SLI, set.OP_SUI:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32())}
			default:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}
			}
		}
	} else if instruction.Info.Type == data.TypeR {
		data := instruction.Data.(*data.DataR)
		return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32()), Register(data.RegisterT.ToUint32())}
	}
	return INVALID_INDEX, nil
}

func (this *ReservationStation) getIndexFromDependencies(target Register, dependencies []Register) Register {
	for index, value := range dependencies {
		if Register(value) == target {
			return Register(index)
		}
	}
	return INVALID_INDEX
}
