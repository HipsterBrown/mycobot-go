package mycobot

import (
	"time"
	"github.com/hipsterbrown/mycobot-go/internal/robot"
)

// Option configures a robot
type Option func(*robot.Base)

// WithBaudRate sets custom baud rate
func WithBaudRate(baud int) Option {
	return func(b *robot.Base) {
		// Baud rate is set during construction
		// This will be applied in NewMyCobot280
	}
}

// WithTimeout sets default command timeout
func WithTimeout(timeout time.Duration) Option {
	return func(b *robot.Base) {
		// Will be implemented when we add timeout support
	}
}
