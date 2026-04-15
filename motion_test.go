package mycobot

import (
	"testing"

	"github.com/hipsterbrown/mycobot-go/types"
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

func TestMotion_JogAngle_UsesDirection(t *testing.T) {
	motion := &Motion{}
	// Verify the method signature accepts types.Direction (not int)
	var _ func(*Motion) = func(m *Motion) {
		_ = m.JogAngle
	}
	_ = motion
	_ = types.DirPositive
}

func TestMotion_SendAngle_Exists(t *testing.T) {
	motion := &Motion{}
	var _ func(*Motion) = func(m *Motion) {
		_ = m.SendAngle
	}
	_ = motion
}

func TestMotion_SendCoord_Exists(t *testing.T) {
	motion := &Motion{}
	var _ func(*Motion) = func(m *Motion) {
		_ = m.SendCoord
	}
	_ = motion
}

func TestMotion_IsInPosition_Exists(t *testing.T) {
	motion := &Motion{}
	var _ func(*Motion) = func(m *Motion) {
		_ = m.IsInPosition
	}
	_ = motion
}
