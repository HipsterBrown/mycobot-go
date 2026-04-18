# Architecture Flatten Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Flatten the mycobot-go API surface, promote the transport engine into an unexported `*base` embedded by model types, and delete the premature abstractions flagged in the /simplify architecture review.

**Architecture:** Convert subsystem structs (`Motion`, `IO`, `Servo`) into plain methods on a shared unexported `*base` type living at the root of the `mycobot` package. Each model type (`MechArm270` today, `MyCobot280` next) becomes a 10-line wrapper that embeds `*base` with a preconfigured `ModelConfig`, so methods are auto-promoted to the public surface. Delete the `Robot` interface, `RobotError`, every unused error sentinel, existence-only tests, and the `internal/robot/` and `internal/errors/` subpackages.

**Tech Stack:** Go 1.22, `go.bug.st/serial` v1.x, `github.com/stretchr/testify`.

**Spec:** `dev/plans/2026-04-18-architecture-flatten-design.md`

---

## File structure after migration

**New files**
- `types/pin.go` — `PinMode`, `PinSignal`, their constants (moved from `io.go`).
- `types/axis.go` — `CoordAxis`, `PositionFlag`, their constants (moved from `motion.go`).
- `base.go` (root) — unexported `base` type, transport loop, every shared opcode method. Absorbs today's `internal/robot/base.go` plus all methods from the subsystem files.
- `base_test.go` — fake-serial unit tests (eight cases).
- `dev/plans/README.md` — one-paragraph explainer for the dev planning dir.

**Modified files**
- `mecharm270.go` — shrinks to a 10-line constructor wrapping `*base`.
- `option.go` — `Option` becomes `func(*base)`; add `WithDefaultTimeout`.
- `errors.go` — sentinels inlined from `internal/errors`, `RobotError` deleted, unused sentinels deleted.
- `integration_test.go` — `arm.Motion.X` / `arm.IO.X` / `arm.Servo.X` rewritten to `arm.X`.
- `mecharm270_test.go` — drop the `assert.NotNil(t, arm.Motion/IO/Servo)` checks.
- `README.md` — collapse subsystem sections into a flat method list; drop broken `docs/plans/` links; document `WithDefaultTimeout`.
- `errors_test.go` — drop tests for deleted types (or delete the file if empty).

**Deleted files/directories**
- `robot.go`
- `motion.go`, `io.go`, `servo.go`
- `motion_test.go`, `io_test.go`, `servo_test.go`
- `internal/robot/` (entire directory — `base.go`, `base_test.go`)
- `internal/errors/` (entire directory)
- `docs/plans/` (moved to `dev/plans/`)

---

## Task 1: Move enum constants to `types/`

**Files:**
- Create: `types/pin.go`, `types/axis.go`
- Modify: `io.go`, `motion.go`
- Delete: `io_test.go`, `motion_test.go` (they reference the deleted local enums directly — `assert.Equal(t, PinMode(0), PinInput)` — and are the existence-check tests already slated for removal in the spec)

Keeps the build green end-of-commit. No behavior change.

- [ ] **Step 1: Create `types/pin.go`.**

```go
package types

// PinMode configures a pin's behavior.
type PinMode int

const (
	PinInput       PinMode = 0
	PinOutput      PinMode = 1
	PinInputPullup PinMode = 2
)

// PinSignal represents a digital pin state.
type PinSignal int

const (
	SignalLow  PinSignal = 0
	SignalHigh PinSignal = 1
)
```

- [ ] **Step 2: Create `types/axis.go`.**

```go
package types

// CoordAxis represents a coordinate axis for single-axis movement.
type CoordAxis int

const (
	AxisX CoordAxis = iota
	AxisY
	AxisZ
	AxisRx
	AxisRy
	AxisRz
)

// PositionFlag specifies whether IsInPosition checks angles or coordinates.
type PositionFlag int

const (
	PositionAngles PositionFlag = 0
	PositionCoords PositionFlag = 1
)
```

- [ ] **Step 3: Update `io.go` — delete local `PinMode`/`PinSignal` + constants (lines 11-26 in current file), reference `types.*` everywhere below.**

Expected edits in `io.go`:
- Delete the `PinMode` type + `PinInput`/`PinOutput`/`PinInputPullup` const block.
- Delete the `PinSignal` type + `SignalLow`/`SignalHigh` const block.
- Change every signature like `mode PinMode` to `mode types.PinMode`, `signal PinSignal` to `signal types.PinSignal`.
- Change every return value like `SignalLow` to `types.SignalLow`, `PinSignal(data[0])` to `types.PinSignal(data[0])`.

- [ ] **Step 4: Update `motion.go` — delete local `PositionFlag`/`CoordAxis` + constants (lines 12-32), reference `types.*` everywhere below.**

Expected edits in `motion.go`:
- Delete the `PositionFlag` type + `PositionAngles`/`PositionCoords` const block.
- Delete the `CoordAxis` type + `AxisX..AxisRz` const block.
- Change signatures like `axis CoordAxis` to `axis types.CoordAxis`, `flag PositionFlag` to `flag types.PositionFlag`.
- Change the comparison `if axis <= AxisZ` to `if axis <= types.AxisZ`.
- Change the conditional `if flag == PositionAngles` to `if flag == types.PositionAngles`.

