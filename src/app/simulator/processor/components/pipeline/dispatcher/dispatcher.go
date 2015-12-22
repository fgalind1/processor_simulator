package dispatcher

import (
	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/registeraliastable"
	"app/simulator/processor/components/reorderbuffer"
	"app/simulator/processor/components/reservationstation"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/operation"
	"app/simulator/processor/models/set"
)

type Dispatcher struct {
	*dispatcher
}

type dispatcher struct {
	index                          uint32
	processor                      iprocessor.IProcessor
	startOperationId               uint32
	registers                      uint32
	reservationStationEntries      uint32
	reorderBufferEntries           uint32
	instructionsDispatchedPerCycle uint32
	instructionsWrittenPerCycle    uint32
	registerAliasTableEntries      uint32
	bus                            *storagebus.StorageBus
	isActive                       bool
}

func New(index uint32, processor iprocessor.IProcessor, startOperationId, registers,
	reservationStationEntries, reorderBufferEntries, instructionsDispatchedPerCycle, instructionsWrittenPerCycle, registerAliasTableEntries uint32) *Dispatcher {
	return &Dispatcher{
		&dispatcher{
			index:                          index,
			processor:                      processor,
			startOperationId:               startOperationId,
			registers:                      registers,
			reservationStationEntries:      reservationStationEntries,
			reorderBufferEntries:           reorderBufferEntries,
			instructionsDispatchedPerCycle: instructionsDispatchedPerCycle,
			instructionsWrittenPerCycle:    instructionsWrittenPerCycle,
			registerAliasTableEntries:      registerAliasTableEntries,
			isActive:                       true,
		},
	}
}

func (this *Dispatcher) Index() uint32 {
	return this.dispatcher.index
}

func (this *Dispatcher) Processor() iprocessor.IProcessor {
	return this.dispatcher.processor
}

func (this *Dispatcher) StartOperationId() uint32 {
	return this.dispatcher.startOperationId
}

func (this *Dispatcher) Registers() uint32 {
	return this.dispatcher.registers
}

func (this *Dispatcher) ReservationStationEntries() uint32 {
	return this.dispatcher.reservationStationEntries
}

func (this *Dispatcher) ReorderBufferEntries() uint32 {
	return this.dispatcher.reorderBufferEntries
}

func (this *Dispatcher) InstructionsFetchedPerCycle() uint32 {
	return this.dispatcher.instructionsDispatchedPerCycle
}

func (this *Dispatcher) InstructionsWrittenPerCycle() uint32 {
	return this.dispatcher.instructionsWrittenPerCycle
}

func (this *Dispatcher) RegisterAliasTableEntries() uint32 {
	return this.dispatcher.registerAliasTableEntries
}

func (this *Dispatcher) Bus() *storagebus.StorageBus {
	return this.dispatcher.bus
}

func (this *Dispatcher) IsActive() bool {
	return this.dispatcher.isActive
}

func (this *Dispatcher) Close() {
	this.dispatcher.isActive = false
}

func (this *Dispatcher) Run(input channel.Channel, output map[info.CategoryEnum]channel.Channel, commonDataBus, recoveryBus channel.Channel) {

	// Create register alias table
	rat := registeraliastable.New(this.Index(), this.RegisterAliasTableEntries())

	// Create re-order buffer
	rob := reorderbuffer.New(this.Index(),
		this.Processor(),
		this.StartOperationId(),
		this.ReorderBufferEntries(),
		this.InstructionsWrittenPerCycle(),
		rat)
	commonDataBusROB := channel.New(commonDataBus.Capacity())

	// Create reservation station
	rs := reservationstation.New(this.Index(), this.Processor(),
		this.Registers(),
		this.ReservationStationEntries(),
		this.InstructionsFetchedPerCycle(),
		rat, rob.Bus())
	commonDataBusRS := channel.New(commonDataBus.Capacity())

	// Create storage bus
	this.dispatcher.bus = rob.Bus()

	// Launch each unit as a goroutine
	logger.Print(" => Initializing dispatcher unit %d", this.Index())

	// Start dispatcher of operations to be executed into reservation station
	go this.runDispatcherToReservationStation(input, rs, rat, rob)
	// Start common bus multiplexer to send ack to reservation station and reorder buffer
	go this.runCommonBusMultiplexer(commonDataBus, commonDataBusRS, commonDataBusROB)

	// Run reservation station
	rs.Run(commonDataBusRS, output)
	// Run re-order buffer
	rob.Run(commonDataBusROB, recoveryBus)
}

