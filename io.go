package mycobot

import (
	"context"
	"fmt"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

// PinMode configures a pin's behavior
type PinMode int

const (
	PinInput       PinMode = 0
	PinOutput      PinMode = 1
	PinInputPullup PinMode = 2
)

// PinSignal represents a digital pin state
type PinSignal int

const (
	SignalLow  PinSignal = 0
	SignalHigh PinSignal = 1
)

// IO provides Atom IO operations (end-effector head, 0x60 range)
type IO struct {
	robot *robot.Base
}

// SetPinMode configures a pin as input, output, or input with pullup
func (io *IO) SetPinMode(ctx context.Context, pin int, mode PinMode) error {
	cmd := protocol.Command{
		Code:     protocol.SetPinMode,
		Data:     []byte{byte(pin), byte(mode)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetDigitalOutput sets a digital pin high or low
func (io *IO) SetDigitalOutput(ctx context.Context, pin int, signal PinSignal) error {
	cmd := protocol.Command{
		Code:     protocol.SetDigitalOutput,
		Data:     []byte{byte(pin), byte(signal)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// GetDigitalInput reads the state of a digital input pin
func (io *IO) GetDigitalInput(ctx context.Context, pin int) (PinSignal, error) {
	cmd := protocol.Command{
		Code:     protocol.GetDigitalInput,
		Data:     []byte{byte(pin)},
		HasReply: true,
	}
	data, err := io.robot.SendCommand(ctx, cmd)
	if err != nil {
		return SignalLow, err
	}
	if len(data) > 0 {
		return PinSignal(data[0]), nil
	}
	return SignalLow, nil
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
