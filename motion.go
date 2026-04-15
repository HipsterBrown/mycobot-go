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
func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction types.Direction, speed types.Speed) error {
	if err := direction.Validate(); err != nil {
		return err
	}
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
func (m *Motion) JogCoord(ctx context.Context, axis CoordAxis, direction types.Direction, speed types.Speed) error {
	if err := direction.Validate(); err != nil {
		return err
	}
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

// SendAngle moves a single joint to the specified angle
func (m *Motion) SendAngle(ctx context.Context, joint types.JointID, angle types.Angle, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(int(angle*100))...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngle,
		Data: data,
	}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// SendCoord moves a single coordinate axis to the specified value
func (m *Motion) SendCoord(ctx context.Context, axis CoordAxis, value float64, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	// XYZ axes (0-2) use * 10, rotation axes (3-5) use * 100
	var encoded int
	if axis <= AxisZ {
		encoded = int(value * 10)
	} else {
		encoded = int(value * 100)
	}

	data := []byte{byte(axis + 1)}
	data = append(data, protocol.EncodeInt16(encoded)...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoord,
		Data: data,
	}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// IsInPosition checks if the robot is at the target position.
// flag: 0 = check angles, 1 = check coordinates
func (m *Motion) IsInPosition(ctx context.Context, data []float64, flag int) (bool, error) {
	var encoded []byte
	if flag == 0 {
		// Angles: encode with * 100
		encoded = protocol.EncodeAngles(data)
	} else {
		// Coordinates: encode with split multiplier
		if len(data) != 6 {
			return false, nil
		}
		encoded = protocol.EncodeCoords(data[0], data[1], data[2], data[3], data[4], data[5])
	}
	encoded = append(encoded, byte(flag))

	cmd := protocol.Command{
		Code: protocol.IsInPosition,
		Data: encoded,
	}
	resp, err := m.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(resp) > 0 {
		return resp[0] == 1, nil
	}
	return false, nil
}
