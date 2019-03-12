package gb

import "errors"

type ram interface{}

const (
	gbByteMask   = 0xFF    // 0b11111111
	gbMaxAddress = 0x10000 // 64 Kb
)

var (
	gbErrUnaligned   = errors.New("gbRAM: address isn't aligned to 8-bit boundary")
	gbErrOutOfBounds = errors.New("gbRAM: address isn't within 64Kb memory bounds")
)

type gbRAM struct {
	mem [gbMaxAddress >> 8]uint8
}

func newGBRAM() *gbRAM {
	return &gbRAM{}
}

func (r *gbRAM) poke(val uint8, addr uint32) error {
	if addr&gbByteMask != 0 {
		return gbErrUnaligned
	}

	if addr >= gbMaxAddress {
		return gbErrOutOfBounds
	}

	r.mem[addr>>8] = val
	return nil
}
