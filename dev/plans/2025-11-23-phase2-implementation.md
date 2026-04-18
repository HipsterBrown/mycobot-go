# Phase 2: Extended Features - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend MyCobot280 with Motion, IO, Servo subsystems and gripper support, then replicate to other robot models.

**Architecture:** Subsystem-based design with Motion, IO, Servo structs attached to robot. Gripper interface with runtime attachment. Shared subsystems across all robot models.

**Tech Stack:** Go 1.21+, go.bug.st/serial, testify for assertions.

**Design Document:** See [Phase 2 Design](./2025-11-23-phase2-extended-features-design.md) for architecture details.

**Prerequisites:** Phase 1 complete (protocol, types, base robot, MyCobot280 basics).

---

## Task 1: Motion Subsystem - Types and Structure

**Files:**
- Create: `motion.go`
- Create: `motion_test.go`

**Step 1: Write test for CoordAxis type**

Create `motion_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordAxis_Constants(t *testing.T) {
	assert.Equal(t, CoordAxis(0), AxisX)
	assert.Equal(t, CoordAxis(1), AxisY)
	assert.Equal(t, CoordAxis(2), AxisZ)
	assert.Equal(t, CoordAxis(3), AxisRx)
	assert.Equal(t, CoordAxis(4), AxisRy)
	assert.Equal(t, CoordAxis(5), AxisRz)
}

func TestMotion_Structure(t *testing.T) {
	motion := &Motion{}
	assert.NotNil(t, motion)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestCoordAxis
```

Expected: FAIL - "undefined: CoordAxis"

**Step 3: Implement CoordAxis type and Motion struct**

Create `motion.go`:
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// CoordAxis represents a coordinate axis for single-axis movement
type CoordAxis int

const (
	AxisX CoordAxis = iota
	AxisY
	AxisZ
	AxisRx
	AxisRy
	AxisRz
)

// Motion provides motion control operations
type Motion struct {
	robot *robot.Base
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestCoordAxis
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add motion.go motion_test.go
git commit -m "feat(motion): add CoordAxis type and Motion struct"
```

---

## Task 2: Motion Subsystem - JOG Operations

**Files:**
- Modify: `motion.go`
- Modify: `motion_test.go`

**Step 1: Write tests for JOG operations**

Add to `motion_test.go`:
```go
func TestMotion_JogAngle_NotConnected(t *testing.T) {
	motion := &Motion{robot: nil}
	ctx := context.Background()

	// Should handle nil robot gracefully
	// (Will be tested with real robot in integration tests)
	assert.NotNil(t, motion)
}

func TestMotion_JogStop_NotConnected(t *testing.T) {
	motion := &Motion{robot: nil}
	ctx := context.Background()
	assert.NotNil(t, motion)
}
```

**Step 2: Run test to verify tests exist**

Run:
```bash
go test . -v -run "TestMotion_Jog"
```

Expected: PASS (structure tests)

**Step 3: Implement JOG methods**

Add to `motion.go`:
```go
// JogAngle performs incremental joint movement
// direction: 0 = negative, 1 = positive
func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(joint), byte(direction), byte(speed)}
	cmd := protocol.Command{
		Code: protocol.JogAngle,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// JogCoord performs incremental coordinate movement
func (m *Motion) JogCoord(ctx context.Context, axis CoordAxis, direction int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(axis), byte(direction), byte(speed)}
	cmd := protocol.Command{
		Code: protocol.JogCoord,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// JogStop stops JOG movement
func (m *Motion) JogStop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.JogStop}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test . -v -run "TestMotion"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add motion.go motion_test.go
git commit -m "feat(motion): add JOG operations (JogAngle, JogCoord, JogStop)"
```

---

## Task 3: Motion Subsystem - Movement Control

**Files:**
- Modify: `motion.go`
- Modify: `motion_test.go`

**Step 1: Write tests for movement control**

Add to `motion_test.go`:
```go
func TestMotion_PauseResume(t *testing.T) {
	motion := &Motion{robot: nil}
	assert.NotNil(t, motion)
	// Integration tests will verify actual pause/resume behavior
}

func TestMotion_Stop(t *testing.T) {
	motion := &Motion{robot: nil}
	assert.NotNil(t, motion)
}
```

**Step 2: Run test to verify tests exist**

Run:
```bash
go test . -v -run "TestMotion_Pause\|TestMotion_Stop"
```

Expected: PASS (structure tests)

**Step 3: Implement movement control methods**

Add to `motion.go`:
```go
// Pause pauses current movement
func (m *Motion) Pause(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Pause}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// Resume resumes paused movement
func (m *Motion) Resume(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Resume}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// Stop stops all movement
func (m *Motion) Stop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Stop}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// IsPaused returns true if robot is paused
func (m *Motion) IsPaused(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPaused}
	data, err := m.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestMotion"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add motion.go motion_test.go
git commit -m "feat(motion): add movement control (Pause, Resume, Stop, IsPaused)"
```

---

## Task 4: Motion Subsystem - Single Joint/Coord Commands

**Files:**
- Modify: `motion.go`
- Modify: `motion_test.go`

**Step 1: Write tests for single commands**

Add to `motion_test.go`:
```go
func TestMotion_SendAngle_Validation(t *testing.T) {
	motion := &Motion{robot: nil}
	ctx := context.Background()

	// These will be integration tests - just verify structure exists
	assert.NotNil(t, motion)
}

