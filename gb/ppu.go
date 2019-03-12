package gb

type ppu interface{}

type gbPPU struct{}

func newGBPPU() *gbPPU {
	return &gbPPU{}
}