- [ ] **Step 5: Delete the test files that reference the moved enums.**

`io_test.go` and `motion_test.go` both assert against the now-gone local `PinMode`/`PinSignal`/`CoordAxis`/`PositionFlag` types. They are the existence-check tests already slated for deletion by the spec — remove them now so Task 1 actually compiles.

```bash
rm io_test.go motion_test.go
```

- [ ] **Step 6: Build + test.**

Run: `cd /Users/nick.hehr/src/mycobot-go && go build ./... && go test ./...`
Expected: all pass.

- [ ] **Step 7: Commit.**

```bash
git add types/pin.go types/axis.go io.go motion.go
git rm io_test.go motion_test.go
git commit -m "$(cat <<'EOF'
refactor(types): move PinMode, PinSignal, CoordAxis, PositionFlag to types package

These enums are domain types, not subsystem-struct members. Collocating
with the rest of the types package sets up the subsystem flatten to
follow: io.go and motion.go no longer need to host them.

Also delete io_test.go and motion_test.go — they were existence-check
tests that asserted equality with the now-relocated enums. Those tests
never caught anything the compiler doesn't, and the files are already
slated for removal by the subsequent flatten.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Move `docs/plans/` → `dev/plans/`

**Files:**
- Move: all files under `docs/plans/` → `dev/plans/`
- Create: `dev/plans/README.md`
- Modify: `README.md`

No Go code touched.

- [ ] **Step 1: `git mv` the historical plan files.**

```bash
cd /Users/nick.hehr/src/mycobot-go
git mv docs/plans/2025-11-23-mycobot-go-implementation.md dev/plans/
git mv docs/plans/2025-11-23-mycobot-go-port-design.md dev/plans/
git mv docs/plans/2025-11-23-phase2-extended-features-design.md dev/plans/
git mv docs/plans/2025-11-23-phase2-implementation.md dev/plans/
git mv docs/plans/2026-04-15-successor-plan-design.md dev/plans/
git mv docs/plans/2026-04-15-successor-plan-implementation.md dev/plans/
```

- [ ] **Step 2: Verify `docs/plans/` is now empty and remove it.**

```bash
ls docs/plans/ 2>&1   # expect: empty or "No such file or directory"
rmdir docs/plans 2>/dev/null || true
ls docs/ 2>&1          # expect: nothing (or any remaining user-facing docs; there should be none today)
rmdir docs 2>/dev/null || true
```

- [ ] **Step 3: Create `dev/plans/README.md`.**

```markdown
# Development planning docs

This directory holds design documents and implementation plans authored during development. Consumer-facing documentation lives in the repository's top-level [README.md](../../README.md).

Naming convention: `YYYY-MM-DD-<topic>-design.md` for design/brainstorm docs, `YYYY-MM-DD-<topic>-implementation.md` for implementation plans.
```

- [ ] **Step 4: Drop the broken `docs/plans/` links from `README.md`.**

Edit `README.md` — delete the entire "## Documentation" section at the bottom (lines ~215-218: heading, the two `[Design document]` / `[Implementation plan]` bullet links, and the blank lines around them). The `## License` section immediately below stays.

- [ ] **Step 5: Build + test.**

Run: `go build ./... && go test ./...`
Expected: pass (nothing code-affecting changed).

- [ ] **Step 6: Commit.**

```bash
git add dev/plans/ README.md
# git mv already staged the old files as deletions
git commit -m "$(cat <<'EOF'
docs: relocate historical planning docs to dev/plans/

Keeps consumer-facing docs in README.md and separates development
planning artifacts so they do not surface as user documentation.
EOF
)"
```

---

## Task 3: Flatten API and promote `base`

**The big one.** Every change below lands in a single commit because the build is not green mid-task — subsystem structs and their methods have to be deleted at the same moment the flattened methods appear on `*base`.

**Expect `go build ./...` to fail between steps 1 and 17.** Do not run it as a checkpoint until Step 18. Unused imports in the scaffolded `base.go` will resolve as methods are added in later steps; subsystem files and the old `internal/robot/` directory must remain until after the flattened methods compile, or cross-references will break the build even harder.

**Files:**
- Create: `base.go` (root)
- Modify: `mecharm270.go`, `option.go`, `mecharm270_test.go`, `integration_test.go`, `README.md`
- Delete: `motion.go`, `io.go`, `servo.go`, `robot.go`, `motion_test.go`, `io_test.go`, `servo_test.go`, `internal/robot/` directory

**Prep:** Before editing, read each current file fully — the move is method-by-method and you need the exact current code to move.

- [ ] **Step 1: Read the current transport engine and all subsystem files.**

Use the Read tool on these in order:
- `internal/robot/base.go` (all of it)
- `mecharm270.go`
- `motion.go`
- `io.go`
- `servo.go`

Note each public method's signature and body — those are what move.

- [ ] **Step 2: Create `base.go` at the repo root from `internal/robot/base.go`.**

Scaffold, filling in method bodies as listed in subsequent steps.

