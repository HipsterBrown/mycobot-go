package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// Compile-time check that MechArm270 implements Robot
var _ Robot = (*MechArm270)(nil)

// MechArm270 represents a MechArm 270 robot
type MechArm270 struct {
	Motion *Motion
	IO     *IO
	Servo  *Servo

	base   *robot.Base
	config ModelConfig
}

// NewMechArm270 creates a new MechArm270 instance
func NewMechArm270(port string, opts ...Option) *MechArm270 {
	config := getModelConfig(types.ModelMechArm270)
	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	for _, opt := range opts {
		opt(base)
	}

	return &MechArm270{
		Motion: &Motion{robot: base},
		IO:     &IO{robot: base},
		Servo:  &Servo{robot: base},
		base:   base,
		config: config,
	}
}

// Open establishes connection to the robot
func (m *MechArm270) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MechArm270) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MechArm270) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MechArm270) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}

// PowerOn powers on all servos
func (m *MechArm270) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MechArm270) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MechArm270) IsPowerOn(ctx context.Context) (bool, error) {
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
func (m *MechArm270) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}
	if err := speed.Validate(); err != nil {
		return err
	}

	data := protocol.EncodeAngles(angles.ToFloat64())
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (m *MechArm270) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}
	return result, nil
}

// SendCoords sends coordinate position to the robot
func (m *MechArm270) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error {
	if err := speed.Validate(); err != nil {
		return err
	}
	if err := mode.Validate(); err != nil {
		return err
	}

	data := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data = append(data, byte(speed), byte(mode))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MechArm270) GetCoords(ctx context.Context) (types.Coord, error) {
	cmd := protocol.Command{Code: protocol.GetCoords}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return types.Coord{}, err
	}

	x, y, z, rx, ry, rz, err := protocol.DecodeCoords(data)
	if err != nil {
		return types.Coord{}, err
	}

	return types.Coord{X: x, Y: y, Z: z, Rx: rx, Ry: ry, Rz: rz}, nil
}

// IsMoving returns true if robot is currently moving
func (m *MechArm270) IsMoving(ctx context.Context) (bool, error) {
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
