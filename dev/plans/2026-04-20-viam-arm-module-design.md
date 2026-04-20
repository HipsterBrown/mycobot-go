# Viam arm module for MyCobot / MechArm (design)

Date: 2026-04-20
Status: Draft, ready for implementation planning
Scope: New repo `github.com/hipsterbrown/viam-mycobot` — a Viam registry module that consumes `github.com/hipsterbrown/mycobot-go` to expose Elephant Robotics arms as Viam `arm.Arm` components.

## Goal

Ship a Go-based Viam module that lets users configure a MechArm 270 (and, incrementally, other MyCobot-family arms) as an `arm.Arm` component on a Viam machine. The module must support Viam's motion service by providing a URDF-based kinematics model and honoring the standard `arm.Arm` interface, including Viam-side IK for `MoveToPosition`.

## Non-goals

- No client-side trajectory blending or dynamics modeling.
- No mesh-based collision geometry — primitives only.
- No automated hardware-in-the-loop CI.
- Registry publishing (`viam module create` / `viam module upload`) is out of scope for this design.

## Decisions locked in during brainstorming

1. **Separate repo**: `github.com/hipsterbrown/viam-mycobot`, imports `mycobot-go`.
2. **Multi-model from day one**: share a single `arm.Arm` implementation across MyCobot variants; MechArm 270 ships first.
3. **Kinematics format**: URDF, embedded via `//go:embed`. SVA JSON conversion deferred.
4. **`MoveToPosition` strategy**: Viam-side IK against the URDF, then `client.SendAngles`. The URDF is the single source of truth for kinematics.
5. **Scaffolding**: start from `viam module generate` output; prune defaults as needed.
6. **Model triple**: `hipsterbrown:mycobot:<model>`, e.g. `hipsterbrown:mycobot:mecharm270`.

## Repo layout

The `viam module generate` Go scaffold puts `main.go` in `cmd/module/` and the resource implementation at the repo root:

```
viam-mycobot/
├── .github/workflows/deploy.yml    # from generator, pruned
├── cmd/module/main.go              # ModularMain — registers every model from models.go
├── mycobot.go                      # arm.Arm implementation, shared across models
├── mycobot_test.go
├── config.go                       # Config struct + Validate
├── config_test.go
├── models.go                       # model registry: triple → mycobot-go ctor + URDF bytes
├── kinematics.go                   # //go:embed URDFs, builds referenceframe.Model per model
├── kinematics/
│   └── mecharm270.urdf             # pruned (no meshes); FK-validated
├── go.mod
├── go.sum
├── Makefile                        # generator default
├── meta.json                       # declares mecharm270 model; others added as they land
├── build.sh
├── setup.sh
├── README.md
└── dev/plans/                      # mirrors mycobot-go convention for design/impl docs
```

### Multi-model sharing

`mycobot.go` holds one `Arm` struct that depends on a narrow Go interface — the opcode subset the Viam arm API consumes from mycobot-go:

```go
type client interface {
    Open(ctx context.Context) error
    Close() error
    PowerOn(ctx context.Context) error
    SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error
    GetAngles(ctx context.Context) (types.Angles, error)
    IsInPosition(ctx context.Context, data []float64, flag types.PositionFlag) (bool, error)
    Stop(ctx context.Context) error
    // DoCommand passthroughs add more methods as needed:
    SetColor(ctx context.Context, r, g, b byte) error
    // ... jog, pin, servo methods consumed only by DoCommand
}
```

`models.go` calls `resource.RegisterComponent` once per supported model. Each registration closes over a `ClientFactory func(port string, opts ...mycobot.Option) client` (wrapping `mycobot.NewMechArm270`, etc.) and the URDF bytes, then delegates to a shared `newArm(ctx, deps, conf, logger, factory, urdf)` in `mycobot.go`.

Adding a new model = one entry in `models.go` + one URDF file in `kinematics/`. The `models.go` entry references the embedded URDF variable by name, so forgetting to add the URDF is a compile error.

## Config surface

```go
type Config struct {
    SerialPort     string `json:"serial_port"`                // required
    BaudRate       int    `json:"baud_rate,omitempty"`        // default: model's DefaultBaud
    UseCRC         bool   `json:"use_crc,omitempty"`          // default: false
    DefaultTimeout string `json:"default_timeout,omitempty"`  // duration string; default: "1s"
}
```

`Validate` checks:
- `serial_port` is non-empty.
- `baud_rate`, if set, is in the model's `SupportedBaud` slice (via a model parameter passed to `Validate`).
- `default_timeout`, if set, parses via `time.ParseDuration`.

`Validate` returns `([]string{}, nil)` — no implicit dependencies on other resources.

Per-model `resource.Registration.Constructor` closes over:
1. The model's mycobot-go client factory.
2. The embedded URDF bytes.
3. The model-specific `SupportedBaud` list (for config validation).

