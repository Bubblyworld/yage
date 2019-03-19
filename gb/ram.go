package gb

import "errors"

type ram interface {
	poke(uint32, uint8) error
	read(uint32) (uint8, error)
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

func (r *gbRAM) read(addr uint32) (uint8, error) {
	if addr >= gbMaxAddress {
		return 0, gbErrOutOfBounds
	}

	return r.mem[addr], nil
}

// readN is a utility function for reading multiple bytes from a given address.
func readN(r ram, addr, n uint32) ([]uint8, error) {
	var res []uint8
	for i := uint32(0); i < n; i++ {
		b, err := r.read(addr + i)
		if err != nil {
			return nil, err
		}
		res = append(res, b)
	}

	return res, nil
}

// pokeN is a utility function for writing multiple bytes to the given address.
func pokeN(r ram, addr uint32, vals []uint8) error {
	for i, val := range vals {
		if err := r.poke(addr+uint32(i), val); err != nil {
			return err
		}
	}

	return nil
}
