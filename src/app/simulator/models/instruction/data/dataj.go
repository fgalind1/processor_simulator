package data

import (
	"errors"
	"fmt"

	"app/simulator/models/bits"
)

type DataJ struct {
	Opcode  bits.Bits // 6 bit
	Address bits.Bits // 26 bit
}

func (this *DataJ) ToUint32() uint32 {
	return bits.Concat(this.Opcode, this.Address).ToUint32()
}

func (this *DataJ) ToString() string {
	return fmt.Sprintf("[Op = %s, Address = %s]",
		this.Opcode.ToString(), this.Address.ToString())
}

func (this *DataJ) ToInterface() interface{} {
	return this
}

func getDataJFromUint32(data uint32) (*DataJ, error) {
	bits := bits.FromUint32(data, 32)
	return &DataJ{
		Opcode:  bits.Slice(31, 26),
		Address: bits.Slice(25, 0),
	}, nil
}

func getDataJFromParts(parts ...uint32) (*DataJ, error) {

	// check length of parts and set parts into the right operands
	if len(parts) != 2 {
		return nil, errors.New(fmt.Sprintf("Data I expecting 3 or 4 parts and got %d", len(parts)))
	}

	return &DataJ{
		Opcode:  bits.FromUint32(parts[0], 6),
		Address: bits.FromUint32(parts[1], 26),
	}, nil
}
