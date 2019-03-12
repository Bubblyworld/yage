package gb

import "errors"

type cpu interface{}

type gbRegisterType int

const (
	// 8-bit registers
	gbRegisterA gbRegisterType = 0 // accumulator register
	gbRegisterF gbRegisterType = 1 // flag register
	gbRegisterB gbRegisterType = 2
	gbRegisterC gbRegisterType = 3
	gbRegisterD gbRegisterType = 4
	gbRegisterE gbRegisterType = 5
	gbRegisterH gbRegisterType = 6
	gbRegisterL gbRegisterType = 7

	// 16-bit stack pointer register
	gbRegisterSP gbRegisterType = 8

	// 16-bit program counter register
	gbRegisterPC gbRegisterType = 9

	// 16-bit combined registers
	gbRegisterAF gbRegisterType = 10
	gbRegisterBC gbRegisterType = 11
	gbRegisterDE gbRegisterType = 12
	gbRegisterHL gbRegisterType = 13
)

var (
	gbErrInvalidRegisterEncoding = errors.New("gbCPU: invalid register encoding")
	gbErrUnknownRegisterEncoding = errors.New("gbCPU: unknown register encoding")
)

func (rt gbRegisterType) is8Bit() bool {
	return rt >= gbRegisterA && rt <= gbRegisterL
}

func (rt gbRegisterType) is16Bit() bool {
	return rt >= gbRegisterSP && rt <= gbRegisterHL
}

func (rt gbRegisterType) isCombined() bool {
	return rt >= gbRegisterAF && rt <= gbRegisterHL
}

func decodeRegisterType(t uint8) (gbRegisterType, error) {
	if t != (t & 0x7) {
		return 0, gbErrInvalidRegisterEncoding
	}

	switch t {
	case 0: // 0b000
		return gbRegisterB, nil

	case 1: // 0b001
		return gbRegisterC, nil

	case 2: // 0b010
		return gbRegisterD, nil

	case 3: // 0b011
		return gbRegisterE, nil

	case 4: // 0b100
		return gbRegisterH, nil

	case 5: // 0b101
		return gbRegisterL, nil

	case 6: // 0b110
		return gbRegisterHL, nil

	case 7: // 0b111
		return gbRegisterA, nil
	}

	panic(gbErrUnknownRegisterEncoding) // should never happen
	return 0, nil
}

const (
	gbFlagCarry     uint8 = 0x1 << 4
	gbFlagHalfCarry uint8 = 0x1 << 5
	gbFlagSubtract  uint8 = 0x1 << 6
	gbFlagZero      uint8 = 0x1 << 7

	gbOpcodeMaskHeader uint8 = 0xc0 // 0b11000000
	gbOpcodeMaskFirst  uint8 = 0x38 // 0b00111000
	gbOpcodeMaskSecond uint8 = 0x7  // 0b00000111

	gbOpcodeHeaderLD uint8 = 0x40 // 0b01000000
)

type gbCPU struct {
	reg8  [8]uint8    // semantically a map[gbRegisterType]uint8
	reg16 [2][2]uint8 // semantically a map[gbRegisterType][2]uint8
}

var (
	gbErrUnknownRegisterType = errors.New("gbCPU: unknown register type")
)

func newGBCPU() *gbCPU {
	return &gbCPU{}
}

func (c *gbCPU) readRegister(t gbRegisterType) [2]uint8 {
	if t.is8Bit() {
		return [2]uint8{c.reg8[t], 0}
	}

	if !t.isCombined() {
		return c.reg16[t-gbRegisterSP]
	}

	switch t {
	case gbRegisterAF:
		return [2]uint8{c.reg8[gbRegisterA], c.reg8[gbRegisterF]}

	case gbRegisterBC:
		return [2]uint8{c.reg8[gbRegisterB], c.reg8[gbRegisterC]}

	case gbRegisterDE:
		return [2]uint8{c.reg8[gbRegisterD], c.reg8[gbRegisterE]}

	case gbRegisterHL:
		return [2]uint8{c.reg8[gbRegisterH], c.reg8[gbRegisterL]}
	}

	panic(gbErrUnknownRegisterType) // should never get here
	return [2]uint8{0, 0}
}

func (c *gbCPU) pokeRegister(val [2]uint8, t gbRegisterType) {
	if t.is8Bit() {
		c.reg8[t] = val[0]
		return
	}

	if !t.isCombined() {
		c.reg16[t-gbRegisterSP] = val
		return
	}

	switch t {
	case gbRegisterAF:
		c.reg8[gbRegisterA] = val[0]
		c.reg8[gbRegisterF] = val[1]
		return

	case gbRegisterBC:
		c.reg8[gbRegisterB] = val[0]
		c.reg8[gbRegisterC] = val[1]
		return

	case gbRegisterDE:
		c.reg8[gbRegisterD] = val[0]
		c.reg8[gbRegisterE] = val[1]
		return

	case gbRegisterHL:
		c.reg8[gbRegisterH] = val[0]
		c.reg8[gbRegisterL] = val[1]
		return
	}

	panic(gbErrUnknownRegisterType) // should never get here
}

func (c *gbCPU) execute(ops []uint8) error {
	if len(ops) == 0 {
		return nil
	}

	opcode := ops[0]
	// data := ops[1:]
	switch opcode & gbOpcodeMaskHeader {
	case gbOpcodeHeaderLD:
		fR, err := decodeRegisterType((opcode & gbOpcodeMaskFirst) >> 3)
		if err != nil {
			panic(err) // should never happen
		}

		sR, err := decodeRegisterType(opcode & gbOpcodeMaskSecond)
		if err != nil {
			panic(err) // should never happen
		}

		// TODO(guy): Implement the 4 cases
		fR = sR
		sR = fR
	}

	return nil
}
