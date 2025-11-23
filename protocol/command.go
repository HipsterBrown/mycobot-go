package protocol

import "fmt"

// Command represents a protocol command to send to the robot
type Command struct {
	Code   byte
	Data   []byte
	UseCRC bool // true for newer models (MyCobot280, 320), false for older
}

// NewCommand creates a command without data
func NewCommand(code byte) Command {
	return Command{
		Code:   code,
		Data:   nil,
		UseCRC: false,
	}
}

// NewCommandWithData creates a command with data payload
func NewCommandWithData(code byte, data []byte) Command {
	return Command{
		Code:   code,
		Data:   data,
		UseCRC: false,
	}
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

// Decode parses wire format response into Response struct
func Decode(data []byte, useCRC bool) (*Response, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("response too short: %d bytes", len(data))
	}

	// Validate header
	if data[0] != Header || data[1] != Header {
		return nil, fmt.Errorf("invalid header: %#x %#x", data[0], data[1])
	}

	length := int(data[2])
	code := data[3]

	// Calculate expected total length
	// Format: [Header Header Length Code Data... Footer/CRC]
	// For non-CRC: length = Code + Data + Footer
	// For CRC: length = Code + Data + (length marker) + CRC (hence one extra byte in formula)
	var expectedLen int
	if useCRC {
		expectedLen = length + 2 // headers(2)
	} else {
		expectedLen = length + 3 // headers(2) + length_byte(1)
	}

	if len(data) < expectedLen {
		return nil, fmt.Errorf("incomplete response: expected %d, got %d", expectedLen, len(data))
	}

	// Validate CRC or footer
	if useCRC {
		expectedCRC := calculateCRC(data[:len(data)-1])
		actualCRC := data[len(data)-1]
		if expectedCRC != actualCRC {
			return nil, fmt.Errorf("CRC mismatch: expected %#x, got %#x", expectedCRC, actualCRC)
		}
	} else {
		footer := data[len(data)-1]
		if footer != Footer {
			return nil, fmt.Errorf("invalid footer: %#x", footer)
		}
	}

	// Extract data payload
	var payload []byte
	var dataLen int
	if useCRC {
		dataLen = length - 3 // subtract code, length indicator, and CRC
	} else {
		dataLen = length - 2 // subtract code and footer
	}
	if dataLen > 0 {
		payload = make([]byte, dataLen)
		copy(payload, data[4:4+dataLen])
	}

	return &Response{
		Code: code,
		Data: payload,
	}, nil
}
