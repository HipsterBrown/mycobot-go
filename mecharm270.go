package mycobot

import "github.com/hipsterbrown/mycobot-go/types"

// MechArm270 is an Elephant Robotics MechArm 270 robot arm.
type MechArm270 struct {
	*base
}

// NewMechArm270 creates a client for a MechArm 270 connected at the given
// serial port. Default configuration is 115200 baud, CRC off; override with
// WithBaudRate, WithCRC, WithDefaultTimeout.
func NewMechArm270(port string, opts ...Option) *MechArm270 {
	cfg := getModelConfig(types.ModelMechArm270)
	b := newBase(port, cfg)
	for _, opt := range opts {
		opt(b)
	}
	return &MechArm270{base: b}
}
