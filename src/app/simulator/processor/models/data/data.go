package data

import (
	"errors"
	"fmt"

	"app/simulator/processor/models/bits"
)

/*
 |------+---------------------------------------------------------------|
 | Type | -31-                    Data (bits)                     -0- |
 |------+------------+--------+--------+--------+-----------+-----------|
 |  R   | Opcode (6) | Rs (5) | Rt (5) | Rd (5) | Shamt (5) | Funct (6) |
 |------+------------+--------+--------+--------+-----------+-----------|
 |  I   | Opcode (6) | Rs (5) | Rt (5) |         Immediate (16)         |
 |------+------------+--------+--------+--------------------------------|
 |  J   | Opcode (6) |                  Address (26)                    |
 |------+------------+--------------------------------------------------|
*/

type TypeEnum string

const (
	TypeR TypeEnum = "R"
	TypeI TypeEnum = "I"
	TypeJ TypeEnum = "J"
)

type Data interface {
	// General interface for different data formats (R, I, J)
	ToUint32() uint32
	ToString() string
	ToInterface() interface{}
}

func GetOpcodeFromUint32(value uint32) uint8 {
	bits := bits.FromUint32(value, 32)
	return uint8(bits.Slice(31, 26).ToUint32())
}

func GetDataFromUint32(datatype TypeEnum, value uint32) (Data, error) {
	switch datatype {
	case TypeR:
		return getDataRFromUint32(value)
	case TypeI:
		return getDataIFromUint32(value)
	case TypeJ:
		return getDataJFromUint32(value)
	default:
		return nil, errors.New(fmt.Sprintf("Invalid data type %v", datatype))
	}
}

func GetDataFromParts(datatype TypeEnum, parts ...uint32) (Data, error) {
	switch datatype {
	case TypeR:
		return getDataRFromParts(parts...)
	case TypeI:
		return getDataIFromParts(parts...)
	case TypeJ:
		return getDataJFromParts(parts...)
	default:
		return nil, errors.New(fmt.Sprintf("Invalid data type %v", datatype))
	}
}
