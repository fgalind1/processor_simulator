package operation

import (
	"app/simulator/processor/models/instruction"
)

type Operation struct {
	*operation
}

type operation struct {
	id                  uint32
	address             uint32
	word                []byte
	instruction         *instruction.Instruction
	renamedDestRegister int32
	predictedAddress    int32
	taken               bool
}

func New(id uint32, address uint32) *Operation {
	return &Operation{
		&operation{
			id:                  id,
			address:             address,
			predictedAddress:    -1,
			taken:               false,
			renamedDestRegister: -1,
		},
	}
}

func Cast(object interface{}) *Operation {
	return object.(*Operation)
}

func (this *Operation) Id() uint32 {
	return this.operation.id
}

func (this *Operation) Address() uint32 {
	return this.operation.address
}

func (this *Operation) Word() []byte {
	return this.operation.word
}

func (this *Operation) Instruction() *instruction.Instruction {
	return this.operation.instruction
}

func (this *Operation) Taken() bool {
	return this.operation.taken
}

func (this *Operation) RenamedDestRegister() int32 {
	return this.operation.renamedDestRegister
}

func (this *Operation) PredictedAddress() int32 {
	return this.operation.predictedAddress
}

func (this *Operation) SetWord(word []byte) {
	this.operation.word = word
}

func (this *Operation) SetInstruction(instruction *instruction.Instruction) {
	this.operation.instruction = instruction
}

func (this *Operation) SetNextPredictedAddress(address uint32) {
	this.operation.predictedAddress = int32(address)
}

func (this *Operation) SetBranchResult(taken bool) {
	this.operation.taken = taken
}

func (this *Operation) SetRenamedDestRegister(register uint32) {
	this.operation.renamedDestRegister = int32(register)
}
