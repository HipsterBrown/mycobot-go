package mycobot

import (
	"context"
	"testing"

	"github.com/hipsterbrown/mycobot-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMechArm270(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0")

	assert.NotNil(t, arm)
	assert.Equal(t, types.ModelMechArm270, arm.config.Model)
	assert.Equal(t, 6, arm.config.JointCount)
	assert.NotNil(t, arm.base)
}

func TestNewMechArm270_WithOptions(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0",
		WithBaudRate(1000000),
	)
	assert.NotNil(t, arm)
}

func TestNewMechArm270_WithCRC(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0", WithCRC())
	assert.NotNil(t, arm)
}

func TestMechArm270_PowerOn_NotConnected(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()
	err := arm.PowerOn(ctx)
	assert.Error(t, err)
}

func TestMechArm270_SendAngles_Validation(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()

	// Wrong angle count
	angles := types.Angles{0, 45, 90}
	err := arm.SendAngles(ctx, angles, types.SpeedMedium)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")
}

func TestMechArm270_SendAngles_OutOfRange(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()

	// 200 > 175 for joint 6
	angles := types.Angles{0, 0, 0, 0, 0, 200}
	err := arm.SendAngles(ctx, angles, types.SpeedMedium)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

