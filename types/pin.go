package types

// PinMode configures a pin's behavior.
type PinMode int

const (
	PinInput       PinMode = 0
	PinOutput      PinMode = 1
	PinInputPullup PinMode = 2
)

// PinSignal represents a digital pin state.
type PinSignal int

const (
	SignalLow  PinSignal = 0
	SignalHigh PinSignal = 1
)
