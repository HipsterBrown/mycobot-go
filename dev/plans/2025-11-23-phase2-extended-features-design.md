# Phase 2: Extended Features - Design Document

**Date:** 2025-11-23
**Status:** Approved
**Phase:** 2 of 4

---

## Overview

**Goal:** Extend MyCobot280 with Motion, IO, Servo subsystems and gripper support, then replicate the pattern to other robot models (MyCobot320, MechArm270, MyPalletizer260).

**Prerequisites:** Phase 1 complete - core foundation with protocol layer, types, base robot, and MyCobot280 with basic motion commands.

**Outcome:** Full-featured robot control for all four supported models with subsystem-based API organization.

---

## Architecture

### Subsystem-Based Design

Each robot model has subsystems attached as fields:

```go
type MyCobot280 struct {
    base    *robot.Base
    config  ModelConfig
    Motion  *Motion   // Motion control subsystem
    IO      *IO       // Digital/PWM IO subsystem
    Servo   *Servo    // Individual servo control subsystem
    Gripper gripper.Gripper  // Optional, attached at runtime
}
```

### Subsystem Implementation Pattern

Each subsystem holds a reference to the base robot for sending commands:

```go
type Motion struct {
    robot *robot.Base
}

type IO struct {
    robot *robot.Base
}

type Servo struct {
    robot *robot.Base
}
```

### Initialization Flow

```go
func NewMyCobot280(port string, opts ...Option) *MyCobot280 {
    config := getModelConfig(types.ModelMyCobot280)
    base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

    // Apply options
    for _, opt := range opts {
        opt(base)
    }

    return &MyCobot280{
        base:    base,
        config:  config,
        Motion:  &Motion{robot: base},
        IO:      &IO{robot: base},
        Servo:   &Servo{robot: base},
        Gripper: nil,  // Attached later via AttachGripper()
    }
}
```

---

## Motion Subsystem

**Package:** `mycobot` (same package as robot)
**File:** `motion.go`

### Responsibilities
- JOG operations (incremental movement)
- Movement control (Pause/Resume/Stop)
- Position queries
- Single joint/coordinate commands

### API

```go
type Motion struct {
    robot *robot.Base
}

// JOG operations - incremental movement
func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction int, speed types.Speed) error
func (m *Motion) JogCoord(ctx context.Context, axis CoordAxis, direction int, speed types.Speed) error
func (m *Motion) JogStop(ctx context.Context) error

// Movement control
func (m *Motion) Pause(ctx context.Context) error
func (m *Motion) Resume(ctx context.Context) error
func (m *Motion) Stop(ctx context.Context) error
func (m *Motion) IsPaused(ctx context.Context) (bool, error)

// Position queries
func (m *Motion) IsInPosition(ctx context.Context, target types.Coord, tolerance float64) (bool, error)

// Single joint/coord commands (complementing SendAngles/SendCoords)
func (m *Motion) SendAngle(ctx context.Context, joint types.JointID, angle types.Angle, speed types.Speed) error
func (m *Motion) SendCoord(ctx context.Context, axis CoordAxis, value float64, speed types.Speed) error
```

### New Types

```go
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
```

### Usage Example

```go
robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// JOG joint 1 forward
robot.Motion.JogAngle(ctx, types.Joint1, 1, types.SpeedSlow)
time.Sleep(1 * time.Second)
robot.Motion.JogStop(ctx)

// Move single joint to position
robot.Motion.SendAngle(ctx, types.Joint2, types.Angle(45.0), types.SpeedMedium)

// Pause/resume movement
robot.SendAngles(ctx, angles, types.SpeedSlow)
time.Sleep(500 * time.Millisecond)
robot.Motion.Pause(ctx)
time.Sleep(1 * time.Second)
robot.Motion.Resume(ctx)
```

---

## IO Subsystem

**Package:** `mycobot`
**File:** `io.go`

### Responsibilities
- Digital I/O (read/write pins)
- PWM output
- LED color control (Atom LED)

### API

```go
type IO struct {
    robot *robot.Base
}

// Digital I/O
func (io *IO) SetDigitalOutput(ctx context.Context, pin int, value bool) error
func (io *IO) GetDigitalInput(ctx context.Context, pin int) (bool, error)

// PWM output
func (io *IO) SetPWMOutput(ctx context.Context, pin int, value int) error

// LED color control (Atom LED on end effector)
func (io *IO) SetColor(ctx context.Context, r, g, b byte) error
```

### New Types

