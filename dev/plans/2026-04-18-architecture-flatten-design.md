# Architecture flatten: design

**Date:** 2026-04-18
**Status:** Draft — pending review
**Scope:** Resolve the first seven items of the architecture review returned by the `/simplify` pass.
**Non-goal:** Change the library's framing/positioning (item 8 of the review). The library remains a Go port of pymycobot for Elephant Robotics arms.

## Context

`mycobot-go` is a Go port of `pymycobot` targeting Elephant Robotics robot arms. It currently supports one model (MechArm 270) and will add a second (MyCobot 280) after gripper control lands. The codebase was scaffolded from a comprehensive design document before a second model existed, which produced several premature abstractions:

- A `Robot` interface with one implementer and a partial method surface.
- Subsystem structs (`Motion`, `IO`, `Servo`) hanging off the concrete model, with their methods and domain enums (`PinMode`, `CoordAxis`, `PositionFlag`) exported from the top-level package.
- A transport engine (`internal/robot.Base`) hidden behind `internal/`, forcing any future model into the top-level `mycobot` package.
- An aspirational error surface: `RobotError` defined but never constructed, and roughly a dozen `ErrInvalid*` sentinels that no call site returns.
- An undocumented 1-second fallback timeout for commands invoked with `context.Background()`.
- Existence-only tests (`_ = m.JogAngle`) that duplicate what the compiler already enforces.
- Planning docs living in `docs/plans/`, surfaced alongside user-facing content in the top-level `docs/` directory.

The goal of this design is to flatten the architecture so adding MyCobot 280 is a small, obvious change, and so the remaining public surface is honest about what it does.

## Goals

1. Make the shared transport + opcode surface the single source of truth for every model, written once.
2. Match `pymycobot`'s flat API shape (`arm.jog_angle(...)`) so a consumer cross-referencing both libraries sees the same vocabulary.
3. Shrink each model file to a thin wrapper that carries model-specific config and any real overrides.
4. Make timeout behavior discoverable (option, not magic).
5. Replace tests that cost maintenance but catch no bugs with ones that exercise the transport logic that has actually broken in the past.
6. Separate development planning artifacts from consumer-facing docs.
7. Delete error types that have no call site rather than wiring them up aspirationally.

## Non-goals

- No change to the wire protocol or opcode codes. `protocol/` is intentionally untouched.
- No change to the `types/` package's existing types (`Angle`, `Speed`, `Coord`, `JointID`, etc.). It only gains new files for enums migrating out of the subsystem structs.
- No premature `transport` or `client` public package. If an advanced user later needs raw-bytes access without a model wrapper, the unexported `*base` can be promoted at that time. YAGNI.
- No narrative rewrite of `README.md`. The API-overview section that documents `arm.Motion.*` / `arm.IO.*` / `arm.Servo.*` must be updated to match the flattened surface (see "Breaking changes" and the migration order), but no tone/framing changes are in scope.

## Breaking changes

This refactor is source-incompatible with the currently-documented API. Callers must migrate:

| Old (0.x) | New |
|---|---|
| `arm.Motion.JogAngle(...)`, `arm.Motion.JogStop(...)`, `arm.Motion.Pause(...)`, etc. | `arm.JogAngle(...)`, `arm.JogStop(...)`, `arm.Pause(...)` |
| `arm.IO.SetColor(...)`, `arm.IO.SetPinMode(...)`, etc. | `arm.SetColor(...)`, `arm.SetPinMode(...)` |
| `arm.Servo.GetEncoders(...)`, `arm.Servo.ReleaseServo(...)`, etc. | `arm.GetEncoders(...)`, `arm.ReleaseServo(...)` |
| `mycobot.PinMode`, `mycobot.PinSignal`, `mycobot.PinInput`, `mycobot.PinOutput`, `mycobot.SignalLow`, `mycobot.SignalHigh` | `types.PinMode`, `types.PinSignal`, `types.PinInput`, etc. |
| `mycobot.CoordAxis`, `mycobot.AxisX..AxisRz`, `mycobot.PositionFlag`, `mycobot.PositionAngles`, `mycobot.PositionCoords` | `types.CoordAxis`, `types.AxisX`, `types.PositionFlag`, etc. |
| `mycobot.Robot` interface | Removed. Consumers define their own if needed. |
| `mycobot.RobotError` struct | Removed. |
| `mycobot.ErrInvalidCommand`, `ErrCommandTimeout`, `ErrInvalidResponse`, `ErrInvalidJoint`, `ErrInvalidSpeed`, `ErrInvalidAngle`, `ErrInvalidCoordinate`, `ErrNoGripper`, `ErrGripperNotSupported`, `ErrNotPowered`, `ErrEmergencyStop`, `ErrServoError`, `ErrConnectionTimeout` | Removed. |

