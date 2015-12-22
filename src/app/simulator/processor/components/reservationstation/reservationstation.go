package reservationstation

import (
	"fmt"

	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/registeraliastable"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/instruction"
	"app/simulator/processor/models/operation"
	"app/simulator/processor/models/set"
)

type OperandType string
type Register int32
type EntryIndex int32

const (
	RAT_OFFSET    = 100
	REGISTER_FREE = -1
	INVALID_INDEX = -1
)

const (
	NilType      OperandType = "NIL"
	MemoryType   OperandType = "MEM"
	RegisterType OperandType = "REG"
	RatType      OperandType = "RAT"
)

type Operand struct {
	Type     OperandType
	Register Register
	RatEntry int32
}

func newMemoryOp(register Register) Operand {
	return Operand{Type: MemoryType, Register: register, RatEntry: INVALID_INDEX}
}

func newRegisterOp(register Register) Operand {
	return Operand{Type: RegisterType, Register: register, RatEntry: INVALID_INDEX}
}

func newRegisterRatOp(register Register, ratEntry int32) Operand {
	return Operand{Type: RatType, Register: register, RatEntry: ratEntry}
}

func newNilDep() Operand {
	return Operand{Type: NilType, Register: Register(INVALID_INDEX), RatEntry: INVALID_INDEX}
}

func (this Operand) IsValid() bool {
	return this.Type != NilType && this.Register != INVALID_INDEX
}

func (this Operand) HasDependency(operand Operand) bool {
	return this.IsValid() && operand.IsValid() && this.Type == operand.Type &&
		(this.Type == MemoryType && this.Register != INVALID_INDEX && this.Register == operand.Register) ||
		(this.Type == RegisterType && this.Register == operand.Register && this.Register != INVALID_INDEX) ||
		(this.Type == RatType && this.Register == operand.Register && this.RatEntry == operand.RatEntry && this.RatEntry != INVALID_INDEX)
}

func (this Operand) String() string {
	switch this.Type {
	case MemoryType:
		return fmt.Sprintf("M(R%d)", this.Register)
	case RegisterType:
		return fmt.Sprintf("R%d", this.Register)
	case RatType:
		return fmt.Sprintf("R%d(RAT%d)", this.Register, this.RatEntry)
	default:
		return fmt.Sprintf("%v", this.Type)
	}
}

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
	instructionsDispatchedStack uint32
	instructionsDispatchedMax   uint32
	registerAliasTable          *registeraliastable.RegisterAliasTable
	bus                         *storagebus.StorageBus
}

type RsEntry struct {
	Operation    *operation.Operation
	Destination  Operand
	Operands     []Operand
	Dependencies []Operand
	Free         bool
	Busy         bool
}

