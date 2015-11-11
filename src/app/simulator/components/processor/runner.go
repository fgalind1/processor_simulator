package processor

import (
	"time"

	"app/logger"
	"app/simulator/consts"
	"app/simulator/models/instruction/instruction"
)

func (this *Processor) ReachedEnd(bytes []byte) bool {
	for _, b := range bytes {
		if b != consts.ENDING_BYTE {
			return false
		}
	}
	return true
}

func (this *Processor) Start() {

	// Create channels for go routines (https://blog.golang.org/pipelines)
	addressChannel := make(chan uint32)
	dataChannel := make(chan byte)
	instructionChannel := make(chan *instruction.Instruction)

	// Run pipeline stages in parallel
	this.processor.dataLog[consts.FETCH_EVENT] = []uint32{}
	go this.Fetch(addressChannel, dataChannel)

	this.processor.dataLog[consts.DECODE_EVENT] = []uint32{}
	go this.Decode(dataChannel, instructionChannel)

	this.processor.dataLog[consts.EXECUTE_EVENT] = []uint32{}
	go this.Execute(instructionChannel)

	logger.Print("\n------------- PROGRAM STARTED --------------\n")
	logger.Print("-------------- Cycle: %04d ----------------", this.processor.cycles)

	// Trigger first instruction for fetching
	go func() {
		addressChannel <- uint32(0)
	}()
}

func (this *Processor) NextCycle() int {
	this.IncrementCycles()
	logger.Print("-------------- Cycle: %04d ----------------", this.processor.cycles)
	if this.HasFinished() {
		return consts.PROGRAM_RUNNING
	} else {
		time.Sleep(consts.STEP_PERIOD)
		this.IncrementCycles()
		time.Sleep(consts.STEP_PERIOD)
		logger.Print("-------------- Cycle: %04d ----------------", this.processor.cycles)
		return consts.PROGRAM_FINISHED
	}
}

func (this *Processor) HasFinished() bool {
	return this.processor.fetchState && this.processor.decodeState && this.processor.executeState
}

func (this *Processor) Wait(cycles uint32) {
	currentCycles := this.processor.cycles
	for this.processor.cycles < currentCycles+cycles {
		time.Sleep(consts.WAIT_PERIOD)
	}
}
