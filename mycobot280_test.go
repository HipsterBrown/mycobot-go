package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/mycobot-go/types"
)

func TestNewMyCobot280(t *testing.T) {
	robot := NewMyCobot280("/dev/ttyUSB0")

	assert.NotNil(t, robot)
	assert.Equal(t, types.ModelMyCobot280, robot.config.Model)
	assert.Equal(t, 6, robot.config.JointCount)
	assert.NotNil(t, robot.base)
}

func TestNewMyCobot280_WithOptions(t *testing.T) {
	robot := NewMyCobot280("/dev/ttyUSB0",
		WithBaudRate(1000000),
	)

	assert.NotNil(t, robot)
	// Baud rate will be tested in integration tests
}
