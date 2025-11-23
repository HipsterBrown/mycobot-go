# MyCobot Go Port Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Port pymycobot Python library to Go, creating a thread-safe, idiomatic Go package for controlling Elephant Robotics myCobot series robotic arms over serial communication.

**Architecture:** Channel-based concurrency with single goroutine owning serial connection per robot. Strongly-typed domain primitives with validation. Interface-based gripper system. Exposed protocol layer for advanced users. Hybrid API with core commands on robot, specialized features grouped into subsystems.

**Tech Stack:** Go 1.21+, go.bug.st/serial for cross-platform serial communication, testify for test assertions.

---

## Phase 1: Core Foundation

### Task 1: Project Setup and Go Module Initialization

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `.gitignore`
- Create: `README.md`

**Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/yourusername/mycobot-go
```

Expected: Creates `go.mod` with module name

**Step 2: Add serial port dependency**

Run:
```bash
go get go.bug.st/serial@latest
```

Expected: Adds serial dependency to go.mod

**Step 3: Add testify for testing**

Run:
```bash
go get github.com/stretchr/testify@latest
```

Expected: Adds testify to go.mod

**Step 4: Create .gitignore**

Create `.gitignore`:
```
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary
*.test

# Output
*.out

# Go workspace file
go.work

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
```

**Step 5: Create basic README**

Create `README.md`:
```markdown
# mycobot-go

Go library for controlling Elephant Robotics myCobot series robotic arms.

## Supported Models

- MyCobot 280
- MyCobot 320
- MechArm 270
- MyPalletizer 260

## Features

- Thread-safe concurrent access
- Context-based timeout/cancellation
- Strongly-typed API
- Interface-based gripper support
- Exposed protocol layer for advanced usage

## Installation

```bash
go get github.com/yourusername/mycobot-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/yourusername/mycobot-go"
    "github.com/yourusername/mycobot-go/types"
)

func main() {
    robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
    ctx := context.Background()

    if err := robot.Open(ctx); err != nil {
        log.Fatal(err)
    }
    defer robot.Close()

    robot.PowerOn(ctx)
    robot.SendAngles(ctx, types.Angles{0, 0, 0, 0, 0, 0}, types.SpeedMedium)
}
```

## Documentation

See [docs/plans/2025-11-23-mycobot-go-port-design.md](docs/plans/2025-11-23-mycobot-go-port-design.md) for architecture details.

## License

MIT
```

**Step 6: Commit project setup**

Run:
```bash
git add go.mod go.sum .gitignore README.md
git commit -m "feat: initialize Go module and project structure"
```

Expected: Clean commit with project foundation

---

### Task 2: Protocol Package - Command Codes

**Files:**
- Create: `protocol/codes.go`
- Create: `protocol/codes_test.go`

**Step 1: Write test for protocol codes existence**

Create `protocol/codes_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./protocol -v
```

Expected: FAIL - "undefined: Header"

**Step 3: Implement protocol codes**

Create `protocol/codes.go`:
```go
package protocol

// Protocol frame markers
const (
	Header byte = 0xFE
	Footer byte = 0xFA
)

// System status commands
const (
	SoftwareVersion      byte = 0x02
	GetRobotID          byte = 0x03
	SetRobotID          byte = 0x04
	PowerOn             byte = 0x10
	PowerOff            byte = 0x11
	IsPowerOn           byte = 0x12
	ReleaseAllServos    byte = 0x13
	IsControllerConnected byte = 0x14
	ReadNextError       byte = 0x15
	SetFreshMode        byte = 0x16
	GetFreshMode        byte = 0x17
	SetFreeMode         byte = 0x1A
	IsFreeMode          byte = 0x1B
)

// MDI mode commands
const (
	GetAngles           byte = 0x20
	SendAngle           byte = 0x21
	SendAngles          byte = 0x22
	GetCoords           byte = 0x23
	SendCoord           byte = 0x24
	SendCoords          byte = 0x25
	Pause               byte = 0x26
	IsPaused            byte = 0x27
	Resume              byte = 0x28
	Stop                byte = 0x29
	IsInPosition        byte = 0x2A
	IsMoving            byte = 0x2B
)

// JOG mode commands
const (
	JogAngle            byte = 0x30
	JogAbsolute         byte = 0x31
	JogCoord            byte = 0x32
	JogIncrement        byte = 0x33
	JogStop             byte = 0x34
)

// Encoder commands
const (
	SetEncoder          byte = 0x3A
	GetEncoder          byte = 0x3B
	SetEncoders         byte = 0x3C
	GetEncoders         byte = 0x3D
)

// Running status and settings
const (
	GetSpeed            byte = 0x40
	SetSpeed            byte = 0x41
	GetJointMinAngle    byte = 0x4A
	GetJointMaxAngle    byte = 0x4B
	SetJointMin         byte = 0x4C
	SetJointMax         byte = 0x4D
)

// Servo control
const (
	IsServoEnable       byte = 0x50
	IsAllServoEnable    byte = 0x51
	SetServoData        byte = 0x52
	GetServoData        byte = 0x53
	SetServoCalibration byte = 0x54
	ReleaseServo        byte = 0x56
	FocusServo          byte = 0x57
)

