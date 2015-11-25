package iprocessor

import (
	"app/simulator/processor/components/clock"
	"app/simulator/processor/components/memory"
	"app/simulator/processor/config"
	"app/simulator/processor/models/set"
)

type IProcessor interface {
	Start()

	///////////////////////////
	//       Internals       //
	///////////////////////////
	Finish()
	InstructionsFetched() []string
	InstructionsFetchedCounter() uint32
	InstructionsCompleted() []uint32
	InstructionsCompletedCounter() uint32
	LogInstructionFetched(address uint32)
	LogInstructionCompleted(operationId uint32)
	LogEvent(unit string, index uint32, operationId uint32, start uint32)
	LogEventStart(unit string, index uint32, operationId uint32)
	LogEventFinish(unit string, index uint32, operationId uint32)
	LogBranchInstruction(conditionalBranch, mispredicted bool)
	RemoveForwardLogs(operationId uint32)
	ReachedEnd(bytes []byte) bool

	///////////////////////////
	//       Metadata        //
	///////////////////////////
	InstructionsMap() map[uint32]string
	InstructionsSet() set.Set
	Config() *config.Config
	DataMemory() *memory.Memory
	InstructionsMemory() *memory.Memory
	RegistersMemory() *memory.Memory
	ProgramCounter() uint32
	SetProgramCounter(value uint32)
	IncrementProgramCounter(offset int32)
	GetGuessByAddress(address uint32) bool
	SetBranchResult(address uint32, taken bool)

	///////////////////////////
	//       Clock           //
	///////////////////////////
	Clock() *clock.Clock
	Cycles() uint32
	DurationMs() uint32
	RunClock()
	PauseClock()
	ContinueClock()
	WaitClock()
	FinishedClock() bool
	NextCycle() int
	Wait(cycles uint32)
}
