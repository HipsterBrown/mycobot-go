package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// CoordAxis represents a coordinate axis for single-axis movement
type CoordAxis int

const (
	AxisX CoordAxis = iota
	AxisY
	AxisZ
	AxisRx
	AxisRy
	AxisRz
)

// Motion provides motion control operations
type Motion struct {
	robot *robot.Base
}

// JogAngle performs incremental joint movement
// direction: 0 = negative, 1 = positive
func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(joint), byte(direction), byte(speed)}
	cmd := protocol.Command{
		Code: protocol.JogAngle,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// JogCoord performs incremental coordinate movement
func (m *Motion) JogCoord(ctx context.Context, axis CoordAxis, direction int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(axis), byte(direction), byte(speed)}
	cmd := protocol.Command{
		Code: protocol.JogCoord,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// JogStop stops JOG movement
func (m *Motion) JogStop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.JogStop}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// Pause pauses current movement
func (m *Motion) Pause(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Pause}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// Resume resumes paused movement
func (m *Motion) Resume(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Resume}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// Stop stops all movement
func (m *Motion) Stop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Stop}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// IsPaused returns true if robot is paused
func (m *Motion) IsPaused(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPaused}
	data, err := m.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