```go
package mycobot

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"go.bug.st/serial"
	internalerrors "github.com/hipsterbrown/mycobot-go/internal/errors"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// headerSentinel is the two-byte frame header used to resync the stream.
var headerSentinel = []byte{protocol.Header, protocol.Header}

// openSerial is the serial-port factory. Production wires this to serial.Open;
// tests override it with a fake. Unexported.
var openSerial = func(port string, mode *serial.Mode) (serial.Port, error) {
	return serial.Open(port, mode)
}

// minFrameLen is header(2) + length(1) + code(1) + footer/crc(1).
const minFrameLen = 5

// base provides the transport loop and every shared opcode method.
type base struct {
	port           string
	baudrate       int
	useCRC         bool
	defaultTimeout time.Duration
	config         ModelConfig

	conn      serial.Port
	cmdChan   chan *command
	closeChan chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

	mu        sync.RWMutex // protects `connected`
	connected bool
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

func newBase(port string, cfg ModelConfig) *base {
	return &base{
		port:     port,
		baudrate: cfg.DefaultBaud,
		useCRC:   cfg.UseCRC,
		config:   cfg,
	}
}

// SetBaudRate overrides the default baud rate. Must be called before Open.
func (b *base) SetBaudRate(baud int) { b.baudrate = baud }

// SetUseCRC toggles CRC framing. Must be called before Open.
func (b *base) SetUseCRC(v bool) { b.useCRC = v }

// SetDefaultTimeout sets the fallback per-command read timeout used when
// the caller's context has no deadline. Must be called before Open.
func (b *base) SetDefaultTimeout(d time.Duration) { b.defaultTimeout = d }
```

- [ ] **Step 3: Port lifecycle methods (`Open`, `Close`, `IsConnected`) from `internal/robot/base.go` to `base.go` with receiver `*base`.**

Copy verbatim — only the receiver name changes from `*Base` to `*base`.

- [ ] **Step 4: Port `commandLoop`, `readResponse`, `extractMatchingFrame` verbatim.**

Same receiver rename. Add the timeout resolution chain in `readResponse`:

Replace the existing:
```go
timeout := 1 * time.Second
if d, ok := ctx.Deadline(); ok {
    timeout = time.Until(d)
}
```

with:
```go
var timeout time.Duration
if d, ok := ctx.Deadline(); ok {
    remaining := time.Until(d)
    if remaining <= 0 {
        return nil, ctx.Err()
    }
    timeout = remaining
} else if b.defaultTimeout > 0 {
    timeout = b.defaultTimeout
} else {
    timeout = 1 * time.Second
}
```

- [ ] **Step 5: Port `SendCommand` verbatim with receiver `*base`.**

- [ ] **Step 6: Port power methods from current `mecharm270.go` onto `*base`.**

Move `PowerOn`, `PowerOff`, `IsPowerOn` verbatim. Receiver changes from `*MechArm270` (with `m.base.SendCommand(...)`) to `*base` (with `b.SendCommand(...)`). Example:

```go
func (b *base) PowerOn(ctx context.Context) error {
	_, err := b.SendCommand(ctx, protocol.Command{Code: protocol.PowerOn})
	return err
}

func (b *base) IsPowerOn(ctx context.Context) (bool, error) {
	data, err := b.SendCommand(ctx, protocol.Command{Code: protocol.IsPowerOn, HasReply: true})
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}
```

- [ ] **Step 7: Port MDI motion methods from current `mecharm270.go` onto `*base`.**

Move `SendAngles`, `GetAngles`, `SendCoords`, `GetCoords`, `IsMoving`. Validation that currently reads `m.config.JointCount` / `m.config.Model` now reads `b.config.JointCount` / `b.config.Model`.

- [ ] **Step 8: Port motion subsystem methods from `motion.go` onto `*base`.**

Move `JogAngle`, `JogCoord`, `JogStop`, `Pause`, `Resume`, `Stop`, `IsPaused`, `SendAngle`, `SendCoord`, `IsInPosition`. Receivers change from `*Motion` to `*base` and the inner `m.robot.SendCommand(...)` becomes `b.SendCommand(...)`. Update method signatures to reference `types.CoordAxis` / `types.PositionFlag` / `types.AxisZ` / `types.PositionAngles` (already done by Task 1 inside `motion.go` itself — just preserve those references on the moved signatures).

- [ ] **Step 9: Port IO subsystem methods from `io.go` onto `*base`.**

Move `SetPinMode`, `SetDigitalOutput`, `GetDigitalInput`, `SetPWMMode`, `SetPWMOutput`, `SetColor`. Receivers change from `*IO` to `*base`. References to `types.PinMode` / `types.PinSignal` / `types.SignalLow` preserved.

**Careful with `SetColor`:** the current signature is `func (io *IO) SetColor(ctx context.Context, r, g, b byte) error` — parameter `b` shadows the new receiver `b *base`. Rename the color parameter:

```go
func (b *base) SetColor(ctx context.Context, red, green, blue byte) error {
	_, err := b.SendCommand(ctx, protocol.Command{
		Code: protocol.SetColor,
		Data: []byte{red, green, blue},
	})
	return err
}
```

Check the integration test — `arm.IO.SetColor(ctx, 0, 255, 0)` in `integration_test.go` is positional, so the rename has no caller impact. README also uses positional args. Safe rename.

