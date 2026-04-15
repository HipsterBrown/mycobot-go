package mycobot

import (
	"errors"
	"fmt"

	internalerrors "github.com/hipsterbrown/mycobot-go/internal/errors"
)

// Standard errors
var (
	// Connection errors (re-exported from internal/errors)
	ErrRobotClosed       = internalerrors.ErrRobotClosed
	ErrNotConnected      = internalerrors.ErrNotConnected
	ErrConnectionTimeout = errors.New("connection timeout")

	// Command errors
	ErrInvalidCommand  = errors.New("invalid command")
	ErrCommandTimeout  = errors.New("command timeout")
	ErrInvalidResponse = errors.New("invalid response from robot")

	// Validation errors
	ErrInvalidJoint      = errors.New("invalid joint ID")
	ErrInvalidSpeed      = errors.New("speed out of range")
	ErrInvalidAngle      = errors.New("angle out of range")
	ErrInvalidCoordinate = errors.New("coordinate out of range")

	// Gripper errors
	ErrNoGripper           = errors.New("no gripper attached")
	ErrGripperNotSupported = errors.New("gripper operation not supported")

	// State errors
	ErrNotPowered    = errors.New("robot not powered on")
	ErrEmergencyStop = errors.New("emergency stop active")
	ErrServoError    = errors.New("servo error detected")
)

// RobotError wraps errors with operation and model context
type RobotError struct {
	Op    string // Operation that failed (e.g., "SendAngles")
	Model string // Robot model (e.g., "MechArm270")
	Err   error  // Underlying error
}

func (e *RobotError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Model, e.Op, e.Err)
}

func (e *RobotError) Unwrap() error {
	return e.Err
}
