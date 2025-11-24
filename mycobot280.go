package mycobot

import (
	"context"

	"github.com/yourusername/mycobot-go/internal/robot"
	"github.com/yourusername/mycobot-go/protocol"
	"github.com/yourusername/mycobot-go/types"
)

// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base   *robot.Base
	config ModelConfig
}

// NewMyCobot280 creates a new MyCobot280 instance
func NewMyCobot280(port string, opts ...Option) *MyCobot280 {
	config := getModelConfig(types.ModelMyCobot280)

	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	// Apply options
	for _, opt := range opts {
		opt(base)
	}

	return &MyCobot280{
		base:   base,
		config: config,
	}
}

// Open establishes connection to the robot
func (m *MyCobot280) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MyCobot280) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MyCobot280) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MyCobot280) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}
