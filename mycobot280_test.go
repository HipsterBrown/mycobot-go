package mycobot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/types"
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

func TestMyCobot280_PowerOn(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// This will fail without hardware, but tests the method exists
	err := robot.PowerOn(ctx)
	// We expect an error because we're not connected
	assert.Error(t, err)
}

func TestMyCobot280_PowerOff(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	err := robot.PowerOff(ctx)
	assert.Error(t, err)
}

func TestMyCobot280_SendAngles_Validation(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// Test with invalid angle count
	angles := types.Angles{0, 45, 90} // Only 3 angles, need 6
	err := robot.SendAngles(ctx, angles, types.SpeedMedium)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")
}

func TestMyCobot280_SendAngles_OutOfRange(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// Angle out of range for MyCobot280
	angles := types.Angles{0, 0, 0, 0, 0, 200} // 200 > 175 for joint 6
	err := robot.SendAngles(ctx, angles, types.SpeedMedium)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}
