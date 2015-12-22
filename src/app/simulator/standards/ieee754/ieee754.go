package ieee754

const (
	TOTAL_BITS_32       = 32
	EXPONENT_BITS_32    = 8
	SIGNIFICAND_BITS_32 = 23
)

func PackFloat754_32(fValue float32) uint32 {
	var fnorm float32
	var shift, sign, exp, significand uint32

	// Special case
	if fValue == 0.0 {
		return 0
	}

	// Check sign and start with normalization
	if fValue < 0 {
		sign = 1
		fnorm = -fValue
	} else {
		sign = 0
		fnorm = fValue
	}

	// Get the normalized form
	shift = 0
	for fnorm >= 2.0 {
		fnorm /= 2.0
		shift++
	}
	for fnorm < 1.0 {
		fnorm *= 2.0
		shift--
	}
	fnorm = fnorm - 1.0

	// Calculate the binary form (non-float) of the significand data
	significand = uint32(fnorm * float32((1<<SIGNIFICAND_BITS_32)+0.5))

	// Get the biased exponent
	exp = shift + ((1 << (EXPONENT_BITS_32 - 1)) - 1)

	return (sign << (TOTAL_BITS_32 - 1)) + (exp << (TOTAL_BITS_32 - EXPONENT_BITS_32 - 1)) + significand
}

func UnPackFloat754_32(value uint32) float32 {
	var result float32
	var shift, bias uint32

	// Special case
	if value == 0 {
		return 0.0
	}

	// Get the significand
	result = float32(value & ((1 << SIGNIFICAND_BITS_32) - 1))
	result /= (1 << SIGNIFICAND_BITS_32)
	result += 1.0

	// Get the exponent
	bias = (1 << (EXPONENT_BITS_32 - 1)) - 1
	shift = ((value >> SIGNIFICAND_BITS_32) & ((1 << EXPONENT_BITS_32) - 1)) - bias
	for shift > 0 {
		result *= 2.0
		shift--
	}
	for shift < 0 {
		result /= 2.0
		shift++
	}

	// Ge the sign
	if ((value >> (TOTAL_BITS_32 - 1)) & 1) > 0 {
		result *= -1.0
	} else {
		result *= 1.0
	}

	return result
}
