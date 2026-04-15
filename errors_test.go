package mycobot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRobotError_Error(t *testing.T) {
	err := &RobotError{
		Op:    "SendAngles",
		Model: "MechArm270",
		Err:   ErrInvalidSpeed,
	}

	msg := err.Error()
	assert.Contains(t, msg, "MechArm270")
	assert.Contains(t, msg, "SendAngles")
	assert.Contains(t, msg, "speed")
}

func TestRobotError_Unwrap(t *testing.T) {
	err := &RobotError{
		Op:    "PowerOn",
		Model: "MechArm270",
		Err:   ErrConnectionTimeout,
	}

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, ErrConnectionTimeout, unwrapped)
}

func TestStandardErrors_Defined(t *testing.T) {
	// Just verify errors are defined
	assert.NotNil(t, ErrRobotClosed)
	assert.NotNil(t, ErrNotConnected)
	assert.NotNil(t, ErrInvalidSpeed)
	assert.NotNil(t, ErrNoGripper)
}
