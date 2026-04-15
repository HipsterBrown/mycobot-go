package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolCodes_Defined(t *testing.T) {
	// Test critical command codes are defined
	assert.Equal(t, byte(0xFE), Header)
	assert.Equal(t, byte(0xFA), Footer)
	assert.Equal(t, byte(0x10), PowerOn)
	assert.Equal(t, byte(0x11), PowerOff)
	assert.Equal(t, byte(0x20), GetAngles)
	assert.Equal(t, byte(0x22), SendAngles)
}

func TestAtomIOCodes_MatchPymycobot(t *testing.T) {
	// Atom IO codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0x60), SetPinMode)
	assert.Equal(t, byte(0x61), SetDigitalOutput)
	assert.Equal(t, byte(0x62), GetDigitalInput)
	assert.Equal(t, byte(0x63), SetPWMMode)
	assert.Equal(t, byte(0x64), SetPWMOutput)
	assert.Equal(t, byte(0x6A), SetColor)
}

func TestGripperCodes_MatchPymycobot(t *testing.T) {
	// Gripper codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0x65), GetGripperValue)
	assert.Equal(t, byte(0x66), SetGripperState)
	assert.Equal(t, byte(0x67), SetGripperValue)
	assert.Equal(t, byte(0x68), SetGripperIni)
	assert.Equal(t, byte(0x69), IsGripperMoving)
}

func TestBasicIOCodes_MatchPymycobot(t *testing.T) {
	// Basic IO codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0xA0), SetBasicOutput)
	assert.Equal(t, byte(0xA1), GetBasicInput)
}
