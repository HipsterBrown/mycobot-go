package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoord_Creation(t *testing.T) {
	coord := Coord{
		X:  100.5,
		Y:  -50.2,
		Z:  200.0,
		Rx: 45.0,
		Ry: -30.5,
		Rz: 90.0,
	}

	assert.Equal(t, 100.5, coord.X)
	assert.Equal(t, -50.2, coord.Y)
	assert.Equal(t, 200.0, coord.Z)
	assert.Equal(t, 45.0, coord.Rx)
	assert.Equal(t, -30.5, coord.Ry)
	assert.Equal(t, 90.0, coord.Rz)
}

func TestCoord_ToSlice(t *testing.T) {
	coord := Coord{X: 1, Y: 2, Z: 3, Rx: 4, Ry: 5, Rz: 6}

	slice := coord.ToSlice()

	expected := []float64{1, 2, 3, 4, 5, 6}
	assert.Equal(t, expected, slice)
}
