package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordAxis_Constants(t *testing.T) {
	assert.Equal(t, CoordAxis(0), AxisX)
	assert.Equal(t, CoordAxis(1), AxisY)
	assert.Equal(t, CoordAxis(2), AxisZ)
	assert.Equal(t, CoordAxis(3), AxisRx)
	assert.Equal(t, CoordAxis(4), AxisRy)
	assert.Equal(t, CoordAxis(5), AxisRz)
}

func TestMotion_Structure(t *testing.T) {
	motion := &Motion{}
	assert.NotNil(t, motion)
}
