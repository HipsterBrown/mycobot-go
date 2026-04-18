package mycobot

import "github.com/hipsterbrown/mycobot-go/internal/robot"

// Option configures a robot
type Option func(*robot.Base)

// WithBaudRate sets custom baud rate
func WithBaudRate(baud int) Option {
	return func(b *robot.Base) {
		b.SetBaudRate(baud)
	}
}

// WithCRC enables CRC mode for firmware that requires it.
// By default, the standard 0xFA footer is used (matching pymycobot defaults).
func WithCRC() Option {
	return func(b *robot.Base) {
		b.SetUseCRC(true)
	}
}
