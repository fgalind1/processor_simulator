package dispatcher

import (
	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/reorderbuffer"
	"app/simulator/processor/components/reservationstation"
	"app/simulator/processor/components/storagebus"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/operation"
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
	instructionsDispatchedPerCycle uint32
	instructionsWrittenPerCycle    uint32
	bus                            *storagebus.StorageBus
	isActive                       bool
}

func New(index uint32, processor iprocessor.IProcessor, startOperationId, registers,
	reservationStationEntries, instructionsDispatchedPerCycle, instructionsWrittenPerCycle uint32) *Dispatcher {
	return &Dispatcher{
		&dispatcher{
			index:                          index,
			processor:                      processor,
			startOperationId:               startOperationId,
			registers:                      registers,
			reservationStationEntries:      reservationStationEntries,
			instructionsDispatchedPerCycle: instructionsDispatchedPerCycle,
			instructionsWrittenPerCycle:    instructionsWrittenPerCycle,
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

func (this *Dispatcher) InstructionsFetchedPerCycle() uint32 {
	return this.dispatcher.instructionsDispatchedPerCycle
}

func (this *Dispatcher) InstructionsWrittenPerCycle() uint32 {
	return this.dispatcher.instructionsWrittenPerCycle
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
	// Create reservation station
	rs := reservationstation.New(this.Index(), this.Processor(),
		this.Registers(),
		this.ReservationStationEntries(),
		this.InstructionsFetchedPerCycle())
	commonDataBusRS := channel.New(commonDataBus.Capacity())
	// Create re-order buffer
	rob := reorderbuffer.New(this.Index(),
		this.Processor(),
		this.StartOperationId(),
		this.ReservationStationEntries(),
		this.InstructionsWrittenPerCycle())
	commonDataBusROB := channel.New(commonDataBus.Capacity())
	// Create storage bus
	this.dispatcher.bus = this.getStorageBus(rob)

	// Launch each unit as a goroutine
	logger.Print(" => Initializing dispatcher unit %d", this.Index())

	// Start dispatcher of operations to be executed into reservation station
	go this.runDispatcherToReservationStation(input, rs)
	// Start common bus multiplexer to send ack to reservation station and reorder buffer
	go this.runCommonBusMultiplexer(commonDataBus, commonDataBusRS, commonDataBusROB)

	// Run reservation station
	rs.Run(commonDataBusRS, output)
	// Run re-order buffer
	rob.Run(commonDataBusROB, recoveryBus)
}

func (this *Dispatcher) runDispatcherToReservationStation(input channel.Channel, rs *reservationstation.ReservationStation) {
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

func (this *Dispatcher) getStorageBus(rob *reorderbuffer.ReorderBuffer) *storagebus.StorageBus {

	return &storagebus.StorageBus{

		// Registers handlers
		LoadRegister: func(index uint32) uint32 {
			return rob.LoadRegister(index)
		},
		StoreRegister: func(op *operation.Operation, index, value uint32) {
			rob.StoreRegister(op, index, value)
		},

		// Data Memory handlers
		LoadData: func(address uint32) uint32 {
			return rob.LoadData(address)
		},
		StoreData: func(op *operation.Operation, address, value uint32) {
			rob.StoreData(op, address, value)
		},

		// Program Counter handlers
		IncrementProgramCounter: func(op *operation.Operation, value int32) {
			rob.IncrementProgramCounter(op, value)
		},
		SetProgramCounter: func(op *operation.Operation, value uint32) {
			rob.SetProgramCounter(op, value)
		},
	}
}
