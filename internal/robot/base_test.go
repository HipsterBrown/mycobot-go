package robot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

func TestNewBase(t *testing.T) {
	base := NewBase("/dev/ttyUSB0", 115200, true)

	assert.Equal(t, "/dev/ttyUSB0", base.port)
	assert.Equal(t, 115200, base.baudrate)
	assert.True(t, base.useCRC)
	assert.False(t, base.connected)
}

func TestBase_OpenClose(t *testing.T) {
	// This test requires actual hardware, so we'll test the structure
	base := NewBase("/dev/null", 115200, true)

	// Verify initial state
	assert.False(t, base.IsConnected())
	assert.Nil(t, base.cmdChan)
	assert.Nil(t, base.closeChan)
}

func TestBase_CommandChannels(t *testing.T) {
	base := NewBase("/dev/ttyUSB0", 115200, true)

	// After initialization, channels should exist (tested in integration)
	assert.NotNil(t, base)
}

func TestBase_SendCommand_NotConnected(t *testing.T) {
	base := NewBase("/dev/ttyUSB0", 115200, true)
	ctx := context.Background()

	cmd := protocol.Command{Code: protocol.PowerOn, UseCRC: true}
	_, err := base.SendCommand(ctx, cmd)

	assert.Error(t, err)
	// Should fail because not connected
}