// Atom IO
const (
	SetColor            byte = 0x6A
	SetDigitalOutput    byte = 0xA0
	GetDigitalInput     byte = 0xA1
	SetPWMOutput        byte = 0xA2
	GetGripperValue     byte = 0x67
	SetGripperState     byte = 0x68
	SetGripperValue     byte = 0x66
	SetGripperIni       byte = 0x69
	IsGripperMoving     byte = 0x6B
)

// Basic IO
const (
	SetBasicOutput      byte = 0xA0
	GetBasicInput       byte = 0xA1
)

// Gripper extended
const (
	InitGripper              byte = 0x38
	SetGripperProtectCurrent byte = 0x39
	GetGripperProtectCurrent byte = 0x37
	SetHTSGripperTorque      byte = 0x35
	GetHTSGripperTorque      byte = 0x36
)
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./protocol -v
```

Expected: PASS

**Step 5: Commit protocol codes**

Run:
```bash
git add protocol/
git commit -m "feat(protocol): add protocol command codes"
```

---

### Task 3: Protocol Package - Command Encoding

**Files:**
- Create: `protocol/command.go`
- Modify: `protocol/codes_test.go`
- Create: `protocol/command_test.go`

**Step 1: Write test for command encoding**

Create `protocol/command_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./protocol -v -run TestCommand_Encode
```

Expected: FAIL - "undefined: Command"

**Step 3: Implement Command struct and Encode**

Create `protocol/command.go`:
```go
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
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./protocol -v -run TestCommand_Encode
```

Expected: PASS (all 3 tests)

**Step 5: Commit command encoding**

Run:
```bash
git add protocol/
git commit -m "feat(protocol): add command encoding with CRC support"
```

---

### Task 4: Protocol Package - Response Decoding

**Files:**
- Modify: `protocol/command.go`
- Modify: `protocol/command_test.go`

**Step 1: Write test for response decoding**

Add to `protocol/command_test.go`:
```go
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./protocol -v -run TestDecode
```

Expected: FAIL - "undefined: Decode"

**Step 3: Implement Decode function**

Add to `protocol/command.go`:
```go
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
	expectedLen := length + 3 // headers + length byte
	if useCRC {
		expectedLen++ // +1 for CRC byte at end
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
	dataLen := length - 2 // subtract code and length bytes
	if dataLen > 0 {
		payload = make([]byte, dataLen)
		copy(payload, data[4:4+dataLen])
	}

	return &Response{
		Code: code,
		Data: payload,
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./protocol -v -run TestDecode
```

Expected: PASS (all 6 tests)

**Step 5: Commit response decoding**

Run:
```bash
git add protocol/
git commit -m "feat(protocol): add response decoding with validation"
```

---

### Task 5: Protocol Package - Data Encoding Helpers

**Files:**
- Modify: `protocol/command.go`
- Modify: `protocol/command_test.go`

**Step 1: Write tests for data encoding helpers**

Add to `protocol/command_test.go`:
```go
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
		0xDC, 0xCF, // -9025
		0x2E, 0xE0, // 12000
	}

	assert.Equal(t, expected, data)
}

