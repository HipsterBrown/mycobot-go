package types

import "fmt"

// Coord represents a 3D coordinate with rotation
type Coord struct {
	X, Y, Z    float64 // Position in mm
	Rx, Ry, Rz float64 // Rotation in degrees
}

// ToSlice converts coordinate to slice for encoding
func (c Coord) ToSlice() []float64 {
	return []float64{c.X, c.Y, c.Z, c.Rx, c.Ry, c.Rz}
}

// NewCoordFromSlice creates Coord from slice
func NewCoordFromSlice(data []float64) (Coord, error) {
	if len(data) != 6 {
		return Coord{}, fmt.Errorf("expected 6 values, got %d", len(data))
	}
	return Coord{
		X:  data[0],
		Y:  data[1],
		Z:  data[2],
		Rx: data[3],
		Ry: data[4],
		Rz: data[5],
	}, nil
}