The repository has one user today (the author, per project memory), and the library is pre-1.0, so breaking changes are acceptable.

## Decisions

| # | Question | Decision |
|---|---|---|
| 1 | Roadmap | Design for sharing. MyCobot 280 is next after gripper. |
| 2 | API shape | Flat — methods on the model type. No subsystem structs. |
| 3 | Shared engine | Unexported `*base` in root `mycobot` package. Models embed it. |
| 4 | Context timeout | Reintroduce `WithDefaultTimeout(d)` option. Fallback chain: `ctx.Deadline()` → `b.defaultTimeout` → 1 s. |
| 5 | Tests | Delete existence tests. Add fake-serial unit tests against the `serial.Port` interface. |
| 6 | Errors | YAGNI. Delete `RobotError` + every unused sentinel. Keep `ErrRobotClosed`, `ErrNotConnected`. |
| 7 | Docs | Move `docs/plans/` → `dev/plans/`. Future brainstorm specs land in `dev/plans/` too. |

## Target package layout

```
mycobot-go/
  mycobot.go          # package doc
  base.go             # unexported `base` type: transport + all shared opcodes
  mecharm270.go       # MechArm270 model (embeds *base)
  config.go           # ModelConfig registry (unchanged)
  option.go           # WithBaudRate, WithCRC, WithDefaultTimeout
  errors.go           # ErrRobotClosed, ErrNotConnected (only)
  base_test.go        # fake-serial unit tests
  mecharm270_test.go  # model-config sanity
  integration_test.go # real hardware (build tag)

  protocol/           # unchanged
  types/              # all domain enums + types
    joint.go / angle.go / speed.go / coord.go / coord_mode.go / direction.go / model.go
    pin.go            # NEW: PinMode, PinSignal
    axis.go           # NEW: CoordAxis, PositionFlag

  dev/
    plans/            # dev planning docs (moved from docs/plans/)
      README.md       # explains this dir
      2026-04-18-architecture-flatten-design.md  # this file

  docs/               # user-facing only (empty after move; README is at root)
```

**Deletions:**
- `robot.go` — the `Robot` interface.
- `motion.go`, `io.go`, `servo.go` — subsystem structs.
- `motion_test.go`, `io_test.go`, `servo_test.go` — existence-only tests.
- `internal/robot/` — contents absorbed into root `base.go`.
- `internal/errors/` — contents absorbed into root `errors.go`.
- `docs/plans/` — moved to `dev/plans/`.

## The `base` type

Single source of truth for the transport loop and every opcode method.

```go
// base.go (unexported)

type base struct {
    port           string
    baudrate       int
    useCRC         bool
    defaultTimeout time.Duration
    config         ModelConfig

    conn      serial.Port
    cmdChan   chan *command     // preserves current buffer size of 32
    closeChan chan struct{}
    closeOnce sync.Once         // guards closeChan close on first Close() call
    wg        sync.WaitGroup
    mu        sync.RWMutex      // protects `connected`
    connected bool
}
```

All concurrency primitives and their contracts are preserved from the current `internal/robot.Base` — `cmdChan` buffer stays at 32, `closeOnce` keeps `Close()` idempotent, `mu` protects only `connected`. The only structural additions are `defaultTimeout` and `config`.

Methods on `*base`:

- **Lifecycle:** `Open(ctx) error`, `Close() error`, `IsConnected() bool`
- **Raw passthrough:** `SendCommand(ctx, protocol.Command) ([]byte, error)`
- **Power:** `PowerOn`, `PowerOff`, `IsPowerOn`
- **MDI motion:** `SendAngles`, `GetAngles`, `SendCoords`, `GetCoords`, `SendAngle`, `SendCoord`, `IsMoving`, `IsInPosition`, `Pause`, `Resume`, `Stop`, `IsPaused`
- **JOG motion:** `JogAngle`, `JogCoord`, `JogStop`
- **Servo:** `ReleaseServo`, `FocusServo`, `IsServoEnabled`, `GetEncoder`, `SetEncoder`, `GetEncoders`, `SetEncoders`, `GetServoData`, `SetServoData`, `SetServoCalibration`, `GetJointMin`, `GetJointMax`
- **IO:** `SetPinMode`, `SetDigitalOutput`, `GetDigitalInput`, `SetPWMMode`, `SetPWMOutput`, `SetColor`

