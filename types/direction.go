package types

import "fmt"

// Direction specifies the direction for JOG movement
type Direction int

const (
	// DirNegative moves in the negative direction
	DirNegative Direction = 0
	// DirPositive moves in the positive direction
	DirPositive Direction = 1
)

// Validate checks if the direction is valid
func (d Direction) Validate() error {
	if d != DirNegative && d != DirPositive {
		return fmt.Errorf("invalid direction %d: must be 0 (negative) or 1 (positive)", d)
	}
	return nil
}