- [ ] **Step 10: Port Servo subsystem methods from `servo.go` onto `*base`.**

Move `ReleaseServo`, `FocusServo`, `IsServoEnabled`, `GetEncoder`, `SetEncoder`, `GetEncoders`, `SetEncoders`, `GetServoData`, `SetServoData`, `SetServoCalibration`, `GetJointMin`, `GetJointMax`. Receivers change from `*Servo` to `*base`.

- [ ] **Step 11: Rewrite `mecharm270.go` to the thin wrapper.**

Replace the entire contents with:

```go
package mycobot

import "github.com/hipsterbrown/mycobot-go/types"

// MechArm270 is an Elephant Robotics MechArm 270 robot arm.
type MechArm270 struct {
	*base
}

// NewMechArm270 creates a client for a MechArm 270 connected at the given
// serial port. Default configuration is 115200 baud, CRC off; override with
// WithBaudRate, WithCRC, WithDefaultTimeout.
func NewMechArm270(port string, opts ...Option) *MechArm270 {
	cfg := getModelConfig(types.ModelMechArm270)
	b := newBase(port, cfg)
	for _, opt := range opts {
		opt(b)
	}
	return &MechArm270{base: b}
}
```

- [ ] **Step 12: Update `option.go` — change the `Option` type and switch setter receivers to `*base`.**

```go
package mycobot

// Option configures a robot client.
type Option func(*base)

// WithBaudRate overrides the default baud rate for the port.
func WithBaudRate(baud int) Option {
	return func(b *base) { b.SetBaudRate(baud) }
}

// WithCRC enables CRC framing for firmware that requires it.
// Default is the 0xFA footer used by MechArm 270 / MyCobot 280.
func WithCRC() Option {
	return func(b *base) { b.SetUseCRC(true) }
}
```

(`WithDefaultTimeout` is added in Task 5, not here.)

- [ ] **Step 13: Rewrite `mecharm270_test.go` to drop subsystem existence checks.**

Delete these three functions (verbatim names in the current file):
- `TestMechArm270_HasMotionSubsystem` (line ~61)
- `TestMechArm270_HasIOSubsystem` (line ~66)
- `TestMechArm270_HasServoSubsystem` (line ~71)

Keep everything else. The constructor-sanity tests, `TestMechArm270_PowerOn_NotConnected`, and both `SendAngles` validation tests still work because `arm.config` / `arm.base` / `arm.PowerOn` / `arm.SendAngles` all remain accessible via embedding promotion.

- [ ] **Step 14: Rewrite `integration_test.go` call sites.**

Replace every `arm.Motion.X` / `arm.IO.X` / `arm.Servo.X` with `arm.X`. Specifically (from current line numbers):
- Line 121: `arm.Motion.JogAngle(...)` → `arm.JogAngle(...)`
- Line 123: `arm.Motion.JogStop(ctx)` → `arm.JogStop(ctx)`
- Line 134: `arm.Servo.GetEncoders(ctx)` → `arm.GetEncoders(ctx)`
- Lines 151, 153: `arm.IO.SetColor(...)` → `arm.SetColor(...)`

No other integration-test changes.

- [ ] **Step 15: Delete subsystem files.**

`io_test.go` and `motion_test.go` were already removed in Task 1; the rest go now.

```bash
rm motion.go io.go servo.go robot.go servo_test.go
```

- [ ] **Step 16: Delete `internal/robot/` directory.**

```bash
rm -r internal/robot
```

- [ ] **Step 17: Update `README.md` — collapse subsystem sections into a flat API.**

Make every edit below. The section numbers refer to the current pre-flatten README.

1. **"Features" list (around line 14).** The bullet `- Subsystem-based design: Motion, IO, and Servo controls exposed as fields` — rewrite to `- Flat API: every opcode is a method on the model type (matches pymycobot's shape)`.

2. **"### Robot" section (around line 64).** The sentence `\`MechArm270\` implements the \`Robot\` interface and exposes three subsystems:` — rewrite to `\`MechArm270\` exposes the full opcode surface as methods directly on the type:`. Delete the three-line subsystem-fields block:
    ```
    // Subsystems
    arm.Motion   // JOG, single joint/axis moves, pause/resume/stop
    arm.IO       // Atom digital pins, PWM, LED color
    arm.Servo    // Individual servo control, encoders, calibration
    ```

3. **"### Motion Subsystem" / "### IO Subsystem" / "### Servo Subsystem" headers.** Rename to `### Motion (MDI + JOG)`, `### Atom IO (end-effector)`, `### Servo / Encoders` respectively. In every code snippet under those headers, rewrite:
   - `arm.Motion.` → `arm.`
   - `arm.IO.` → `arm.`
   - `arm.Servo.` → `arm.`
   - `mycobot.AxisX` → `types.AxisX` (and `AxisY`, `AxisZ`, `AxisRx`, `AxisRy`, `AxisRz`)
   - `mycobot.PinOutput` → `types.PinOutput` (and `PinInput`, `PinInputPullup`)
   - `mycobot.SignalHigh` → `types.SignalHigh` (and `SignalLow`)
   - `mycobot.PositionAngles` → `types.PositionAngles`
   - `mycobot.PositionCoords` → `types.PositionCoords`