func (this *Dispatcher) runDispatcherToReservationStation(input channel.Channel,
	rs *reservationstation.ReservationStation, rat *registeraliastable.RegisterAliasTable, rob *reorderbuffer.ReorderBuffer) {

	incomingQueue := map[uint32]*operation.Operation{}
	currentOperationId := this.StartOperationId()

	// For each operation received to schedule, process it
	for {
		value, running := <-input.Channel()
		if !running || !this.IsActive() {
			logger.Print(" => Flushing dispatcher unit %d (dispatcher to RS)", this.Index())
			return
		}
		op := operation.Cast(value)

		// Add to current operation
		incomingQueue[op.Id()] = op

		// Send to incoming channel pending ops (if available)
		for op, exists := incomingQueue[currentOperationId]; exists; op, exists = incomingQueue[currentOperationId] {

			// Allocate in ROB if there is spacde, otherwise stall
			rob.Allocate(op)

			// Rename register in case of WAR & WAR hazards
			if this.RegisterAliasTableEntries() > 0 {
				_, destRegister := rs.GetDestinationDependency(op.Id(), op.Instruction())
				if destRegister != -1 {
					found, _ := rat.AddMap(uint32(destRegister), op.Id())
					if !found {
						// Need to stall for an available RAT entry
						logger.Collect(" => [DI%d][%03d]: No entry available in RAT. Wait for one...", this.Index(), op.Id())
						break
					}

					// Rename to physical registers
					this.renameRegisters(op.Id(), op, rat)
				}
			}

			//Redirect input operations to the required execution unit channels
			logger.Collect(" => [DI%d][%03d]: Scheduling to RS: %s, %s", this.Index(), op.Id(), op.Instruction().Info.ToString(), op.Instruction().Data.ToString())
			rs.Schedule(op)
			currentOperationId += 1
		}
		input.Release()
	}
}

func (this *Dispatcher) runCommonBusMultiplexer(input, output1, output2 channel.Channel) {
	// For each result got from execution units in the common data bus send to RS and ROB
	for {
		value, running := <-input.Channel()
		if !running {
			output1.Close()
			output2.Close()
			logger.Print(" => Flushing dispatcher unit %d (CDB Mux)", this.Index())
			return
		}
		output1.Add(value)
		output2.Add(value)
		input.Release()
	}
}

func (this *Dispatcher) renameRegisters(operationId uint32, op *operation.Operation, rat *registeraliastable.RegisterAliasTable) {

	if op.Instruction().Info.Type == data.TypeI {
		data := op.Instruction().Data.(*data.DataI)
		if !op.Instruction().Info.IsBranch() {
			opcode := op.Instruction().Info.Opcode
			if opcode != set.OP_SW && opcode != set.OP_SLI && opcode != set.OP_SUI {
				reg, _ := rat.GetPhysicalRegister(operationId, data.RegisterD.ToUint32())
				op.SetRenamedDestRegister(reg)
			}
		}
		op.Instruction().Data = data
	} else if op.Instruction().Info.Type == data.TypeR {
		data := op.Instruction().Data.(*data.DataR)
		reg, _ := rat.GetPhysicalRegister(operationId, data.RegisterD.ToUint32())
		op.SetRenamedDestRegister(reg)
	}
}
