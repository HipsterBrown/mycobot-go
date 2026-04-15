package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIO_Structure(t *testing.T) {
	io := &IO{}
	assert.NotNil(t, io)
}

func TestPinMode_Constants(t *testing.T) {
	assert.Equal(t, PinMode(0), PinInput)
	assert.Equal(t, PinMode(1), PinOutput)
	assert.Equal(t, PinMode(2), PinInputPullup)
}

func TestPinSignal_Constants(t *testing.T) {
	assert.Equal(t, PinSignal(0), SignalLow)
	assert.Equal(t, PinSignal(1), SignalHigh)
}

func TestIO_MethodsExist(t *testing.T) {
	io := &IO{}
	_ = io.SetPinMode
	_ = io.SetDigitalOutput
	_ = io.GetDigitalInput
	_ = io.SetPWMMode
	_ = io.SetPWMOutput
	_ = io.SetColor
}
