package gb

import (
	"errors"
)

type cpu interface {
	// load reads and decodes the memory poined to by the PC register into an
	// opcode. It must not perform any mutating operations, such as
	// incrementing the PC counter.
	load(ram) (*gbOpcode, error)

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
	gbRegisterUnknown gbRegisterType = 0

	// 8-bit registers
	gbRegisterA gbRegisterType = 1 // accumulator register
	gbRegisterF gbRegisterType = 2 // flag register
	gbRegisterB gbRegisterType = 3
	gbRegisterC gbRegisterType = 4
	gbRegisterD gbRegisterType = 5
	gbRegisterE gbRegisterType = 6
	gbRegisterH gbRegisterType = 7
	gbRegisterL gbRegisterType = 8

	// 16-bit stack pointer register
	gbRegisterSP gbRegisterType = 9

	// 16-bit program counter register
	gbRegisterPC gbRegisterType = 10

	// 16-bit combined registers
	gbRegisterAF gbRegisterType = 11
	gbRegisterBC gbRegisterType = 12
	gbRegisterDE gbRegisterType = 13
	gbRegisterHL gbRegisterType = 14
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

	case 7: // 0b111
		return gbRegisterA
	}

	return gbRegisterUnknown
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
	gbErrUnknownRegisterType    = errors.New("gbCPU: unknown register type")
	gbErrUnknownOpcode          = errors.New("gbCPU: unknown opcode")
	gbErrIncompatibleOpcodeSize = errors.New("gbCPU: incompatible opcode sizes returned by decode")
)

func newGBCPU() *gbCPU {
	return &gbCPU{}
}

func (c *gbCPU) readRegister(t gbRegisterType) uint16 {
	if t.is8Bit() {
		return uint16(c.reg8[t-1])
	}

	if t.is16Bit() && !t.isCombined() {
		return c.reg16[t-gbRegisterSP-1]
	}

	switch t {
	case gbRegisterAF:
		return uint16(c.readRegister(gbRegisterA)<<8) +
			uint16(c.readRegister(gbRegisterF))

	case gbRegisterBC:
		return uint16(c.readRegister(gbRegisterB)<<8) +
			uint16(c.readRegister(gbRegisterC))

	case gbRegisterDE:
		return uint16(c.readRegister(gbRegisterD)<<8) +
			uint16(c.readRegister(gbRegisterE))

	case gbRegisterHL:
		return uint16(c.readRegister(gbRegisterH)<<8) +
			uint16(c.readRegister(gbRegisterL))
	}

	panic(gbErrUnknownRegisterType) // should never get here
	return 0
}

func (c *gbCPU) pokeRegister(val uint16, t gbRegisterType) {
	if t.is8Bit() {
		c.reg8[t-1] = uint8(val & 0xFF)
		return
	}

	if t.is16Bit() && !t.isCombined() {
		c.reg16[t-gbRegisterSP-1] = val
		return
	}

	switch t {
	case gbRegisterAF:
		c.reg8[gbRegisterA-1] = uint8(val >> 8)
		c.reg8[gbRegisterF-1] = uint8(val & 0xFF)
		return

	case gbRegisterBC:
		c.reg8[gbRegisterB-1] = uint8(val >> 8)
		c.reg8[gbRegisterC-1] = uint8(val & 0xFF)
		return

	case gbRegisterDE:
		c.reg8[gbRegisterD-1] = uint8(val >> 8)
		c.reg8[gbRegisterE-1] = uint8(val & 0xFF)
		return

	case gbRegisterHL:
		c.reg8[gbRegisterH-1] = uint8(val >> 8)
		c.reg8[gbRegisterL-1] = uint8(val & 0xFF)
		return
	}

	panic(gbErrUnknownRegisterType) // should never get here
}

func (c *gbCPU) load(r ram) (*gbOpcode, error) {
	addr := uint32(c.readRegister(gbRegisterPC))
	op, err := r.read(addr)
	if err != nil {
		return nil, err
	}

	ops := []uint8{op}
	opcode, n, err := decode(ops)
	if err != nil && err != gbErrWrongOpcodeSize {
		return nil, err
	}
	if err == nil {
		return opcode, nil
	}
	if n < 0 {
		panic(gbErrIncompatibleOpcodeSize) // should never get here
	}

	// Opcode requires more data.
	opsn, err := readN(r, addr+1, uint32(n))
	if err != nil {
		return nil, err
	}

	opcode, n, err = decode(append(ops, opsn...))
	if n != 0 {
		panic(gbErrIncompatibleOpcodeSize)
	}

	return opcode, err
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

	// TODO(guy): Update PC register where necessary.
}

// runInstructionCycle performs a full fetch, decode and execute cycle.
func runInstructionCycle(c cpu, r ram) error {
	opcode, err := c.load(r)
	if err != nil {
		return err
	}

	return c.execute(r, opcode)
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

	vals, err := readN(r, addr, 1)
	if !only8Bit {
		vals, err = readN(r, addr, 2)
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