4. **"## Package Structure" section (around line 187–213).** Rewrite the tree to match the new layout. Specifically: delete lines for `robot.go`, `motion.go`, `io.go`, `servo.go`, `internal/robot/base.go`, `internal/errors/`. Add `base.go` at the root with a comment like "transport + opcode methods shared across models." Leave `protocol/`, `types/`, and config/option/errors lines as-is (types/ gets new `pin.go` and `axis.go` — add them).

5. **"## Documentation" section at the bottom (the two `[Design document]` / `[Implementation plan]` links).** Already removed in Task 2. Confirm it's gone before committing.

- [ ] **Step 18: Build + unit test.**

```bash
go build ./...
go test ./...
```

Expected: all packages compile, all unit tests pass. If a method signature was missed in the move, the compiler will flag it — fix before continuing.

- [ ] **Step 19: Run integration tests against the arm.**

```bash
MYCOBOT_PORT=/dev/ttyUSB0 go test . -tags=integration -v -count=1
```

(Adjust `MYCOBOT_PORT` to your actual device, e.g. `/dev/cu.usbserial-*` on macOS.) Expected: same pass/fail profile as before this task. If anything that passed before now fails, the regression is in this commit — bisect method by method.

- [ ] **Step 20: Commit.**

```bash
git add -A
# -A because we both created base.go and deleted multiple files
git commit -m "$(cat <<'EOF'
refactor: flatten subsystem structs into methods on unexported *base

Motion, IO, and Servo were premature abstractions — each held no state,
carved an arbitrary subset of opcodes, and forced every model to expose
them as fields. Collapse them into plain methods on a shared unexported
*base type that model types embed. MechArm270 shrinks to a 10-line
wrapper; MyCobot 280 will follow the same shape.

Also delete the Robot interface (one implementer, partial surface),
promote internal/robot/base.go to root, introduce an openSerial
factory seam for future fake-serial tests, and update the context
timeout chain to clamp already-expired contexts before SetReadTimeout.

Breaking API change; pre-1.0 so acceptable. README updated in the
same commit so in-tree examples match.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Trim errors

**Files:**
- Modify: `errors.go`, `base.go`, `errors_test.go`
- Delete: `internal/errors/` directory

Only safe after Task 3 because `internal/robot/base.go` (which imported `internal/errors`) no longer exists.

- [ ] **Step 1: Inline the two surviving sentinels into root `errors.go`.**

Replace the entire contents of `errors.go` with:

```go
package mycobot

import "errors"

// Connection errors.
var (
	// ErrRobotClosed is returned when SendCommand is called after Close.
	ErrRobotClosed = errors.New("robot connection closed")
	// ErrNotConnected is returned when SendCommand is called before Open.
	ErrNotConnected = errors.New("robot not connected")
)
```

Delete: the `internalerrors "github.com/hipsterbrown/mycobot-go/internal/errors"` import, the `ErrConnectionTimeout`, `ErrInvalidCommand`, `ErrCommandTimeout`, `ErrInvalidResponse`, `ErrInvalidJoint`, `ErrInvalidSpeed`, `ErrInvalidAngle`, `ErrInvalidCoordinate`, `ErrNoGripper`, `ErrGripperNotSupported`, `ErrNotPowered`, `ErrEmergencyStop`, `ErrServoError` sentinels, and the entire `RobotError` struct + `Error()` + `Unwrap()` methods.

- [ ] **Step 2: Update `base.go` — drop the `internal/errors` import.**

In `base.go`, change:
```go
internalerrors "github.com/hipsterbrown/mycobot-go/internal/errors"
```
to nothing (remove the import), and change references to `internalerrors.ErrNotConnected` / `internalerrors.ErrRobotClosed` to plain `ErrNotConnected` / `ErrRobotClosed` (same package now).

- [ ] **Step 3: Delete `internal/errors/` directory.**

```bash
rm -r internal/errors
# If internal/ is now empty, remove it too:
rmdir internal 2>/dev/null || true
```

- [ ] **Step 4: Trim `errors_test.go`.**

Delete `TestRobotError_Error` and `TestRobotError_Unwrap`. If nothing else remains, delete the file entirely:

```bash
cat errors_test.go   # inspect remaining content
# if only imports + nothing else: rm errors_test.go
```

- [ ] **Step 5: Sanity grep for lingering references to deleted symbols.**

```bash
grep -rn "RobotError\|ErrInvalidCommand\|ErrCommandTimeout\|ErrInvalidResponse\|ErrInvalidJoint\|ErrInvalidSpeed\|ErrInvalidAngle\|ErrInvalidCoordinate\|ErrNoGripper\|ErrGripperNotSupported\|ErrNotPowered\|ErrEmergencyStop\|ErrServoError\|ErrConnectionTimeout\|internalerrors\|internal/errors" . 2>/dev/null | grep -v "dev/plans\|\.claude/" || echo "clean"
```

Expected: prints only `clean`. If any file still references a deleted symbol, fix it before continuing.

- [ ] **Step 6: Build + test.**

```bash
go build ./...
go test ./...
```

Expected: all pass. If any file still imports `internal/errors`, fix it.

- [ ] **Step 7: Commit.**

```bash
git add -A
git commit -m "$(cat <<'EOF'
refactor(errors): delete RobotError and aspirational sentinels

