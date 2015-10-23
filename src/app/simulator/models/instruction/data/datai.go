package data

import (
	"errors"
	"fmt"

	"app/simulator/models/bits"
)

type DataI struct {
	Opcode    bits.Bits // 6 bit
	RegisterD bits.Bits // 5 bit
	RegisterS bits.Bits // 5 bit
	Immediate bits.Bits // 16 bit
}

func (this *DataI) ToUint32() uint32 {
	return bits.Concat(this.Opcode, this.RegisterD, this.RegisterS, this.Immediate).ToUint32()
}

func (this *DataI) ToString() string {
	return fmt.Sprintf("[Op = %s, Rd = %s, Rs = %s, Im = %s]",
		this.Opcode.ToString(), this.RegisterD.ToString(), this.RegisterS.ToString(), this.Immediate.ToString())
}

func (this *DataI) ToInterface() interface{} {
	return this
}

func getDataIFromUint32(data uint32) (*DataI, error) {
	bits := bits.FromUint32(data, 32)
	return &DataI{
		Opcode:    bits.Slice(31, 26),
		RegisterD: bits.Slice(25, 21),
		RegisterS: bits.Slice(20, 16),
		Immediate: bits.Slice(15, 0),
	}, nil
}

func getDataIFromParts(parts ...uint32) (*DataI, error) {

	// check length of parts and set parts into the right operands
	var values []uint32
	if len(parts) == 4 {
		values = []uint32{parts[0], parts[1], parts[2], parts[3]}
	} else if len(parts) == 3 {
		values = []uint32{parts[0], parts[1], 0, parts[2]}
	} else {
		return nil, errors.New(fmt.Sprintf("Data I expecting 3 or 4 parts and got %d", len(parts)))
	}

	return &DataI{
		Opcode:    bits.FromUint32(values[0], 6),
		RegisterD: bits.FromUint32(values[1], 5),
		RegisterS: bits.FromUint32(values[2], 5),
		Immediate: bits.FromUint32(values[3], 16),
	}, nil
}
