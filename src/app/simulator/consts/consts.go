package consts

import "time"

const (
	PROGRAM_FINISHED = 1
	PROGRAM_RUNNING  = 0

	BITS_PER_BYTE  = 8
	BITS_PER_WORD  = 32
	BYTES_PER_WORD = BITS_PER_WORD / BITS_PER_BYTE

	ENDING_BYTE = 0x77

	STATUS_REGISTER = 0
	FLAG_PARITY     = 2
	FLAG_ZERO       = 6
	FLAG_SIGN       = 7
	FLAG_OVERFLOW   = 11

	FETCH_CYCLES   = 1
	DECODE_CYCLES  = 1
	EXECUTE_CYCLES = 1

	FETCH_EVENT   = "F"
	DECODE_EVENT  = "D"
	EXECUTE_EVENT = "E"

	ACTIVE   = true
	INACTIVE = false

	WAIT_PERIOD = 5 * time.Millisecond
	MENU_DELAY  = 10 * time.Millisecond
	STEP_PERIOD = 50 * time.Millisecond
)
