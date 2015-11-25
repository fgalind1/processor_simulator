package decoder

import (
	"errors"
	"fmt"

	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/instruction"
	"app/simulator/processor/models/operation"
)

type Decoder struct {
	*decoder
}

type decoder struct {
	index     uint32
	processor iprocessor.IProcessor
	isActive  bool
}

func New(index uint32, processor iprocessor.IProcessor) *Decoder {
	return &Decoder{
		&decoder{
			index:     index,
			processor: processor,
			isActive:  true,
		},
	}
}

func (this *Decoder) Index() uint32 {
	return this.decoder.index
}

func (this *Decoder) Processor() iprocessor.IProcessor {
	return this.decoder.processor
}

func (this *Decoder) IsActive() bool {
	return this.decoder.isActive
}

func (this *Decoder) Close() {
	this.decoder.isActive = false
}

func (this *Decoder) Run(input, output channel.Channel) {
	// Launch each unit as a goroutine
	logger.Print(" => Initializing decoder unit %d", this.Index())
	go func() {
		for {
			value, running := <-input.Channel()
			if !running || !this.IsActive() {
				logger.Print(" => Flushing decoder unit %d", this.Index())
				return
			}
			op := operation.Cast(value)
			// Iterate instructions received via the channel
			instruction, err := this.decodeInstruction(op)
			if err != nil {
				logger.Error(err.Error())
				break
			}
			// Send data to output
			op.SetInstruction(instruction)
			output.Add(op)
			// Release one item from input Channel
			input.Release()
		}
	}()
}

func (this *Decoder) decodeInstruction(op *operation.Operation) (*instruction.Instruction, error) {
	startCycles := this.Processor().Cycles()

	// Do decode once a data instruction is received
	instruction, err := this.Processor().InstructionsSet().GetInstructionFromBytes(op.Word())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed decoding instruction %#04X. %s]", op.Word(), err.Error()))
	}
	logger.Collect(" => [DE%d][%03d]: %#04X = %s, %s", this.Index(), op.Id(), op.Word(), instruction.Info.ToString(), instruction.Data.ToString())

	// Wait cycles of a decode stage
	this.Processor().Wait(consts.DECODE_CYCLES)

	// Log event
	if this.IsActive() {
		this.Processor().LogEvent(consts.DECODE_EVENT, this.Index(), op.Id(), startCycles)
	}
	return instruction, nil
}
