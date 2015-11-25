package data

import (
	"errors"
	"fmt"

	"app/simulator/processor/models/bits"
)

type DataR struct {
	Opcode    bits.Bits // 6 bit
	RegisterD bits.Bits // 5 bit
	RegisterS bits.Bits // 5 bit
	RegisterT bits.Bits // 5 bit
	Shamt     bits.Bits // 5 bit
	Funct     bits.Bits // 6 bit
}

func (this *DataR) ToUint32() uint32 {
	return bits.Concat(this.Opcode, this.RegisterD, this.RegisterS, this.RegisterT, this.Shamt, this.Funct).ToUint32()
}

func (this *DataR) ToString() string {
	return fmt.Sprintf("[Op = %s, Rd = %s, Rs = %s, Rt = %s]",
		this.Opcode.ToString(), this.RegisterD.ToString(), this.RegisterS.ToString(), this.RegisterT.ToString())
}

func (this *DataR) ToInterface() interface{} {
	return this
}

func getDataRFromUint32(data uint32) (*DataR, error) {
	bits := bits.FromUint32(data, 32)
	return &DataR{
		Opcode:    bits.Slice(31, 26),
		RegisterD: bits.Slice(25, 21),
		RegisterS: bits.Slice(20, 16),
		RegisterT: bits.Slice(15, 11),
		Shamt:     bits.Slice(10, 6),
		Funct:     bits.Slice(5, 0),
	}, nil
}

func getDataRFromParts(parts ...uint32) (*DataR, error) {

	// check length of parts and set parts into the right operands
	if len(parts) != 4 {
		return nil, errors.New(fmt.Sprintf("Data R expecting 4 parts and got %d", len(parts)))
	}

	return &DataR{
		Opcode:    bits.FromUint32(parts[0], 6),
		RegisterD: bits.FromUint32(parts[1], 5),
		RegisterS: bits.FromUint32(parts[2], 5),
		RegisterT: bits.FromUint32(parts[3], 5),
		Shamt:     bits.FromUint32(0, 5),
		Funct:     bits.FromUint32(0, 6),
	}, nil
}
