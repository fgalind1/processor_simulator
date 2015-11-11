package processor

import (
	"errors"
	"fmt"
	"strings"

	"app/logger"
	"app/simulator/consts"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
)

/**************************************************************************************************************/

func (this *Processor) Fetch(addressChannel chan uint32, dataChannel chan byte) error {
	// Iterate addresses received via the channel
	for address := range addressChannel {

		// Do fetch once a new address is received
		data := this.InstructionsMemory().Load(address, consts.BYTES_PER_WORD)
		if this.ReachedEnd(data) {
			this.processor.fetchState = consts.INACTIVE
			break
		}

		this.IncrementInstructionsExecuted()
		this.LogEvent(consts.FETCH_EVENT)
		msg := fmt.Sprintf(" => [F]: INS[%#04X] = %#04X", address, data)
		value, ok := this.InstructionsMap()[address]
		if ok {
			msg = fmt.Sprintf("%s // %s", msg, strings.TrimSpace(strings.Split(value, "=>")[1]))
		}
		logger.Print(msg)

		// Wait cycles of a fetch stage
		this.Wait(consts.FETCH_CYCLES)
		this.LogEvent(consts.FETCH_EVENT)

		// Send data content to decoder input channel
		dataChannel <- data[0]
		dataChannel <- data[1]
		dataChannel <- data[2]
		dataChannel <- data[3]

		// Add next instruction for fetching
		go func() {
			addressChannel <- this.getNextInstructionAddress(address)
		}()
	}
	return nil
}

func (this *Processor) getNextInstructionAddress(currentAddress uint32) uint32 {
	return currentAddress + consts.BYTES_PER_WORD
}

/**************************************************************************************************************/

func (this *Processor) Decode(dataChannel <-chan byte, instructionChannel chan *instruction.Instruction) error {
	// Iterate data received via the channel
	word := []byte{}
	for data := range dataChannel {
		word = append(word, data)
		// Process until a full word is received
		if len(word) == consts.BYTES_PER_WORD {
			this.LogEvent(consts.DECODE_EVENT)

			// Do decode once a data instruction is received
			instruction, err := this.InstructionsSet().GetInstructionFromBytes(word)
			if err != nil {
				return errors.New(fmt.Sprintf("Failed decoding instruction %#04X. %s]", word, err.Error()))
			}
			logger.Print(" => [D]: %#04X = %s, %s", word, instruction.Info.ToString(), instruction.Data.ToString())
			word = []byte{}

			// Wait cycles of a decode stage
			this.Wait(consts.DECODE_CYCLES)
			this.LogEvent(consts.DECODE_EVENT)

			// Send data content to execution units input channel
			instructionChannel <- instruction

			// Check when this stage is done
			if this.processor.fetchState == consts.INACTIVE {
				this.processor.decodeState = consts.INACTIVE
				break
			}
		}
	}
	return nil
}

/**************************************************************************************************************/

func (this *Processor) Execute(instructionChannel <-chan *instruction.Instruction) error {
	// Iterate instructions received via the channel
	for instruction := range instructionChannel {
		this.LogEvent(consts.EXECUTE_EVENT)

		// Do decode once a data instruction is received
		err := this.doExecute(instruction)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed executing instruction. %s]", err.Error()))
		}
		logger.Print(" => [E]: Executed %s, %s", instruction.Info.ToString(), instruction.Data.ToString())

		// Wait cycles of a execution stage
		this.Wait(consts.EXECUTE_CYCLES)
		this.LogEvent(consts.EXECUTE_EVENT)

		// Check when this stage is done
		if this.processor.decodeState == consts.INACTIVE {
			this.processor.executeState = consts.INACTIVE
			break
		}
	}
	return nil
}

func (this *Processor) doExecute(instruction *instruction.Instruction) error {
	var err error
	switch instruction.Info.Category {
	case info.Aritmetic:
		err = this.AluUnit().Process(instruction)
	case info.Data:
		err = this.DataUnit().Process(instruction)
	case info.Control:
		err = this.ControlUnit().Process(instruction)
	default:
		err = errors.New("Invalid instruction info category")
	}
	this.IncrementProgramCounter(consts.BYTES_PER_WORD)
	return err
}
