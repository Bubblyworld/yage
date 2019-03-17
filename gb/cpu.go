package gb

import "errors"

type cpu interface {
	// load reads and decodes the memory poined to by the PC register into an
	// opcode. It must not perform any mutating operations, such as
	// incrementing the PC counter.
	load() (*gbOpcode, error)

	// execute performs the given opcode, updating memory, registers and
	// peripherals as needed.
	execute(ram, *gbOpcode) error

	// readRegister returns the value in the given register. If the register is
	// 8-bit, the least-significant bits hold the actual value of the register.
	readRegister(gbRegisterType) uint16

	// pokeRegister assigns the given value to the given register. If the register
	// is 8-bit, the least-significant bits of the value are assigned to it.
	pokeRegister(uint16, gbRegisterType)
}

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

func decodeRegisterType(t uint8) gbRegisterType {
	t = t & 0x7 // use only 3 least-significant bits

	switch t {
	case 0: // 0b000
		return gbRegisterB

	case 1: // 0b001
		return gbRegisterC

	case 2: // 0b010
		return gbRegisterD

	case 3: // 0b011
		return gbRegisterE

	case 4: // 0b100
		return gbRegisterH

	case 5: // 0b101
		return gbRegisterL

	case 6: // 0b110
		return gbRegisterHL

	case 7: // 0b111
		return gbRegisterA
	}

	panic(gbErrUnknownRegisterEncoding) // should never happen
	return 0
}

const (
	gbFlagCarry     uint8 = 0x1 << 4
	gbFlagHalfCarry uint8 = 0x1 << 5
	gbFlagSubtract  uint8 = 0x1 << 6
	gbFlagZero      uint8 = 0x1 << 7
)

type gbCPU struct {
	reg8  [8]uint8  // semantically a map[gbRegisterType]uint8
	reg16 [2]uint16 // semantically a map[gbRegisterType]uint16
}

var (
	gbErrUnknownRegisterType = errors.New("gbCPU: unknown register type")
	gbErrUnknownOpcode       = errors.New("gbCPU: unknown opcode")
)

func newGBCPU() *gbCPU {
	return &gbCPU{}
}

func (c *gbCPU) readRegister(t gbRegisterType) uint16 {
	if t.is8Bit() {
		return uint16(c.reg8[t])
	}

	if !t.isCombined() {
		return c.reg16[t-gbRegisterSP]
	}

	switch t {
	case gbRegisterAF:
		return (uint16(c.reg8[gbRegisterA]) << 8) + uint16(c.reg8[gbRegisterF])

	case gbRegisterBC:
		return (uint16(c.reg8[gbRegisterB]) << 8) + uint16(c.reg8[gbRegisterC])

	case gbRegisterDE:
		return (uint16(c.reg8[gbRegisterD]) << 8) + uint16(c.reg8[gbRegisterE])

	case gbRegisterHL:
		return (uint16(c.reg8[gbRegisterH]) << 8) + uint16(c.reg8[gbRegisterL])
	}

	panic(gbErrUnknownRegisterType) // should never get here
	return 0
}

func (c *gbCPU) pokeRegister(val uint16, t gbRegisterType) {
	if t.is8Bit() {
		c.reg8[t] = uint8(val & 0xFF)
		return
	}

	if !t.isCombined() {
		c.reg16[t-gbRegisterSP] = val
		return
	}

	switch t {
	case gbRegisterAF:
		c.reg8[gbRegisterA] = uint8(val >> 8)
		c.reg8[gbRegisterF] = uint8(val & 0xFF)
		return

	case gbRegisterBC:
		c.reg8[gbRegisterB] = uint8(val >> 8)
		c.reg8[gbRegisterC] = uint8(val & 0xFF)
		return

	case gbRegisterDE:
		c.reg8[gbRegisterD] = uint8(val >> 8)
		c.reg8[gbRegisterE] = uint8(val & 0xFF)
		return

	case gbRegisterHL:
		c.reg8[gbRegisterH] = uint8(val >> 8)
		c.reg8[gbRegisterL] = uint8(val & 0xFF)
		return
	}

	panic(gbErrUnknownRegisterType) // should never get here
}

func (c *gbCPU) load() (*gbOpcode, error) {
	// TODO(guy): implement
	return nil, nil
}

func (c *gbCPU) execute(r ram, op *gbOpcode) error {
	switch op.tipe {
	case gbOpcodeLDRRp:
		to := decodeRegisterType(op.first)
		from := decodeRegisterType(op.second)
		pokeRegisterIntoRegister(c, from, to)
		return nil

	case gbOpcodeLDRHl:
		to := decodeRegisterType(op.first)
		addr := uint32(c.readRegister(gbRegisterHL))
		return pokeRAMIntoRegister(c, r, to, addr, true)

	case gbOpcodeLDHlR:
		from := decodeRegisterType(op.second)
		addr := uint32(c.readRegister(gbRegisterHL))
		return pokeRegisterIntoRAM(c, r, from, addr, true)

	default:
		return gbErrUnknownOpcode
	}
}

func pokeRegisterIntoRAM(c cpu, r ram, t gbRegisterType,
	addr uint32, only8Bit bool) error {

	val := c.readRegister(t)
	first := uint8(val & 0xFF)
	second := uint8(val >> 8)

	err := r.poke(addr, first)
	if err != nil {
		return err
	}

	if !only8Bit {
		err = r.poke(addr+1, second)
	}

	return err
}

func pokeRAMIntoRegister(c cpu, r ram, t gbRegisterType,
	addr uint32, only8Bit bool) error {

	vals, err := r.read(addr, 1)
	if !only8Bit {
		vals, err = r.read(addr, 2)
	}
	if err != nil {
		return err
	}

	val := uint16(vals[0])
	if !only8Bit {
		// TODO(guy): Check endianness here against spec
		val = uint16(vals[0]<<8) + uint16(vals[1])
	}
	gbErrUnknownOpcode = errors.New("gbCPU: unknown opcode")

	c.pokeRegister(val, t)
	return nil
}

func pokeRegisterIntoRegister(c cpu, from, to gbRegisterType) {
	// TODO(guy): Check endianness here.
	c.pokeRegister(c.readRegister(from), to)
}
