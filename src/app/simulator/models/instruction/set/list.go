package set

import (
	"app/simulator/models/instruction/data"
	"app/simulator/models/instruction/info"
)

const (
	OP_ADD   = 0x00
	OP_ADDU  = 0x01
	OP_SUB   = 0x02
	OP_SUBU  = 0x03
	OP_ADDI  = 0x04
	OP_ADDIU = 0x05
	OP_CMP   = 0x06
	OP_MUL   = 0x07

	OP_LW  = 0x10
	OP_SW  = 0x11
	OP_LLI = 0x12
	OP_SLI = 0x13
	OP_LUI = 0x14
	OP_SUI = 0x15

	OP_BEQ = 0x20
	OP_BNE = 0x21
	OP_BLT = 0x22
	OP_BGT = 0x23
	OP_J   = 0x24
)

func Init() Set {
	return []*info.Info{
		info.New(OP_ADD, "add", info.Aritmetic, data.TypeR),
		info.New(OP_ADDU, "addu", info.Aritmetic, data.TypeR),
		info.New(OP_SUB, "sub", info.Aritmetic, data.TypeR),
		info.New(OP_SUBU, "subu", info.Aritmetic, data.TypeI),
		info.New(OP_ADDI, "addi", info.Aritmetic, data.TypeI),
		info.New(OP_ADDIU, "addiu", info.Aritmetic, data.TypeR),
		info.New(OP_CMP, "cmp", info.Aritmetic, data.TypeR),
		info.New(OP_MUL, "mul", info.Aritmetic, data.TypeR),

		info.New(OP_LW, "lw", info.Data, data.TypeI),
		info.New(OP_SW, "sw", info.Data, data.TypeI),
		info.New(OP_LLI, "lli", info.Data, data.TypeI),
		info.New(OP_SLI, "sli", info.Data, data.TypeI),
		info.New(OP_LUI, "lui", info.Data, data.TypeI),
		info.New(OP_SUI, "sui", info.Data, data.TypeI),

		info.New(OP_BEQ, "beq", info.Control, data.TypeI),
		info.New(OP_BNE, "bne", info.Control, data.TypeI),
		info.New(OP_BLT, "blt", info.Control, data.TypeI),
		info.New(OP_BGT, "bgt", info.Control, data.TypeI),
		info.New(OP_J, "j", info.Control, data.TypeJ),
	}
}
