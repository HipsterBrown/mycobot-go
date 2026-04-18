package mycobot

import (
	"context"
	"fmt"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// IO provides Atom IO operations (end-effector head, 0x60 range)
type IO struct {
	robot *robot.Base
}

// SetPinMode configures a pin as input, output, or input with pullup
func (io *IO) SetPinMode(ctx context.Context, pin int, mode types.PinMode) error {
	cmd := protocol.Command{
		Code:     protocol.SetPinMode,
		Data:     []byte{byte(pin), byte(mode)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetDigitalOutput sets a digital pin high or low
func (io *IO) SetDigitalOutput(ctx context.Context, pin int, signal types.PinSignal) error {
	cmd := protocol.Command{
		Code:     protocol.SetDigitalOutput,
		Data:     []byte{byte(pin), byte(signal)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// GetDigitalInput reads the state of a digital input pin
func (io *IO) GetDigitalInput(ctx context.Context, pin int) (types.PinSignal, error) {
	cmd := protocol.Command{
		Code:     protocol.GetDigitalInput,
		Data:     []byte{byte(pin)},
		HasReply: true,
	}
	data, err := io.robot.SendCommand(ctx, cmd)
	if err != nil {
		return types.SignalLow, err
	}
	if len(data) > 0 {
		return types.PinSignal(data[0]), nil
	}
	return types.SignalLow, nil
}

// SetPWMMode configures a pin for PWM output
func (io *IO) SetPWMMode(ctx context.Context, pin int) error {
	cmd := protocol.Command{
		Code:     protocol.SetPWMMode,
		Data:     []byte{byte(pin)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetPWMOutput sets PWM frequency and duty cycle on a channel
func (io *IO) SetPWMOutput(ctx context.Context, channel int, freq int, dutyCycle int) error {
	if dutyCycle < 0 || dutyCycle > 256 {
		return fmt.Errorf("duty cycle %d out of range [0, 256]", dutyCycle)
	}

	cmd := protocol.Command{
		Code:     protocol.SetPWMOutput,
		Data:     []byte{byte(channel), byte(freq >> 8), byte(freq & 0xFF), byte(dutyCycle)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetColor sets the RGB color of the Atom LED
func (io *IO) SetColor(ctx context.Context, r, g, b byte) error {
	cmd := protocol.Command{
		Code:     protocol.SetColor,
		Data:     []byte{r, g, b},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}
