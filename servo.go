package mycobot

import (
	"context"
	"encoding/binary"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// Servo provides servo control operations
type Servo struct {
	robot *robot.Base
}

// ReleaseServo powers off a single servo, allowing free movement
func (s *Servo) ReleaseServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code:     protocol.ReleaseServo,
		Data:     []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// FocusServo powers on a single servo
func (s *Servo) FocusServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code:     protocol.FocusServo,
		Data:     []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// IsServoEnabled checks if a specific servo is powered on
func (s *Servo) IsServoEnabled(ctx context.Context, joint types.JointID) (bool, error) {
	cmd := protocol.Command{
		Code:     protocol.IsServoEnable,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// GetEncoder reads the encoder value for a single joint (0-4096)
func (s *Servo) GetEncoder(ctx context.Context, joint types.JointID) (int, error) {
	cmd := protocol.Command{
		Code:     protocol.GetEncoder,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		return int(binary.BigEndian.Uint16(data[:2])), nil
	}
	return 0, nil
}

// SetEncoder sets the encoder value for a single joint (0-4096)
func (s *Servo) SetEncoder(ctx context.Context, joint types.JointID, value int) error {
	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(value)...)

	cmd := protocol.Command{
		Code:     protocol.SetEncoder,
		Data:     data,
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetEncoders reads all encoder values
func (s *Servo) GetEncoders(ctx context.Context) ([]int, error) {
	cmd := protocol.Command{Code: protocol.GetEncoders, HasReply: true}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	count := len(data) / 2
	encoders := make([]int, count)
	for i := 0; i < count; i++ {
		encoders[i] = int(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
	}
	return encoders, nil
}

// SetEncoders sets all encoder values simultaneously
func (s *Servo) SetEncoders(ctx context.Context, encoders []int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	var data []byte
	for _, enc := range encoders {
		data = append(data, protocol.EncodeInt16(enc)...)
	}
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code:     protocol.SetEncoders,
		Data:     data,
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetServoData reads a servo parameter
func (s *Servo) GetServoData(ctx context.Context, joint types.JointID, dataID byte) (int, error) {
	cmd := protocol.Command{
		Code:     protocol.GetServoData,
		Data:     []byte{byte(joint), dataID},
		HasReply: true,
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 1 {
		return int(data[0]), nil
	}
	return 0, nil
}

// SetServoData writes a servo parameter
func (s *Servo) SetServoData(ctx context.Context, joint types.JointID, dataID byte, value int) error {
	cmd := protocol.Command{
		Code:     protocol.SetServoData,
		Data:     []byte{byte(joint), dataID, byte(value)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// SetServoCalibration sets the current position as angle zero for a joint.
// This writes to non-volatile memory on the servo.
func (s *Servo) SetServoCalibration(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code:     protocol.SetServoCalibration,
		Data:     []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetJointMin reads the minimum angle limit for a joint from firmware
func (s *Servo) GetJointMin(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code:     protocol.GetJointMinAngle,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		value := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(value) / 100.0, nil
	}
	return 0, nil
}

// GetJointMax reads the maximum angle limit for a joint from firmware
func (s *Servo) GetJointMax(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code:     protocol.GetJointMaxAngle,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		value := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(value) / 100.0, nil
	}
	return 0, nil
}