```go
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
```

### Usage Example

```go
robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// Set LED to green
robot.IO.SetColor(ctx, 0, 255, 0)

// Digital output
robot.IO.SetDigitalOutput(ctx, 1, true)

// Read digital input
pressed, _ := robot.IO.GetDigitalInput(ctx, 2)

// PWM output (0-255)
robot.IO.SetPWMOutput(ctx, 3, 128)
```

**Note:** Pin numbers and capabilities are model-specific but the API is consistent across models.

---

## Servo Subsystem

**Package:** `mycobot`
**File:** `servo.go`

### Responsibilities
- Individual servo control
- Encoder operations
- Servo calibration
- Joint limit management

### API

```go
type Servo struct {
    robot *robot.Base
}

// Servo state control
func (s *Servo) ReleaseServo(ctx context.Context, joint types.JointID) error
func (s *Servo) FocusServo(ctx context.Context, joint types.JointID) error
func (s *Servo) IsServoEnabled(ctx context.Context, joint types.JointID) (bool, error)

// Encoder operations
func (s *Servo) GetEncoder(ctx context.Context, joint types.JointID) (int, error)
func (s *Servo) SetEncoder(ctx context.Context, joint types.JointID, value int) error
func (s *Servo) GetEncoders(ctx context.Context) ([]int, error)
func (s *Servo) SetEncoders(ctx context.Context, encoders []int) error

// Servo data and calibration
func (s *Servo) GetServoData(ctx context.Context, joint types.JointID) (int, error)
func (s *Servo) SetServoData(ctx context.Context, joint types.JointID, value int) error
func (s *Servo) SetServoCalibration(ctx context.Context, joint types.JointID, value int) error

// Joint limits
func (s *Servo) GetJointMin(ctx context.Context, joint types.JointID) (types.Angle, error)
func (s *Servo) GetJointMax(ctx context.Context, joint types.JointID) (types.Angle, error)
func (s *Servo) SetJointMin(ctx context.Context, joint types.JointID, angle types.Angle) error
func (s *Servo) SetJointMax(ctx context.Context, joint types.JointID, angle types.Angle) error
```

### Usage Example

```go
robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// Release a specific servo (makes it freely movable)
robot.Servo.ReleaseServo(ctx, types.Joint3)

// Get encoder value
encoder, _ := robot.Servo.GetEncoder(ctx, types.Joint1)

// Set joint limits
robot.Servo.SetJointMin(ctx, types.Joint2, types.Angle(-90))
robot.Servo.SetJointMax(ctx, types.Joint2, types.Angle(90))

// Get all encoders
encoders, _ := robot.Servo.GetEncoders(ctx)
```

**Note:** Encoder values are servo-specific integers, not angles. Calibration is advanced functionality for fine-tuning servo behavior.

---

## Gripper Interface

**Package:** `gripper` (new package)
**Files:** `gripper/gripper.go`, `gripper/pro.go`

### Responsibilities
- Abstract gripper operations
- Allow hot-swapping different gripper types
- Provide consistent API across gripper models

### Interface Design

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

### ProGripper Implementation

```go
package gripper

// ProGripper represents the Elephant Robotics Pro Gripper
type ProGripper struct {
    robot Commander
}

// NewProGripper creates a new ProGripper instance
func NewProGripper() *ProGripper {
    return &ProGripper{}
}

// Initialize prepares the gripper
func (g *ProGripper) Initialize(ctx context.Context, robot Commander) error {
    g.robot = robot
    // Send initialization command
    cmd := protocol.Command{Code: protocol.InitGripper}
    _, err := g.robot.SendCommand(ctx, cmd)
    return err
}

// Implements all other Gripper interface methods...
```

### Robot Integration

```go
// In mycobot280.go

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

### Usage Example

```go
robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// Attach gripper
proGripper := gripper.NewProGripper()
robot.AttachGripper(ctx, proGripper)

// Use gripper
robot.Gripper.Open(ctx)
time.Sleep(1 * time.Second)
robot.Gripper.Close(ctx)

// Set specific value (0-100)
robot.Gripper.SetValue(ctx, 50)

// Check if moving
moving, _ := robot.Gripper.IsMoving(ctx)