func New(index uint32, processor iprocessor.IProcessor, registers uint32, reservationStationEntries uint32,
	instructionsDispatchedPerCycle uint32, rat *registeraliastable.RegisterAliasTable, robBus *storagebus.StorageBus) *ReservationStation {
	rs := &ReservationStation{
		&reservationStation{
			index:                       index,
			processor:                   processor,
			entries:                     make([]RsEntry, reservationStationEntries),
			input:                       channel.New(reservationStationEntries),
			lock:                        make(chan bool, 1),
			isActive:                    true,
			instructionsDispatchedStack: 0,
			instructionsDispatchedMax:   instructionsDispatchedPerCycle,
			registerAliasTable:          rat,
			bus:                         robBus,
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

func (this *ReservationStation) AddInstructionsFetchedStack() {

	this.reservationStation.instructionsDispatchedStack += 1

	if this.reservationStation.instructionsDispatchedStack == this.reservationStation.instructionsDispatchedMax {
		this.Processor().Wait(consts.DISPATCH_CYCLES)
		this.reservationStation.instructionsDispatchedStack = 0
	}
}

func (this *ReservationStation) InstructionsFetchedStack() uint32 {
	return this.reservationStation.instructionsDispatchedStack
}

func (this *ReservationStation) RegisterAliasTable() *registeraliastable.RegisterAliasTable {
	return this.reservationStation.registerAliasTable
}

func (this *ReservationStation) Bus() *storagebus.StorageBus {
	return this.reservationStation.bus
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
			this.reservationStation.instructionsDispatchedStack = 0
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
		dest, valueOperands, memoryOperands := this.getComponentsFromInstruction(op.Instruction())

		// Convert to operand objects
		ops := []Operand{}
		for _, register := range memoryOperands {
			ops = append(ops, newMemoryOp(register))
		}
		for _, register := range valueOperands {

			ratEntry, ok := this.RegisterAliasTable().GetPhysicalRegister(op.Id()-1, uint32(register))
			if ok {
				ops = append(ops, newRegisterRatOp(register, int32(ratEntry)))
			} else {
				ops = append(ops, newRegisterOp(register))
			}
		}

		// Rat Dest
		regDestRat := newNilDep()
		if dest != INVALID_INDEX {
			if op.RenamedDestRegister() != INVALID_INDEX {
				regDestRat = newRegisterRatOp(dest, op.RenamedDestRegister())
			} else {
				regDestRat = newRegisterOp(dest)
			}
		}

		dependencies := this.getDependencies(op.Id(), regDestRat, ops)
		logger.Collect(" => [RS%d][%03d]: Adding op to entry %d [D: %v, O's: %v, V's: %v] ..",
			this.Index(), op.Id(), entryIndex, regDestRat, ops, dependencies)

		// Store entry depency into reservation station
		this.Entries()[entryIndex] = RsEntry{
			Operation:    op,
			Destination:  regDestRat,
			Operands:     ops,
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

		op := operation.Cast(value)
		commonDataBus.Release()
		this.Lock() <- true

		dest, _, _ := this.getComponentsFromInstruction(op.Instruction())
		entryIndex := this.getEntryIndexFromOperationId(op.Id())
		entryOp := this.Entries()[entryIndex]
		logger.Collect(" => [RS%d][%03d]: Operation completed, releasing entry %d", this.Index(), op.Id(), entryIndex)
		if entryIndex != INVALID_INDEX {
			// Release entry
			this.Entries()[entryIndex].Busy = false
			this.Entries()[entryIndex].Free = true
			// Release entry from reservation station queue
			this.Input().Release()
		}
		// Release destination register (as RAT if enabled)
		if dest != INVALID_INDEX {
			logger.Collect(" => [RS%d][%03d]: Register %v resolved", this.Index(), op.Id(), entryOp.Destination)
			this.releaseOperation(entryOp.Destination)
		}

		// Release operands registers
		for _, operand := range entryOp.Operands {
			if operand.IsValid() && (operand.Type == MemoryType || len(this.RegisterAliasTable().Entries()) == 0) {
				logger.Collect(" => [RS%d][%03d]: Register %v resolved", this.Index(), op.Id(), operand)
				this.releaseOperation(operand)
			}
		}
		<-this.Lock()
	}
}

func (this *ReservationStation) dispatchOperation(entryIndex EntryIndex, op *operation.Operation) {

	// Lock a slot in max quantity of instructions dispatcher per cycle
	this.AddInstructionsFetchedStack()
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

func (this *ReservationStation) releaseOperation(operand Operand) {

	// Remove dependencies from entries
	for entryIndex, entry := range this.Entries() {
		if !entry.Free && !entry.Busy {
			dependencies := this.getDependencies(entry.Operation.Id(), entry.Destination, entry.Operands)
			this.Entries()[entryIndex].Dependencies = dependencies
			//logger.Collect(" => [RS%d][%03d]: Dependencies op of entry %d [O: %v, V's: %v]...",
			//	this.Index(), entry.Operation.Id(), entryIndex, entry.Operands, dependencies)
			// If no dependencies anymore, release and execute
			if len(this.Entries()[entryIndex].Dependencies) == 0 {
				this.dispatchOperation(EntryIndex(entryIndex), entry.Operation)
			}
		}
	}
}

func (this *ReservationStation) getDependencies(operationId uint32, destRegister Operand, targetOperands []Operand) []Operand {
	dependencies := []Operand{}

	// If renaming registers enabled (RAT) is not enabled, check target destination
	if len(this.RegisterAliasTable().Entries()) == 0 {
		targetOperands = append(targetOperands, destRegister)
	}

	//targetRegisters := operandRegisters //append(operandRegisters, destRegister)
	for _, entry := range this.Entries() {
		if !entry.Free && entry.Operation.Id() < operationId {

			// Operands dependencies (RAW) True-dependencies.
			// WAR/WAW should not exist at the moment if register renaming is enabled
			for _, targetOperand := range targetOperands {

				// Check entry operands against target registers
				for _, entryOperand := range entry.Operands {
					if entryOperand.Type == MemoryType || len(this.RegisterAliasTable().Entries()) == 0 {

						// Check if there are no registers dependencies
						if targetOperand.HasDependency(entryOperand) {

							if entryOperand.Type == MemoryType {
								// Check address dependencies
								//if this.checkAddressDependencies(entry.Operation) {
								dependencies = append(dependencies, targetOperand)
								//}

							} else {
								dependencies = append(dependencies, targetOperand)
							}
						}
					}
				}

				// Check entry destination against target registers
				if targetOperand.HasDependency(entry.Destination) {
					dependencies = append(dependencies, targetOperand)
				}
			}
		}
	}
	return dependencies
}

func (this *ReservationStation) GetDestinationDependency(operationId uint32, instruction *instruction.Instruction) (bool, Register) {

	dest, _, _ := this.getComponentsFromInstruction(instruction)

	// No destination register on instruction
	if dest == INVALID_INDEX {
		return false, dest
	}

	for _, entry := range this.Entries() {
		if !entry.Free && entry.Operation.Id() < operationId {

			// Check entry destination and operands against target registers
			for _, entryRegister := range append(entry.Operands, entry.Destination) {

				// Dest dependency (WAR or WAW) False-dependencie
				if entryRegister.IsValid() && entryRegister.Register == dest {
					return true, dest
				}
			}
		}
	}

	return false, dest
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

func (this *ReservationStation) getComponentsFromInstruction(instruction *instruction.Instruction) (Register, []Register, []Register) {
	// Return destination, value/operand pointers, memory register pointers,

	if instruction.Info.Type == data.TypeI {
		data := instruction.Data.(*data.DataI)
		if instruction.Info.IsBranch() {
			return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32()), Register(data.RegisterS.ToUint32())}, []Register{}
		} else {
			switch instruction.Info.Opcode {
			case set.OP_LW:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}, []Register{Register(data.RegisterS.ToUint32())}
			case set.OP_LLI, set.OP_LUI:
				return Register(data.RegisterD.ToUint32()), []Register{}, []Register{}
			case set.OP_SW:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterS.ToUint32()), Register(data.RegisterD.ToUint32())}, []Register{Register(data.RegisterD.ToUint32())}
			case set.OP_SLI, set.OP_SUI:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32())}, []Register{Register(data.RegisterD.ToUint32())}
			default:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}, []Register{}
			}
		}
	} else if instruction.Info.Type == data.TypeR {
		data := instruction.Data.(*data.DataR)
		return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32()), Register(data.RegisterT.ToUint32())}, []Register{}
	}
	return INVALID_INDEX, nil, nil
}

