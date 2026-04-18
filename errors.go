package mycobot

import "errors"

// Connection errors.
var (
	// ErrRobotClosed is returned when SendCommand is called after Close.
	ErrRobotClosed = errors.New("robot connection closed")
	// ErrNotConnected is returned when SendCommand is called before Open.
	ErrNotConnected = errors.New("robot not connected")
)
