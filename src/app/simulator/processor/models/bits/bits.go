package bits

type Bits []bool

func FromUint32(value uint32, size uint8) Bits {
	bits := make([]bool, size)
	i := uint8(0)
	for value >= 1 && i < size {
		bits[i] = (value & 0x1) > 0
		value >>= 1
		i++
	}
	return bits
}

func Concat(parts ...Bits) Bits {
	bits := []bool{}
	for i, _ := range parts {
		for _, value := range parts[len(parts)-i-1] {
			bits = append(bits, value)
		}
	}
	return bits
}

func (this Bits) Slice(end uint8, start uint8) Bits {
	return this[start : end+1]
}

func (this Bits) ToUint32() uint32 {
	value := uint32(0)
	for i, set := range this {
		if set {
			value += (1 << uint8(i))
		}
	}
	return value
}

func (this Bits) ToString() string {
	str := ""
	for i, _ := range this {
		value := this[len(this)-i-1]
		if value {
			str += "1"
		} else {
			str += "0"
		}
	}
	return str
}
