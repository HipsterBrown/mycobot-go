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

	resp, n, err := Decode(data, false)
	require.NoError(t, err)

	assert.Equal(t, byte(0x20), resp.Code)
	assert.Equal(t, []byte{0x01, 0x02, 0x03}, resp.Data)
	assert.Equal(t, len(data), n)
}

func TestDecode_ValidResponseWithCRC(t *testing.T) {
	// Build valid CRC response
	packet := []byte{0xFE, 0xFE, 0x04, 0x12, 0xAA}
	crc := calculateCRC(packet)
	data := append(packet, crc)

	resp, n, err := Decode(data, true)
	require.NoError(t, err)

	assert.Equal(t, byte(0x12), resp.Code)
	assert.Equal(t, []byte{0xAA}, resp.Data)
	assert.Equal(t, len(data), n)
}

func TestDecode_ReturnsFrameSizeWithTrailingBytes(t *testing.T) {
	// Frame followed by the start of a second frame in the same buffer.
	frame := []byte{0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA}
	trailing := []byte{0xFE, 0xFE, 0x02}
	data := append(append([]byte{}, frame...), trailing...)

	resp, n, err := Decode(data, false)
	require.NoError(t, err)
	assert.Equal(t, len(frame), n)
	assert.Equal(t, byte(0x12), resp.Code)
}

func TestDecode_InvalidHeader(t *testing.T) {
	data := []byte{0xAA, 0xFE, 0x02, 0x10, 0xFA}

	_, _, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header")
}

func TestDecode_TooShort(t *testing.T) {
	data := []byte{0xFE, 0xFE}

	_, _, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestDecode_InvalidCRC(t *testing.T) {
	data := []byte{0xFE, 0xFE, 0x03, 0x10, 0xFF} // wrong CRC

	_, _, err := Decode(data, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CRC mismatch")
}

func TestDecode_InvalidFooter(t *testing.T) {
	data := []byte{0xFE, 0xFE, 0x02, 0x10, 0xAA} // wrong footer

	_, _, err := Decode(data, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid footer")
}

func TestEncodeInt16(t *testing.T) {
	tests := []struct {
		input    int
		expected []byte
	}{
		{0, []byte{0x00, 0x00}},
		{100, []byte{0x00, 0x64}},
		{-100, []byte{0xFF, 0x9C}},
		{4500, []byte{0x11, 0x94}},
		{-4500, []byte{0xEE, 0x6C}},
	}

	for _, tt := range tests {
		result := EncodeInt16(tt.input)
		assert.Equal(t, tt.expected, result, "EncodeInt16(%d)", tt.input)
	}
}

func TestEncodeAngles(t *testing.T) {
	angles := []float64{0, 45.5, -90.25, 120}

	data := EncodeAngles(angles)

	// Each angle encoded as int16 (angle * 100)
	// 0 -> 0, 45.5 -> 4550, -90.25 -> -9025, 120 -> 12000
	expected := []byte{
		0x00, 0x00, // 0
		0x11, 0xC6, // 4550
		0xDC, 0xBF, // -9025
		0x2E, 0xE0, // 12000
	}

	assert.Equal(t, expected, data)
}

func TestDecodeAngles(t *testing.T) {
	// Data represents: [0, 45.5, -90.25, 120]
	data := []byte{
		0x00, 0x00, // 0
		0x11, 0xC6, // 4550
		0xDC, 0xBF, // -9025
		0x2E, 0xE0, // 12000
	}

	angles, err := DecodeAngles(data)
	require.NoError(t, err)

	expected := []float64{0, 45.5, -90.25, 120}
	assert.Equal(t, len(expected), len(angles))

	for i, exp := range expected {
		assert.InDelta(t, exp, angles[i], 0.01, "angle %d", i)
	}
}

func TestDecodeAngles_InvalidLength(t *testing.T) {
	data := []byte{0x00} // odd length

	_, err := DecodeAngles(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid angle data length")
}

func TestEncodeCoords(t *testing.T) {
	// pymycobot encodes XYZ with _coord2int (value * 10)
	// and Rx/Ry/Rz with _angle2int (value * 100)
	data := EncodeCoords(100.5, -50.2, 200.0, 45.0, -30.5, 90.25)

	expected := []byte{
		// XYZ: value * 10, big-endian int16
		0x03, 0xED, // 100.5 * 10 = 1005
		0xFE, 0x0A, // -50.2 * 10 = -502
		0x07, 0xD0, // 200.0 * 10 = 2000
		// Rx/Ry/Rz: value * 100, big-endian int16
		0x11, 0x94, // 45.0 * 100 = 4500
		0xF4, 0x16, // -30.5 * 100 = -3050
		0x23, 0x41, // 90.25 * 100 = 9025
	}

	assert.Equal(t, expected, data)
}

func TestDecodeCoords(t *testing.T) {
	data := []byte{
		// XYZ: encoded with * 10
		0x03, 0xED, // 1005
		0xFE, 0x0A, // -502
		0x07, 0xD0, // 2000
		// Rx/Ry/Rz: encoded with * 100
		0x11, 0x94, // 4500
		0xF4, 0x16, // -3050
		0x23, 0x41, // 9025
	}

	x, y, z, rx, ry, rz, err := DecodeCoords(data)
	require.NoError(t, err)

	assert.InDelta(t, 100.5, x, 0.1)
	assert.InDelta(t, -50.2, y, 0.1)
	assert.InDelta(t, 200.0, z, 0.1)
	assert.InDelta(t, 45.0, rx, 0.01)
	assert.InDelta(t, -30.5, ry, 0.01)
	assert.InDelta(t, 90.25, rz, 0.01)
}

func TestDecodeCoords_InvalidLength(t *testing.T) {
	// Only 8 bytes instead of 12
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	_, _, _, _, _, _, err := DecodeCoords(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 12 bytes")
}
