package gb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// prepareForOpcodes provides blank *gbCPU and *gbRAM instances that are set-up
// with the program counter register pointing to memory containing the given
// opcode binaries.
func prepareForOpcodes(t *testing.T, opcodes []uint8) (*gbCPU, *gbRAM) {
	c := newGBCPU()
	r := newGBRAM()

	// Write opcodes to a memory location and point (PC) to it.
	assert.NoError(t, pokeN(r, 0x100, opcodes))
	c.pokeRegister(0x100, gbRegisterPC)

	return c, r
}

// Test8BitLD_RR tests the 8-bit [LD R,R'] opcodes.
func Test8BitLD_R_R(t *testing.T) {
	opcodeHeader := uint8(1)
	opcodeRegs := []uint8{0, 1, 2, 3, 4, 5, 7}

	// Random values for the test registers/memory.
	const (
		v1 uint16 = 0x24
		v2 uint16 = 0x42
	)

	testFn := func(r1, r2 uint8) func(*testing.T) {
		return func(t *testing.T) {
			t1 := decodeRegisterType(r1)
			t2 := decodeRegisterType(r2)
			opcode := (opcodeHeader << 6) + (r1 << 3) + r2
			c, r := prepareForOpcodes(t, []uint8{opcode})

			// Write something to the source and dest registers for test.
			c.pokeRegister(v1, t1)
			c.pokeRegister(v2, t2)

			// Run a full instruction cycle on the CPU.
			assert.NoError(t, runInstructionCycle(c, r))
			assert.Equal(t, v2, c.readRegister(t1))
			assert.Equal(t, v2, c.readRegister(t2))
		}
	}

	for _, r1 := range opcodeRegs {
		for _, r2 := range opcodeRegs {
			name := fmt.Sprintf("01 %03b %03b", r1, r2)
			t.Run(name, testFn(r1, r2))
		}
	}
}

// Test8BitLD_HL_R tests the 8-bit [LD (HL),R] opcodes.
func Test8BitLD_HL_R(t *testing.T) {
	opcodeHeader := uint8(1)
	opcodeHLReg := uint8(6)
	opcodeRegs := []uint8{0, 1, 2, 3, 4, 5, 7}

	// Random values for the test registers/memory.
	const (
		v1   uint16 = 0x24
		v2   uint8  = 0x42
		addr uint32 = 0x204
	)

	testFn := func(r uint8) func(*testing.T) {
		return func(t *testing.T) {
			rt := decodeRegisterType(r)
			opcode := (opcodeHeader << 6) + (opcodeHLReg << 3) + r
			c, r := prepareForOpcodes(t, []uint8{opcode})

			// Write something to the source register.
			c.pokeRegister(v1, rt)

			// Write something to memory and set (HL) to its address.
			assert.NoError(t, r.poke(addr, v2))
			c.pokeRegister(uint16(addr), gbRegisterHL)

			// We expect the value in memory afterwards to be equal to R, but
			// if it's the H or L register we need to be careful because we
			// clobbered it.
			expectedMem := uint8(v1)
			if rt == gbRegisterH {
				expectedMem = uint8(c.readRegister(gbRegisterH))
			}
			if rt == gbRegisterL {
				expectedMem = uint8(c.readRegister(gbRegisterL))
			}

			// Run a full instruction cycle on the CPU.
			assert.NoError(t, runInstructionCycle(c, r))
			mem, err := r.read(addr)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, expectedMem, mem)
		}
	}

	for _, r := range opcodeRegs {
		name := fmt.Sprintf("01 110 %03b", r)
		t.Run(name, testFn(r))
	}
}

// Test8BitLD_R_HL tests the 8-bit [LD R,(HL)] opcodes.
func Test8BitLD_R_HL(t *testing.T) {
	opcodeHeader := uint8(1)
	opcodeHLReg := uint8(6)
	opcodeRegs := []uint8{0, 1, 2, 3, 4, 5, 7}

	// Random values for the test registers/memory.
	const (
		v1   uint16 = 0x24
		v2   uint8  = 0x42
		addr uint32 = 0x204
	)

	testFn := func(r uint8) func(*testing.T) {
		return func(t *testing.T) {
			rt := decodeRegisterType(r)
			opcode := (opcodeHeader << 6) + (r << 3) + opcodeHLReg
			c, r := prepareForOpcodes(t, []uint8{opcode})

			// Write something to the dest register.
			c.pokeRegister(v1, rt)

			// Write something to memory and set (HL) to its address.
			assert.NoError(t, r.poke(addr, v2))
			c.pokeRegister(uint16(addr), gbRegisterHL)

			// Run a full instruction cycle on the CPU.
			assert.NoError(t, runInstructionCycle(c, r))
			assert.Equal(t, uint16(v2), c.readRegister(rt))
		}
	}

	for _, r := range opcodeRegs {
		name := fmt.Sprintf("01 110 %03b", r)
		t.Run(name, testFn(r))
	}
}
