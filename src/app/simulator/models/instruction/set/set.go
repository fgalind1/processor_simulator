package set

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"app/simulator/models/instruction/data"
	"app/simulator/models/instruction/info"
	"app/simulator/models/instruction/instruction"
)

type Set []*info.Info

func (this Set) GetInstructionInfoFromName(name string) (*info.Info, error) {
	for _, info := range this {
		if strings.ToLower(info.Name) == strings.ToLower(name) {
			return info, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("No instruction was found with name: %s", name))
}

func (this Set) GetInstructionInfoFromOpcode(opcode uint8) (*info.Info, error) {
	for _, info := range this {
		if info.Opcode == opcode {
			return info, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("No instruction was found with name: %d", opcode))
}

func (this Set) GetInstructionFromString(line string) (*instruction.Instruction, error) {

	line = strings.Replace(line, "\t", " ", -1)
	line = strings.Replace(line, ",", " ", -1)
	line = strings.TrimSpace(line)
	items := strings.Split(line, " ")

	// Search opcode in the instruction set
	info, err := this.GetInstructionInfoFromName(items[0])
	if err != nil {
		return nil, err
	}

	if len(items) <= 1 {
		return nil, errors.New(fmt.Sprintf("Only one operand found in the instruction: %s. Expecting more than one operand", line))
	}

	// Check all operands (except opcode/operation) is a numeric value
	operands := []uint32{uint32(info.Opcode)}
	for _, value := range items[1:] {
		if strings.TrimSpace(value) == "" {
			continue
		}
		integer, err := strconv.Atoi(strings.Replace(value, "R", "", -1))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Expecting an integer or RX operand and found: %s", value))
		}
		operands = append(operands, uint32(integer))
	}

	// Get data object from operands
	data, err := data.GetDataFromParts(info.Type, operands...)
	if err != nil {
		return nil, err
	}

	return instruction.New(info, data), nil
}

func (this Set) GetInstructionFromBytes(bytes []byte) (*instruction.Instruction, error) {

	if len(bytes) != 4 {
		return nil, errors.New(fmt.Sprintf("Expecting 4 bytes and got %d", len(bytes)))
	}

	value := uint32(bytes[0])<<24 + uint32(bytes[1])<<16 + uint32(bytes[2])<<8 + uint32(bytes[3])
	opcode := data.GetOpcodeFromUint32(value)

	// Search opcode in the instruction set
	info, err := this.GetInstructionInfoFromOpcode(opcode)
	if err != nil {
		return nil, err
	}

	// Get data object from operands
	data, err := data.GetDataFromUint32(info.Type, value)
	if err != nil {
		return nil, err
	}

	return instruction.New(info, data), nil
}
