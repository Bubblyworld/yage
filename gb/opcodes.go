package gb

import "errors"

type gbOpcodeType int

const (
	gbOpcodeHeader00 uint8 = 0x00
	gbOpcodeHeader01 uint8 = 0x01
	gbOpcodeHeader10 uint8 = 0x10
	gbOpcodeHeader11 uint8 = 0x11

	gbOpcodePart000 uint8 = 0x0
	gbOpcodePart001 uint8 = 0x1
	gbOpcodePart010 uint8 = 0x2
	gbOpcodePart011 uint8 = 0x3
	gbOpcodePart100 uint8 = 0x4
	gbOpcodePart101 uint8 = 0x5
	gbOpcodePart110 uint8 = 0x6
	gbOpcodePart111 uint8 = 0x7

	gbOpcodeMaskHeader uint8 = 0xc0 // 0b11000000
	gbOpcodeMaskFirst  uint8 = 0x38 // 0b00111000
	gbOpcodeMaskSecond uint8 = 0x7  // 0b00000111

	// 8-bit IO instructions
	gbOpcodeLDRRp  gbOpcodeType = 1  // [ LD R, R'   ]
	gbOpcodeLDRHl  gbOpcodeType = 2  // [ LD R, (HL) ]
	gbOpcodeLDHlR  gbOpcodeType = 3  // [ LD (HL), R ]
	gbOpcodeLDRN   gbOpcodeType = 4  // [ LD R, n ]
	gbOpcodeLDHlN  gbOpcodeType = 5  // [ LD (HL), n ]
	gbOpcodeLDABc  gbOpcodeType = 6  // [ LD A, (BC) ]
	gbOpcodeLDBcA  gbOpcodeType = 7  // [ LD (BC), A ]
	gbOpcodeLDADe  gbOpcodeType = 8  // [ LD A, (DE) ]
	gbOpcodeLDDeA  gbOpcodeType = 9  // [ LD (DE), A ]
	gbOpcodeLDAC   gbOpcodeType = 10 // [ LD A, (0xFF00+C) ]
	gbOpcodeLDCA   gbOpcodeType = 11 // [ LD (0xFF00+C), A ]
	gbOpcodeLDAN   gbOpcodeType = 12 // [ LD A, (0xFF00+n) ]
	gbOpcodeLDNA   gbOpcodeType = 13 // [ LD (0xFF00+n), A ]
	gbOpcodeLDANn  gbOpcodeType = 14 // [ LD A, (nn) ]
	gbOpcodeLDNnA  gbOpcodeType = 15 // [ LD (nn), A ]
	gbOpcodeLDAHlI gbOpcodeType = 16 // [ LD A, (HLI) ]
	gbOpcodeLDHlIA gbOpcodeType = 17 // [ LD (HLI), A ]
	gbOpcodeLDAHlD gbOpcodeType = 18 // [ LD A, (HLD) ]
	gbOpcodeLDHlDA gbOpcodeType = 19 // [ LD (HLD), A ]
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

		// The 01 opcodes are all a single byte.
		if len(o.data) > 0 {
			return nil, -len(o.data), gbErrWrongOpcodeSize
		}

		if o.first == gbOpcodePart110 {
			o.tipe = gbOpcodeLDHlR
			o.cycles = 2
			return &o, 0, nil
		}

		if o.second == gbOpcodePart110 {
			o.tipe = gbOpcodeLDRHl
			o.cycles = 2
			return &o, 0, nil
		}

		if fR != gbRegisterUnknown && sR != gbRegisterUnknown {
			o.tipe = gbOpcodeLDRRp
			o.cycles = 1
			return &o, 0, nil
		}
	}

	return nil, 0, gbErrInvalidOpcode
}
