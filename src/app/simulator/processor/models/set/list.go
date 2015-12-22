package set

import (
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
)

const (
	OP_ADD   = 0x00
	OP_ADDI  = 0x01
	OP_ADDU  = 0x02
	OP_ADDIU = 0x03
	OP_SUB   = 0x04
	OP_SUBI  = 0x05
	OP_SUBU  = 0x06
	OP_MUL   = 0x07

	OP_SHL  = 0x08
	OP_SHLI = 0x09
	OP_SHR  = 0x0A
	OP_SHRI = 0x0B

	OP_CMP  = 0x0C
	OP_AND  = 0x0D
	OP_ANDI = 0x0E
	OP_OR   = 0x0F
	OP_ORI  = 0x10

	OP_FADD = 0x12
	OP_FSUB = 0x13
	OP_FMUL = 0x14
	OP_FDIV = 0x15

	OP_LW  = 0x20
	OP_SW  = 0x21
	OP_LLI = 0x22
	OP_SLI = 0x23
	OP_LUI = 0x24
	OP_SUI = 0x25

	OP_BEQ = 0x30
	OP_BNE = 0x31
	OP_BLT = 0x32
	OP_BGT = 0x33
	OP_J   = 0x34
)

func Init() Set {
	return []*info.Info{
		info.New(OP_ADD, "add", info.Aritmetic, data.TypeR, 2),
		info.New(OP_ADDI, "addi", info.Aritmetic, data.TypeI, 2),
		info.New(OP_ADDU, "addu", info.Aritmetic, data.TypeR, 2),
		info.New(OP_ADDIU, "addiu", info.Aritmetic, data.TypeR, 2),
		info.New(OP_SUB, "sub", info.Aritmetic, data.TypeR, 2),
		info.New(OP_SUBI, "subi", info.Aritmetic, data.TypeI, 2),
		info.New(OP_SUBU, "subu", info.Aritmetic, data.TypeI, 2),
		info.New(OP_MUL, "mul", info.Aritmetic, data.TypeR, 4),

		info.New(OP_SHL, "shl", info.Aritmetic, data.TypeR, 2),
		info.New(OP_SHLI, "shli", info.Aritmetic, data.TypeR, 2),
		info.New(OP_SHR, "shr", info.Aritmetic, data.TypeR, 2),
		info.New(OP_SHRI, "shri", info.Aritmetic, data.TypeR, 2),

		info.New(OP_CMP, "cmp", info.Aritmetic, data.TypeR, 2),
		info.New(OP_AND, "and", info.Aritmetic, data.TypeR, 2),
		info.New(OP_ANDI, "andi", info.Aritmetic, data.TypeR, 2),
		info.New(OP_OR, "or", info.Aritmetic, data.TypeR, 2),
		info.New(OP_ORI, "ori", info.Aritmetic, data.TypeR, 2),

		info.New(OP_FADD, "fadd", info.FloatingPoint, data.TypeR, 8),
		info.New(OP_FSUB, "fsub", info.FloatingPoint, data.TypeR, 8),
		info.New(OP_FMUL, "fmul", info.FloatingPoint, data.TypeR, 8),
		info.New(OP_FDIV, "fdiv", info.FloatingPoint, data.TypeR, 8),

		info.New(OP_LW, "lw", info.LoadStore, data.TypeI, 2),
		info.New(OP_SW, "sw", info.LoadStore, data.TypeI, 2),
		info.New(OP_LLI, "lli", info.LoadStore, data.TypeI, 1),
		info.New(OP_SLI, "sli", info.LoadStore, data.TypeI, 1),
		info.New(OP_LUI, "lui", info.LoadStore, data.TypeI, 1),
		info.New(OP_SUI, "sui", info.LoadStore, data.TypeI, 1),

		info.New(OP_BEQ, "beq", info.Control, data.TypeI, 1),
		info.New(OP_BNE, "bne", info.Control, data.TypeI, 1),
		info.New(OP_BLT, "blt", info.Control, data.TypeI, 1),
		info.New(OP_BGT, "bgt", info.Control, data.TypeI, 1),
		info.New(OP_J, "j", info.Control, data.TypeJ, 1),
	}
}
