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

func TestDecode_ValidResponse(t *testing.T) {
	// Simulate response: [FE FE 05 20 01 02 03 FA]
	data := []byte{0xFE, 0xFE, 0x05, 0x20, 0x01, 0x02, 0x03, 0xFA}

	resp, err := Decode(data, false)
	require.NoError(t, err)

	assert.Equal(t, byte(0x20), resp.Code)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, resp.Data)
}

func TestDecode_ValidResponseWithCRC(t *testing.T) {
	// Build valid CRC response
	packet := []byte{0xFE, 0xFE, 0x04, 0x12, 0xAA}
	crc := calculateCRC(packet)
	data := append(packet, crc)

	resp, err := Decode(data, true)
	require.NoError(t, err)

	assert.Equal(t, byte(0x12), resp.Code)
	assert.Equal(t, []byte{0xAA}, resp.Data)
}

func TestDecode_InvalidHeader(t *testing.T) {
	data := []byte{0xAA, 0xFE, 0x02, 0x10, 0xFA}

	_, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header")
}

func TestDecode_TooShort(t *testing.T) {
	data := []byte{0xFE, 0xFE}

	_, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecode_InvalidCRC(t *testing.T) {
	data := []byte{0xFE, 0xFE, 0x03, 0x10, 0xFF} // wrong CRC

	_, err := Decode(data, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CRC mismatch")
}

func TestDecode_InvalidFooter(t *testing.T) {
	data := []byte{0xFE, 0xFE, 0x02, 0x10, 0xAA} // wrong footer

	_, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid footer")
}