And delegates to a shared constructor in `mycobot.go`.

## Arm API wiring

`mycobot.go` implements the full `arm.Arm` interface:

| Viam method | Implementation |
|---|---|
| `EndPosition(ctx, extra)` | Compute FK on cached URDF model using current `JointPositions` → `spatialmath.Pose`. |
| `MoveToPosition(ctx, pose, extra)` | Run Viam-side IK against URDF (via the motionplan/referenceframe packages), then drive to the resulting joints via `MoveToJointPositions` (shares the same completion-polling path). |
| `MoveToJointPositions(ctx, positions, extra)` | Convert `[]referenceframe.Input` from radians to degrees, validate each against model `JointLimits`, call `client.SendAngles` at the resolved speed (see "Speed resolution" below), then block on a single per-move completion poll (see "Firmware quirk handling") until `IsInPosition` is true, ctx is cancelled, or `Stop` is called. |
| `MoveThroughJointPositions(ctx, positions, options, extra)` | Loop over waypoints calling `MoveToJointPositions`. No firmware-level blending. |
| `JointPositions(ctx, extra)` | `client.GetAngles` → convert degrees to radians → `[]referenceframe.Input`. |
| `Stop(ctx, extra)` | `client.Stop`; clear in-flight-move flag. |
| `IsMoving(ctx)` | Local `atomic.Bool` updated by `SendAngles` caller and a background poll goroutine (see "Firmware quirk handling" below). |
| `Kinematics(ctx)` | Return cached `referenceframe.Model` built from embedded URDF. |
| `Geometries(ctx, extra)` | Return geometries from the kinematics model. |
| `CurrentInputs(ctx)` | Delegate to `JointPositions`. |
| `GoToInputs(ctx, inputs ...[]referenceframe.Input)` | Delegate to `MoveToJointPositions` for each input set. |
| `Name()` / `Close(ctx)` / `DoCommand(ctx, cmd)` / `Reconfigure(ctx, deps, conf)` | Standard. |

### Firmware quirk handling

`mycobot-go` carries a known firmware quirk: query replies (`GET_ANGLES`, `GET_ENCODERS`, `IS_MOVING`) stall while the arm is in motion. Calling `IsMoving` directly on the wire would block the entire command loop during a move.

Workaround: `IsMoving` reads a module-owned `atomic.Bool` rather than querying the firmware. The flag is:
- Set to `true` by `MoveToJointPositions` / `MoveToPosition` / `MoveThroughJointPositions` immediately before calling `SendAngles`.
- Cleared by a short-lived per-move goroutine that polls `client.IsInPosition(ctx, target, PositionAngles)` at a modest interval (default 100 ms) until it returns true or ctx is cancelled.

This matches how Viam motion service expects `IsMoving` to behave (cheap, non-blocking) and avoids the stall trap.

### Speed resolution

`mycobot-go`'s `SendAngles` requires a `types.Speed` (0-100). The module resolves it per move in this order:

1. If the call's `extra map[string]interface{}` has a numeric `"speed"` key, clamp to 0-100 and use it.
2. Otherwise, use the module's current default speed.

Default speed starts at `types.SpeedMedium` (50). A `DoCommand` entry `{"command": "set_default_speed", "speed": <int>}` updates it atomically at runtime. No `speed` field is added to `Config` — speed is per-move, not a static machine parameter.

### Angle units

`mycobot-go` exposes degrees; Viam `referenceframe.Input` is radians for revolute joints. `mycobot.go` is the only place conversion happens.

### DoCommand passthroughs

Expose non-arm-API mycobot-go features via `DoCommand`'s `command` key:

- `set_color` — `{r, g, b}` → `client.SetColor`
- `set_pin_mode` / `set_digital_output` / `get_digital_input`
- `jog_angle` / `jog_coord` / `jog_stop`
- `release_servo` / `focus_servo`
- `set_default_speed` — updates the default move speed (see "Speed resolution" above)

Each command parses its own args and returns a result map. Unknown commands return an error.

## Kinematics

### Source

Start from Elephant Robotics' public URDF for the MechArm 270 (published in their `mycobot_ros` repo on GitHub). Prune to Viam's needs:

1. Keep `<link>` / `<joint>` elements with their `<origin>`, `<axis>`, `<limit>`.
2. Drop `<visual>` and mesh references — Viam doesn't render them, and dropping meshes keeps the repo binary-free and avoids license attribution on mesh STL files.
3. Keep `<collision>` blocks but replace mesh refs with primitive `<geometry>` shapes (box / cylinder) sized from published arm dimensions. This gives motion planning usable collision volumes without shipping binary meshes.
4. Reconcile joint limits with `mycobot-go`'s `modelConfigs` (±165° for J1–J5, ±175° for J6 on MechArm 270). If upstream URDF differs, the mycobot-go values win — they match firmware-reported limits.

