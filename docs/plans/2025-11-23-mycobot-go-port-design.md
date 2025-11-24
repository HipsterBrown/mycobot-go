# MyCobot Go Port Design

**Date**: 2025-11-23
**Status**: Approved
**Target Models**: MyCobot 280, MyCobot 320, MechArm 270, MyPalletizer 260

## Overview

Port the pymycobot Python library to Go, creating a thread-safe, idiomatic Go package for controlling Elephant Robotics myCobot series robotic arms and grippers over serial communication.

## Design Decisions

### 1. Package Structure (Modular Foundation)

Build a core framework supporting the four specified models initially, with clean extensibility for additional models:

```
mycobot-go/
├── protocol/              # Low-level serial protocol (optional for users)
│   ├── command.go        # Command encoding/decoding
│   ├── codes.go          # Protocol command codes
│   └── crc.go            # CRC calculation for newer models
├── types/                # Strongly-typed domain types
│   ├── angle.go          # Angle with validation
│   ├── coord.go          # Coordinate types
│   ├── joint.go          # Joint IDs (1-6)
│   ├── speed.go          # Speed (0-100) with validation
│   └── common.go         # Shared enums and constants
├── gripper/              # Gripper interfaces and implementations
│   ├── gripper.go        # Core Gripper interface
│   ├── pro.go            # ProGripper implementation
│   ├── adaptive.go       # Adaptive gripper with extended features
│   └── electric.go       # Basic electric gripper
├── internal/             # Non-exported shared code
│   ├── serial/           # Serial communication wrapper
│   └── queue/            # Command queue implementation
├── mycobot280.go         # MyCobot 280 implementation
├── mycobot320.go         # MyCobot 320 implementation
├── mecharm270.go         # MechArm 270 implementation
├── mypalletizer260.go    # MyPalletizer 260 implementation
├── robot.go              # Shared Robot interface and base types
└── examples/             # Usage examples
```

**Rationale**: Focused on immediate needs while providing clean extensibility for future robot models.

### 2. Concurrency Model (Channel-Based)

Single goroutine per robot instance owns the serial connection. Other goroutines send commands via channels.

**Benefits**:
- Naturally thread-safe without mutex ceremony
- Context support for timeouts/cancellation
- Matches physical reality (serial hardware is sequential)
- Easy to test and reason about

**Implementation**:
```go
type baseRobot struct {
    cmdChan   chan *command
    closeChan chan struct{}
    conn      *serial.Port
}

func (b *baseRobot) commandLoop() {
    for {
        select {
        case cmd := <-b.cmdChan:
            // Process command
        case <-b.closeChan:
            return
        }
    }
}
```

### 3. API Design (Hybrid Approach)

Core motion commands directly on robot, specialized features grouped:

```go
// Common operations - direct access
robot.SendAngles(ctx, angles, speed)
robot.GetCoords(ctx)
robot.PowerOn(ctx)

// Specialized subsystems - organized
robot.Motion.JogAngle(ctx, joint, direction, speed)
robot.IO.SetDigitalOutput(ctx, pin, value)
robot.Servo.Release(ctx, joint)
robot.Gripper.SetValue(ctx, value)
```

**Rationale**: Balances discoverability (common commands easy to find) with organization (specialized features logically grouped).

### 4. Error Handling (Explicit Errors)

Every fallible operation returns `(result, error)`:

```go
angles, err := robot.GetAngles(ctx)
if err != nil {
    return err
}
```

**Rationale**: Go-idiomatic, forces proper error handling, enables graceful failure recovery in robotics applications.

### 5. Protocol Layer (Exposed but Optional)

Separate `protocol` package for advanced users and debugging:

```go
// Normal use - high-level API
robot.SendAngles(ctx, angles, speed)

// Advanced - direct protocol access
import "github.com/hipsterbrown/mycobot-go/protocol"
cmd := protocol.NewCommand(protocol.SEND_ANGLES, data)
robot.SendCommand(ctx, cmd)
```

**Rationale**: Keeps public API clean while enabling power users to implement new commands or debug protocol issues.

### 6. Type Safety (Strongly Typed with Validation)

Custom types with compile-time safety and runtime validation:

```go
type JointID int
type Speed int // 0-100
type Angle float64 // validated range per joint

robot.SendAngle(ctx, Joint1, Angle(45.5), Speed(50))
```

**Rationale**: Catches errors early, self-documenting API, prevents common mistakes.

### 7. Gripper Architecture (Interface-Based)

Grippers are optional and can be attached/detached anytime:

```go
type Gripper interface {
    Initialize(ctx context.Context, robot Commander) error
    SetValue(ctx context.Context, value Value) error
    GetValue(ctx context.Context) (Value, error)
    GetStatus(ctx context.Context) (Status, error)
    IsMoving(ctx context.Context) (bool, error)
    Release(ctx context.Context) error
}

type TorqueGripper interface {
    Gripper
    SetTorque(ctx context.Context, torque Torque) error
    GetTorque(ctx context.Context) (Torque, error)
}

// Usage
robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// Robot works without gripper
robot.SendAngles(ctx, angles, speed)

// Attach gripper when needed
robot.AttachGripper(ctx, gripper.NewProGripper())
robot.Gripper.SetValue(ctx, 50)
```

**Rationale**: Matches physical reality (grippers are optional), supports multiple gripper types through interfaces, type-safe capabilities via interface composition.

## Core Components

### Types Package

Strongly-typed domain primitives:

```go
// JointID represents a robot joint (1-6)
type JointID int

const (
    Joint1 JointID = 1
    Joint2 JointID = 2
    // ...
)

// Speed (0-100) with validation
type Speed int

const (
    SpeedMin    Speed = 0
    SpeedSlow   Speed = 25
    SpeedMedium Speed = 50
    SpeedFast   Speed = 75
    SpeedMax    Speed = 100
)

func (s Speed) Validate() error {
    if s < 0 || s > 100 {
        return fmt.Errorf("speed %d out of range [0,100]", s)
    }
    return nil
}

// Angle with model-specific validation
type Angle float64

func (a Angle) ValidateForJoint(joint JointID, model Model) error {
    limits := getJointLimits(model, joint)
    if float64(a) < limits.Min || float64(a) > limits.Max {
        return fmt.Errorf("angle %.2f out of range", a)
    }
    return nil
}

// Coord represents 3D position + rotation
type Coord struct {
    X, Y, Z    float64  // Position in mm
    Rx, Ry, Rz float64  // Rotation in degrees
}

// Angles represents all joint angles
type Angles []Angle
```

### Robot Interface

Base interface all models implement:

```go
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

### Subsystems

Specialized features organized into logical groups:

**Motion Subsystem**:
```go
type Motion struct {
    robot *baseRobot
}

func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction Direction, speed types.Speed) error
func (m *Motion) JogCoord(ctx context.Context, axis Axis, direction Direction, speed types.Speed) error
func (m *Motion) JogStop(ctx context.Context) error
func (m *Motion) SyncSendAngles(ctx context.Context, angles types.Angles, speed types.Speed, timeout time.Duration) error
```

**IO Subsystem**:
```go
type IO struct {
    robot *baseRobot
}

func (i *IO) SetDigitalOutput(ctx context.Context, pin Pin, value PinState) error
func (i *IO) GetDigitalInput(ctx context.Context, pin Pin) (PinState, error)
func (i *IO) SetPWMOutput(ctx context.Context, pin Pin, value PWMValue) error
func (i *IO) SetPinMode(ctx context.Context, pin Pin, mode PinMode) error
```

**Servo Subsystem**:
```go
type Servo struct {
    robot *baseRobot
}

func (s *Servo) IsEnabled(ctx context.Context, joint types.JointID) (bool, error)
func (s *Servo) Release(ctx context.Context, joint types.JointID) error
func (s *Servo) Focus(ctx context.Context, joint types.JointID) error
func (s *Servo) SetData(ctx context.Context, joint types.JointID, data ServoData) error
func (s *Servo) GetEncoder(ctx context.Context, joint types.JointID) (int, error)
```

### Protocol Package

Low-level command encoding/decoding:

```go
// Command codes from Python ProtocolCode class
const (
    Header byte = 0xFE
    Footer byte = 0xFA

    PowerOn      byte = 0x10
    GetAngles    byte = 0x20
    SendAngles   byte = 0x22
    // ... all protocol codes
)

// Command represents a protocol command
type Command struct {
    Code   byte
    Data   []byte
    UseCRC bool  // Model-specific
}

func (c Command) Encode() ([]byte, error) {
    // Build packet: [Header Header Length Code Data... Footer/CRC]
    packet := []byte{Header, Header, byte(len(c.Data) + 2), c.Code}
    packet = append(packet, c.Data...)

    if c.UseCRC {
        packet = append(packet, calculateCRC(packet))
    } else {
        packet = append(packet, Footer)
    }

    return packet, nil
}

// Response parsing
func Decode(data []byte, useCRC bool) (*Response, error) {
    // Validate header, length, CRC/footer
    // Extract payload
}

// Helper functions
func EncodeAngles(angles []float64) []byte
func DecodeAngles(data []byte) ([]float64, error)
func EncodeInt16(value int) []byte
```

### Model Configurations

Each robot model has specific parameters:

```go
type ModelConfig struct {
    Model         Model
    JointCount    int
    JointLimits   []JointLimit
    UseCRC        bool
    DefaultBaud   int
    SupportedBaud []int
}

