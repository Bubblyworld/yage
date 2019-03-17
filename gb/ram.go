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

func (r *gbRAM) read(addr, n uint32) ([]uint8, error) {
	if addr+n >= gbMaxAddress {
		return nil, gbErrOutOfBounds
	}

	return r.mem[addr : addr+n], nil
}