func TestDecodeAngles(t *testing.T) {
	// Data represents: [0, 45.5, -90.25, 120]
	data := []byte{
		0x00, 0x00, // 0
		0x11, 0xC6, // 4550
		0xDC, 0xCF, // -9025
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./protocol -v -run "TestEncodeInt16|TestEncodeAngles|TestDecodeAngles"
```

Expected: FAIL - "undefined: EncodeInt16"

**Step 3: Implement encoding helper functions**

Add to `protocol/command.go`:
```go
import (
	"encoding/binary"
	"fmt"
)

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

// EncodeCoords encodes coordinates (x, y, z, rx, ry, rz) to wire format
func EncodeCoords(x, y, z, rx, ry, rz float64) []byte {
	coords := []float64{x, y, z, rx, ry, rz}
	return EncodeAngles(coords)
}

// DecodeCoords decodes wire format back to coordinates
func DecodeCoords(data []byte) (x, y, z, rx, ry, rz float64, err error) {
	coords, err := DecodeAngles(data)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	if len(coords) != 6 {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid coord data: expected 6 values, got %d", len(coords))
	}

	return coords[0], coords[1], coords[2], coords[3], coords[4], coords[5], nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./protocol -v -run "TestEncodeInt16|TestEncodeAngles|TestDecodeAngles"
```

Expected: PASS (all tests)

**Step 5: Commit encoding helpers**

Run:
```bash
git add protocol/
git commit -m "feat(protocol): add data encoding/decoding helpers for angles and coords"
```

---

### Task 6: Types Package - Basic Domain Types

**Files:**
- Create: `types/joint.go`
- Create: `types/joint_test.go`
- Create: `types/speed.go`
- Create: `types/speed_test.go`

**Step 1: Write tests for JointID**

Create `types/joint_test.go`:
```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJointID_Constants(t *testing.T) {
	assert.Equal(t, JointID(1), Joint1)
	assert.Equal(t, JointID(2), Joint2)
	assert.Equal(t, JointID(3), Joint3)
	assert.Equal(t, JointID(4), Joint4)
	assert.Equal(t, JointID(5), Joint5)
	assert.Equal(t, JointID(6), Joint6)
}

func TestJointID_Validate(t *testing.T) {
	tests := []struct {
		joint       JointID
		jointCount  int
		expectError bool
	}{
		{Joint1, 6, false},
		{Joint6, 6, false},
		{JointID(0), 6, true},
		{JointID(7), 6, true},
		{Joint4, 4, false},
		{Joint5, 4, true},
	}

	for _, tt := range tests {
		err := tt.joint.Validate(tt.jointCount)
		if tt.expectError {
			assert.Error(t, err, "joint=%d, count=%d", tt.joint, tt.jointCount)
		} else {
			assert.NoError(t, err, "joint=%d, count=%d", tt.joint, tt.jointCount)
		}
	}
}
```

**Step 2: Write tests for Speed**

Create `types/speed_test.go`:
```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpeed_Constants(t *testing.T) {
	assert.Equal(t, Speed(0), SpeedMin)
	assert.Equal(t, Speed(25), SpeedSlow)
	assert.Equal(t, Speed(50), SpeedMedium)
	assert.Equal(t, Speed(75), SpeedFast)
	assert.Equal(t, Speed(100), SpeedMax)
}

func TestSpeed_Validate(t *testing.T) {
	tests := []struct {
		speed       Speed
		expectError bool
	}{
		{SpeedMin, false},
		{SpeedMax, false},
		{SpeedMedium, false},
		{Speed(50), false},
		{Speed(-1), true},
		{Speed(101), true},
	}

	for _, tt := range tests {
		err := tt.speed.Validate()
		if tt.expectError {
			assert.Error(t, err, "speed=%d", tt.speed)
		} else {
			assert.NoError(t, err, "speed=%d", tt.speed)
		}
	}
}
```

**Step 3: Run tests to verify they fail**

Run:
```bash
go test ./types -v
```

Expected: FAIL - package types not found

**Step 4: Implement JointID type**

Create `types/joint.go`:
```go
package types

import "fmt"

// JointID represents a robot joint identifier (1-based indexing)
type JointID int

const (
	Joint1 JointID = 1
	Joint2 JointID = 2
	Joint3 JointID = 3
	Joint4 JointID = 4
	Joint5 JointID = 5
	Joint6 JointID = 6
)

// Validate checks if joint ID is valid for given joint count
func (j JointID) Validate(jointCount int) error {
	if j < 1 || int(j) > jointCount {
		return fmt.Errorf("invalid joint ID %d for robot with %d joints", j, jointCount)
	}
	return nil
}

// Index returns 0-based index for array access
func (j JointID) Index() int {
	return int(j) - 1
}
```

**Step 5: Implement Speed type**

Create `types/speed.go`:
```go
package types

import "fmt"

// Speed represents robot movement speed (0-100)
type Speed int

const (
	SpeedMin    Speed = 0
	SpeedSlow   Speed = 25
	SpeedMedium Speed = 50
	SpeedFast   Speed = 75
	SpeedMax    Speed = 100
)

// Validate checks if speed is in valid range
func (s Speed) Validate() error {
	if s < 0 || s > 100 {
		return fmt.Errorf("speed %d out of range [0, 100]", s)
	}
	return nil
}
```

**Step 6: Run tests to verify they pass**

Run:
```bash
go test ./types -v
```

Expected: PASS

**Step 7: Commit basic types**

Run:
```bash
git add types/
git commit -m "feat(types): add JointID and Speed domain types with validation"
```

---

### Task 7: Types Package - Angle and Coord Types

**Files:**
- Create: `types/angle.go`
- Create: `types/angle_test.go`
- Create: `types/coord.go`
- Create: `types/coord_test.go`
- Create: `types/model.go`

**Step 1: Write tests for Angle**

Create `types/angle_test.go`:
```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAngle_ValidateForJoint(t *testing.T) {
	tests := []struct {
		name        string
		angle       Angle
		joint       JointID
		model       Model
		expectError bool
	}{
		{"valid within range", Angle(45), Joint1, ModelMyCobot280, false},
		{"valid at min", Angle(-165), Joint1, ModelMyCobot280, false},
		{"valid at max", Angle(165), Joint1, ModelMyCobot280, false},
		{"below min", Angle(-170), Joint1, ModelMyCobot280, true},
		{"above max", Angle(170), Joint1, ModelMyCobot280, true},
		{"different joint", Angle(170), Joint6, ModelMyCobot280, false}, // Joint6 has different limits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.angle.ValidateForJoint(tt.joint, tt.model)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAngles_Validate(t *testing.T) {
	validAngles := Angles{0, 45, -90, 30, -45, 90}
	invalidLengthAngles := Angles{0, 45, -90} // only 3 angles
	invalidValueAngles := Angles{0, 45, -200, 30, -45, 90} // -200 out of range

	err := validAngles.Validate(6, ModelMyCobot280)
	assert.NoError(t, err)

	err = invalidLengthAngles.Validate(6, ModelMyCobot280)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")

	err = invalidValueAngles.Validate(6, ModelMyCobot280)
	assert.Error(t, err)
}
```

**Step 2: Write tests for Coord**

Create `types/coord_test.go`:
```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoord_Creation(t *testing.T) {
	coord := Coord{
		X:  100.5,
		Y:  -50.2,
		Z:  200.0,
		Rx: 45.0,
		Ry: -30.5,
		Rz: 90.0,
	}

	assert.Equal(t, 100.5, coord.X)
	assert.Equal(t, -50.2, coord.Y)
	assert.Equal(t, 200.0, coord.Z)
	assert.Equal(t, 45.0, coord.Rx)
	assert.Equal(t, -30.5, coord.Ry)
	assert.Equal(t, 90.0, coord.Rz)
}

func TestCoord_ToSlice(t *testing.T) {
	coord := Coord{X: 1, Y: 2, Z: 3, Rx: 4, Ry: 5, Rz: 6}

	slice := coord.ToSlice()

	expected := []float64{1, 2, 3, 4, 5, 6}
	assert.Equal(t, expected, slice)
}
```

**Step 3: Run tests to verify they fail**

Run:
```bash
go test ./types -v -run "TestAngle|TestCoord"
```

Expected: FAIL - "undefined: Angle"

**Step 4: Implement Model type**

Create `types/model.go`:
```go
package types

// Model represents a robot model type
type Model string

const (
	ModelMyCobot280      Model = "MyCobot280"
	ModelMyCobot320      Model = "MyCobot320"
	ModelMechArm270      Model = "MechArm270"
	ModelMyPalletizer260 Model = "MyPalletizer260"
)

// JointLimit defines min/max angle for a joint
type JointLimit struct {
	MinAngle float64
	MaxAngle float64
}

// getJointLimits returns joint limits for a specific model and joint
func getJointLimits(model Model, joint JointID) JointLimit {
	limits := modelJointLimits[model]
	if int(joint) <= len(limits) {
		return limits[joint.Index()]
	}
	// Default safe limits if not found
	return JointLimit{MinAngle: -165, MaxAngle: 165}
}

var modelJointLimits = map[Model][]JointLimit{
	ModelMyCobot280: {
		{MinAngle: -165, MaxAngle: 165}, // Joint 1
		{MinAngle: -165, MaxAngle: 165}, // Joint 2
		{MinAngle: -165, MaxAngle: 165}, // Joint 3
		{MinAngle: -165, MaxAngle: 165}, // Joint 4
		{MinAngle: -165, MaxAngle: 165}, // Joint 5
		{MinAngle: -175, MaxAngle: 175}, // Joint 6
	},
	ModelMyCobot320: {
		{MinAngle: -170, MaxAngle: 170}, // Joint 1
		{MinAngle: -137, MaxAngle: 137}, // Joint 2
		{MinAngle: -150, MaxAngle: 150}, // Joint 3
		{MinAngle: -145, MaxAngle: 145}, // Joint 4
		{MinAngle: -165, MaxAngle: 165}, // Joint 5
		{MinAngle: -180, MaxAngle: 180}, // Joint 6
	},
	ModelMechArm270: {
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -175, MaxAngle: 175},
	},
	ModelMyPalletizer260: {
		{MinAngle: -165, MaxAngle: 165},
		{MinAngle: -90, MaxAngle: 90},
		{MinAngle: -90, MaxAngle: 90},
		{MinAngle: -165, MaxAngle: 165},
	},
}
```

**Step 5: Implement Angle type**

Create `types/angle.go`:
```go
package types

import "fmt"

// Angle represents a joint angle in degrees
type Angle float64

// ValidateForJoint checks if angle is valid for specific joint and model
func (a Angle) ValidateForJoint(joint JointID, model Model) error {
	limits := getJointLimits(model, joint)
	if float64(a) < limits.MinAngle || float64(a) > limits.MaxAngle {
		return fmt.Errorf("angle %.2f out of range [%.2f, %.2f] for joint %d on %s",
			a, limits.MinAngle, limits.MaxAngle, joint, model)
	}
	return nil
}

// Angles represents a set of joint angles
type Angles []Angle

// Validate checks if all angles are valid for the given model
func (a Angles) Validate(jointCount int, model Model) error {
	if len(a) != jointCount {
		return fmt.Errorf("expected %d angles, got %d", jointCount, len(a))
	}

	for i, angle := range a {
		joint := JointID(i + 1)
		if err := angle.ValidateForJoint(joint, model); err != nil {
			return err
		}
	}

	return nil
}

// ToFloat64 converts Angles to []float64
func (a Angles) ToFloat64() []float64 {
	result := make([]float64, len(a))
	for i, angle := range a {
		result[i] = float64(angle)
	}
	return result
}
```

**Step 6: Implement Coord type**

Create `types/coord.go`:
```go
package types

// Coord represents a 3D coordinate with rotation
type Coord struct {
	X, Y, Z    float64 // Position in mm
	Rx, Ry, Rz float64 // Rotation in degrees
}

// ToSlice converts coordinate to slice for encoding
func (c Coord) ToSlice() []float64 {
	return []float64{c.X, c.Y, c.Z, c.Rx, c.Ry, c.Rz}
}

// NewCoordFromSlice creates Coord from slice
func NewCoordFromSlice(data []float64) (Coord, error) {
	if len(data) != 6 {
		return Coord{}, fmt.Errorf("expected 6 values, got %d", len(data))
	}
	return Coord{
		X:  data[0],
		Y:  data[1],
		Z:  data[2],
		Rx: data[3],
		Ry: data[4],
		Rz: data[5],
	}, nil
}
```

Add import to `types/coord.go`:
```go
import "fmt"
```

**Step 7: Run tests to verify they pass**

Run:
```bash
go test ./types -v
```

Expected: PASS

**Step 8: Commit angle and coord types**

Run:
```bash
git add types/
git commit -m "feat(types): add Angle, Angles, and Coord types with model-specific validation"
```

---

### Task 8: Error Types and Base Robot Structure

**Files:**
- Create: `errors.go`
- Create: `errors_test.go`
- Create: `robot.go`

**Step 1: Write tests for error types**

Create `errors_test.go`:
```go
package mycobot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRobotError_Error(t *testing.T) {
	err := &RobotError{
		Op:    "SendAngles",
		Model: "MyCobot280",
		Err:   ErrInvalidSpeed,
	}

	msg := err.Error()
	assert.Contains(t, msg, "MyCobot280")
	assert.Contains(t, msg, "SendAngles")
	assert.Contains(t, msg, "speed")
}

func TestRobotError_Unwrap(t *testing.T) {
	err := &RobotError{
		Op:    "PowerOn",
		Model: "MyCobot280",
		Err:   ErrConnectionTimeout,
	}

	unwrapped := errors.Unwrap(err)
	assert.Equal(t, ErrConnectionTimeout, unwrapped)
}

func TestStandardErrors_Defined(t *testing.T) {
	// Just verify errors are defined
	assert.NotNil(t, ErrRobotClosed)
	assert.NotNil(t, ErrNotConnected)
	assert.NotNil(t, ErrInvalidSpeed)
	assert.NotNil(t, ErrNoGripper)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestRobotError
```

Expected: FAIL - "undefined: RobotError"

**Step 3: Implement error types**

Create `errors.go`:
```go
package mycobot

import (
	"errors"
	"fmt"
)

// Standard errors
var (
	// Connection errors
	ErrRobotClosed       = errors.New("robot connection closed")
	ErrNotConnected      = errors.New("robot not connected")
	ErrConnectionTimeout = errors.New("connection timeout")

	// Command errors
	ErrInvalidCommand  = errors.New("invalid command")
	ErrCommandTimeout  = errors.New("command timeout")
	ErrInvalidResponse = errors.New("invalid response from robot")

	// Validation errors
	ErrInvalidJoint      = errors.New("invalid joint ID")
	ErrInvalidSpeed      = errors.New("speed out of range")
	ErrInvalidAngle      = errors.New("angle out of range")
	ErrInvalidCoordinate = errors.New("coordinate out of range")

	// Gripper errors
	ErrNoGripper           = errors.New("no gripper attached")
	ErrGripperNotSupported = errors.New("gripper operation not supported")

	// State errors
	ErrNotPowered    = errors.New("robot not powered on")
	ErrEmergencyStop = errors.New("emergency stop active")
	ErrServoError    = errors.New("servo error detected")
)

// RobotError wraps errors with operation and model context
type RobotError struct {
	Op    string // Operation that failed (e.g., "SendAngles")
	Model string // Robot model (e.g., "MyCobot280")
	Err   error  // Underlying error
}

func (e *RobotError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Model, e.Op, e.Err)
}

func (e *RobotError) Unwrap() error {
	return e.Err
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestRobotError
```

Expected: PASS

**Step 5: Implement Robot interface**

Create `robot.go`:
```go
package mycobot

import (
	"context"

	"github.com/yourusername/mycobot-go/types"
)

// Robot is the base interface all robot models implement
type Robot interface {
	// Connection management
	Open(ctx context.Context) error
	Close() error
	IsConnected() bool

	// Core motion commands
	SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error
	GetAngles(ctx context.Context) (types.Angles, error)
	SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error
	GetCoords(ctx context.Context) (types.Coord, error)

	// Power and status
	PowerOn(ctx context.Context) error
	PowerOff(ctx context.Context) error
	IsPowerOn(ctx context.Context) (bool, error)

	// Movement queries
	IsMoving(ctx context.Context) (bool, error)
	IsInPosition(ctx context.Context, target types.Coord, tolerance float64) (bool, error)
}
```

**Step 6: Commit errors and robot interface**

Run:
```bash
git add errors.go errors_test.go robot.go
git commit -m "feat: add error types and Robot interface definition"
```

---

### Task 9: Base Robot Structure with Model Configuration

**Files:**
- Create: `internal/robot/base.go`
- Create: `internal/robot/base_test.go`
- Create: `config.go`
- Create: `config_test.go`

**Step 1: Write tests for model configuration**

Create `config_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/mycobot-go/types"
)

func TestModelConfig_MyCobot280(t *testing.T) {
	config := getModelConfig(types.ModelMyCobot280)

	assert.Equal(t, types.ModelMyCobot280, config.Model)
	assert.Equal(t, 6, config.JointCount)
	assert.Equal(t, 115200, config.DefaultBaud)
	assert.True(t, config.UseCRC)
	assert.Len(t, config.JointLimits, 6)
}

func TestModelConfig_AllModels(t *testing.T) {
	models := []types.Model{
		types.ModelMyCobot280,
		types.ModelMyCobot320,
		types.ModelMechArm270,
		types.ModelMyPalletizer260,
	}

	for _, model := range models {
		config := getModelConfig(model)
		assert.NotNil(t, config)
		assert.Equal(t, model, config.Model)
		assert.Greater(t, config.JointCount, 0)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestModelConfig
```

Expected: FAIL - "undefined: getModelConfig"

**Step 3: Implement model configuration**

Create `config.go`:
```go
package mycobot

import "github.com/yourusername/mycobot-go/types"

// ModelConfig defines model-specific parameters
type ModelConfig struct {
	Model         types.Model
	JointCount    int
	JointLimits   []types.JointLimit
	UseCRC        bool
	DefaultBaud   int
	SupportedBaud []int
}

func getModelConfig(model types.Model) ModelConfig {
	return modelConfigs[model]
}

var modelConfigs = map[types.Model]ModelConfig{
	types.ModelMyCobot280: {
		Model:      types.ModelMyCobot280,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -175, MaxAngle: 175},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMyCobot320: {
		Model:      types.ModelMyCobot320,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -170, MaxAngle: 170},
			{MinAngle: -137, MaxAngle: 137},
			{MinAngle: -150, MaxAngle: 150},
			{MinAngle: -145, MaxAngle: 145},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -180, MaxAngle: 180},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMechArm270: {
		Model:      types.ModelMechArm270,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -175, MaxAngle: 175},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
	types.ModelMyPalletizer260: {
		Model:      types.ModelMyPalletizer260,
		JointCount: 4,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165},
			{MinAngle: -90, MaxAngle: 90},
			{MinAngle: -90, MaxAngle: 90},
			{MinAngle: -165, MaxAngle: 165},
		},
		UseCRC:        true,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
}
```

**Step 4: Write tests for base robot structure**

Create `internal/robot/base_test.go`:
```go
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
```

**Step 5: Implement base robot structure**

Create `internal/robot/base.go`:
```go
package robot

import (
	"context"
	"sync"

	"go.bug.st/serial"
	"github.com/yourusername/mycobot-go/protocol"
)

// Base provides common robot functionality
type Base struct {
	port      string
	baudrate  int
	useCRC    bool
	connected bool

	// Serial connection (owned by command loop goroutine)
	conn *serial.Port

	// Command queue
	cmdChan   chan *command
	closeChan chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

	mu sync.RWMutex // Protects connected flag
}

type command struct {
	ctx      context.Context
	request  protocol.Command
	response chan *response
}

type response struct {
	data []byte
	err  error
}

// NewBase creates a new base robot
func NewBase(port string, baudrate int, useCRC bool) *Base {
	return &Base{
		port:      port,
		baudrate:  baudrate,
		useCRC:    useCRC,
		connected: false,
	}
}

// IsConnected returns true if robot is connected
func (b *Base) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}
```

**Step 6: Run tests to verify they pass**

Run:
```bash
go test . -v -run TestModelConfig
go test ./internal/robot -v
```

Expected: PASS (all tests)

**Step 7: Commit base robot structure**

Run:
```bash
git add config.go config_test.go internal/
git commit -m "feat: add base robot structure and model configuration"
```

---

### Task 10: Channel-Based Command Loop and Serial Communication

**Files:**
- Modify: `internal/robot/base.go`
- Modify: `internal/robot/base_test.go`

**Step 1: Write tests for Open/Close**

Add to `internal/robot/base_test.go`:
```go
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
```

**Step 2: Write tests for SendCommand**

Add to `internal/robot/base_test.go`:
```go
func TestBase_SendCommand_NotConnected(t *testing.T) {
	base := NewBase("/dev/ttyUSB0", 115200, true)
	ctx := context.Background()

	cmd := protocol.Command{Code: protocol.PowerOn, UseCRC: true}
	_, err := base.SendCommand(ctx, cmd)

	assert.Error(t, err)
	// Should fail because not connected
}
```

**Step 3: Implement Open method**

Add to `internal/robot/base.go`:
```go
import (
	"context"
	"fmt"
	"time"
)

// Open establishes serial connection and starts command loop
func (b *Base) Open(ctx context.Context) error {
	conn, err := serial.Open(b.port, &serial.Mode{
		BaudRate: b.baudrate,
	})
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	b.conn = conn
	b.cmdChan = make(chan *command, 32)
	b.closeChan = make(chan struct{})

	b.mu.Lock()
	b.connected = true
	b.mu.Unlock()

	b.wg.Add(1)
	go b.commandLoop()

	return nil
}
```

**Step 4: Implement command loop**

Add to `internal/robot/base.go`:
```go
// commandLoop runs in dedicated goroutine, owns serial connection
func (b *Base) commandLoop() {
	defer b.wg.Done()
	defer b.conn.Close()
	defer func() {
		b.mu.Lock()
		b.connected = false
		b.mu.Unlock()
	}()

	for {
		select {
		case <-b.closeChan:
			return

		case cmd := <-b.cmdChan:
			// Check if command context is already cancelled
			if err := cmd.ctx.Err(); err != nil {
				cmd.response <- &response{err: err}
				continue
			}

			// Set CRC mode from base config
			cmd.request.UseCRC = b.useCRC

			// Encode and write command
			data, err := cmd.request.Encode()
			if err != nil {
				cmd.response <- &response{err: fmt.Errorf("encode failed: %w", err)}
				continue
			}

			if _, err := b.conn.Write(data); err != nil {
				cmd.response <- &response{err: fmt.Errorf("write failed: %w", err)}
				continue
			}

			// Read response with context deadline
			respData, err := b.readResponse(cmd.ctx, cmd.request.Code)
			cmd.response <- &response{data: respData, err: err}
		}
	}
}
```

**Step 5: Implement readResponse**

Add to `internal/robot/base.go`:
```go
import (
	"io"
	"github.com/yourusername/mycobot-go/protocol"
)

// readResponse reads and decodes response from serial port
func (b *Base) readResponse(ctx context.Context, expectedCode byte) ([]byte, error) {
	// Read response with timeout
	deadline := time.Now().Add(1 * time.Second)
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	}

	// Read until we have enough data for a minimal packet
	buf := make([]byte, 256)
	totalRead := 0

	for time.Now().Before(deadline) {
		b.conn.SetReadDeadline(deadline)
		n, err := b.conn.Read(buf[totalRead:])
		if err != nil && err != io.EOF {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("read timeout")
			}
			continue
		}

		totalRead += n

		// Try to decode what we have
		if totalRead >= 5 {
			resp, err := protocol.Decode(buf[:totalRead], b.useCRC)
			if err == nil {
				return resp.Data, nil
			}
		}
	}

	return nil, fmt.Errorf("read timeout after %d bytes", totalRead)
}
```

**Step 6: Implement SendCommand and Close**

Add to `internal/robot/base.go`:
```go
import (
	mycobot "github.com/yourusername/mycobot-go"
)

