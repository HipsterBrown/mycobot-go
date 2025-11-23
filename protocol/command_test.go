package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_EncodeNoData(t *testing.T) {
	cmd := Command{
		Code:   PowerOn,
		Data:   nil,
		UseCRC: false,
	}

	data, err := cmd.Encode()
	require.NoError(t, err)

	// Expected: [FE FE 02 10 FA]
	// Header Header Length Code Footer
	expected := []byte{0xFE, 0xFE, 0x02, 0x10, 0xFA}
	assert.Equal(t, expected, data)
}

func TestCommand_EncodeWithData(t *testing.T) {
	cmd := Command{
		Code:   SendAngles,
		Data:   []byte{0x01, 0x02, 0x03},
		UseCRC: false,
	}

	data, err := cmd.Encode()
	require.NoError(t, err)

	// Expected: [FE FE 05 22 01 02 03 FA]
	// Length = data(3) + 2 = 5
	expected := []byte{0xFE, 0xFE, 0x05, 0x22, 0x01, 0x02, 0x03, 0xFA}
	assert.Equal(t, expected, data)
}

func TestCommand_EncodeWithCRC(t *testing.T) {
	cmd := Command{
		Code:   PowerOn,
		Data:   nil,
		UseCRC: true,
	}

	data, err := cmd.Encode()
	require.NoError(t, err)

	// Expected: [FE FE 03 10 CRC]
	// Length = 2 + 1(crc) = 3
	assert.Equal(t, byte(0xFE), data[0])
	assert.Equal(t, byte(0xFE), data[1])
	assert.Equal(t, byte(0x03), data[2])
	assert.Equal(t, byte(0x10), data[3])

	// CRC is XOR of all previous bytes
	expectedCRC := byte(0xFE) ^ byte(0xFE) ^ byte(0x03) ^ byte(0x10)
	assert.Equal(t, expectedCRC, data[4])
}
