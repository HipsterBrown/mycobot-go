package errors

import "errors"

// Connection errors used by internal/robot package
var (
	ErrRobotClosed  = errors.New("robot connection closed")
	ErrNotConnected = errors.New("robot not connected")
)
