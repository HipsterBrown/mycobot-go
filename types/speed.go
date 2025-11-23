package types

import "fmt"

// Speed represents robot movement speed (0-100)
type Speed int

const (
	SpeedMin    Speed = 0
	SpeedSlow   Speed = 25
	SpeedMedium Speed = 50
	SpeedFast   Speed = 75
	SpeedMax    Speed = 100
)

// Validate checks if speed is in valid range
func (s Speed) Validate() error {
	if s < 0 || s > 100 {
		return fmt.Errorf("speed %d out of range [0, 100]", s)
	}
	return nil
}
