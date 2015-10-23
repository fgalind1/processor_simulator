package instruction

import (
	"app/simulator/models/instruction/data"
	"app/simulator/models/instruction/info"
)

type Instruction struct {
	Info *info.Info
	Data data.Data
}

func New(info *info.Info, data data.Data) *Instruction {
	return &Instruction{
		Info: info,
		Data: data,
	}
}

func (this *Instruction) ToUint32() uint32 {
	return this.Data.ToUint32()
}
