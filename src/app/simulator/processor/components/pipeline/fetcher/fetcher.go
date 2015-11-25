package fetcher

import (
	"fmt"
	"strings"

	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/branchpredictor"
	"app/simulator/processor/components/channel"
	"app/simulator/processor/config"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/operation"
)

type Fetcher struct {
	*fetcher
}

type fetcher struct {
	index                       uint32
	processor                   iprocessor.IProcessor
	branchPredictor             *branchpredictor.BranchPredictor
	instructionsFetchedPerCycle uint32
	isActive                    bool
}

func New(index uint32, processor iprocessor.IProcessor, instructionsFetchedPerCycle uint32, branchPredictorType config.PredictorType) *Fetcher {
	return &Fetcher{
		&fetcher{
			index:                       index,
			processor:                   processor,
			instructionsFetchedPerCycle: instructionsFetchedPerCycle,
			branchPredictor:             branchpredictor.New(branchPredictorType, index, processor),
			isActive:                    true,
		},
	}
}

func (this *Fetcher) Index() uint32 {
	return this.fetcher.index
}

func (this *Fetcher) Processor() iprocessor.IProcessor {
	return this.fetcher.processor
}

func (this *Fetcher) BranchPredictor() *branchpredictor.BranchPredictor {
	return this.fetcher.branchPredictor
}

func (this *Fetcher) InstructionsFetchedPerCycle() uint32 {
	return this.fetcher.instructionsFetchedPerCycle
}

func (this *Fetcher) IsActive() bool {
	return this.fetcher.isActive
}

func (this *Fetcher) Close() {
	this.fetcher.isActive = false
}

func (this *Fetcher) Run(input, output channel.Channel) {
	logger.Print(" => Initializing fetcher unit %d", this.Index())
	// Launch each unit as a goroutine
	go func() {
		for {
			value, running := <-input.Channel()
			if !running || !this.IsActive() {
				logger.Print(" => Flushing fetcher unit %d", this.Index())
				return
			}
			// Release item from input Channel
			input.Release()

			// Initial operation (address)
			op := operation.Cast(value)

			// Load instructions data from memory
			data := this.Processor().InstructionsMemory().Load(op.Address(), consts.BYTES_PER_WORD*this.InstructionsFetchedPerCycle())

			// Fetch instructions
			startCycles := this.Processor().Cycles()
			operations, err := this.fetchInstructions(op, data, input)
			if err != nil {
				logger.Error(err.Error())
				break
			}

			// Wait cycles of a fetch stage
			this.Processor().Wait(consts.FETCH_CYCLES)

			// After wait cycle, notify decode channel with new operations
			for _, op := range operations {
				if this.IsActive() {
					this.Processor().LogEvent(consts.FETCH_EVENT, this.Index(), op.Id(), startCycles)
					output.Add(op)
				}
			}
		}
	}()
}

func (this *Fetcher) fetchInstructions(op *operation.Operation, bytes []byte, input channel.Channel) ([]*operation.Operation, error) {

	initialAddress := op.Address()
	totalInstructions := len(bytes) / consts.BYTES_PER_WORD
	ops := []*operation.Operation{}

	// Analyze each instruction loaded
	for i := 0; i < totalInstructions; i += 1 {

		data := bytes[i*consts.BYTES_PER_WORD : (i+1)*consts.BYTES_PER_WORD]

		// Check program reach end
		if this.Processor().ReachedEnd(data) {
			this.processor.Finish()
			return ops, nil
		}

		// Do fetch once a new address is received
		msg := fmt.Sprintf(" => [FE%d][%03d]: INS[%#04X] = %#04X", this.Index(), this.Processor().InstructionsFetchedCounter(), op.Address(), data)
		value, ok := this.Processor().InstructionsMap()[op.Address()]
		if ok {
			msg = fmt.Sprintf("%s // %s", msg, strings.TrimSpace(strings.Split(value, "=>")[1]))
		}
		logger.Collect(msg)

		// Log event
		this.Processor().LogInstructionFetched(op.Address())

		// Update data into operation and add to array for post-events
		op.SetWord([]byte{data[0], data[1], data[2], data[3]})

		// Add operation to be sent to decode channel
		ops = append(ops, op)

		// Add next instruction for fetching (as many instructions as it supports per cycle)
		needsWait, instruction := this.BranchPredictor().PreDecodeInstruction(op.Address())
		if needsWait {
			logger.Collect(" => [FE%d][%03d]: Wait detected, no fetching more instructions this cycle", this.Index(), this.Processor().InstructionsFetchedCounter()-1)
			// Add next instruction in a go routine as it need to be stalled
			go func() {
				address, _, err := this.BranchPredictor().GetNextAddress(op.Address(), instruction)
				newOp := operation.New(this.Processor().InstructionsFetchedCounter(), address)
				if err == nil {
					input.Add(newOp)
				}
			}()
			return ops, nil
		} else {
			address, predicted, err := this.BranchPredictor().GetNextAddress(op.Address(), instruction)
			// Set current operation added to be decoded the predicted address
			if predicted {
				ops[len(ops)-1].SetNextPredictedAddress(address)
			}

			// Create new operation object
			op = operation.New(this.Processor().InstructionsFetchedCounter(), address)
			// If is the last instruction from the package or the predicted address is outside of the address package
			if err == nil && (i >= totalInstructions-1 || initialAddress+(uint32(i+1)*consts.BYTES_PER_WORD) != op.Address()) {
				input.Add(op)
				return ops, nil
			}
		}
	}
	return ops, nil
}