// SendCommand queues a command and waits for response
func (b *Base) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	if !b.IsConnected() {
		return nil, mycobot.ErrNotConnected
	}

	responseChan := make(chan *response, 1)

	select {
	case b.cmdChan <- &command{
		ctx:      ctx,
		request:  cmd,
		response: responseChan,
	}:
		// Command queued successfully

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-b.closeChan:
		return nil, mycobot.ErrRobotClosed
	}

	// Wait for response
	select {
	case resp := <-responseChan:
		return resp.data, resp.err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close shuts down the command loop gracefully
func (b *Base) Close() error {
	b.closeOnce.Do(func() {
		close(b.closeChan)
	})
	b.wg.Wait()
	return nil
}
```

**Step 7: Run tests to verify they pass**

Run:
```bash
go test ./internal/robot -v
```

Expected: PASS (unit tests)

**Step 8: Commit command loop implementation**

Run:
```bash
git add internal/robot/
git commit -m "feat: add channel-based command loop and serial communication"
```

---

### Task 11: MyCobot280 Basic Implementation

**Files:**
- Create: `mycobot280.go`
- Create: `mycobot280_test.go`
- Create: `option.go`

**Step 1: Write tests for MyCobot280 construction**

Create `mycobot280_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/mycobot-go/types"
)