### Validation and model config

Commands whose valid domain depends on the model (`SendAngles` — 6 vs 12 joints; `SendAngle` / `GetJointMin` / `GetJointMax` — joint ID range) read from `b.config` directly. We do not gate opcode availability at the `*base` layer; if a model's firmware does not implement an opcode, calling it will time out, and that is OK — the same is true of pymycobot.

### Timeout resolution

The per-read timeout in the command loop is resolved as:

```
if ctx has deadline:
    remaining := time.Until(deadline)
    if remaining <= 0: return ctx.Err() immediately (already expired)
    timeout := remaining
else if b.defaultTimeout > 0:
    timeout := b.defaultTimeout
else:
    timeout := 1 * time.Second
```

The 1 s final fallback is preserved so callers who never set a deadline or option don't hang on misbehaving firmware. The expired-context guard is a small correctness fix over the current code, which would call `SetReadTimeout` with a negative duration.

### CRC mode resolution

The existing per-command mutation in `commandLoop` is preserved: before `Encode()`, `cmd.request.UseCRC` is set to `b.useCRC`. Callers never need to set `UseCRC` on a `protocol.Command`; the transport layer owns it. Protocol-level tests in `protocol/command_test.go` that construct `Command{UseCRC: true}` directly are unaffected because they exercise `Encode` / `Decode` in isolation, not through `commandLoop`.

### Serial port factory seam

To allow fake-serial unit tests, a package-level unexported variable in `base.go` holds the serial-open function:

```go
// default: real serial driver
var openSerial = func(port string, mode *serial.Mode) (serial.Port, error) {
    return serial.Open(port, mode)
}
```

Tests that need a fake set `openSerial` to a factory returning their fake and restore it via `t.Cleanup`. The `newBase` signature stays `newBase(port string, cfg ModelConfig) *base`; the seam does not leak into any public or model-facing signature.

### `SendCommand` passthrough

The current `MechArm270.SendCommand(ctx, protocol.Command) ([]byte, error)` method is an explicit re-export of the transport's low-level entry point. After flattening, `*base` has `SendCommand` directly and Go's embedding promotion exposes it on every model type. This is intentional: advanced users who need to issue a raw opcode retain access, on every current and future model, with no per-model re-export boilerplate.

## Model types

```go
// mecharm270.go

type MechArm270 struct {
    *base
}

func NewMechArm270(port string, opts ...Option) *MechArm270 {
    cfg := getModelConfig(types.ModelMechArm270)
    b := newBase(port, cfg)
    for _, opt := range opts {
        opt(b)
    }
    return &MechArm270{base: b}
}
```

That is the whole file, plus any model-specific method overrides when they materialize. Shared methods (`PowerOn`, `SendAngles`, etc.) are available on `*MechArm270` via Go's embedding promotion — `arm.PowerOn(ctx)` continues to work.

When `MyCobot 280` arrives: a `mycobot280.go` file with the same 10-line shape and a different `ModelConfig`. Gripper methods that exist only on certain models are defined on the model type itself, not on `*base`.

## Robot interface

Deleted. The interface had one implementer and carved a subset of the real method surface, so programming against it yielded less than programming against the concrete type. Go convention: "accept interfaces, return concrete types" — libraries return concrete types, consumers define interfaces at their call sites.

## Small reshuffles

### Constants migrate to `types/`

- `PinMode` (Input / Output / InputPullup), `PinSignal` (Low / High) → `types/pin.go`
- `CoordAxis` (X / Y / Z / Rx / Ry / Rz), `PositionFlag` (Angles / Coords) → `types/axis.go`

Call sites using `mycobot.PinOutput` become `types.PinOutput`, etc. Mechanical find/replace.

### Errors collapse to root

`internal/errors/` is inlined into root `errors.go`. The package exports exactly:

