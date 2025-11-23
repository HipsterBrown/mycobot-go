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