These types were defined but never constructed. Keep only the two
sentinels that base.go actually returns (ErrNotConnected,
ErrRobotClosed). Inline them from internal/errors/ and delete the
subpackage — the extra hop served no purpose once internal/robot/
is gone. Sentinels come back one at a time, lazily, when a real
call site returns them.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Add `WithDefaultTimeout` option

TDD: the behavior test comes before the wiring.

**Files:**
- Modify: `option.go`, `base_test.go` (create if missing), `README.md`

- [ ] **Step 1: Ensure `base_test.go` exists at the repo root.**

If Task 6 hasn't run yet, create the skeleton:

```go
package mycobot

import "testing"
```

(A full `base_test.go` with the fake-serial helpers arrives in Task 6. For Task 5 we can get away with inspecting `base` fields directly since they are package-internal.)

- [ ] **Step 2: Write a failing test for `WithDefaultTimeout`.**

Append to `base_test.go`:

```go
func TestWithDefaultTimeout_setsField(t *testing.T) {
	b := newBase("/dev/null", getModelConfig(types.ModelMechArm270))
	WithDefaultTimeout(250 * time.Millisecond)(b)

	if b.defaultTimeout != 250*time.Millisecond {
		t.Fatalf("defaultTimeout = %v, want 250ms", b.defaultTimeout)
	}
}
```

Add required imports: `"testing"`, `"time"`, `"github.com/hipsterbrown/mycobot-go/types"`.

- [ ] **Step 3: Run the test to confirm it fails.**

```bash
go test ./... -run TestWithDefaultTimeout
```

Expected: FAIL with `undefined: WithDefaultTimeout`.

- [ ] **Step 4: Implement `WithDefaultTimeout` in `option.go`.**

Add to `option.go`:

```go
import "time"

// WithDefaultTimeout sets the fallback per-command read timeout used when
// the caller's context has no deadline. If both are absent, the transport
// falls back to 1 second.
func WithDefaultTimeout(d time.Duration) Option {
	return func(b *base) { b.SetDefaultTimeout(d) }
}
```

- [ ] **Step 5: Run the test again — should pass.**

```bash
go test ./... -run TestWithDefaultTimeout
```

Expected: PASS.

- [ ] **Step 6: Build + run full test suite.**

```bash
go build ./...
go test ./...
```

Expected: pass.

- [ ] **Step 7: Document the option in `README.md`.**

Add to the "### Options" section:

```go
// Custom fallback timeout when ctx has no deadline (default: 1s)
arm := mycobot.NewMechArm270("/dev/ttyUSB0", mycobot.WithDefaultTimeout(500*time.Millisecond))
```

- [ ] **Step 8: Commit.**

```bash
git add option.go base_test.go README.md
git commit -m "$(cat <<'EOF'
feat(option): add WithDefaultTimeout for deadline-less contexts

Replaces the undocumented 1s fallback with a configurable, visible
option. context.Background() still works (falls back to 1s) but the
value is now discoverable in the API surface. Tightens the timeout
resolution chain so deadline → default → 1s is explicit.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: Fake-serial unit tests

**Files:**
- Modify: `base_test.go` (grow with fake + eight test cases)

No production-code changes — the `openSerial` seam already landed in Task 3.

- [ ] **Step 1: Append the fake `serial.Port` and helpers to `base_test.go`.**

Task 5 already created `base_test.go` with a `TestWithDefaultTimeout_setsField` test and its imports. Do not replace the file — append the helpers below, merging imports into the existing import block. After the merge the top of the file should have one import block containing: `bytes`, `context`, `errors`, `io`, `sync`, `testing`, `time`, `go.bug.st/serial`, and `github.com/hipsterbrown/mycobot-go/types`.

Helper code to append after the import block:

```go
// fakePort implements serial.Port for tests.
type fakePort struct {
	mu         sync.Mutex
	readBuf    *bytes.Buffer // bytes the fake delivers to Read
	writeBuf   bytes.Buffer  // bytes the code-under-test wrote
	flushCount int           // increments on ResetInputBuffer
	readErr    error         // if non-nil, Read returns it
	readDelay  time.Duration // simulate blocking before delivering bytes
	closed     bool
}

func newFakePort(reply []byte) *fakePort {
	return &fakePort{readBuf: bytes.NewBuffer(reply)}
}

func (f *fakePort) Read(p []byte) (int, error) {
	f.mu.Lock()
	if f.readErr != nil {
		err := f.readErr
		f.mu.Unlock()
		return 0, err
	}
	if f.readDelay > 0 {
		d := f.readDelay
		f.mu.Unlock()
		time.Sleep(d)
		f.mu.Lock()
	}
	if f.readBuf.Len() == 0 {
		f.mu.Unlock()
		return 0, io.EOF
	}
	n, _ := f.readBuf.Read(p)
	f.mu.Unlock()
	return n, nil
}

func (f *fakePort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.writeBuf.Write(p)
}

