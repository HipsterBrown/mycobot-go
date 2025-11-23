package protocol

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