/*func (this *ReservationStation) checkAddressDependencies(instruction *instruction.Instruction) bool {
	// Return destination, value/operand pointers, memory register pointers,

	if instruction.Info.Type == data.TypeI {
		data := instruction.Data.(*data.DataI)
		if instruction.Info.IsBranch() {
			return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32()), Register(data.RegisterS.ToUint32())}, []Register{}
		} else {
			switch instruction.Info.Opcode {
			case set.OP_LW:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}, []Register{Register(data.RegisterS.ToUint32())}
			case set.OP_LLI, set.OP_LUI:
				return Register(data.RegisterD.ToUint32()), []Register{}, []Register{}
			case set.OP_SW:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterS.ToUint32()), Register(data.RegisterD.ToUint32())}, []Register{Register(data.RegisterD.ToUint32())}
			case set.OP_SLI, set.OP_SUI:
				return Register(INVALID_INDEX), []Register{Register(data.RegisterD.ToUint32())}, []Register{Register(data.RegisterD.ToUint32())}
			default:
				return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32())}, []Register{}
			}
		}
	} else if instruction.Info.Type == data.TypeR {
		data := instruction.Data.(*data.DataR)
		return Register(data.RegisterD.ToUint32()), []Register{Register(data.RegisterS.ToUint32()), Register(data.RegisterT.ToUint32())}, []Register{}
	}
	return INVALID_INDEX, nil, nil
}
*/