func (f *fakePort) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakePort) SetReadTimeout(time.Duration) error { return nil }
func (f *fakePort) SetMode(*serial.Mode) error          { return nil }
func (f *fakePort) ResetInputBuffer() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flushCount++
	return nil
}
func (f *fakePort) ResetOutputBuffer() error   { return nil }
func (f *fakePort) SetDTR(bool) error           { return nil }
func (f *fakePort) SetRTS(bool) error           { return nil }
func (f *fakePort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}
func (f *fakePort) Drain() error { return nil }
func (f *fakePort) Break(time.Duration) error { return nil }

// installFakePort swaps the package-level serial factory with one that
// returns `fake` and restores the original on test cleanup.
func installFakePort(t *testing.T, fake *fakePort) {
	t.Helper()
	prev := openSerial
	openSerial = func(string, *serial.Mode) (serial.Port, error) { return fake, nil }
	t.Cleanup(func() { openSerial = prev })
}

// openTestArm is a helper: fresh fake, fresh MechArm270 already opened.
func openTestArm(t *testing.T, reply []byte) (*MechArm270, *fakePort) {
	t.Helper()
	fake := newFakePort(reply)
	installFakePort(t, fake)
	arm := NewMechArm270("/dev/fake")
	if err := arm.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = arm.Close() })
	return arm, fake
}
```

Build check:

```bash
go build ./...
```

Expected: pass. If the `serial.Port` interface signature has drifted in your `go.bug.st/serial` version, add or remove methods to match — the compiler error tells you which.

- [ ] **Step 2: Test 1 — `HasReply=false` skips the read.**

```go
func TestBase_NoReplyCommandReturnsImmediately(t *testing.T) {
	arm, fake := openTestArm(t, nil) // no reply injected

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := arm.PowerOn(ctx) // HasReply=false
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("PowerOn: %v", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("PowerOn took %v, expected near-instant return", elapsed)
	}
	if fake.writeBuf.Len() == 0 {
		t.Error("expected bytes to be written")
	}
}
```

Run: `go test ./... -run TestBase_NoReplyCommandReturnsImmediately`
Expected: PASS.

- [ ] **Step 3: Test 2 — `HasReply=true` decodes a reply.**

```go
func TestBase_ReplyCommandDecodes(t *testing.T) {
	// GetAngles reply: FE FE 0E 20 <6 int16 angles as 12 bytes> FA
	// Angles = [0, 90.0, -45.5, 180.0, -180.0, 0.01]
	// Encoded int16 *100 = [0, 9000=0x2328, -4550=0xEE3A, 18000=0x4650, -18000=0xB9B0, 1=0x0001]
	reply := []byte{
		0xFE, 0xFE, 0x0E, 0x20,
		0x00, 0x00,
		0x23, 0x28,
		0xEE, 0x3A,
		0x46, 0x50,
		0xB9, 0xB0,
		0x00, 0x01,
		0xFA,
	}
	arm, _ := openTestArm(t, reply)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	angles, err := arm.GetAngles(ctx)
	if err != nil {
		t.Fatalf("GetAngles: %v", err)
	}
	if len(angles) != 6 {
		t.Fatalf("got %d angles, want 6", len(angles))
	}
	want := []float64{0, 90.0, -45.5, 180.0, -180.0, 0.01}
	for i, a := range angles {
		if diff := float64(a) - want[i]; diff < -0.01 || diff > 0.01 {
			t.Errorf("angle[%d] = %v, want %v", i, a, want[i])
		}
	}
}
```

Run: `go test ./... -run TestBase_ReplyCommandDecodes`
Expected: PASS.

- [ ] **Step 4: Test 3 — stale-frame resync (wrong code, then right code).**

```go
func TestBase_SkipsStaleFrameWithWrongCode(t *testing.T) {
	// First: a stale PowerOn reply (code 0x10, 5 bytes).
	// Second: the GetAngles reply (code 0x20) we actually want.
	stale := []byte{0xFE, 0xFE, 0x02, 0x10, 0xFA}
	wanted := []byte{
		0xFE, 0xFE, 0x0E, 0x20,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0xFA,
	}
	arm, _ := openTestArm(t, append(stale, wanted...))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	angles, err := arm.GetAngles(ctx)
	if err != nil {
		t.Fatalf("GetAngles: %v", err)
	}
	if len(angles) != 6 {
		t.Fatalf("got %d angles, want 6", len(angles))
	}
}
```

Run: `go test ./... -run TestBase_SkipsStaleFrameWithWrongCode`
Expected: PASS.

- [ ] **Step 5: Test 4 — garbage prefix before header.**

```go
func TestBase_SkipsGarbageBeforeHeader(t *testing.T) {
	reply := append(
		[]byte{0xAA, 0xBB, 0xCC},
		0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA, // IsPowerOn reply: payload = 0x01
	)
	arm, _ := openTestArm(t, reply)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	on, err := arm.IsPowerOn(ctx)
	if err != nil {
		t.Fatalf("IsPowerOn: %v", err)
	}
	if !on {
		t.Error("expected power on")
	}
}
```

Run: `go test ./... -run TestBase_SkipsGarbageBeforeHeader`
Expected: PASS.

- [ ] **Step 6: Test 5 — terminal read error (non-EOF).**

```go
func TestBase_ReadErrorIsTerminal(t *testing.T) {
	fake := newFakePort(nil)
	fake.readErr = errors.New("simulated device lost")
	installFakePort(t, fake)

	arm := NewMechArm270("/dev/fake")
	if err := arm.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer arm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	_, err := arm.IsPowerOn(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("returned after %v, expected immediate (<500ms)", elapsed)
	}
}
```

Run: `go test ./... -run TestBase_ReadErrorIsTerminal`
Expected: PASS — the call returns in <500ms, not the 2s context timeout.

- [ ] **Step 7: Test 6 — `WithDefaultTimeout` honored when ctx has no deadline.**

```go
func TestBase_WithDefaultTimeoutHonored(t *testing.T) {
	fake := newFakePort(nil) // no reply, so read will time out
	installFakePort(t, fake)

	arm := NewMechArm270("/dev/fake", WithDefaultTimeout(80*time.Millisecond))
	if err := arm.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer arm.Close()

	start := time.Now()
	_, err := arm.IsPowerOn(context.Background()) // no deadline
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed < 50*time.Millisecond || elapsed > 300*time.Millisecond {
		t.Errorf("elapsed = %v, expected ~80ms (±range)", elapsed)
	}
}
```

Run: `go test ./... -run TestBase_WithDefaultTimeoutHonored`
Expected: PASS.

- [ ] **Step 8: Test 7 — `ResetInputBuffer` called once per command.**

```go
func TestBase_FlushesBeforeEveryWrite(t *testing.T) {
	reply := []byte{0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA}
	// We'll issue two IsPowerOn calls, so preload two copies.
	arm, fake := openTestArm(t, append(reply, reply...))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if _, err := arm.IsPowerOn(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := arm.IsPowerOn(ctx); err != nil {
		t.Fatal(err)
	}

	fake.mu.Lock()
	got := fake.flushCount
	fake.mu.Unlock()
	if got != 2 {
		t.Errorf("flushCount = %d, want 2 (one per command)", got)
	}
}
```

Run: `go test ./... -run TestBase_FlushesBeforeEveryWrite`
Expected: PASS.

- [ ] **Step 9: Test 8 — incomplete frame arrives, completes on second read.**

```go
func TestBase_AssemblesFrameAcrossReads(t *testing.T) {
	// Craft a fake that delivers a valid frame but in two fragments.
	// This requires a slightly richer fake, so inline the setup.
	full := []byte{0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA}
	first := full[:3]
	second := full[3:]

	fake := newFakePort(first) // first chunk pre-loaded
	// After the first Read consumes `first`, the second chunk is staged.
	// The simplest way: we wrap the fake so the reader has a background
	// goroutine that appends `second` after a short delay.
	go func() {
		time.Sleep(20 * time.Millisecond)
		fake.mu.Lock()
		fake.readBuf.Write(second)
		fake.mu.Unlock()
	}()
	installFakePort(t, fake)

	arm := NewMechArm270("/dev/fake")
	if err := arm.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer arm.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	on, err := arm.IsPowerOn(ctx)
	if err != nil {
		t.Fatalf("IsPowerOn: %v", err)
	}
	if !on {
		t.Error("expected on")
	}
}
```

Note: if `fake.Read` returns `io.EOF` when the buffer is empty, the test will short-circuit. Ensure the fake's `Read` behavior matches: when `readBuf.Len() == 0`, return `(0, io.EOF)` — the transport treats EOF as "no bytes this round, try again." Verify by running.

Run: `go test ./... -run TestBase_AssemblesFrameAcrossReads`
Expected: PASS.

- [ ] **Step 10: Full test suite.**

```bash
go build ./...
go test ./...
```

Expected: all unit tests pass.

- [ ] **Step 11: Re-run integration suite against the arm.**

```bash
MYCOBOT_PORT=/dev/ttyUSB0 go test . -tags=integration -v -count=1
```

Expected: same pass/fail profile as before Task 6. This is the second integration checkpoint from the spec.

- [ ] **Step 12: Commit.**

```bash
git add base_test.go
git commit -m "$(cat <<'EOF'
test(base): add fake-serial unit tests for transport and opcode logic

Eight cases covering the areas that have actually regressed before:
HasReply skip path, frame resync, stale-code filtering, garbage
prefix tolerance, terminal read errors, WithDefaultTimeout honored,
input-buffer drain per write, multi-read frame assembly.

Replaces the deleted existence-check tests with regression coverage
that does not need hardware.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Completion checklist

- [ ] All six tasks committed.
- [ ] `go build ./...` green.
- [ ] `go test ./...` green.
- [ ] Integration suite run at least once on real hardware after Task 3 and again after Task 6.
- [ ] `README.md` reflects the flat API.
- [ ] `docs/plans/` no longer exists; historical plans live in `dev/plans/`.
- [ ] No references to `Robot` interface, `RobotError`, `mycobot.Motion`, `mycobot.IO`, `mycobot.Servo`, `internal/robot`, or `internal/errors` remain in the tree.

## Rollback plan

Each task is a single commit. If a task breaks integration on hardware and the regression is not obvious, `git revert <sha>` of that task's commit restores the prior state without touching the earlier tasks. Task 3 is the most fragile — if a specific method move is suspect, `git log -p base.go` after the revert lets you re-apply that task incrementally by copying methods over one at a time.
