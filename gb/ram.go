package gb

import "errors"

type ram interface {
	poke(uint32, uint8) error
	read(uint32, uint32) ([]uint8, error)
}

const (
	gbByteMask   = 0xFF    // 0b11111111
	gbMaxAddress = 0x10000 // 64 Kb
)

var (
	gbErrOutOfBounds = errors.New("gbRAM: address isn't within 64Kb memory bounds")
)

type gbRAM struct {
	mem [gbMaxAddress]uint8
}

func newGBRAM() *gbRAM {
	return &gbRAM{}
}

func (r *gbRAM) poke(addr uint32, val uint8) error {
	if addr >= gbMaxAddress {
		return gbErrOutOfBounds
	}

	r.mem[addr] = val
	return nil
}

// TODO(guy): read should read a single byte, and there should be a utility
// function for reading multiple bytes (for consistency)
func (r *gbRAM) read(addr, n uint32) ([]uint8, error) {
	if addr+n >= gbMaxAddress {
		return nil, gbErrOutOfBounds
	}

	return r.mem[addr : addr+n], nil
}

// write is a utility function for writing multiple bytes to the given memory
// location.
func write(r ram, addr uint32, vals []uint8) error {
	for i, val := range vals {
		if err := r.poke(addr+uint32(i), val); err != nil {
			return err
		}
	}

	return nil
}
