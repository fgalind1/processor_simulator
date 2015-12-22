package processor

import (
	"app/logger"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/components/pipeline/decoder"
	"app/simulator/processor/components/pipeline/dispatcher"
	"app/simulator/processor/components/pipeline/executor"
	"app/simulator/processor/components/pipeline/fetcher"
	"app/simulator/processor/config"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/operation"
)

func (this *Processor) Start() {

	logger.CleanBuffer()
	logger.Print("\n Simulation:")
	logger.Print("\n => Starting program...")

	// Launch pipeline units and execute instruction 0x0000
	recoveryChannel := channel.New(1)
	flushFunc := this.StartPipelineUnits(this.Config(), recoveryChannel, 0, 0x0000)

	// Launch clock
	go this.RunClock()

	// Launch recovery
	go this.RunRecovery(recoveryChannel, flushFunc)

	logger.Print(" => Program is running...")
	logger.SetVerboseQuiet(true)
}

func (this *Processor) StartPipelineUnits(config *config.Config, recoveryChannel channel.Channel, operationId, address uint32) func() {

	// Initialize channels
	addressChannel := channel.New(config.InstructionsFetchedPerCycle()) // Branch Predictor -> Fetch
	instructionChannel := channel.New(config.InstructionsQueue())       // Fetch -> Decode
	operationChannel := channel.New(config.InstructionsDecodedQueue())  // Decode -> Dispatch
	executionChannels := map[info.CategoryEnum]channel.Channel{         // Dispatch -> Execute
		info.Aritmetic:     channel.New(config.AluUnits()),
		info.LoadStore:     channel.New(config.LoadStoreUnits()),
		info.Control:       channel.New(config.BranchUnits()),
		info.FloatingPoint: channel.New(config.FpuUnits()),
	}
	commonDataBusChannel := channel.New(channel.INFINITE) // Execute -> Dispatcher (RS & ROB)

	closeUnitsHandlers := []func(){}

	/////////////////////////////////////////////////////////////////////////////
	// Run stage units in parallel                                             //
	//  - Units are executed as go routines https://blog.golang.org/pipelines) //
	/////////////////////////////////////////////////////////////////////////////

	// ---------- Fetch ------------ //
	fe := fetcher.New(uint32(0), this, config.InstructionsFetchedPerCycle(), config.BranchPredictorType())
	fe.Run(addressChannel, instructionChannel)
	closeUnitsHandlers = append(closeUnitsHandlers, fe.Close)

	// ---------- Decode ------------ //
	for index := uint32(0); index < config.DecoderUnits(); index++ {
		de := decoder.New(index, this)
		de.Run(instructionChannel, operationChannel)
		closeUnitsHandlers = append(closeUnitsHandlers, de.Close)
	}

	// ----- Dispatch / RS / ROB ---- //
	di := dispatcher.New(uint32(0), this, operationId, config.TotalRegisters(),
		config.ReservationStationEntries(), config.ReorderBufferEntries(), config.InstructionsDispatchedPerCycle(),
		config.InstructionsWrittenPerCycle(), config.RegisterAliasTableEntries())
	di.Run(operationChannel, executionChannels, commonDataBusChannel, recoveryChannel)
	closeUnitsHandlers = append(closeUnitsHandlers, di.Close)

	// ------- Execute (Alu) -------- //
	for index := uint32(0); index < config.AluUnits(); index++ {
		ex := executor.New(index, this, di.Bus(), info.Aritmetic)
		ex.Run(executionChannels, commonDataBusChannel)
		closeUnitsHandlers = append(closeUnitsHandlers, ex.Close)
	}

	// ---- Execute (Load Store) ---- //
	for index := uint32(0); index < config.LoadStoreUnits(); index++ {
		ex := executor.New(index, this, di.Bus(), info.LoadStore)
		ex.Run(executionChannels, commonDataBusChannel)
		closeUnitsHandlers = append(closeUnitsHandlers, ex.Close)
	}

	// ------ Execute (Branch) ------ //
	for index := uint32(0); index < config.BranchUnits(); index++ {
		ex := executor.New(index, this, di.Bus(), info.Control)
		ex.Run(executionChannels, commonDataBusChannel)
		closeUnitsHandlers = append(closeUnitsHandlers, ex.Close)
	}

	// ------- Execute (FPU) -------- //
	for index := uint32(0); index < config.FpuUnits(); index++ {
		ex := executor.New(index, this, di.Bus(), info.FloatingPoint)
		ex.Run(executionChannels, commonDataBusChannel)
		closeUnitsHandlers = append(closeUnitsHandlers, ex.Close)
	}

	// Set instruction for fetching to start pipeline
	go addressChannel.Add(operation.New(operationId, address))

	// Return flush function for all channels
	return func() {
		// Close units & channels
		for _, closeUnitHandler := range closeUnitsHandlers {
			closeUnitHandler()
		}

		// Close channels
		addressChannel.Close()
		instructionChannel.Close()
		operationChannel.Close()
		executionChannels[info.Aritmetic].Close()
		executionChannels[info.LoadStore].Close()
		executionChannels[info.Control].Close()
		executionChannels[info.FloatingPoint].Close()
		commonDataBusChannel.Close()
	}
}

func (this *Processor) RunRecovery(recoveryChannel channel.Channel, flushFunc func()) {
	for value := range recoveryChannel.Channel() {
		op := operation.Cast(value)
		logger.Collect(" => Recovering at OpId: %d and Address: %#04X", op.Id(), op.Address())

		logger.SetVerboseQuiet(true)
		// Flush pipeline
		flushFunc()
		// Clean logs
		this.RemoveForwardLogs(op.Id() - 1)
		// Clear speculative jumps
		this.ClearSpeculativeJumps()
		// Start pipeline from the recovery address
		flushFunc = this.StartPipelineUnits(this.Config(), recoveryChannel, op.Id(), op.Address())
		// Release value from channel
		recoveryChannel.Release()
	}
}