func TestNewMyCobot280(t *testing.T) {
	robot := NewMyCobot280("/dev/ttyUSB0")

	assert.NotNil(t, robot)
	assert.Equal(t, types.ModelMyCobot280, robot.config.Model)
	assert.Equal(t, 6, robot.config.JointCount)
	assert.NotNil(t, robot.base)
}

func TestNewMyCobot280_WithOptions(t *testing.T) {
	robot := NewMyCobot280("/dev/ttyUSB0",
		WithBaudRate(1000000),
	)

	assert.NotNil(t, robot)
	// Baud rate will be tested in integration tests
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestNewMyCobot280
```

Expected: FAIL - "undefined: NewMyCobot280"

**Step 3: Implement Option pattern**

Create `option.go`:
```go
package mycobot

import (
	"time"
	"github.com/yourusername/mycobot-go/internal/robot"
)

// Option configures a robot
type Option func(*robot.Base)

// WithBaudRate sets custom baud rate
func WithBaudRate(baud int) Option {
	return func(b *robot.Base) {
		// Baud rate is set during construction
		// This will be applied in NewMyCobot280
	}
}

// WithTimeout sets default command timeout
func WithTimeout(timeout time.Duration) Option {
	return func(b *robot.Base) {
		// Will be implemented when we add timeout support
	}
}
```

**Step 4: Implement MyCobot280**

Create `mycobot280.go`:
```go
package mycobot

import (
	"context"

	"github.com/yourusername/mycobot-go/internal/robot"
	"github.com/yourusername/mycobot-go/protocol"
	"github.com/yourusername/mycobot-go/types"
)

// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base   *robot.Base
	config ModelConfig
}

