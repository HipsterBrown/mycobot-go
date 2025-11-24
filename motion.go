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
