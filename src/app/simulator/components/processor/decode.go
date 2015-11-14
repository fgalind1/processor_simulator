package processor

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/components/channel"
	"app/simulator/consts"
	"app/simulator/models/instruction/instruction"
)

func (this *Processor) LaunchDecoderUnits(decoderUnits uint32, input, output channel.Channel) {
	for index := uint32(0); index < decoderUnits; index++ {
		go func(index uint32) {
			// Launch each unit as a goroutine
			logger.Print(" => Initializing decoder unit %d", index)
			for value := range input.Channel() {
				// Iterate instructions received via the channel
				this.decodeInstruction(index, value.(instruction.Word), output)
				// Release one item from input Channel
				input.Release()
			}
		}(index)
	}
}

func (this *Processor) decodeInstruction(index uint32, word instruction.Word, output channel.Channel) error {
	startCycles := this.Cycles()
	data := []byte{word[0], word[1], word[2], word[3]}

	// Do decode once a data instruction is received
	instruction, err := this.InstructionsSet().GetInstructionFromBytes(data)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed decoding instruction %#04X. %s]", data, err.Error()))
	}
	logger.Collect(" => [D%d]: %#04X = %s, %s", index, data, instruction.Info.ToString(), instruction.Data.ToString())

	// Wait cycles of a decode stage
	this.Wait(consts.DECODE_CYCLES)

	// Log and sent data to output
	this.LogEvent(consts.DECODE_EVENT, consts.DECODE_EVENT, index, startCycles)
	output.Add(instruction)
	return nil
}
