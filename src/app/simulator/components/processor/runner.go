package processor

import (
	"strings"

	"app/logger"
	"app/simulator/consts"
)

func (this *Processor) HasNext(bytes []byte) bool {
	for _, b := range bytes {
		if b != consts.ENDING_BYTE {
			return true
		}
	}
	return false
}

func (this *Processor) RunNext() int {

	if this.ProgramCounter() == 0 {
		logger.Print("\n------------- PROGRAM STARTED --------------\n")
	}
	logger.Print("---------------- PC: %#04X ----------------", this.ProgramCounter())

	value, ok := this.InstructionsMap()[this.ProgramCounter()]
	if ok {
		logger.Print(" // %s", strings.TrimSpace(strings.Split(value, "=>")[1]))
	}

	///////////////////////////
	//      Fetch stage      //
	///////////////////////////
	data, err := this.Fetch(this.ProgramCounter())
	if err != nil {
		logger.Error("Failed fetching instruction at PC = %#04X. %s]", this.ProgramCounter(), err.Error())
		return consts.INSTRUCTION_FAIL
	}
	// Check if the program ends
	if !this.HasNext(data) {
		logger.Print("\n------------- PROGRAM FINISHED -------------")
		return consts.INSTRUCTION_REACHED_END
	}
	logger.Print(" => [F]: [I = % X]", data)

	///////////////////////////
	//      Decode stage     //
	///////////////////////////
	instruction, err := this.Decode(data)
	if err != nil {
		logger.Error("Failed decoding instruction at PC = %#04X. %s]", this.ProgramCounter(), err.Error())
		return consts.INSTRUCTION_FAIL
	}
	logger.Print(" => [D]: %s", instruction.Info.ToString())
	logger.Print(" => [D]: %s", instruction.Data.ToString())

	///////////////////////////
	//      Execute stage    //
	///////////////////////////
	err = this.Execute(instruction)
	if err != nil {
		logger.Error("Failed executing instruction at PC = %#04X. %s]", this.ProgramCounter(), err.Error())
		return consts.INSTRUCTION_FAIL
	}

	this.IncrementProgramCounter(consts.BYTES_PER_WORD)
	logger.Print("")
	return consts.INSTRUCTION_OK
}