### Embedding & loading

```go
//go:embed kinematics/mecharm270.urdf
var mecharm270URDF []byte
```

On each `resource.Registration.Constructor` invocation, the URDF parses once into a `referenceframe.Model` via `urdf.ParseModelXMLString(bytes, modelName)` and is cached on the `Arm` struct. `Kinematics()` returns the cached model.

### Validation

`mycobot_test.go` includes an FK round-trip check per model:

1. Load the URDF into a `referenceframe.Model`.
2. For a table of joint configurations (zero pose, two canonical non-zero configurations), compute FK via `referenceframe.ComputePosition`.
3. Assert resulting poses are finite and within the arm's published reach envelope.

This catches gross URDF authoring errors without requiring hardware. Firmware-vs-URDF TCP comparison is deferred to ad-hoc manual testing against real hardware.

### Per-model files

- `kinematics/mecharm270.urdf` — lands with v0 of the module.
- Other MyCobot variants land incrementally as `mycobot-go` adds support. Each new model introduces one file and one `models.go` entry.

### `meta.json` per-model declaration

`meta.json` declares one `models` entry per supported model triple (standard Viam pattern). v0 ships a single entry for `hipsterbrown:mycobot:mecharm270`; each future model adds another entry plus its URDF. The binary is shared — all models run from the same `cmd/module/main.go`.

## Lifecycle and error handling

### Construction

The shared constructor:
1. Builds the mycobot-go client via the injected factory with user-configured options (`WithBaudRate`, `WithCRC`, `WithDefaultTimeout`).
2. Calls `client.Open(ctx)`.
3. Calls `client.PowerOn(ctx)`.
4. Parses the URDF into a `referenceframe.Model`.
5. Returns the `Arm` struct.

Any error from steps 1–4 is returned to Viam; the runtime will retry reconfigure.

### Reconfigure

`Reconfigure` compares the new `Config` against the current one. If any transport-affecting field changed (`serial_port`, `baud_rate`, `use_crc`, `default_timeout`), close the old client, construct a new one, and re-open. Otherwise no-op (the kinematics model is immutable per model).

Implementation note: the comparison uses a single `transportKey()` helper that derives a comparable value from the `Config`. Adding a transport-affecting field later is a one-line change in that helper — callers don't have to update a list.

### Close

`Close(ctx)` calls `client.Close()` and cancels the background move-completion polling goroutine, if any.

### Errors

Wrap mycobot-go errors with `fmt.Errorf("mycobot <op>: %w", err)` so callers preserve `errors.Is(err, mycobot.ErrNotConnected)` and friends. Out-of-range joint positions return a `resource.Name`-qualified error *before* hitting the wire. Context cancellation during a move calls `client.Stop` before returning `ctx.Err()` — we don't leave the arm mid-trajectory when the caller bails.

### Concurrency

`mycobot-go`'s base already serializes serial I/O via its command goroutine, so the module adds no second mutex on wire access. Module-owned state:
- `atomic.Bool` for `IsMoving`.
- Per-move polling goroutine, cancelled by `Close` or `Stop`.
- `sync.Mutex` around cached reconfigure inputs (only touched on `Reconfigure`).

## Testing

Three layers, all runnable via `go test ./...`:

1. **Config** (`config_test.go`) — `Validate` happy paths and every failure case: missing `serial_port`, unsupported baud, unparseable timeout.
2. **Arm with fake client** (`mycobot_test.go`) — define the narrow `client` interface, inject a fake that records calls. Cover:
   - `MoveToJointPositions` converts radians → degrees correctly and calls `SendAngles` with the right `types.Speed`.
   - Joint-limit validation rejects out-of-range targets before any wire call.
   - `Stop` sends `client.Stop` and clears the `IsMoving` flag.
   - `DoCommand` dispatches to the correct mycobot-go method for each `command` value.
   - `Reconfigure` closes and re-opens when serial params change, no-ops otherwise.
3. **Kinematics** (`mycobot_test.go`) — FK round-trip table per registered model.

Integration testing against real hardware is manual and out of scope for this module's CI.

## Build and publish

- `Makefile` and `build.sh` from the generator produce `module.tar.gz`.
- `.github/workflows/deploy.yml` uses Viam's cloud-build action on tag push.
- Registry publishing (`viam module create` / `viam module upload`) is operator work outside this design.

## Open items deferred

- SVA JSON kinematics (may replace URDF later for Viam-native consistency).
- Additional MyCobot / MyPalletizer models as `mycobot-go` grows support.
- Hardware-in-the-loop CI, if it ever becomes worth the infrastructure.
- Firmware-vs-URDF TCP calibration check as an optional opt-in tool.
