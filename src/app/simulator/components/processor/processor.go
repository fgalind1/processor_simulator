package processor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/logger"
	"app/simulator/components/channel"
	"app/simulator/components/clock"
	"app/simulator/components/memory"
	"app/simulator/components/storagebus"
	"app/simulator/config"
	"app/simulator/consts"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/set"
	"app/utils"
)

type Processor struct {
	*processor
}

type processor struct {
	// internals
	done                   bool
	instructionsDispatched []string
	instructionsCompleted  []string
	dataLog                map[string][]LogEvent

	addressChannel     channel.Channel
	instructionChannel channel.Channel
	operationChannel   channel.Channel
	executionChannels  map[info.CategoryEnum]channel.Channel

	// metadata
	instructionsMap map[uint32]string
	instructionsSet set.Set
	config          *config.Config

	// clock
	clockUnit *clock.Clock

	// data/memory
	programCounter    uint32
	registerMemory    *memory.Memory
	instructionMemory *memory.Memory
	dataMemory        *memory.Memory
	storageBus        *storagebus.StorageBus
}

///////////////////////////
//       Internals       //
///////////////////////////

func (this *Processor) InstructionsDispatched() []string {
	return this.processor.instructionsDispatched
}

func (this *Processor) InstructionsCompleted() []string {
	return this.processor.instructionsCompleted
}

func (this *Processor) LogInstructionDispatched(address uint32) {
	value, ok := this.InstructionsMap()[address]
	if ok {
		value = strings.Split(value, "=>")[1]
	} else {
		value = fmt.Sprintf(" %#04X", address)
	}
	this.processor.instructionsDispatched = append(this.processor.instructionsDispatched, value)
}

func (this *Processor) LogInstructionCompleted() {
	value := this.processor.instructionsDispatched[len(this.processor.instructionsCompleted)]
	this.processor.instructionsCompleted = append(this.processor.instructionsCompleted, value)
}

func (this *Processor) LogEvent(stage string, unit string, index uint32, start uint32) {
	id := fmt.Sprintf("%2s%d", unit, index)
	_, ok := this.processor.dataLog[stage]
	if !ok {
		this.processor.dataLog[stage] = []LogEvent{}
	}
	this.processor.dataLog[stage] = append(this.processor.dataLog[stage], NewEvent(id, start, this.Cycles()))
}

///////////////////////////
//       Metadata        //
///////////////////////////

func (this *Processor) InstructionsMap() map[uint32]string {
	return this.processor.instructionsMap
}

func (this *Processor) InstructionsSet() set.Set {
	return this.processor.instructionsSet
}

func (this *Processor) Config() *config.Config {
	return this.processor.config
}

///////////////////////////
//      Data/Memory      //
///////////////////////////

func (this *Processor) StorageBus() *storagebus.StorageBus {
	return this.processor.storageBus
}

func (this *Processor) DataMemory() *memory.Memory {
	return this.processor.dataMemory
}

func (this *Processor) InstructionsMemory() *memory.Memory {
	return this.processor.instructionMemory
}

func (this *Processor) RegistersMemory() *memory.Memory {
	return this.processor.registerMemory
}

func (this *Processor) ProgramCounter() uint32 {
	return this.processor.programCounter
}

func (this *Processor) SetProgramCounter(value uint32) {
	this.processor.programCounter = value
}

func (this *Processor) IncrementProgramCounter(offset int32) {
	if offset < 0 {
		this.processor.programCounter -= uint32(offset * -1)
	} else {
		this.processor.programCounter += uint32(offset)
	}
}

///////////////////////////
//       Clock           //
///////////////////////////

func (this *Processor) Clock() *clock.Clock {
	return this.processor.clockUnit
}

func (this *Processor) Cycles() uint32 {
	return this.processor.clockUnit.Cycles()
}

func (this *Processor) DurationMs() uint32 {
	return this.processor.clockUnit.DurationMs()
}

func (this *Processor) RunClock() {
	this.processor.clockUnit.Run()
}

func (this *Processor) PauseClock() {
	this.processor.clockUnit.Pause()
}

func (this *Processor) ContinueClock() {
	this.processor.clockUnit.Continue()
}

func (this *Processor) WaitClock() {
	this.processor.clockUnit.Wait()
}

func (this *Processor) FinishedClock() bool {
	return this.processor.clockUnit.Finished()
}

func (this *Processor) NextCycle() int {
	if this.FinishedClock() {
		logger.Print(" => Program has finished\n")
		return consts.PROGRAM_FINISHED
	}

	this.WaitClock()
	return consts.PROGRAM_RUNNING
}

