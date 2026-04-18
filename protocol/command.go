package protocol

import (
	"encoding/binary"
	"fmt"
)

// Command represents a protocol command to send to the robot
type Command struct {
	Code     byte
	Data     []byte
	UseCRC   bool // when true, uses XOR checksum instead of 0xFA footer
	HasReply bool // when true, the caller expects a response frame from the firmware
}

// Encode converts the command to wire format
// Format: [Header Header Length Code Data... Footer/CRC]
func (c Command) Encode() ([]byte, error) {
	dataLen := len(c.Data)
	length := dataLen + 2 // code + length byte

	if c.UseCRC {
		length++ // +1 for CRC
	}

	// Build packet
	packet := make([]byte, 0, length+3) // headers + length + code + data + footer/crc
	packet = append(packet, Header, Header, byte(length), c.Code)

	if dataLen > 0 {
		packet = append(packet, c.Data...)
	}

	if c.UseCRC {
		crc := calculateCRC(packet)
		packet = append(packet, crc)
	} else {
		packet = append(packet, Footer)
	}

	return packet, nil
}

// calculateCRC computes XOR checksum for newer robot models
func calculateCRC(data []byte) byte {
	var crc byte = 0
	for _, b := range data {
		crc ^= b
	}
	return crc
}

// Response represents a decoded response from the robot
type Response struct {
	Code byte
	Data []byte
}

// Decode parses the next wire-format frame from data. It returns the decoded
// Response, the number of bytes the frame occupies (so callers can advance a
// streaming buffer), and an error if the frame is malformed or incomplete.
func Decode(data []byte, useCRC bool) (*Response, int, error) {
	if len(data) < 5 {
		return nil, 0, fmt.Errorf("response too short: %d bytes", len(data))
	}

	if data[0] != Header || data[1] != Header {
		return nil, 0, fmt.Errorf("invalid header: %#x %#x", data[0], data[1])
	}

	length := int(data[2])

	// Format: [Header Header Length Code Data... Footer/CRC]
	// Non-CRC length counts Code + Data + Footer;
	// CRC length additionally counts the CRC byte.
	var frameSize int
	if useCRC {
		frameSize = length + 2 // headers(2)
	} else {
		frameSize = length + 3 // headers(2) + length_byte(1)
	}

	if len(data) < frameSize {
		return nil, 0, fmt.Errorf("incomplete response: expected %d, got %d", frameSize, len(data))
	}

	frame := data[:frameSize]

	if useCRC {
		expectedCRC := calculateCRC(frame[:frameSize-1])
		actualCRC := frame[frameSize-1]
		if expectedCRC != actualCRC {
			return nil, 0, fmt.Errorf("CRC mismatch: expected %#x, got %#x", expectedCRC, actualCRC)
		}
	} else if frame[frameSize-1] != Footer {
		return nil, 0, fmt.Errorf("invalid footer: %#x", frame[frameSize-1])
	}

	var dataLen int
	if useCRC {
		dataLen = length - 3 // code + length indicator + CRC
	} else {
		dataLen = length - 2 // code + footer
	}
	var payload []byte
	if dataLen > 0 {
		payload = make([]byte, dataLen)
		copy(payload, frame[4:4+dataLen])
	}

	return &Response{Code: frame[3], Data: payload}, frameSize, nil
}

// EncodeInt16 encodes an integer as big-endian int16
func EncodeInt16(value int) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(value))
	return buf
}

// EncodeAngles encodes angles in degrees to wire format
// Each angle is converted to int16 (angle * 100) for precision
func EncodeAngles(angles []float64) []byte {
	data := make([]byte, 0, len(angles)*2)
	for _, angle := range angles {
		value := int16(angle * 100)
		data = append(data, EncodeInt16(int(value))...)
	}
	return data
}

// DecodeAngles decodes wire format back to angles in degrees
func DecodeAngles(data []byte) ([]float64, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("invalid angle data length: %d", len(data))
	}

	count := len(data) / 2
	angles := make([]float64, count)

	for i := 0; i < count; i++ {
		value := int16(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
		angles[i] = float64(value) / 100.0
	}

	return angles, nil
}

// EncodeCoords encodes coordinates (x, y, z, rx, ry, rz) to wire format.
// XYZ positions are encoded as int16 (value * 10) per pymycobot _coord2int.
// Rx/Ry/Rz rotations are encoded as int16 (value * 100) per pymycobot _angle2int.
func EncodeCoords(x, y, z, rx, ry, rz float64) []byte {
	data := make([]byte, 0, 12)
	// XYZ: multiply by 10
	for _, v := range []float64{x, y, z} {
		data = append(data, EncodeInt16(int(v*10))...)
	}
	// Rx/Ry/Rz: multiply by 100
	for _, v := range []float64{rx, ry, rz} {
		data = append(data, EncodeInt16(int(v*100))...)
	}
	return data
}

// DecodeCoords decodes wire format back to coordinates.
// XYZ are decoded with divisor 10, Rx/Ry/Rz with divisor 100.
func DecodeCoords(data []byte) (x, y, z, rx, ry, rz float64, err error) {
	if len(data) != 12 {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid coord data: expected 12 bytes, got %d", len(data))
	}

	// XYZ: divide by 10
	xyz := make([]float64, 3)
	for i := 0; i < 3; i++ {
		value := int16(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
		xyz[i] = float64(value) / 10.0
	}

	// Rx/Ry/Rz: divide by 100
	rot := make([]float64, 3)
	for i := 0; i < 3; i++ {
		value := int16(binary.BigEndian.Uint16(data[6+i*2 : 6+i*2+2]))
		rot[i] = float64(value) / 100.0
	}

	return xyz[0], xyz[1], xyz[2], rot[0], rot[1], rot[2], nil
}
