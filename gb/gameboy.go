package gb

type Gameboy struct {
	cpu cpu
	ppu ppu
	ram ram
}

func NewGameboy() *Gameboy {
	return &Gameboy{
		cpu: newGBCPU(),
		ppu: newGBPPU(),
		ram: newGBRAM(),
	}
}

// Step moves the gameboy state forward by a single quartz-cycle.
func (g *Gameboy) Step() {
}
