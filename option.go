package mycobot

import "time"

// Option configures a robot client.
type Option func(*base)

// WithBaudRate overrides the default baud rate for the port.
func WithBaudRate(baud int) Option {
	return func(b *base) { b.SetBaudRate(baud) }
}

// WithCRC enables CRC framing for firmware that requires it.
// Default is the 0xFA footer used by MechArm 270 / MyCobot 280.
func WithCRC() Option {
	return func(b *base) { b.SetUseCRC(true) }
}

// WithDefaultTimeout sets the fallback per-command read timeout used when
// the caller's context has no deadline. If both are absent, the transport
// falls back to 1 second.
func WithDefaultTimeout(d time.Duration) Option {
	return func(b *base) { b.SetDefaultTimeout(d) }
}
