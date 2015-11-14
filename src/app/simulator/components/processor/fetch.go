package processor

import (
	"fmt"
	"strings"
	"time"

	"app/logger"
	"app/simulator/components/channel"
	"app/simulator/consts"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
)

func (this *Processor) LaunchFetchUnit(input, output channel.Channel) {
	// Launch each unit as a goroutine
	logger.Print(" => Initializing fetch unit %d", 0)
	go func() {
		for value := range input.Channel() {
			// Check program reach end
			data := this.InstructionsMemory().Load(value.(uint32), consts.BYTES_PER_WORD)
			if this.reachedEnd(data) {
				this.processor.done = true
				break
			}

			this.fetchInstruction(value.(uint32), data, input, output)
			// Release one item from input Channel
			input.Release()
		}
	}()
}

func (this *Processor) fetchInstruction(address uint32, data []byte, input, output channel.Channel) error {
	startCycles := this.Cycles()
	this.LogInstructionDispatched(address)

	// Do fetch once a new address is received
	msg := fmt.Sprintf(" => [F0]: INS[%#04X] = %#04X", address, data)
	value, ok := this.InstructionsMap()[address]
	if ok {
		msg = fmt.Sprintf("%s // %s", msg, strings.TrimSpace(strings.Split(value, "=>")[1]))
	}
	logger.Collect(msg)

	// Wait cycles of a fetch stage
	this.Wait(consts.FETCH_CYCLES)
	this.LogEvent(consts.FETCH_EVENT, consts.FETCH_EVENT, 0, startCycles)

	// Send data content to decoder input channel and release current channel
	output.Add(instruction.Word{data[0], data[1], data[2], data[3]})

	// Add next instruction for fetching
	go func() {
		nextAddress, _ := this.getNextInstructionAddress(address, data)
		nextData := this.InstructionsMemory().Load(nextAddress, consts.BYTES_PER_WORD)
		if this.reachedEnd(nextData) {
			this.processor.done = true
		}
		input.Add(nextAddress)
	}()
	return nil
}

func (this *Processor) reachedEnd(bytes []byte) bool {
	for _, b := range bytes {
		if b != consts.ENDING_BYTE {
			return false
		}
	}
	return true
}

func (this *Processor) getNextInstructionAddress(currentAddress uint32, data []byte) (uint32, error) {
	// Pre-decode to see if it is a branch instruction
	instruction, _ := this.InstructionsSet().GetInstructionFromBytes(data)
	// Branching/Control instructions
	if instruction.Info.Category == info.Control {
		// Stall until previous instruction finishes
		logger.Collect(" => [F0]: Branch detected, wait to finish instructions on the queue...")
		for len(this.InstructionsDispatched()) != len(this.InstructionsCompleted()) {
			this.Wait(1)
			// Let execute stage add instruction into InstructionsCompleted queue first and then compare lenghts
			time.Sleep(this.Config().CyclePeriodMs() / 2)
		}
		logger.Collect(" => [F0]: Waited for address resolution and got %#04X", this.ProgramCounter())
		return this.ProgramCounter(), nil
	} else {
		return currentAddress + consts.BYTES_PER_WORD, nil
	}
}