```go
var (
    ErrRobotClosed  = errors.New("robot connection closed")
    ErrNotConnected = errors.New("robot not connected")
)
```

Everything else (`RobotError` struct, `ErrInvalidCommand`, `ErrCommandTimeout`, `ErrInvalidResponse`, `ErrInvalidJoint`, `ErrInvalidSpeed`, `ErrInvalidAngle`, `ErrInvalidCoordinate`, `ErrNoGripper`, `ErrGripperNotSupported`, `ErrNotPowered`, `ErrEmergencyStop`, `ErrServoError`, `ErrConnectionTimeout`) is deleted. New sentinels return on demand, one at a time, when a call site actually returns them.

`errors_test.go` keeps only tests for the two surviving sentinels (or is deleted if those tests were just trivial equality checks).

### Docs move

- `docs/plans/` → `dev/plans/` via `git mv`.
- `dev/plans/README.md` — new, one paragraph: "Development planning docs. Consumer docs are in the top-level README."
- Top-level `README.md` — drop the two links at the bottom to the old plan files.

## Test strategy

### Delete

- `motion_test.go`, `io_test.go`, `servo_test.go` — existence checks.
- The `_ = arm.Method` style asserts inside `mecharm270_test.go`.
- `errors_test.go` tests for deleted types.

### Keep as-is

- `config_test.go`
- `types/*_test.go`
- `protocol/*_test.go`
- `integration_test.go`

### Add `base_test.go`

Uses a fake `serial.Port` implementation. The fake holds an injected reply buffer and a captured write buffer, implements the six `serial.Port` methods, and is injected via the factory seam described in the `base` section.

**Target test cases:**

1. `HasReply=false` skips read — write `PowerOn`, inject no reply, call returns nil quickly.
2. `HasReply=true` decodes reply — inject `FE FE 0E 20 <12 bytes> FA`, `GetAngles` returns six decoded angles.
3. Stale-frame resync — inject a well-formed frame with wrong code followed by the real one; we skip the wrong code and return the right payload.
4. Garbage prefix — inject `0xAA 0xBB FE FE ... FA`; `readResponse` skips to the header and decodes.
5. Split frame — inject in two `Read` chunks; decode eventually succeeds.
6. Terminal read error — fake returns a non-EOF error; `readResponse` returns it immediately (no busy loop).
7. `WithDefaultTimeout` honored — open with 50 ms default, call with `context.Background()`, assert deadline ≈ 50 ms (not 1 s).
8. `ResetInputBuffer` called before each write — fake tracks flush count; assert exactly one per command.

## Migration order

Six commits on a single feature branch, each keeping `go build ./...` and `go test ./...` green. The order matters: the errors cleanup (step 4) deliberately comes after the base promotion (step 3) to avoid a circular-import problem. If errors were inlined first, `internal/robot/base.go` (which still imports `internal/errors`) would either lose its symbol or have to import the root `mycobot` package — but `mycobot` already imports `internal/robot`, creating a cycle.

1. **Move enum constants to `types/`.**
   - `PinMode`, `PinSignal`, `PinInput`, `PinOutput`, `PinInputPullup`, `SignalLow`, `SignalHigh` → new `types/pin.go`.
   - `CoordAxis`, `AxisX..AxisRz`, `PositionFlag`, `PositionAngles`, `PositionCoords` → new `types/axis.go`.
   - Update call sites in `io.go`, `motion.go`, `integration_test.go` to reference `types.*`.
   - Pure rename. Build green at end of commit.

2. **Move `docs/plans/` → `dev/plans/`.**
   - `dev/plans/` already exists (it contains this spec).
   - `git mv docs/plans/* dev/plans/` for each historical plan file.
   - Add `dev/plans/README.md` with one paragraph: "Development planning docs. Consumer docs are in the top-level README."
   - Edit top-level `README.md` to drop the two "Design document" / "Implementation plan" links at the bottom.
   - No code change.

