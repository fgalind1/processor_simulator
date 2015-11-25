package bits

import (
	"testing"
)

func TestNew(t *testing.T) {
	// Dec: 118
	// Bin: 1110110
	bits := FromUint32(118, 7)
	if len(bits) != 7 {
		t.Errorf("Lenght expected %d - Got %d", 7, len(bits))
	}

	for _, bit := range []int{0, 3} {
		if bits[bit] {
			t.Errorf("Bit %d expected on low and got high", bit)
		}
	}

	for _, bit := range []int{1, 2, 4, 5, 6} {
		if !bits[bit] {
			t.Errorf("Bit %d expected on high and got low", bit)
		}
	}
}

func TestNewOverflow(t *testing.T) {
	// Dec: 118
	// Bin: 10110
	bits := FromUint32(118, 5) // last three bits are not decoded
	if len(bits) != 5 {
		t.Errorf("Lenght expected %d - Got %d", 5, len(bits))
	}

	for _, bit := range []int{0, 3} {
		if bits[bit] {
			t.Errorf("Bit %d expected on low and got high", bit)
		}
	}

	for _, bit := range []int{1, 2, 4} {
		if !bits[bit] {
			t.Errorf("Bit %d expected on high and got low", bit)
		}
	}
}

func TestToUint32(t *testing.T) {
	// Dec: 118
	// Bin: 1110110
	bits := FromUint32(118, 7)
	value := bits.ToUint32()

	if value != 118 {
		t.Errorf("Expecting %d and got %d", 118, value)
	}
}

func TestSlice(t *testing.T) {
	// Dec: 118
	// Bin: 1110110
	bits := FromUint32(118, 7)

	// Slice bit 4 to bit 1: 1011
	slice := bits.Slice(4, 1)
	value := slice.ToUint32()

	if value != 11 {
		t.Errorf("Expecting %d and got %d", 11, value)
	}
}

func TestConcat(t *testing.T) {
	// Bin: 1010 110
	// Dec: 10   6
	bits1 := FromUint32(10, 4)
	bits0 := FromUint32(6, 3)

	bits := Concat(bits1, bits0)
	value := bits.ToUint32()

	if value != 86 {
		t.Errorf("Expecting %d and got %d", 86, value)
	}
}
