package info

import (
	"fmt"

	"app/simulator/models/instruction/data"
)

type CategoryEnum string

const (
	Aritmetic CategoryEnum = "Artitmetic"
	Data      CategoryEnum = "Data"
	Control   CategoryEnum = "Control"
)

type Info struct {
	Opcode   uint8
	Name     string
	Category CategoryEnum
	Type     data.TypeEnum
}

func New(opcode uint8, name string, category CategoryEnum, datatype data.TypeEnum) *Info {
	return &Info{
		Opcode:   opcode,
		Name:     name,
		Category: category,
		Type:     datatype,
	}
}

func (this Info) ToString() string {
	return fmt.Sprintf("[%s - %v - %v]", this.Name, this.Category, this.Type)
}
