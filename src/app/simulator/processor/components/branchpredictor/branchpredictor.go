package branchpredictor

import (
	"errors"
	"time"

	"app/logger"
	"app/simulator/iprocessor"
	"app/simulator/processor/components/pipeline/executor/branch"
	"app/simulator/processor/config"
	"app/simulator/processor/consts"
	"app/simulator/processor/models/info"
	"app/simulator/processor/models/instruction"
)

type BranchPredictor struct {
	*branchPredictor
}

type branchPredictor struct {
	predictorType config.PredictorType
	index         uint32
	processor     iprocessor.IProcessor
}

func New(predictorType config.PredictorType, index uint32, processor iprocessor.IProcessor) *BranchPredictor {
	return &BranchPredictor{
		&branchPredictor{
			predictorType: predictorType,
			index:         index,
			processor:     processor,
		},
	}
}

func (this *BranchPredictor) PredictorType() config.PredictorType {
	return this.branchPredictor.predictorType
}

func (this *BranchPredictor) Index() uint32 {
	return this.branchPredictor.index
}

func (this *BranchPredictor) Processor() iprocessor.IProcessor {
	return this.branchPredictor.processor
}

func (this *BranchPredictor) PreDecodeInstruction(address uint32) (bool, *instruction.Instruction) {

	// Pre-decode to see if it is a branch instruction
	data := this.Processor().InstructionsMemory().Load(address, consts.BYTES_PER_WORD)
	instruction, _ := this.Processor().InstructionsSet().GetInstructionFromBytes(data)

	// Check if next instruction will need to wait because of a branch instruction
	needsWait, _ := this.needsWait(instruction.Info)
	return needsWait, instruction
}

func (this *BranchPredictor) GetNextAddress(address uint32, instruction *instruction.Instruction) (uint32, bool, error) {

	nextData := this.Processor().InstructionsMemory().Load(address, consts.BYTES_PER_WORD)
	opId := this.Processor().InstructionsFetchedCounter() - 1

	// Check if next instruction is valid
	if this.Processor().ReachedEnd(nextData) {
		this.waitQueueInstructions()
		this.processor.Finish()
		logger.Collect(" => [BP%d][%03d]: Program reached the end", this.Index(), opId)
		return 0, false, errors.New("Program reached the end")
	}

	// If it needs to wait
	needsWait, predicted := this.needsWait(instruction.Info)
	if needsWait {
		// Stall until previous instruction finishes
		logger.Collect(" => [BP%d][%03d]: Branch detected, wait to finish queue (%d out of %d)...",
			this.Index(), this.Processor().InstructionsFetchedCounter()-1, this.Processor().InstructionsCompletedCounter(), opId)
		this.waitQueueInstructions()
		logger.Collect(" => [BP%d][%03d]: Waited for address resolution and got %#04X", this.Index(), opId, this.Processor().ProgramCounter())
		return this.Processor().ProgramCounter(), false, nil
	} else {
		newAddress := this.guessAddress(address, instruction)
		if instruction.Info.IsBranch() {
			logger.Collect(" => [BP%d][%03d]: Predicted address: %#04X", this.Index(), opId, newAddress)
		}
		return newAddress, predicted, nil
	}
}

func (this *BranchPredictor) waitQueueInstructions() {
	for !this.isInstructionCompleted(this.Processor().InstructionsFetchedCounter() - 1) {
		logger.Collect(" => [BP%d][%03d]: Branch detected, wait to finish queue (%d out of %d)...",
			this.Index(), this.Processor().InstructionsFetchedCounter()-1, this.Processor().InstructionsCompletedCounter(), this.Processor().InstructionsFetchedCounter())
		this.Processor().Wait(1)
		// Let execute stage add instruction into InstructionsCompleted queue first and then compare lenghts
		time.Sleep(this.Processor().Config().CyclePeriod() / 2)
	}
}

func (this *BranchPredictor) isInstructionCompleted(operationId uint32) bool {
	for _, completedId := range this.Processor().InstructionsCompleted() {
		if completedId == operationId {
			return true
		}
	}
	return false
}

//////////////////////////////////////////////
//      PreditorType specific methods       //
//////////////////////////////////////////////

func (this *BranchPredictor) needsWait(info *info.Info) (bool, bool) {
	needsWait := info.IsConditionalBranch() && this.PredictorType() == config.StallPredictor
	speculativeExecution := info.IsConditionalBranch() && this.PredictorType() != config.StallPredictor
	return needsWait, speculativeExecution
}

func (this *BranchPredictor) guessAddress(currentAddress uint32, instruction *instruction.Instruction) uint32 {
	if instruction.Info.IsUnconditionalBranch() {
		// These are always taken
		return branch.ComputeAddressTypeJ(instruction.Data)
	}
	if instruction.Info.IsConditionalBranch() {
		offset := branch.ComputeOffsetTypeI(instruction.Data)

		switch this.PredictorType() {
		case config.AlwaysTakenPredictor:
			return uint32(int32(currentAddress) + offset + consts.BYTES_PER_WORD)
		case config.NeverTakenPredictor:
			return currentAddress + consts.BYTES_PER_WORD
		case config.BackwardTakenPredictor:
			if offset < 0 {
				return uint32(int32(currentAddress) + offset + consts.BYTES_PER_WORD)
			} else {
				return currentAddress + consts.BYTES_PER_WORD
			}
		case config.ForwardTakenPredictor:
			if offset > 0 {
				return uint32(int32(currentAddress) + offset + consts.BYTES_PER_WORD)
			} else {
				return currentAddress + consts.BYTES_PER_WORD
			}
		case config.OneBitPredictor:
			taken := this.Processor().GetGuessByAddress(currentAddress)
			if taken {
				return uint32(int32(currentAddress) + offset + consts.BYTES_PER_WORD)
			}
			return currentAddress + consts.BYTES_PER_WORD
		}
	}
	return currentAddress + consts.BYTES_PER_WORD
}