var modelConfigs = map[Model]ModelConfig{
    ModelMyCobot280: {
        Model:      ModelMyCobot280,
        JointCount: 6,
        JointLimits: []JointLimit{
            {MinAngle: -165, MaxAngle: 165}, // Joint 1
            {MinAngle: -165, MaxAngle: 165}, // Joint 2
            {MinAngle: -165, MaxAngle: 165}, // Joint 3
            {MinAngle: -165, MaxAngle: 165}, // Joint 4
            {MinAngle: -165, MaxAngle: 165}, // Joint 5
            {MinAngle: -175, MaxAngle: 175}, // Joint 6
        },
        UseCRC:        true,
        DefaultBaud:   115200,
        SupportedBaud: []int{115200, 1000000},
    },
    // ... other models
}
```

### Concrete Robot Implementation

```go
type MyCobot280 struct {
    *baseRobot
    config ModelConfig

    Motion  *Motion
    IO      *IO
    Servo   *Servo
    Gripper gripper.Gripper
}

func NewMyCobot280(port string, opts ...Option) *MyCobot280 {
    config := modelConfigs[ModelMyCobot280]

    base := &baseRobot{
        port:     port,
        baudrate: config.DefaultBaud,
        model:    config,
    }

    for _, opt := range opts {
        opt(base)
    }

    return &MyCobot280{
        baseRobot: base,
        config:    config,
        Motion:    &Motion{robot: base},
        IO:        &IO{robot: base},
        Servo:     &Servo{robot: base},
    }
}

// Gripper management
func (m *MyCobot280) AttachGripper(ctx context.Context, g gripper.Gripper) error {
    if err := g.Initialize(ctx, m.baseRobot); err != nil {
        return err
    }
    m.Gripper = g
    return nil
}

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

### Error Types

```go
var (
    // Connection errors
    ErrRobotClosed       = errors.New("robot connection closed")
    ErrNotConnected      = errors.New("robot not connected")
    ErrConnectionTimeout = errors.New("connection timeout")

    // Command errors
    ErrInvalidCommand    = errors.New("invalid command")
    ErrCommandTimeout    = errors.New("command timeout")
    ErrInvalidResponse   = errors.New("invalid response from robot")

    // Validation errors
    ErrInvalidJoint      = errors.New("invalid joint ID")
    ErrInvalidSpeed      = errors.New("speed out of range")
    ErrInvalidAngle      = errors.New("angle out of range")
    ErrInvalidCoordinate = errors.New("coordinate out of range")

    // Gripper errors
    ErrNoGripper         = errors.New("no gripper attached")
    ErrGripperNotSupported = errors.New("gripper operation not supported")

    // State errors
    ErrNotPowered        = errors.New("robot not powered on")
    ErrEmergencyStop     = errors.New("emergency stop active")
    ErrServoError        = errors.New("servo error detected")
)

type RobotError struct {
    Op    string
    Model string
    Err   error
}
```

## Channel-Based Concurrency Implementation

```go
type baseRobot struct {
    port     string
    baudrate int
    conn     *serial.Port

    // Command queue
    cmdChan   chan *command
    closeChan chan struct{}
    closeOnce sync.Once
    wg        sync.WaitGroup
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

func (b *baseRobot) Open(ctx context.Context) error {
    conn, err := serial.OpenPort(&serial.Config{
        Name: b.port,
        Baud: b.baudrate,
        ReadTimeout: 100 * time.Millisecond,
    })
    if err != nil {
        return err
    }

    b.conn = conn
    b.cmdChan = make(chan *command, 32)
    b.closeChan = make(chan struct{})

    b.wg.Add(1)
    go b.commandLoop()

    return nil
}

func (b *baseRobot) commandLoop() {
    defer b.wg.Done()
    defer b.conn.Close()

    for {
        select {
        case <-b.closeChan:
            return

        case cmd := <-b.cmdChan:
            if err := cmd.ctx.Err(); err != nil {
                cmd.response <- &response{err: err}
                continue
            }

            data, err := cmd.request.Encode()
            if err != nil {
                cmd.response <- &response{err: err}
                continue
            }

            b.conn.Write(data)
            respData, err := b.readResponse(cmd.ctx, cmd.request.Code)
            cmd.response <- &response{data: respData, err: err}
        }
    }
}

func (b *baseRobot) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
    responseChan := make(chan *response, 1)

    select {
    case b.cmdChan <- &command{ctx: ctx, request: cmd, response: responseChan}:
    case <-ctx.Done():
        return nil, ctx.Err()
    case <-b.closeChan:
        return nil, ErrRobotClosed
    }

    select {
    case resp := <-responseChan:
        return resp.data, resp.err
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (b *baseRobot) Close() error {
    b.closeOnce.Do(func() {
        close(b.closeChan)
    })
    b.wg.Wait()
    return nil
}
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/hipsterbrown/mycobot-go"
    "github.com/hipsterbrown/mycobot-go/types"
)

func main() {
    robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
    ctx := context.Background()

    if err := robot.Open(ctx); err != nil {
        log.Fatal(err)
    }
    defer robot.Close()

    // Power on
    robot.PowerOn(ctx)
    time.Sleep(2 * time.Second)

    // Get current position
    angles, _ := robot.GetAngles(ctx)
    log.Printf("Current angles: %v", angles)

    // Move to home
    homeAngles := types.Angles{0, 0, 0, 0, 0, 0}
    robot.SendAngles(ctx, homeAngles, types.SpeedMedium)

    // Wait for completion
    for {
        moving, _ := robot.IsMoving(ctx)
        if !moving {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }
}
```

