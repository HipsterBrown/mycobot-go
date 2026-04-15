package types

import "fmt"

// CoordMode specifies the interpolation mode for coordinate movement
type CoordMode int

const (
	// CoordModeAngular uses angular interpolation (mode 0)
	CoordModeAngular CoordMode = 0
	// CoordModeLinear uses linear interpolation (mode 1)
	CoordModeLinear CoordMode = 1
)

// Validate checks if the mode is valid
func (m CoordMode) Validate() error {
	if m != CoordModeAngular && m != CoordModeLinear {
		return fmt.Errorf("invalid coord mode %d: must be 0 (angular) or 1 (linear)", m)
	}
	return nil
}
