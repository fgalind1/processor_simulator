package instruction

import (
	"app/simulator/processor/models/data"
	"app/simulator/processor/models/info"
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
