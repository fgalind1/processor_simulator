package ieee754

import (
	"testing"
)

func TestZero32(t *testing.T) {
	packed := PackFloat754_32(0.0)
	unpacked := UnPackFloat754_32(0)

	if packed != 0 {
		t.Errorf("Expecting 0 and got %v", packed)
	}
	if unpacked != 0 {
		t.Errorf("Expecting 0 and got %v", unpacked)
	}
}

func TestPack32(t *testing.T) {
	// Input: 458.90393f
	packed := PackFloat754_32(458.90393)
	// Expected: 01000011111001010111001110110100 / 0x43E573B4 / 1139110836
	if packed != 1139110836 {
		t.Errorf("Expecting 1139110836 and got %v", packed)
	}
}

func TestUnPack32(t *testing.T) {
	// Input: 00110101101001111001010000111011 / 0x35A7943B / 900174907
	unpacked := UnPackFloat754_32(900174907)
	// Expected: 0.000001248561 / 1.248561E-6
	if unpacked != 900174907 {
		t.Errorf("Expecting 900174907 and got %v", unpacked)
	}
}
