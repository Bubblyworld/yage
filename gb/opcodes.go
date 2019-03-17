package gb

import "errors"

type gbOpcodeType int

const (
	gbOpcodeHeader00 uint8 = 0x00 // 0b00000000
	gbOpcodeHeader01 uint8 = 0x01 // 0b00000001
	gbOpcodeHeader10 uint8 = 0x10 // 0b00000010
	gbOpcodeHeader11 uint8 = 0x11 // 0b00000011

	gbOpcodeMaskHeader uint8 = 0xc0 // 0b11000000
	gbOpcodeMaskFirst  uint8 = 0x38 // 0b00111000
	gbOpcodeMaskSecond uint8 = 0x7  // 0b00000111

	// 8-bit IO instructions
	gbOpcodeLDRRp gbOpcodeType = 1 // R <- Rp
	gbOpcodeLDRHl gbOpcodeType = 2 // R <- (HL)
	gbOpcodeLDHlR gbOpcodeType = 3 // (HL) <- R
)

var (
	gbErrInvalidOpcode   = errors.New("gbOpcode: data isn't a valid opcode")
	gbErrWrongOpcodeSize = errors.New("gbOpcode: wrong amount of data given for opcode")
)

type gbOpcode struct {
	header uint8   // bits 7,6 of opcode
	first  uint8   // bits 5,4,3 of opcode
	second uint8   // bits 2,1,0 of opcode
	data   []uint8 // remaining bits of opcode (if any)

	tipe   gbOpcodeType
	cycles int // cycles measures in units of 4 quartz cycles
}

// decode attempts to decode the given data into an opcode. Some opcodes are
// larger in size than others - if there isn't enough data to fully decode one,
// the returned integer provides the number of missing bytes. Similarly, if
// there is too much data, the returned integer is negative in the number of
// additional bytes provided.
// TODO(guy): Handle this with an explicit error type instead.
func decode(ops []uint8) (*gbOpcode, int, error) {
	if len(ops) == 0 {
		return nil, 0, gbErrInvalidOpcode
	}

	o := gbOpcode{
		header: (ops[0] & gbOpcodeMaskHeader) >> 6,
		first:  (ops[0] & gbOpcodeMaskFirst) >> 3,
		second: (ops[0] & gbOpcodeMaskSecond),
		data:   ops[1:],
	}

	switch o.header {
	case gbOpcodeHeader01:
		fR := decodeRegisterType(o.first)
		sR := decodeRegisterType(o.second)

		if fR == gbRegisterHL && sR == gbRegisterHL {
			return nil, 0, gbErrInvalidOpcode
		}

		if len(o.data) > 0 {
			return nil, -len(o.data), gbErrWrongOpcodeSize
		}

		if fR == gbRegisterHL {
			o.tipe = gbOpcodeLDHlR
			o.cycles = 2
			return &o, 0, nil
		}

		if sR == gbRegisterHL {
			o.tipe = gbOpcodeLDRHl
			o.cycles = 2
			return &o, 0, nil
		}

		o.tipe = gbOpcodeLDRRp
		o.cycles = 1
		return &o, 0, nil
	}

	return nil, 0, gbErrInvalidOpcode
}