// NewMyCobot280 creates a new MyCobot280 instance
func NewMyCobot280(port string, opts ...Option) *MyCobot280 {
	config := getModelConfig(types.ModelMyCobot280)

	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	// Apply options
	for _, opt := range opts {
		opt(base)
	}

	return &MyCobot280{
		base:   base,
		config: config,
	}
}

// Open establishes connection to the robot
func (m *MyCobot280) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MyCobot280) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MyCobot280) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MyCobot280) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}
```

**Step 5: Run tests to verify they pass**

Run:
```bash
go test . -v -run TestNewMyCobot280
```

Expected: PASS

**Step 6: Commit MyCobot280 implementation**

Run:
```bash
git add mycobot280.go mycobot280_test.go option.go
git commit -m "feat: add MyCobot280 basic implementation with connection management"
```

---

### Task 12: Core Motion Commands Implementation

**Files:**
- Modify: `mycobot280.go`
- Modify: `mycobot280_test.go`

**Step 1: Write tests for PowerOn/PowerOff**

Add to `mycobot280_test.go`:
```go
func TestMyCobot280_PowerOn(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// This will fail without hardware, but tests the method exists
	err := robot.PowerOn(ctx)
	// We expect an error because we're not connected
	assert.Error(t, err)
}

func TestMyCobot280_PowerOff(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	err := robot.PowerOff(ctx)
	assert.Error(t, err)
}
```

**Step 2: Write tests for GetAngles/SendAngles**

Add to `mycobot280_test.go`:
```go
func TestMyCobot280_SendAngles_Validation(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// Test with invalid angle count
	angles := types.Angles{0, 45, 90} // Only 3 angles, need 6
	err := robot.SendAngles(ctx, angles, types.SpeedMedium)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")
}

