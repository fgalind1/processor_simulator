package info

import (
	"fmt"

	"app/simulator/processor/models/data"
)

type CategoryEnum string

const (
	Aritmetic CategoryEnum = "Artitmetic"
	LoadStore CategoryEnum = "Load Store"
	Control   CategoryEnum = "Control"
)

type Info struct {
	Opcode   uint8
	Name     string
	Category CategoryEnum
	Type     data.TypeEnum
	Cycles   uint8
}

func New(opcode uint8, name string, category CategoryEnum, datatype data.TypeEnum, cycles uint8) *Info {
	return &Info{
		Opcode:   opcode,
		Name:     name,
		Category: category,
		Type:     datatype,
		Cycles:   cycles,
	}
}

func (this Info) ToString() string {
	return fmt.Sprintf("[%s - %v - %v]", this.Name, this.Category, this.Type)
}

func (this Info) IsBranch() bool {
	return this.Category == Control
}

func (this Info) IsConditionalBranch() bool {
	return this.Category == Control && this.Type != data.TypeJ
}

func (this Info) IsUnconditionalBranch() bool {
	return this.Category == Control && this.Type == data.TypeJ
}
