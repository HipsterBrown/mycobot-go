# mycobot-go

Go library for controlling [Elephant Robotics](https://www.elephantrobotics.com/) robot arms over serial. A Go port of [pymycobot](https://github.com/elephantrobotics/pymycobot).

## Supported Models

- MechArm 270

## Features

- Thread-safe serial communication via a dedicated command goroutine
- Context-based timeout and cancellation on all commands
- Strongly-typed API with validated domain types (angles, speed, coordinates, joint IDs)
- Flat API: every opcode is a method on the model type (matches pymycobot's shape)
- Exposed protocol layer for advanced/raw command usage
- Correct wire protocol encoding verified against pymycobot source

## Installation

```bash
go get github.com/hipsterbrown/mycobot-go
```

Requires Go 1.22+.

## Quick Start

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
    arm := mycobot.NewMechArm270("/dev/ttyUSB0")
    ctx := context.Background()

    if err := arm.Open(ctx); err != nil {
        log.Fatal(err)
    }
    defer arm.Close()

    if err := arm.PowerOn(ctx); err != nil {
        log.Fatal(err)
    }
    time.Sleep(500 * time.Millisecond) // allow servos to engage

    // Move all joints to zero
    err := arm.SendAngles(ctx, types.Angles{0, 0, 0, 0, 0, 0}, types.SpeedMedium)
    if err != nil {
        log.Fatal(err)
    }
}
```

## API Overview

### Robot

`MechArm270` exposes the full opcode surface as methods directly on the type:

```go
arm := mycobot.NewMechArm270("/dev/ttyUSB0")

// Core methods on the robot
arm.Open(ctx)
arm.Close()
arm.PowerOn(ctx)
arm.PowerOff(ctx)
arm.IsPowerOn(ctx)
arm.SendAngles(ctx, angles, speed)
arm.GetAngles(ctx)
arm.SendCoords(ctx, coord, speed, mode)
arm.GetCoords(ctx)
arm.IsMoving(ctx)
```

### Options

```go
// Custom baud rate (default: 115200)
arm := mycobot.NewMechArm270("/dev/ttyUSB0", mycobot.WithBaudRate(1000000))

// Enable CRC mode for firmware that requires it (default: off, uses 0xFA footer)
arm := mycobot.NewMechArm270("/dev/ttyUSB0", mycobot.WithCRC())
```

### Motion (MDI + JOG)

```go
// JOG (incremental movement)
arm.JogAngle(ctx, types.Joint1, types.DirPositive, types.SpeedSlow)
arm.JogCoord(ctx, types.AxisX, types.DirNegative, types.SpeedSlow)
arm.JogStop(ctx)

// Single joint/axis control
arm.SendAngle(ctx, types.Joint3, types.Angle(45.0), types.SpeedMedium)
arm.SendCoord(ctx, types.AxisZ, 200.0, types.SpeedMedium)

// Movement control
arm.Pause(ctx)
arm.Resume(ctx)
arm.Stop(ctx)
arm.IsPaused(ctx)

// Position check
arm.IsInPosition(ctx, angleValues, types.PositionAngles)
arm.IsInPosition(ctx, coordValues, types.PositionCoords)
```

### Atom IO (end-effector)

```go
arm.SetPinMode(ctx, pin, types.PinOutput)
arm.SetDigitalOutput(ctx, pin, types.SignalHigh)
signal, _ := arm.GetDigitalInput(ctx, pin)
arm.SetColor(ctx, 255, 0, 0) // red LED
```

### Servo / Encoders

```go
arm.ReleaseServo(ctx, types.Joint1) // free movement
arm.FocusServo(ctx, types.Joint1)   // re-engage
enabled, _ := arm.IsServoEnabled(ctx, types.Joint1)

// Encoder values (0-4096)
enc, _ := arm.GetEncoder(ctx, types.Joint1)
encoders, _ := arm.GetEncoders(ctx)

// Joint limits from firmware
min, _ := arm.GetJointMin(ctx, types.Joint1)
max, _ := arm.GetJointMax(ctx, types.Joint1)
```

### Coordinate Modes

`SendCoords` accepts a mode parameter for trajectory interpolation:

```go
// Angular interpolation (joint space)
arm.SendCoords(ctx, coord, types.SpeedMedium, types.CoordModeAngular)

// Linear interpolation (cartesian space)
arm.SendCoords(ctx, coord, types.SpeedMedium, types.CoordModeLinear)
```

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests (real hardware)

Integration tests are guarded behind a build tag and require a MechArm 270 connected via USB serial.

```bash
MYCOBOT_PORT=/dev/ttyUSB0 go test . -tags=integration -v -count=1
```

Set `MYCOBOT_PORT` to your serial port (e.g., `/dev/ttyUSB0` on Linux, `/dev/cu.usbserial-*` on macOS). Tests will skip automatically if the env var is not set.

The integration tests will:
1. Connect and verify power state
2. Read joint angles and coordinates
3. Send all joints to zero position and verify round-trip
4. JOG a joint and stop
5. Read encoder values
6. Flash the Atom LED green

**Safety**: Ensure the arm has clearance to move to the zero position before running tests.

## Package Structure

```
mycobot-go/
  base.go            # transport + opcode methods shared across models
  mecharm270.go      # MechArm270 robot implementation
  config.go          # Model-specific configuration
  option.go          # Functional options (WithBaudRate, WithCRC)
  errors.go          # Error types
  protocol/          # Wire protocol encoding/decoding
    codes.go         # Command byte constants (from pymycobot ProtocolCode)
    command.go       # Frame encoding, angle/coord encoding
  types/             # Domain types with validation
    joint.go         # JointID (1-based)
    angle.go         # Angle, Angles
    speed.go         # Speed (0-100)
    coord.go         # Coord (X, Y, Z, Rx, Ry, Rz)
    coord_mode.go    # CoordMode (angular/linear)
    direction.go     # Direction (negative/positive)
    model.go         # Model type, joint limits
    pin.go           # PinMode, PinSignal
    axis.go          # CoordAxis, PositionFlag
  internal/
    errors/          # Internal error sentinels
```

## License

MIT