func (this *Processor) Wait(cycles uint32) {
	currentCycles := this.Cycles()
	for this.Cycles() < currentCycles+cycles {
		time.Sleep(consts.WAIT_PERIOD)
	}
}

///////////////////////////
//      New & Start      //
///////////////////////////

func New(assemblyFileName string, config *config.Config) (*Processor, error) {

	p := &Processor{
		&processor{
			done: false,
			instructionsDispatched: []string{},
			instructionsCompleted:  []string{},
			dataLog:                map[string][]LogEvent{},

			addressChannel:     channel.New(1),                               // Branch Predictor -> Fetch
			instructionChannel: channel.New(config.InstructionsBufferSize()), // Fetch -> Decode
			operationChannel:   channel.New(1),                               // Decode -> Execute
			executionChannels:  map[info.CategoryEnum]channel.Channel{},      // Execute -> Execution Units

			instructionsMap: map[uint32]string{},
			instructionsSet: set.Init(),
			config:          config,

			programCounter:    0,
			registerMemory:    memory.New(config.RegistersMemorySize()),
			instructionMemory: memory.New(config.InstructionsMemorySize()),
			dataMemory:        memory.New(config.DataMemorySize()),
		},
	}

	logger.Print(" => Bytes per word: %d", consts.BYTES_PER_WORD)
	logger.Print(" => Registers:      %d", config.RegistersMemorySize()/consts.BYTES_PER_WORD)
	logger.Print(" => Instr Memory:   %d Bytes", config.InstructionsMemorySize())
	logger.Print(" => Data Memory:    %d Bytes", config.DataMemorySize())

	p.storageBus = &storagebus.StorageBus{
		LoadRegister: func(address uint32) uint32 {
			return p.RegistersMemory().LoadUint32(address * consts.BYTES_PER_WORD)
		},
		StoreRegister: func(address, value uint32) {
			p.RegistersMemory().StoreUint32(address*consts.BYTES_PER_WORD, value)
		},

		LoadData:  p.DataMemory().LoadUint32,
		StoreData: p.DataMemory().StoreUint32,

		ProgramCounter:          p.ProgramCounter,
		SetProgramCounter:       p.SetProgramCounter,
		IncrementProgramCounter: p.IncrementProgramCounter,
	}

	// Instanciate functional units
	instructionsFinished := func() bool {
		return p.processor.done && len(p.InstructionsDispatched()) == len(p.InstructionsCompleted())
	}
	p.processor.clockUnit = clock.New(config.CyclePeriodMs(), instructionsFinished)

	err := p.loadInstructionsMemory(assemblyFileName)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (this *Processor) loadInstructionsMemory(assemblyFileName string) error {

	logger.Print(" => Reading hex file: %s", assemblyFileName)
	lines, err := utils.ReadLines(assemblyFileName)
	if err != nil {
		return err
	}

	address := uint32(0)
	for _, line := range lines {

		// Split hex value and humand readable comment
		parts := strings.Split(line, "//")

		// Save human readable for debugging purposes
		if len(parts) > 1 {
			this.InstructionsMap()[address] = strings.TrimSpace(parts[1])
		}

		// Save hex value into instructions memory
		bytes, err := hex.DecodeString(strings.TrimSpace(parts[0]))
		if err != nil {
			return errors.New(fmt.Sprintf("Failed parsing instruction (hex) value: %s. %s", parts[0], err.Error()))
		}
		this.InstructionsMemory().Store(address, bytes...)

		// Increment address
		address += uint32(len(bytes))
	}
	this.InstructionsMemory().Store(address, []byte{consts.ENDING_BYTE, consts.ENDING_BYTE, consts.ENDING_BYTE, consts.ENDING_BYTE}...)
	return nil
}

func (this *Processor) Start() {

	logger.CleanBuffer()
	logger.Print("\n Simulation:")
	logger.Print("\n => Starting program...")

	// Run pipeline stages in parallel as go routines (https://blog.golang.org/pipelines)
	this.LaunchFetchUnit(this.processor.addressChannel, this.processor.instructionChannel)
	this.LaunchDecoderUnits(this.Config().DecoderUnits(), this.processor.instructionChannel, this.processor.operationChannel)
	this.LaunchExecuteUnits(this.Config().AluUnits(), this.Config().BranchUnits(), this.Config().LoadStoreUnits(),
		this.processor.operationChannel, this.processor.executionChannels)

	// Launch clock
	go this.RunClock()

	// Set first instruction for fetching
	go func() {
		this.processor.addressChannel.Add(uint32(0))
	}()
	logger.Print(" => Program is running...")
}