### Gripper Usage

```go
import "github.com/hipsterbrown/mycobot-go/gripper"

robot := mycobot.NewMyCobot280("/dev/ttyUSB0")
robot.Open(ctx)

// Attach gripper
proGripper := gripper.NewProGripper()
robot.AttachGripper(ctx, proGripper)

// Open gripper
robot.Gripper.SetValue(ctx, gripper.Value(100))
time.Sleep(1 * time.Second)

// Close gripper
robot.Gripper.SetValue(ctx, gripper.Value(0))

// Detach when done
robot.DetachGripper(ctx)
```

### Concurrent Operations

```go
var wg sync.WaitGroup

wg.Add(2)
go func() {
    defer wg.Done()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    angles, _ := robot.GetAngles(ctx)
    log.Printf("Angles: %v", angles)
}()

go func() {
    defer wg.Done()
    coords, _ := robot.GetCoords(context.Background())
    log.Printf("Coords: %v", coords)
}()

wg.Wait()
```

### Advanced Protocol Usage

```go
import "github.com/hipsterbrown/mycobot-go/protocol"

// Send custom command not yet in high-level API
cmd := protocol.NewCommandWithData(0xAB, []byte{0x01, 0x02})
response, err := robot.SendCommand(ctx, cmd)
```

## Implementation Phases

### Phase 1: Core Foundation
- Protocol package (encoding/decoding, CRC)
- Types package (validated types)
- Base robot with channel-based concurrency
- Serial communication wrapper
- Error types
- MyCobot280 with core motion commands

### Phase 2: Extended Features
- Motion subsystem (JOG operations)
- IO subsystem
- Servo subsystem
- Basic gripper interface and ProGripper
- All four model implementations

### Phase 3: Advanced Features
- Advanced gripper implementations
- Synchronous operations with timeout
- Complete protocol command coverage
- Examples and documentation

### Phase 4: Polish
- Comprehensive tests
- Benchmarks
- Full documentation
- CI/CD pipeline

## Dependencies

**Runtime**:
- `go.bug.st/serial` - Serial port communication (pure Go, cross-platform)
- Standard library only otherwise

**Testing**:
- `github.com/stretchr/testify` - Test assertions

## Testing Strategy

**Unit Tests**: Mock serial communication and command sender:
```go
type mockCommander struct {
    commands []protocol.Command
    responses map[byte][]byte
}

func TestMyCobot280_SendAngles(t *testing.T) {
    robot := NewMyCobot280("/dev/null")
    mock := &mockCommander{responses: make(map[byte][]byte)}
    robot.baseRobot.commander = mock

    err := robot.SendAngles(ctx, angles, speed)
    assert.NoError(t, err)
    assert.Equal(t, protocol.SendAngles, mock.commands[0].Code)
}
```

**Integration Tests**: Require actual hardware, run with `MYCOBOT_PORT` env var set.

## Success Criteria

1. ✅ Thread-safe concurrent access from multiple goroutines
2. ✅ Context support for timeouts and cancellation
3. ✅ Strongly-typed API with compile-time safety
4. ✅ All four robot models supported
5. ✅ Gripper support with interface-based design
6. ✅ Protocol layer accessible for advanced users
7. ✅ Idiomatic Go code that feels native
8. ✅ Comprehensive error handling
9. ✅ Extensive documentation and examples
10. ✅ >80% test coverage

## Migration from pymycobot

| Python | Go |
|--------|-----|
| `from pymycobot import MyCobot280` | `import "github.com/user/mycobot-go"` |
| `mc = MyCobot280(port, baud)` | `robot := mycobot.NewMyCobot280(port, mycobot.WithBaudRate(baud))` |
| `mc.send_angles([0,0,0,0,0,0], 50)` | `robot.SendAngles(ctx, types.Angles{0,0,0,0,0,0}, types.Speed(50))` |
| `angles = mc.get_angles()` | `angles, err := robot.GetAngles(ctx)` |
| No explicit error handling | Explicit `error` return values |
| Optional thread locking | Thread-safe by default |
| No timeout control | Context-based timeouts everywhere |