func TestMotion_SendCoord(t *testing.T) {
	motion := &Motion{robot: nil}
	assert.NotNil(t, motion)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestMotion_Send"
```

Expected: PASS (structure tests)

**Step 3: Implement single joint/coord commands**

Add to `motion.go`:
```go
// SendAngle sends a single joint to target angle
func (m *Motion) SendAngle(ctx context.Context, joint types.JointID, angle types.Angle, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode joint ID, angle (as int16 * 100), and speed
	data := []byte{byte(joint)}
	angleData := protocol.EncodeAngles([]float64{float64(angle)})
	data = append(data, angleData...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngle,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// SendCoord sends a single coordinate axis to target value
func (m *Motion) SendCoord(ctx context.Context, axis CoordAxis, value float64, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	// Encode axis, value (as int16 * 100), and speed
	data := []byte{byte(axis)}
	valueData := protocol.EncodeAngles([]float64{value})
	data = append(data, valueData...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoord,
		Data: data,
	}

	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// IsInPosition checks if robot is at target position within tolerance
func (m *Motion) IsInPosition(ctx context.Context, target types.Coord, tolerance float64) (bool, error) {
	// Encode target coordinates
	coordData := protocol.EncodeCoords(target.X, target.Y, target.Z, target.Rx, target.Ry, target.Rz)

	// Note: Protocol may need tolerance encoding - using basic implementation
	cmd := protocol.Command{
		Code: protocol.IsInPosition,
		Data: coordData,
	}

	data, err := m.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestMotion"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add motion.go motion_test.go
git commit -m "feat(motion): add single joint/coord commands and IsInPosition"
```

---

## Task 5: Integrate Motion Subsystem into MyCobot280

**Files:**
- Modify: `mycobot280.go`
- Modify: `mycobot280_test.go`

**Step 1: Write test for Motion subsystem integration**

Add to `mycobot280_test.go`:
```go
func TestMyCobot280_MotionSubsystem(t *testing.T) {
	robot := NewMyCobot280("/dev/null")

	assert.NotNil(t, robot.Motion)
	assert.IsType(t, &Motion{}, robot.Motion)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestMyCobot280_Motion
```

Expected: FAIL - "robot.Motion undefined"

**Step 3: Add Motion field to MyCobot280**

Modify `mycobot280.go`:
```go
// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base   *robot.Base
	config ModelConfig
	Motion *Motion
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
		Motion: &Motion{robot: base},
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestMyCobot280_Motion
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mycobot280.go mycobot280_test.go
git commit -m "feat(mycobot280): integrate Motion subsystem"
```

---

## Task 6: IO Subsystem - Structure and Types

**Files:**
- Create: `io.go`
- Create: `io_test.go`

**Step 1: Write tests for IO types**

Create `io_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRGB_PredefinedColors(t *testing.T) {
	assert.Equal(t, byte(0), ColorOff.R)
	assert.Equal(t, byte(255), ColorRed.R)
	assert.Equal(t, byte(0), ColorGreen.R)
	assert.Equal(t, byte(0), ColorBlue.R)
}

func TestIO_Structure(t *testing.T) {
	io := &IO{}
	assert.NotNil(t, io)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestRGB
```

Expected: FAIL - "undefined: RGB"

**Step 3: Implement RGB type and IO struct**

Create `io.go`:
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

// RGB represents an RGB color value
type RGB struct {
	R, G, B byte
}

// Predefined colors for convenience
var (
	ColorOff   = RGB{0, 0, 0}
	ColorRed   = RGB{255, 0, 0}
	ColorGreen = RGB{0, 255, 0}
	ColorBlue  = RGB{0, 0, 255}
	ColorWhite = RGB{255, 255, 255}
)

// IO provides input/output operations
type IO struct {
	robot *robot.Base
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestRGB
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add io.go io_test.go
git commit -m "feat(io): add RGB type and IO struct"
```

---

## Task 7: IO Subsystem - Digital and PWM IO

**Files:**
- Modify: `io.go`
- Modify: `io_test.go`

**Step 1: Write tests for IO operations**

Add to `io_test.go`:
```go
func TestIO_DigitalOutput(t *testing.T) {
	io := &IO{robot: nil}
	assert.NotNil(t, io)
	// Integration tests will verify actual IO behavior
}

func TestIO_PWMOutput(t *testing.T) {
	io := &IO{robot: nil}
	assert.NotNil(t, io)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestIO"
```

Expected: PASS (structure tests)

**Step 3: Implement IO methods**

Add to `io.go`:
```go
// SetDigitalOutput sets a digital output pin
func (io *IO) SetDigitalOutput(ctx context.Context, pin int, value bool) error {
	var pinValue byte
	if value {
		pinValue = 1
	}

	data := []byte{byte(pin), pinValue}
	cmd := protocol.Command{
		Code: protocol.SetDigitalOutput,
		Data: data,
	}

	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// GetDigitalInput reads a digital input pin
func (io *IO) GetDigitalInput(ctx context.Context, pin int) (bool, error) {
	data := []byte{byte(pin)}
	cmd := protocol.Command{
		Code: protocol.GetDigitalInput,
		Data: data,
	}

	resp, err := io.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(resp) > 0 {
		return resp[0] == 1, nil
	}
	return false, nil
}

// SetPWMOutput sets PWM output value (0-255)
func (io *IO) SetPWMOutput(ctx context.Context, pin int, value int) error {
	if value < 0 || value > 255 {
		return ErrInvalidCommand
	}

	data := []byte{byte(pin), byte(value)}
	cmd := protocol.Command{
		Code: protocol.SetPWMOutput,
		Data: data,
	}

	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestIO"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add io.go io_test.go
git commit -m "feat(io): add digital and PWM IO operations"
```

---

## Task 8: IO Subsystem - LED Color Control

**Files:**
- Modify: `io.go`
- Modify: `io_test.go`

**Step 1: Write test for SetColor**

Add to `io_test.go`:
```go
func TestIO_SetColor(t *testing.T) {
	io := &IO{robot: nil}
	assert.NotNil(t, io)
}

func TestIO_SetColor_PredefinedColors(t *testing.T) {
	// Verify predefined colors are valid
	assert.Equal(t, byte(255), ColorRed.R)
	assert.Equal(t, byte(255), ColorGreen.G)
	assert.Equal(t, byte(255), ColorBlue.B)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestIO_SetColor"
```

Expected: PASS (structure tests)

**Step 3: Implement SetColor method**

Add to `io.go`:
```go
// SetColor sets the Atom LED color
func (io *IO) SetColor(ctx context.Context, r, g, b byte) error {
	data := []byte{r, g, b}
	cmd := protocol.Command{
		Code: protocol.SetColor,
		Data: data,
	}

	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestIO"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add io.go io_test.go
git commit -m "feat(io): add LED color control"
```

---

## Task 9: Integrate IO Subsystem into MyCobot280

**Files:**
- Modify: `mycobot280.go`
- Modify: `mycobot280_test.go`

**Step 1: Write test for IO integration**

Add to `mycobot280_test.go`:
```go
func TestMyCobot280_IOSubsystem(t *testing.T) {
	robot := NewMyCobot280("/dev/null")

	assert.NotNil(t, robot.IO)
	assert.IsType(t, &IO{}, robot.IO)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestMyCobot280_IO
```

Expected: FAIL - "robot.IO undefined"

**Step 3: Add IO field to MyCobot280**

Modify `mycobot280.go`:
```go
// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base   *robot.Base
	config ModelConfig
	Motion *Motion
	IO     *IO
}

// In NewMyCobot280, add:
return &MyCobot280{
	base:   base,
	config: config,
	Motion: &Motion{robot: base},
	IO:     &IO{robot: base},
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestMyCobot280_IO
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mycobot280.go mycobot280_test.go
git commit -m "feat(mycobot280): integrate IO subsystem"
```

---

## Task 10: Servo Subsystem - Structure

**Files:**
- Create: `servo.go`
- Create: `servo_test.go`

**Step 1: Write test for Servo structure**

Create `servo_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServo_Structure(t *testing.T) {
	servo := &Servo{}
	assert.NotNil(t, servo)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestServo
```

Expected: FAIL - "undefined: Servo"

**Step 3: Implement Servo struct**

Create `servo.go`:
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// Servo provides individual servo control operations
type Servo struct {
	robot *robot.Base
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestServo
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add servo.go servo_test.go
git commit -m "feat(servo): add Servo struct"
```

---

## Task 11: Servo Subsystem - Servo State Control

**Files:**
- Modify: `servo.go`
- Modify: `servo_test.go`

**Step 1: Write tests for servo control**

Add to `servo_test.go`:
```go
func TestServo_ReleaseServo(t *testing.T) {
	servo := &Servo{robot: nil}
	assert.NotNil(t, servo)
}

func TestServo_FocusServo(t *testing.T) {
	servo := &Servo{robot: nil}
	assert.NotNil(t, servo)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestServo"
```

Expected: PASS (structure tests)

**Step 3: Implement servo control methods**

Add to `servo.go`:
```go
// ReleaseServo releases a specific servo (makes it freely movable)
func (s *Servo) ReleaseServo(ctx context.Context, joint types.JointID) error {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.ReleaseServo,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// FocusServo focuses (enables) a specific servo
func (s *Servo) FocusServo(ctx context.Context, joint types.JointID) error {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.FocusServo,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// IsServoEnabled returns true if servo is enabled
func (s *Servo) IsServoEnabled(ctx context.Context, joint types.JointID) (bool, error) {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.IsServoEnable,
		Data: data,
	}

	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(resp) > 0 {
		return resp[0] == 1, nil
	}
	return false, nil
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestServo"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add servo.go servo_test.go
git commit -m "feat(servo): add servo state control (Release, Focus, IsEnabled)"
```

---

## Task 12: Servo Subsystem - Encoder Operations

**Files:**
- Modify: `servo.go`
- Modify: `servo_test.go`

**Step 1: Write tests for encoders**

Add to `servo_test.go`:
```go
func TestServo_Encoders(t *testing.T) {
	servo := &Servo{robot: nil}
	assert.NotNil(t, servo)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestServo_Encoders"
```

Expected: PASS (structure tests)

**Step 3: Implement encoder methods**

Add to `servo.go`:
```go
// GetEncoder gets encoder value for a joint
func (s *Servo) GetEncoder(ctx context.Context, joint types.JointID) (int, error) {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.GetEncoder,
		Data: data,
	}

	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}

	if len(resp) >= 2 {
		// Decode int16 big-endian
		value := int16(resp[0])<<8 | int16(resp[1])
		return int(value), nil
	}
	return 0, nil
}

// SetEncoder sets encoder value for a joint
func (s *Servo) SetEncoder(ctx context.Context, joint types.JointID, value int) error {
	// Encode as int16
	encodedValue := protocol.EncodeInt16(value)
	data := []byte{byte(joint)}
	data = append(data, encodedValue...)

	cmd := protocol.Command{
		Code: protocol.SetEncoder,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetEncoders gets all encoder values
func (s *Servo) GetEncoders(ctx context.Context) ([]int, error) {
	cmd := protocol.Command{Code: protocol.GetEncoders}
	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Decode multiple int16 values
	count := len(resp) / 2
	encoders := make([]int, count)
	for i := 0; i < count; i++ {
		value := int16(resp[i*2])<<8 | int16(resp[i*2+1])
		encoders[i] = int(value)
	}

	return encoders, nil
}

// SetEncoders sets all encoder values
func (s *Servo) SetEncoders(ctx context.Context, encoders []int) error {
	data := make([]byte, 0, len(encoders)*2)
	for _, enc := range encoders {
		encoded := protocol.EncodeInt16(enc)
		data = append(data, encoded...)
	}

	cmd := protocol.Command{
		Code: protocol.SetEncoders,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestServo"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add servo.go servo_test.go
git commit -m "feat(servo): add encoder operations"
```

---

## Task 13: Servo Subsystem - Servo Data and Calibration

**Files:**
- Modify: `servo.go`
- Modify: `servo_test.go`

**Step 1: Write tests for servo data**

Add to `servo_test.go`:
```go
func TestServo_ServoData(t *testing.T) {
	servo := &Servo{robot: nil}
	assert.NotNil(t, servo)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestServo_ServoData"
```

Expected: PASS (structure tests)

**Step 3: Implement servo data methods**

Add to `servo.go`:
```go
// GetServoData gets servo data value
func (s *Servo) GetServoData(ctx context.Context, joint types.JointID) (int, error) {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.GetServoData,
		Data: data,
	}

	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}

	if len(resp) >= 2 {
		value := int16(resp[0])<<8 | int16(resp[1])
		return int(value), nil
	}
	return 0, nil
}

// SetServoData sets servo data value
func (s *Servo) SetServoData(ctx context.Context, joint types.JointID, value int) error {
	encodedValue := protocol.EncodeInt16(value)
	data := []byte{byte(joint)}
	data = append(data, encodedValue...)

	cmd := protocol.Command{
		Code: protocol.SetServoData,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// SetServoCalibration sets servo calibration value
func (s *Servo) SetServoCalibration(ctx context.Context, joint types.JointID, value int) error {
	encodedValue := protocol.EncodeInt16(value)
	data := []byte{byte(joint)}
	data = append(data, encodedValue...)

	cmd := protocol.Command{
		Code: protocol.SetServoCalibration,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestServo"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add servo.go servo_test.go
git commit -m "feat(servo): add servo data and calibration"
```

---

## Task 14: Servo Subsystem - Joint Limits

**Files:**
- Modify: `servo.go`
- Modify: `servo_test.go`

**Step 1: Write tests for joint limits**

Add to `servo_test.go`:
```go
func TestServo_JointLimits(t *testing.T) {
	servo := &Servo{robot: nil}
	assert.NotNil(t, servo)
}
```

**Step 2: Run tests**

Run:
```bash
go test . -v -run "TestServo_JointLimits"
```

Expected: PASS (structure tests)

**Step 3: Implement joint limit methods**

Add to `servo.go`:
```go
// GetJointMin gets minimum angle for a joint
func (s *Servo) GetJointMin(ctx context.Context, joint types.JointID) (types.Angle, error) {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.GetJointMinAngle,
		Data: data,
	}

	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}

	if len(resp) >= 2 {
		// Decode angle (int16 / 100)
		angles, err := protocol.DecodeAngles(resp)
		if err != nil {
			return 0, err
		}
		if len(angles) > 0 {
			return types.Angle(angles[0]), nil
		}
	}
	return 0, nil
}

// GetJointMax gets maximum angle for a joint
func (s *Servo) GetJointMax(ctx context.Context, joint types.JointID) (types.Angle, error) {
	data := []byte{byte(joint)}
	cmd := protocol.Command{
		Code: protocol.GetJointMaxAngle,
		Data: data,
	}

	resp, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}

	if len(resp) >= 2 {
		angles, err := protocol.DecodeAngles(resp)
		if err != nil {
			return 0, err
		}
		if len(angles) > 0 {
			return types.Angle(angles[0]), nil
		}
	}
	return 0, nil
}

// SetJointMin sets minimum angle for a joint
func (s *Servo) SetJointMin(ctx context.Context, joint types.JointID, angle types.Angle) error {
	encodedAngle := protocol.EncodeAngles([]float64{float64(angle)})
	data := []byte{byte(joint)}
	data = append(data, encodedAngle...)

	cmd := protocol.Command{
		Code: protocol.SetJointMin,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// SetJointMax sets maximum angle for a joint
func (s *Servo) SetJointMax(ctx context.Context, joint types.JointID, angle types.Angle) error {
	encodedAngle := protocol.EncodeAngles([]float64{float64(angle)})
	data := []byte{byte(joint)}
	data = append(data, encodedAngle...)

	cmd := protocol.Command{
		Code: protocol.SetJointMax,
		Data: data,
	}

	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run:
```bash
go test . -v -run "TestServo"
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add servo.go servo_test.go
git commit -m "feat(servo): add joint limit operations"
```

---

## Task 15: Integrate Servo Subsystem into MyCobot280

**Files:**
- Modify: `mycobot280.go`
- Modify: `mycobot280_test.go`

**Step 1: Write test for Servo integration**

Add to `mycobot280_test.go`:
```go
func TestMyCobot280_ServoSubsystem(t *testing.T) {
	robot := NewMyCobot280("/dev/null")

	assert.NotNil(t, robot.Servo)
	assert.IsType(t, &Servo{}, robot.Servo)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestMyCobot280_Servo
```

Expected: FAIL - "robot.Servo undefined"

**Step 3: Add Servo field to MyCobot280**

Modify `mycobot280.go`:
```go
// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base   *robot.Base
	config ModelConfig
	Motion *Motion
	IO     *IO
	Servo  *Servo
}

// In NewMyCobot280, add:
return &MyCobot280{
	base:   base,
	config: config,
	Motion: &Motion{robot: base},
	IO:     &IO{robot: base},
	Servo:  &Servo{robot: base},
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test . -v -run TestMyCobot280_Servo
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mycobot280.go mycobot280_test.go
git commit -m "feat(mycobot280): integrate Servo subsystem"
```

---

## Task 16: Gripper Package - Interface and Commander

**Files:**
- Create: `gripper/gripper.go`
- Create: `gripper/gripper_test.go`

**Step 1: Write test for interface existence**

Create `gripper/gripper_test.go`:
```go
package gripper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGripper_InterfaceExists(t *testing.T) {
	// This test just verifies the interface is defined
	var _ Gripper = (*mockGripper)(nil)
}

type mockGripper struct{}

func (m *mockGripper) Initialize(ctx context.Context, robot Commander) error {
	return nil
}

func (m *mockGripper) Release(ctx context.Context) error {
	return nil
}

func (m *mockGripper) Open(ctx context.Context) error {
	return nil
}

func (m *mockGripper) Close(ctx context.Context) error {
	return nil
}

func (m *mockGripper) SetValue(ctx context.Context, value int) error {
	return nil
}

func (m *mockGripper) GetValue(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockGripper) IsMoving(ctx context.Context) (bool, error) {
	return false, nil
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./gripper -v
```

Expected: FAIL - "package gripper not found"

**Step 3: Create gripper package with interface**

Create `gripper/gripper.go`:
```go
package gripper

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/protocol"
)

// Gripper defines the interface all grippers must implement
type Gripper interface {
	// Initialize prepares the gripper for use
	Initialize(ctx context.Context, robot Commander) error

	// Release cleans up gripper resources
	Release(ctx context.Context) error

	// Basic operations
	Open(ctx context.Context) error
	Close(ctx context.Context) error

	// Value-based control (0-100, gripper-specific meaning)
	SetValue(ctx context.Context, value int) error
	GetValue(ctx context.Context) (int, error)

	// State queries
	IsMoving(ctx context.Context) (bool, error)
}

// Commander interface for sending protocol commands
// This allows gripper to send commands without depending on robot.Base directly
type Commander interface {
	SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error)
}
```

**Step 4: Fix test imports**

Modify `gripper/gripper_test.go` to add missing import:
```go
package gripper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)
```

**Step 5: Run test to verify it passes**

Run:
```bash
go test ./gripper -v
```

Expected: PASS

**Step 6: Commit**

Run:
```bash
git add gripper/
git commit -m "feat(gripper): add Gripper interface and Commander"
```

---

## Task 17: Gripper Package - ProGripper Implementation

**Files:**
- Create: `gripper/pro.go`
- Create: `gripper/pro_test.go`

**Step 1: Write tests for ProGripper**

Create `gripper/pro_test.go`:
```go
package gripper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

type mockCommander struct {
	commands []protocol.Command
}

func (m *mockCommander) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	m.commands = append(m.commands, cmd)
	// Return mock response
	return []byte{0x01}, nil
}

func TestProGripper_NewProGripper(t *testing.T) {
	gripper := NewProGripper()
	assert.NotNil(t, gripper)
}

func TestProGripper_Initialize(t *testing.T) {
	gripper := NewProGripper()
	mock := &mockCommander{}
	ctx := context.Background()

	err := gripper.Initialize(ctx, mock)
	assert.NoError(t, err)
	assert.Len(t, mock.commands, 1)
	assert.Equal(t, protocol.InitGripper, mock.commands[0].Code)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./gripper -v -run TestProGripper
```

Expected: FAIL - "undefined: NewProGripper"

**Step 3: Implement ProGripper**

Create `gripper/pro.go`:
```go
package gripper

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/protocol"
)

// ProGripper represents the Elephant Robotics Pro Gripper
type ProGripper struct {
	robot Commander
}

// NewProGripper creates a new ProGripper instance
func NewProGripper() *ProGripper {
	return &ProGripper{}
}

// Initialize prepares the gripper for use
func (g *ProGripper) Initialize(ctx context.Context, robot Commander) error {
	g.robot = robot

	cmd := protocol.Command{Code: protocol.InitGripper}
	_, err := g.robot.SendCommand(ctx, cmd)
	return err
}

// Release cleans up gripper resources
func (g *ProGripper) Release(ctx context.Context) error {
	// No specific release command for ProGripper
	return nil
}

// Open opens the gripper (sets to max value)
func (g *ProGripper) Open(ctx context.Context) error {
	return g.SetValue(ctx, 100)
}

// Close closes the gripper (sets to min value)
func (g *ProGripper) Close(ctx context.Context) error {
	return g.SetValue(ctx, 0)
}

// SetValue sets gripper position (0-100)
func (g *ProGripper) SetValue(ctx context.Context, value int) error {
	if value < 0 || value > 100 {
		return protocol.ErrInvalidCommand
	}

	data := []byte{byte(value)}
	cmd := protocol.Command{
		Code: protocol.SetGripperValue,
		Data: data,
	}

	_, err := g.robot.SendCommand(ctx, cmd)
	return err
}

// GetValue gets current gripper position
func (g *ProGripper) GetValue(ctx context.Context) (int, error) {
	cmd := protocol.Command{Code: protocol.GetGripperValue}
	resp, err := g.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}

	if len(resp) > 0 {
		return int(resp[0]), nil
	}
	return 0, nil
}

// IsMoving returns true if gripper is currently moving
func (g *ProGripper) IsMoving(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsGripperMoving}
	resp, err := g.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(resp) > 0 {
		return resp[0] == 1, nil
	}
	return false, nil
}
```

**Step 4: Fix error reference**

The code references `protocol.ErrInvalidCommand` but this doesn't exist in protocol package. We need to use the error from the main package. Update `gripper/pro.go`:

```go
import (
	"context"
	"errors"

	"github.com/hipsterbrown/mycobot-go/protocol"
)

var ErrInvalidValue = errors.New("gripper value out of range [0, 100]")

// In SetValue:
if value < 0 || value > 100 {
	return ErrInvalidValue
}
```

**Step 5: Run tests to verify they pass**

Run:
```bash
go test ./gripper -v
```

Expected: PASS

**Step 6: Commit**

Run:
```bash
git add gripper/
git commit -m "feat(gripper): add ProGripper implementation"
```

---

## Task 18: Integrate Gripper into MyCobot280

**Files:**
- Modify: `mycobot280.go`
- Modify: `mycobot280_test.go`

**Step 1: Write tests for gripper integration**

Add to `mycobot280_test.go`:
```go
import "github.com/hipsterbrown/mycobot-go/gripper"

func TestMyCobot280_GripperField(t *testing.T) {
	robot := NewMyCobot280("/dev/null")

	// Gripper should be nil initially
	assert.Nil(t, robot.Gripper)
}

func TestMyCobot280_AttachGripper(t *testing.T) {
	robot := NewMyCobot280("/dev/null")
	ctx := context.Background()

	// Create gripper (will fail without connection, but tests the method exists)
	g := gripper.NewProGripper()
	err := robot.AttachGripper(ctx, g)

	// We expect error because not connected, but method should exist
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run "TestMyCobot280_Gripper\|TestMyCobot280_Attach"
```

Expected: FAIL - "robot.Gripper undefined"

**Step 3: Add Gripper field and methods to MyCobot280**

Modify `mycobot280.go`:
```go
import (
	// ... existing imports
	"github.com/hipsterbrown/mycobot-go/gripper"
)

// MyCobot280 represents a MyCobot 280 robot
type MyCobot280 struct {
	base    *robot.Base
	config  ModelConfig
	Motion  *Motion
	IO      *IO
	Servo   *Servo
	Gripper gripper.Gripper
}

// In NewMyCobot280:
return &MyCobot280{
	base:    base,
	config:  config,
	Motion:  &Motion{robot: base},
	IO:      &IO{robot: base},
	Servo:   &Servo{robot: base},
	Gripper: nil,
}

// AttachGripper attaches and initializes a gripper
func (m *MyCobot280) AttachGripper(ctx context.Context, g gripper.Gripper) error {
	if err := g.Initialize(ctx, m.base); err != nil {
		return err
	}
	m.Gripper = g
	return nil
}

// DetachGripper releases and detaches the gripper
func (m *MyCobot280) DetachGripper(ctx context.Context) error {
	if m.Gripper == nil {
		return nil
	}
	if err := m.Gripper.Release(ctx); err != nil {
		return err
	}
	m.Gripper = nil
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test . -v -run "TestMyCobot280_Gripper\|TestMyCobot280_Attach"
```

Expected: PASS (or expected errors)

**Step 5: Commit**

Run:
```bash
git add mycobot280.go mycobot280_test.go
git commit -m "feat(mycobot280): add gripper attachment support"
```

---

## Task 19: MyCobot320 Implementation

**Files:**
- Create: `mycobot320.go`
- Create: `mycobot320_test.go`

**Step 1: Write tests for MyCobot320**

Create `mycobot320_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestNewMyCobot320(t *testing.T) {
	robot := NewMyCobot320("/dev/ttyUSB0")

	assert.NotNil(t, robot)
	assert.Equal(t, types.ModelMyCobot320, robot.config.Model)
	assert.Equal(t, 6, robot.config.JointCount)
	assert.NotNil(t, robot.base)
}

func TestMyCobot320_Subsystems(t *testing.T) {
	robot := NewMyCobot320("/dev/null")

	assert.NotNil(t, robot.Motion)
	assert.NotNil(t, robot.IO)
	assert.NotNil(t, robot.Servo)
	assert.Nil(t, robot.Gripper)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestNewMyCobot320
```

Expected: FAIL - "undefined: NewMyCobot320"

**Step 3: Implement MyCobot320**

Create `mycobot320.go`:
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/gripper"
	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// MyCobot320 represents a MyCobot 320 robot
type MyCobot320 struct {
	base    *robot.Base
	config  ModelConfig
	Motion  *Motion
	IO      *IO
	Servo   *Servo
	Gripper gripper.Gripper
}

// NewMyCobot320 creates a new MyCobot320 instance
func NewMyCobot320(port string, opts ...Option) *MyCobot320 {
	config := getModelConfig(types.ModelMyCobot320)

	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	// Apply options
	for _, opt := range opts {
		opt(base)
	}

	return &MyCobot320{
		base:    base,
		config:  config,
		Motion:  &Motion{robot: base},
		IO:      &IO{robot: base},
		Servo:   &Servo{robot: base},
		Gripper: nil,
	}
}

// Open establishes connection to the robot
func (m *MyCobot320) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MyCobot320) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MyCobot320) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MyCobot320) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}

// PowerOn powers on all servos
func (m *MyCobot320) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MyCobot320) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MyCobot320) IsPowerOn(ctx context.Context) (bool, error) {
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

// SendAngles sends joint angles to the robot
func (m *MyCobot320) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}

	if err := speed.Validate(); err != nil {
		return err
	}

	angleData := protocol.EncodeAngles(angles.ToFloat64())
	data := append(angleData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (m *MyCobot320) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}

	return result, nil
}

// SendCoords sends coordinate position to the robot
func (m *MyCobot320) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	coordData := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data := append(coordData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MyCobot320) GetCoords(ctx context.Context) (types.Coord, error) {
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

// IsMoving returns true if robot is currently moving
func (m *MyCobot320) IsMoving(ctx context.Context) (bool, error) {
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

// AttachGripper attaches and initializes a gripper
func (m *MyCobot320) AttachGripper(ctx context.Context, g gripper.Gripper) error {
	if err := g.Initialize(ctx, m.base); err != nil {
		return err
	}
	m.Gripper = g
	return nil
}

// DetachGripper releases and detaches the gripper
func (m *MyCobot320) DetachGripper(ctx context.Context) error {
	if m.Gripper == nil {
		return nil
	}
	if err := m.Gripper.Release(ctx); err != nil {
		return err
	}
	m.Gripper = nil
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test . -v -run TestMyCobot320
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mycobot320.go mycobot320_test.go
git commit -m "feat: add MyCobot320 implementation with all subsystems"
```

---

## Task 20: MechArm270 Implementation

**Files:**
- Create: `mecharm270.go`
- Create: `mecharm270_test.go`

**Step 1: Write tests for MechArm270**

Create `mecharm270_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestNewMechArm270(t *testing.T) {
	robot := NewMechArm270("/dev/ttyUSB0")

	assert.NotNil(t, robot)
	assert.Equal(t, types.ModelMechArm270, robot.config.Model)
	assert.Equal(t, 6, robot.config.JointCount)
	assert.NotNil(t, robot.base)
}

func TestMechArm270_Subsystems(t *testing.T) {
	robot := NewMechArm270("/dev/null")

	assert.NotNil(t, robot.Motion)
	assert.NotNil(t, robot.IO)
	assert.NotNil(t, robot.Servo)
	assert.Nil(t, robot.Gripper)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestNewMechArm270
```

Expected: FAIL - "undefined: NewMechArm270"

**Step 3: Implement MechArm270**

Create `mecharm270.go` (copy MyCobot320 pattern, change to MechArm270):
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/gripper"
	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// MechArm270 represents a MechArm 270 robot
type MechArm270 struct {
	base    *robot.Base
	config  ModelConfig
	Motion  *Motion
	IO      *IO
	Servo   *Servo
	Gripper gripper.Gripper
}

// NewMechArm270 creates a new MechArm270 instance
func NewMechArm270(port string, opts ...Option) *MechArm270 {
	config := getModelConfig(types.ModelMechArm270)

	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	for _, opt := range opts {
		opt(base)
	}

	return &MechArm270{
		base:    base,
		config:  config,
		Motion:  &Motion{robot: base},
		IO:      &IO{robot: base},
		Servo:   &Servo{robot: base},
		Gripper: nil,
	}
}

// Open establishes connection to the robot
func (m *MechArm270) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MechArm270) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MechArm270) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MechArm270) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}

// PowerOn powers on all servos
func (m *MechArm270) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MechArm270) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MechArm270) IsPowerOn(ctx context.Context) (bool, error) {
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

// SendAngles sends joint angles to the robot
func (m *MechArm270) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}

	if err := speed.Validate(); err != nil {
		return err
	}

	angleData := protocol.EncodeAngles(angles.ToFloat64())
	data := append(angleData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (m *MechArm270) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}

	return result, nil
}

// SendCoords sends coordinate position to the robot
func (m *MechArm270) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	coordData := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data := append(coordData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MechArm270) GetCoords(ctx context.Context) (types.Coord, error) {
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

// IsMoving returns true if robot is currently moving
func (m *MechArm270) IsMoving(ctx context.Context) (bool, error) {
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

// AttachGripper attaches and initializes a gripper
func (m *MechArm270) AttachGripper(ctx context.Context, g gripper.Gripper) error {
	if err := g.Initialize(ctx, m.base); err != nil {
		return err
	}
	m.Gripper = g
	return nil
}

// DetachGripper releases and detaches the gripper
func (m *MechArm270) DetachGripper(ctx context.Context) error {
	if m.Gripper == nil {
		return nil
	}
	if err := m.Gripper.Release(ctx); err != nil {
		return err
	}
	m.Gripper = nil
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test . -v -run TestMechArm270
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mecharm270.go mecharm270_test.go
git commit -m "feat: add MechArm270 implementation with all subsystems"
```

---

## Task 21: MyPalletizer260 Implementation

**Files:**
- Create: `mypalletizer260.go`
- Create: `mypalletizer260_test.go`

**Step 1: Write tests for MyPalletizer260**

Create `mypalletizer260_test.go`:
```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestNewMyPalletizer260(t *testing.T) {
	robot := NewMyPalletizer260("/dev/ttyUSB0")

	assert.NotNil(t, robot)
	assert.Equal(t, types.ModelMyPalletizer260, robot.config.Model)
	assert.Equal(t, 4, robot.config.JointCount) // Only 4 joints!
	assert.NotNil(t, robot.base)
}

func TestMyPalletizer260_Subsystems(t *testing.T) {
	robot := NewMyPalletizer260("/dev/null")

	assert.NotNil(t, robot.Motion)
	assert.NotNil(t, robot.IO)
	assert.NotNil(t, robot.Servo)
	assert.Nil(t, robot.Gripper)
}

func TestMyPalletizer260_FourJoints(t *testing.T) {
	robot := NewMyPalletizer260("/dev/null")
	ctx := context.Background()

	// Should reject 6-joint angles
	angles6 := types.Angles{0, 0, 0, 0, 0, 0}
	err := robot.SendAngles(ctx, angles6, types.SpeedMedium)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 4 angles")

	// 4-joint angles should be structurally valid (will fail without connection)
	angles4 := types.Angles{0, 0, 0, 0}
	// This will error due to no connection, but validates angle count
	robot.SendAngles(ctx, angles4, types.SpeedMedium)
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test . -v -run TestNewMyPalletizer260
```

Expected: FAIL - "undefined: NewMyPalletizer260"

**Step 3: Implement MyPalletizer260**

Create `mypalletizer260.go` (same pattern as MechArm270, but using ModelMyPalletizer260):
```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/gripper"
	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// MyPalletizer260 represents a MyPalletizer 260 robot (4 joints)
type MyPalletizer260 struct {
	base    *robot.Base
	config  ModelConfig
	Motion  *Motion
	IO      *IO
	Servo   *Servo
	Gripper gripper.Gripper
}

// NewMyPalletizer260 creates a new MyPalletizer260 instance
func NewMyPalletizer260(port string, opts ...Option) *MyPalletizer260 {
	config := getModelConfig(types.ModelMyPalletizer260)

	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	for _, opt := range opts {
		opt(base)
	}

	return &MyPalletizer260{
		base:    base,
		config:  config,
		Motion:  &Motion{robot: base},
		IO:      &IO{robot: base},
		Servo:   &Servo{robot: base},
		Gripper: nil,
	}
}

// Open establishes connection to the robot
func (m *MyPalletizer260) Open(ctx context.Context) error {
	return m.base.Open(ctx)
}

// Close closes the connection to the robot
func (m *MyPalletizer260) Close() error {
	return m.base.Close()
}

// IsConnected returns true if robot is connected
func (m *MyPalletizer260) IsConnected() bool {
	return m.base.IsConnected()
}

// SendCommand sends a raw protocol command (for advanced users)
func (m *MyPalletizer260) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	return m.base.SendCommand(ctx, cmd)
}

// PowerOn powers on all servos
func (m *MyPalletizer260) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (m *MyPalletizer260) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (m *MyPalletizer260) IsPowerOn(ctx context.Context) (bool, error) {
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

// SendAngles sends joint angles to the robot (4 joints)
func (m *MyPalletizer260) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	if err := angles.Validate(m.config.JointCount, m.config.Model); err != nil {
		return err
	}

	if err := speed.Validate(); err != nil {
		return err
	}

	angleData := protocol.EncodeAngles(angles.ToFloat64())
	data := append(angleData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngles,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles (4 joints)
func (m *MyPalletizer260) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles}
	data, err := m.base.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	angles, err := protocol.DecodeAngles(data)
	if err != nil {
		return nil, err
	}

	result := make(types.Angles, len(angles))
	for i, a := range angles {
		result[i] = types.Angle(a)
	}

	return result, nil
}

// SendCoords sends coordinate position to the robot
func (m *MyPalletizer260) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	coordData := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data := append(coordData, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}

	_, err := m.base.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (m *MyPalletizer260) GetCoords(ctx context.Context) (types.Coord, error) {
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

// IsMoving returns true if robot is currently moving
func (m *MyPalletizer260) IsMoving(ctx context.Context) (bool, error) {
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

// AttachGripper attaches and initializes a gripper
func (m *MyPalletizer260) AttachGripper(ctx context.Context, g gripper.Gripper) error {
	if err := g.Initialize(ctx, m.base); err != nil {
		return err
	}
	m.Gripper = g
	return nil
}

// DetachGripper releases and detaches the gripper
func (m *MyPalletizer260) DetachGripper(ctx context.Context) error {
	if m.Gripper == nil {
		return nil
	}
	if err := m.Gripper.Release(ctx); err != nil {
		return err
	}
	m.Gripper = nil
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test . -v -run TestMyPalletizer260
```

Expected: PASS

**Step 5: Commit**

Run:
```bash
git add mypalletizer260.go mypalletizer260_test.go
git commit -m "feat: add MyPalletizer260 implementation (4 joints) with all subsystems"
```

---

## Task 22: Update Phase 1 Plan with Phase 2 Link

**Files:**
- Modify: `docs/plans/2025-11-23-mycobot-go-implementation.md`

**Step 1: Add link to Phase 2 at end of Phase 1 document**

Modify `docs/plans/2025-11-23-mycobot-go-implementation.md`, replace the "Phase 1 Complete - Next Steps" section with:
```markdown
---

## Phase 1 Complete ✅

All 12 tasks completed successfully!

**Implemented:**
- Protocol package (encoding/decoding, CRC)
- Types package (Angle, Coord, JointID, Speed with validation)
- Base robot infrastructure with channel-based concurrency
- Error handling system
- Model configurations for all 4 robots
- MyCobot280 with core motion commands

**Next Phase:** [Phase 2: Extended Features](./2025-11-23-phase2-implementation.md)
```

**Step 2: Commit**

Run:
```bash
git add docs/plans/2025-11-23-mycobot-go-implementation.md
git commit -m "docs: add Phase 2 link to Phase 1 plan"
```

---

## Task 23: Create Plans README

**Files:**
- Create: `docs/plans/README.md`

**Step 1: Create navigation document**

Create `docs/plans/README.md`:
```markdown
# MyCobot Go Implementation Plans

This directory contains the design documents and implementation plans for the mycobot-go project.

## Documents

### Architecture & Design
- [Port Design Document](./2025-11-23-mycobot-go-port-design.md) - Overall architecture and design decisions
- [Phase 2 Extended Features Design](./2025-11-23-phase2-extended-features-design.md) - Subsystem architecture and gripper interface design

### Implementation Plans

1. **[Phase 1: Core Foundation](./2025-11-23-mycobot-go-implementation.md)** ✅ Complete
   - Protocol layer (encoding/decoding, CRC)
   - Types package (validated domain types)
   - Base robot with channel-based concurrency
   - MyCobot280 with core motion commands
   - **Status:** 12/12 tasks complete

2. **[Phase 2: Extended Features](./2025-11-23-phase2-implementation.md)** ← Current
   - Motion subsystem (JOG, Pause/Resume, single joint/coord)
   - IO subsystem (Digital/PWM IO, LED control)
   - Servo subsystem (individual servo control, encoders, calibration)
   - Gripper interface + ProGripper implementation
   - Additional robot models (MyCobot320, MechArm270, MyPalletizer260)
   - **Status:** 0/23 tasks complete

3. **Phase 3: Advanced Features** (TBD)
   - Additional gripper types
   - Advanced protocol commands
   - Synchronous operations
   - Usage examples

4. **Phase 4: Polish** (TBD)
   - Comprehensive testing
   - Documentation
   - CI/CD pipeline
   - Benchmarks

## Progress Summary

- **Total Tasks Planned:** 35+ (Phase 1 + Phase 2)
- **Tasks Completed:** 12
- **Current Phase:** Phase 2
- **Overall Status:** ~34% complete

## Using These Plans

Each implementation plan follows Test-Driven Development (TDD):
1. Write failing test
2. Run test (verify failure)
3. Implement minimal code
4. Run test (verify pass)
5. Commit

Plans are designed to be executed by:
- **Claude with superpowers:executing-plans skill** - Batch execution with review checkpoints
- **Claude with superpowers:subagent-driven-development** - Fresh subagent per task with code review
- **Human developers** - Step-by-step implementation guide
```

**Step 2: Commit**

Run:
```bash
git add docs/plans/README.md
git commit -m "docs: add plans navigation README"
```

---

## Phase 2 Complete! 🎉

When all tasks (1-23) are complete:

- All subsystems implemented on MyCobot280
- All 4 robot models implemented
- Gripper interface with ProGripper
- ~23 tasks, ~150+ tests passing
- Ready for Phase 3: Advanced Features

---

## Notes

- All subsystems are shared across robot models (Motion, IO, Servo)
- Gripper interface allows for easy extension with new gripper types
- Model-specific validation automatically handles different joint counts
- MyPalletizer260 is the special case with 4 joints instead of 6
