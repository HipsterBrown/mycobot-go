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

// PowerOn powers on all servos
func (m *MyCobot280) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MyCobot280) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MyCobot280) IsPowerOn(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPowerOn}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// SendAngles sends joint angles to the robot
func (m *MyCobot280) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	// Validate angles
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}

	// Validate speed
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode angles
	angleData := protocol.EncodeAngles(angles.ToFloat64())

	// Append speed
	data := append(angleData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (m *MyCobot280) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	// Convert to types.Angles
	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}

	return result, nil
}

// SendCoords sends coordinate position to the robot
func (m *MyCobot280) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error {
	// Validate speed
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode coordinates
	coordData := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)

	// Append speed
	data := append(coordData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MyCobot280) GetCoords(ctx context.Context) (types.Coord, error) {
	cmd := protocol.Command{Code: protocol.GetCoords}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return types.Coord{}, err
	}

	x, y, z, rx, ry, rz, err := protocol.DecodeCoords(data)
	if err != nil {
		return types.Coord{}, err
	}

	return types.Coord{
		X:  x,
		Y:  y,
		Z:  z,
		Rx: rx,
		Ry: ry,
		Rz: rz,
	}, nil
}

// IsMoving returns true if robot is currently moving
func (m *MyCobot280) IsMoving(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsMoving}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
