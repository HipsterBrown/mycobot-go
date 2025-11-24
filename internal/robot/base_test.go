package robot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBase(t *testing.T) {
	base := NewBase("/dev/ttyUSB0", 115200, true)

	assert.Equal(t, "/dev/ttyUSB0", base.port)
	assert.Equal(t, 115200, base.baudrate)
	assert.True(t, base.useCRC)
	assert.False(t, base.connected)
}