// Detach when done
robot.DetachGripper(ctx)
```

---

## Additional Robot Models

### Goal
Replicate the MyCobot280 pattern for:
- MyCobot320 (6 joints)
- MechArm270 (6 joints)
- MyPalletizer260 (4 joints)

### Implementation Pattern

Each model gets its own file with identical structure to MyCobot280:

```go
// mycobot320.go
type MyCobot320 struct {
    base    *robot.Base
    config  ModelConfig
    Motion  *Motion
    IO      *IO
    Servo   *Servo
    Gripper gripper.Gripper
}

func NewMyCobot320(port string, opts ...Option) *MyCobot320 {
    config := getModelConfig(types.ModelMyCobot320)
    base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

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

// All same methods: Open, Close, PowerOn, SendAngles, etc.
```

### Key Differences Per Model

**Model configurations** (already exist in `config.go` from Phase 1):
- MyCobot280: 6 joints, specific angle limits per joint
- MyCobot320: 6 joints, different angle limits
- MechArm270: 6 joints, similar to MyCobot280
- MyPalletizer260: **4 joints** (different joint count)

**Validation:** The existing `angles.Validate(m.config.JointCount, m.config.Model)` automatically handles model-specific differences.

**Subsystems:** All models share the same subsystem implementations (`motion.go`, `io.go`, `servo.go`). Only the robot wrapper changes.

### Files to Create

**MyCobot320:**
- `mycobot320.go`
- `mycobot320_test.go`

**MechArm270:**
- `mecharm270.go`
- `mecharm270_test.go`

**MyPalletizer260:**
- `mypalletizer260.go`
- `mypalletizer260_test.go`

### Pattern Benefits

1. **Code reuse:** Subsystems are shared across all models
2. **Consistency:** Same API for all models
3. **Type safety:** Each model is a distinct type
4. **Easy testing:** Test pattern once, replicate to other models

---

## Package Structure

After Phase 2, the package structure will be:

```
mycobot-go/
├── mycobot280.go          # MyCobot280 implementation
├── mycobot320.go          # MyCobot320 implementation
├── mecharm270.go          # MechArm270 implementation
├── mypalletizer260.go     # MyPalletizer260 implementation
├── motion.go              # Motion subsystem (shared)
├── io.go                  # IO subsystem (shared)
├── servo.go               # Servo subsystem (shared)
├── config.go              # Model configurations (Phase 1)
├── errors.go              # Error types (Phase 1)
├── robot.go               # Robot interface (Phase 1)
├── option.go              # Option pattern (Phase 1)
├── types/                 # Domain types (Phase 1)
│   ├── angle.go
│   ├── coord.go
│   ├── joint.go
│   ├── speed.go
│   └── model.go
├── protocol/              # Protocol layer (Phase 1)
│   ├── command.go
│   └── codes.go
├── internal/
│   ├── robot/             # Base robot (Phase 1)
│   │   └── base.go
│   └── errors/            # Internal errors (Phase 1)
│       └── errors.go
└── gripper/               # Gripper package (NEW)
    ├── gripper.go         # Interface + Commander
    └── pro.go             # ProGripper implementation
```

---

## Implementation Order

Following TDD and incremental development:

1. **Motion subsystem** on MyCobot280
2. **IO subsystem** on MyCobot280
3. **Servo subsystem** on MyCobot280
4. **Gripper interface + ProGripper**
5. **MyCobot320** (clone pattern)
6. **MechArm270** (clone pattern)
7. **MyPalletizer260** (clone pattern with 4 joints)

This sequence allows:
- Testing each subsystem independently
- Establishing patterns before replication
- Early detection of design issues
- Incremental value delivery

---

## Testing Strategy

**Unit Tests:**
- Each subsystem tested independently
- Mock robot.Base for command validation
- Test validation logic (joint ranges, speed ranges)

**Integration Tests:**
- Require actual hardware
- Test full command flow with real robot
- Verify subsystem interactions

**Model Tests:**
- Each model has identical test pattern
- Validate model-specific joint counts and limits
- Test constructor and basic operations

---

## Success Criteria

Phase 2 complete when:

1. ✅ Motion subsystem implemented and tested
2. ✅ IO subsystem implemented and tested
3. ✅ Servo subsystem implemented and tested
4. ✅ Gripper interface defined
5. ✅ ProGripper implementation working
6. ✅ All four robot models implemented
7. ✅ All tests passing (unit + integration where possible)
8. ✅ Each model can attach/detach grippers
9. ✅ Documentation updated with examples

---

## Next Phase

**Phase 3: Advanced Features**
- Additional gripper types (Adaptive, Electric, Pneumatic)
- Advanced protocol commands
- Synchronous operations with timeout
- Usage examples and guides