3. **Flatten API and promote `base`.** This is the largest commit; everything below must land together to keep the build green.
   - **Move + rename:** `internal/robot/base.go` → `mycobot/base.go`. Rename type `Base` → unexported `base`. Add `defaultTimeout` and `config` fields. Preserve the `cmdChan` buffer of 32, `closeOnce` semantics, `mu` scope, and `commandLoop`'s per-command `UseCRC` mutation.
   - **Add factory seam:** package-level `var openSerial = ...` in `base.go` (see "Serial port factory seam" above).
   - **Absorb subsystem methods onto `*base`:** every method currently defined on `MechArm270`, `Motion`, `IO`, `Servo` moves to `*base`. Validation that needs joint count/limits reads from `b.config`.
   - **Update `option.go`:** `Option` becomes `func(*base)`. `WithBaudRate` and `WithCRC` switch their parameter type accordingly.
   - **Shrink `mecharm270.go`** to the 10-line wrapper from "Model types".
   - **Delete:** `motion.go`, `io.go`, `servo.go`, `robot.go`, `internal/robot/` (now empty after the move).
   - **Update tests:** `mecharm270_test.go` drops the `arm.Motion`/`arm.IO`/`arm.Servo` `assert.NotNil` checks and any other existence assertions; `integration_test.go` rewrites every `arm.Motion.X` / `arm.IO.X` / `arm.Servo.X` call site to `arm.X`. Delete `motion_test.go`, `io_test.go`, `servo_test.go`.
   - **Update `README.md`:** the "Motion Subsystem", "IO Subsystem", and "Servo Subsystem" sections collapse into a flat method list. Same content, no `arm.X.Y` indirection. The "API Overview" snippet at the top loses the `arm.Motion` / `arm.IO` / `arm.Servo` field documentation.
   - Build + integration green at end of commit.

4. **Trim errors.** Now safe because `internal/robot/` no longer exists.
   - Delete `RobotError` and `errors_test.go` cases that test it.
   - Delete every unused sentinel listed in "Breaking changes".
   - Inline the two surviving sentinels (`ErrRobotClosed`, `ErrNotConnected`) directly into root `errors.go`; delete `internal/errors/` subpackage.
   - Update `base.go`'s import to drop `internal/errors`.

5. **Add `WithDefaultTimeout`** option in `option.go`, wire it into `base.defaultTimeout`, and verify the resolution chain in `commandLoop`. Document in README options section.

6. **Add fake-serial `base_test.go`** with the eight tests from the test strategy section. No production code changes beyond the `openSerial` seam already in place from step 3.

### Regression checks

- `go build ./...` and `go test ./...` green at the end of every step.
- Integration suite against the MechArm 270 after step 3 and again after step 6 — these are the steps that touch transport-adjacent code.

### Validation that the order works

- Step 1 changes only domain enums and their references; no transport touched.
- Step 2 is filesystem-only.
- Step 3 keeps `internal/errors` intact; `base.go` after the move still imports it.
- Step 4 only proceeds because `internal/robot/` no longer references `internal/errors` — `base.go` is now in root and imports the local `errors.go` it just gained sentinels in.
- Steps 5 and 6 only touch additive code on top of the new `*base`.

## Risks

- **Integration test flakiness.** The existing integration tests have timing sensitivity (see the `MechArm 270 firmware quirks` note in project memory). The refactor does not change transport behavior, so pass/fail patterns should match before/after. If they diverge, step 3 is the regression point to bisect.
- **Source-incompatible API change.** Every documented call pattern that goes through `arm.Motion` / `arm.IO` / `arm.Servo`, every reference to the moved enums (`mycobot.PinMode`, `mycobot.CoordAxis`, etc.), the `Robot` interface, `RobotError`, and the deleted sentinels all break. The repo has one user today (the author, per project memory) and the library is pre-1.0; breaking changes are acceptable. README and integration tests are updated in the same commit (step 3) so the in-tree examples stay correct.
- **Order-sensitive migration.** Steps 3 and 4 must land in this order — see "Validation that the order works" — or the build cycles. Steps 1, 2, 5, 6 are independent and reorderable, but the recommended order is what's listed.

## Open questions

None at present. All seven review items have decisions. Anything discovered during implementation gets raised for confirmation rather than decided inline.

## References

- Original review output: conversation transcript 2026-04-17 (four-agent `/simplify` review).
- Prior planning docs: `dev/plans/2025-11-23-mycobot-go-port-design.md`, `dev/plans/2026-04-15-successor-plan-design.md` (after move).
- pymycobot source checked out at `../pymycobot/` — `generate.py`, `mecharm270.py`, `common.py`.
- Project memory: `~/.claude/projects/-Users-nick-hehr-src-mycobot-go/memory/project_mecharm270_firmware_quirks.md`.
