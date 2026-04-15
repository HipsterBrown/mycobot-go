# Successor Plan: Protocol Remediation and MechArm 270 Implementation

**Status**: Approved  
**Date**: 2026-04-15  
**Supersedes**: All prior plans in `docs/plans/2025-11-23-*`

## Background

The original plans (November 2025) were written from Elephant Robotics documentation rather than the pymycobot source code. A critical evaluation against pymycobot's `common.py` and `generate.py` revealed protocol-level bugs in the existing implementation that would cause hardware communication failures. Additionally, the only available hardware for testing is a MechArm 270.

This plan fixes the protocol layer first, then builds a complete MechArm 270 implementation as the sole supported model.

### Problems Found in Current Implementation

**Protocol bugs (will cause hardware failures):**

- Coordinate encoding uses `* 100` for all values. pymycobot uses `* 10` for XYZ position and `* 100` for Rx/Ry/Rz rotation. XYZ values would be sent 10x too large.
- Gripper command codes are misaligned with pymycobot's `ProtocolCode` class (every code is off by 1-2).
- Atom IO codes (0x60 range) are using Basic IO codes (0xA0 range). Digital pin and PWM commands would address the base panel instead of the end-effector.
- `SET_PIN_MODE (0x60)` and `SET_PWM_MODE (0x63)` are missing entirely.

**Design gaps:**

- `SendCoords` is missing a `mode` parameter for angular vs linear interpolation.
- All models hardcoded to `UseCRC: true`. pymycobot defaults to `0xFA` footer without CRC.
- Atom IO and Basic IO are conflated into a single subsystem.
- JOG direction parameter is an untyped `int` with no defined values.

### Design Decisions

- **MechArm 270 only.** Delete MyCobot280 and all other model configs. One correct model is better than four broken ones. Other models can be added in a future plan when hardware is available.
- **No gripper.** No working gripper hardware to test against. Gripper interface and commands are deferred.
- **Minimum viable API.** Only implement what's needed to move the arm reliably: motion, Atom IO, and servo control. Sync moves, global speed, free mode, Basic IO, and TCP/Socket are deferred until real usage demands them.
- **Fix-then-build.** Fix all protocol bugs in isolation (Phase 0) before touching the API surface (Phase 1). This ensures the byte-level foundation is correct and testable before building higher-level features.

## Phase 0: Protocol Remediation

Goal: make the bytes on the wire match what pymycobot sends, verified by unit tests. No API changes — just correctness fixes to the protocol and config layers.

### 0.1 — Regenerate command codes from pymycobot source

Replace the contents of `protocol/codes.go` with values taken directly from pymycobot's `ProtocolCode` class in `common.py`. Key corrections:

| Group | Current (wrong) | Corrected |
|-------|-----------------|-----------|
| Gripper GET_VALUE | 0x67 | 0x65 |
| Gripper SET_STATE | 0x68 | 0x66 |
| Gripper SET_VALUE | 0x66 | 0x67 |
| Gripper CALIBRATION | 0x69 | 0x68 |
| Gripper IS_MOVING | 0x6B | 0x69 |
| Atom SET_PIN_MODE | missing | 0x60 |
| Atom SET_DIGITAL_OUTPUT | 0xA0 | 0x61 |
| Atom GET_DIGITAL_INPUT | 0xA1 | 0x62 |
| Atom SET_PWM_MODE | missing | 0x63 |
| Atom SET_PWM_OUTPUT | 0xA2 | 0x64 |

Basic IO codes (0xA0, 0xA1) stay defined but are clearly labeled as Basic IO, separate from Atom IO.

### 0.2 — Fix coordinate encoding

Split `EncodeCoords` into proper encoding with different multipliers:

- XYZ positions: multiply by 10 (pymycobot's `_coord2int`)
- Rx/Ry/Rz rotations: multiply by 100 (pymycobot's `_angle2int`)

Replace the current `DecodeCoords` with the corresponding reverse: divide by 10 for XYZ, divide by 100 for rotations. The current implementation that delegates to `EncodeAngles` for all values is removed.

`EncodeAngles` / `DecodeAngles` remain unchanged (correctly use `* 100` / `/ 100`).

### 0.3 — Fix CRC default

Change all model configs to `UseCRC: false`. pymycobot's default frame format uses the `0xFA` footer, not CRC. CRC support stays in the code but becomes opt-in via a `WithCRC()` option for users who know their firmware requires it.

### 0.4 — Update tests

All existing protocol tests updated to reflect corrected encoding. Coordinate encode/decode tests specifically verify the split multiplier. For example, `EncodeCoords(200, 100, 50, 45, 90, 0)` should produce XYZ bytes at x10 scale and rotation bytes at x100 scale.

## Phase 1: MechArm 270 Implementation

Goal: replace MyCobot280 with a fully functional MechArm270, fix the `SendCoords` API, implement remaining subsystems (complete Motion, Atom IO, Servo), and test against real hardware.

### 1.1 — Delete MyCobot280, build MechArm270

Delete `mycobot280.go` and `mycobot280_test.go`. Remove model configs for MyCobot280, MyCobot320, and MyPalletizer260 from `config.go` — only MechArm270 remains. Create `mecharm270.go` with the same structure as the old MyCobot280 (constructor, Open/Close, power commands, motion commands) targeting the MechArm270 config.

### 1.2 — Fix Robot interface and SendCoords signature

Add a `mode` parameter to `SendCoords` in the `Robot` interface:

```go
SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error
```

`CoordMode` is a new type in `types/`:

- `CoordModeAngular = 0` — angular interpolation
- `CoordModeLinear = 1` — linear interpolation

### 1.3 — Complete Motion subsystem

The existing `motion.go` has JOG ops and Pause/Resume/Stop. Add the remaining methods:

- `SendAngle(ctx, joint, angle, speed)` — move a single joint
- `SendCoord(ctx, axis, value, speed)` — move along a single axis
- `IsInPosition(ctx, data, mode)` — verify current position matches target

Add a `Direction` type to replace the bare `int` parameter in JOG methods:

- `DirNegative = 0`
- `DirPositive = 1`

### 1.4 — Atom IO subsystem

Create `io.go` with an `IO` struct scoped to Atom IO only (0x60 range codes):

- `SetPinMode(ctx, pin, mode)` — configure pin as input/output/pullup
- `SetDigitalOutput(ctx, pin, signal)` — set pin high/low
- `GetDigitalInput(ctx, pin)` — read pin state
- `SetPWMMode(ctx, pin)` — configure pin for PWM
- `SetPWMOutput(ctx, channel, freq, dutyCycle)` — set PWM output
- `SetColor(ctx, r, g, b)` — control Atom LED

Pin mode and signal values get proper types rather than bare ints.

### 1.5 — Servo subsystem

Create `servo.go` with a `Servo` struct:

- `ReleaseServo(ctx, joint)` / `FocusServo(ctx, joint)` — individual servo power control
- `IsServoEnabled(ctx, joint)` — check if servo is active
- `GetEncoder(ctx, joint)` / `SetEncoder(ctx, joint, value)` — single encoder read/write
- `GetEncoders(ctx)` / `SetEncoders(ctx, encoders, speed)` — all encoders
- `GetServoData(ctx, joint, dataID)` / `SetServoData(ctx, joint, dataID, value)` — servo parameters
- `SetServoCalibration(ctx, joint)` — set current position as zero
- `GetJointMin(ctx, joint)` / `GetJointMax(ctx, joint)` — read joint limits from firmware

### 1.6 — Wire MechArm270 to subsystems

The `MechArm270` struct exposes subsystems as fields:

```go
type MechArm270 struct {
    Motion *Motion
    IO     *IO
    Servo  *Servo
    // base, config (unexported)
}
```

Subsystems are initialized in the constructor, sharing the same `robot.Base` instance.

### 1.7 — Integration test harness

Create a test file with build tag `//go:build integration` that tests against a real MechArm 270 over serial. The harness:

- Reads port from env var (`MYCOBOT_PORT`)
- Connects, verifies power state
- Sends a known-safe set of angles, reads them back
- Tests JOG with immediate stop
- Reads encoders
- Skips automatically when no hardware is present

## Out of Scope

Deferred to future plans when hardware or usage demands it:

- **Other robot models** — MyCobot280, MyCobot320, MyPalletizer260
- **Gripper subsystem** — interface, ProGripper, all gripper commands
- **Basic IO** — base panel pins (0xA0 range)
- **TCP/Socket connections** — serial only
- **Sync move methods** — `sync_send_angles`, `sync_send_coords`
- **Global speed** — `get_speed`, `set_speed`
- **Free mode** — `set_free_mode`, `is_free_mode`
- **Release/focus all servos** — convenience methods on Robot interface (available per-servo through Servo subsystem)
- **Coordinate validation** — model-specific XYZ range checking (depends on arm kinematics)

## Reference

- Original design: `docs/plans/2025-11-23-mycobot-go-port-design.md`
- Original Phase 1 implementation: `docs/plans/2025-11-23-mycobot-go-implementation.md`
- Original Phase 2 design: `docs/plans/2025-11-23-phase2-extended-features-design.md`
- Original Phase 2 implementation: `docs/plans/2025-11-23-phase2-implementation.md`
- pymycobot source (authoritative for protocol): https://github.com/elephantrobotics/pymycobot
- Protocol codes: pymycobot `common.py`, class `ProtocolCode`
- Encoding logic: pymycobot `common.py`, class `DataProcessor` (`_angle2int`, `_coord2int`, `_encode_int16`)
- Command framing: pymycobot `common.py`, method `_mesg`