func TestMyCobot280_SendAngles_OutOfRange(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// Angle out of range for MyCobot280
	angles := types.Angles{0, 0, 0, 0, 0, 200} // 200 > 175 for joint 6
	err := robot.SendAngles(ctx, angles, types.SpeedMedium)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}
```

**Step 3: Implement PowerOn/PowerOff**

Add to `mycobot280.go`:
```go
// PowerOn powers on all servos
func (m *MyCobot280) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MyCobot280) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MyCobot280) IsPowerOn(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPowerOn}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
```

**Step 4: Implement SendAngles/GetAngles**

Add to `mycobot280.go`:
```go
// SendAngles sends joint angles to the robot
func (m *MyCobot280) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	// Validate angles
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}

	// Validate speed
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode angles
	angleData := protocol.EncodeAngles(angles.ToFloat64())

	// Append speed
	data := append(angleData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (m *MyCobot280) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	// Convert to types.Angles
	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}

	return result, nil
}
```

**Step 5: Implement SendCoords/GetCoords**

Add to `mycobot280.go`:
```go
// SendCoords sends coordinate position to the robot
func (m *MyCobot280) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error {
	// Validate speed
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode coordinates
	coordData := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)

	// Append speed
	data := append(coordData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MyCobot280) GetCoords(ctx context.Context) (types.Coord, error) {
	cmd := protocol.Command{Code: protocol.GetCoords}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return types.Coord{}, err
	}

	x, y, z, rx, ry, rz, err := protocol.DecodeCoords(data)
	if err != nil {
		return types.Coord{}, err
	}

	return types.Coord{
		X:  x,
		Y:  y,
		Z:  z,
		Rx: rx,
		Ry: ry,
		Rz: rz,
	}, nil
}
```

**Step 6: Implement IsMoving**

Add to `mycobot280.go`:
```go
// IsMoving returns true if robot is currently moving
func (m *MyCobot280) IsMoving(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsMoving}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
```

**Step 7: Run tests to verify they pass**

Run:
```bash
go test . -v
```

Expected: PASS (validation tests)

**Step 8: Commit core motion commands**

Run:
```bash
git add mycobot280.go mycobot280_test.go
git commit -m "feat: add core motion commands (PowerOn, SendAngles, GetAngles, SendCoords, GetCoords, IsMoving)"
```

---

## Phase 1 Complete - Next Steps

The implementation plan continues with:

**Phase 2: Extended Features** (35+ tasks)
- Motion, IO, Servo subsystems
- Gripper interface and implementations
- All four robot models

**Phase 3: Advanced Features** (20+ tasks)
- Advanced gripper types
- Full protocol coverage
- Examples

**Phase 4: Polish** (15+ tasks)
- Comprehensive testing
- Documentation
- CI/CD

**Total estimated tasks: 80+**

Each task follows the same pattern:
1. Write failing test
2. Run test (verify failure)
3. Implement minimal code
4. Run test (verify pass)
5. Commit

This ensures Test-Driven Development (TDD) throughout implementation.
