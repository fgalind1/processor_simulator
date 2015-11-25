package processor

import (
	"fmt"
	"strings"
	"time"

	"app/logger"
	"app/simulator/processor/components/clock"
	"app/simulator/processor/components/memory"
	"app/simulator/processor/config"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/set"
)

type Processor struct {
	*processor
}

type processor struct {
	// internals
	done                  bool
	instructionsFetched   []string
	instructionsCompleted []uint32
	dataLog               map[uint32][]LogEvent

	// Branch stats
	conditionalBranches   uint32
	unconditionalBranches uint32
	mispredictedBranches  uint32

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
}

///////////////////////////
//       Internals       //
///////////////////////////

func (this *Processor) Finish() {
	logger.Collect(" => Stopping processor...")
	this.processor.done = true
}

func (this *Processor) InstructionsFetched() []string {
	return this.processor.instructionsFetched
}

func (this *Processor) InstructionsFetchedCounter() uint32 {
	return uint32(len(this.processor.instructionsFetched))
}

func (this *Processor) InstructionsCompleted() []uint32 {
	return this.processor.instructionsCompleted
}

func (this *Processor) InstructionsCompletedCounter() uint32 {
	return uint32(len(this.processor.instructionsCompleted))
}

func (this *Processor) LogInstructionFetched(address uint32) {
	value, ok := this.InstructionsMap()[address]
	if ok {
		value = strings.Split(value, "=>")[1]
	} else {
		value = fmt.Sprintf(" %#04X", address)
	}
	this.processor.instructionsFetched = append(this.processor.instructionsFetched, value)
}

func (this *Processor) LogInstructionCompleted(operationId uint32) {
	this.processor.instructionsCompleted = append(this.processor.instructionsCompleted, operationId)
}

func (this *Processor) LogEvent(unit string, index uint32, operationId uint32, start uint32) {

	_, ok := this.processor.dataLog[operationId]
	if !ok {
		this.processor.dataLog[operationId] = []LogEvent{}
	}

	event := NewEvent(fmt.Sprintf("%2s%d", unit, index), start, this.Cycles())
	this.processor.dataLog[operationId] = append(this.processor.dataLog[operationId], event)
}

func (this *Processor) LogEventStart(unit string, index uint32, operationId uint32) {

	_, ok := this.processor.dataLog[operationId]
	if !ok {
		this.processor.dataLog[operationId] = []LogEvent{}
	}

	found := false
	eventId := fmt.Sprintf("%2s%d", unit, index)
	for i, _ := range this.processor.dataLog[operationId] {
		if eventId == this.processor.dataLog[operationId][i].Id {
			this.processor.dataLog[operationId][i].Start = this.Cycles()
			found = true
			break
		}
	}

	if !found {
		this.LogEvent(unit, index, operationId, this.Cycles())
	}
}

func (this *Processor) LogEventFinish(unit string, index uint32, operationId uint32) {
	eventId := fmt.Sprintf("%2s%d", unit, index)
	for i, _ := range this.processor.dataLog[operationId] {
		if eventId == this.processor.dataLog[operationId][i].Id {
			this.processor.dataLog[operationId][i].End = this.Cycles()
			break
		}
	}
}

func (this *Processor) LogBranchInstruction(conditionalBranch, mispredicted bool) {
	if conditionalBranch {
		this.processor.conditionalBranches += 1
	} else {
		this.processor.unconditionalBranches += 1
	}
	if mispredicted {
		this.processor.mispredictedBranches += 1
	}
}

func (this *Processor) RemoveForwardLogs(operationId uint32) {
	// Remove forward ops from instructionsFetched
	if uint32(len(this.processor.instructionsFetched)) > operationId+1 {
		this.processor.instructionsFetched = this.processor.instructionsFetched[:operationId+1]
	}

	// Remove forward ops from instructionsCompleted
	if uint32(len(this.processor.instructionsCompleted)) > operationId+1 {
		this.processor.instructionsCompleted = this.processor.instructionsCompleted[:operationId+1]
	}

	// Remove forward ops from data log
	opsIdToDelete := []uint32{}
	for opId, _ := range this.processor.dataLog {
		if opId > operationId {
			opsIdToDelete = append(opsIdToDelete, opId)
		}
	}
	for _, opId := range opsIdToDelete {
		delete(this.processor.dataLog, opId)
	}
}

func (this *Processor) ReachedEnd(bytes []byte) bool {
	for _, b := range bytes {
		if b != consts.ENDING_BYTE {
			return false
		}
	}
	return true
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
