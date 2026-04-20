# Viam arm module for MyCobot / MechArm — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Also: follow @superpowers:test-driven-development for every task that touches Go code, and @superpowers:verification-before-completion before claiming any task "done".

**Goal:** Ship a new `github.com/hipsterbrown/viam-mycobot` Viam registry module that exposes MechArm 270 (and future MyCobot-family arms) as Viam `arm.Arm` components, backed by `github.com/hipsterbrown/mycobot-go`, with URDF kinematics for motion planning.

**Architecture:** Scaffold via `viam module generate` into a new sibling repo `~/src/viam-mycobot/`. The repo follows Viam's standard Go layout: `cmd/module/main.go` + resource implementation at repo root. One shared `Arm` struct implements the full `arm.Arm` interface, depending on a narrow `client` interface that abstracts `mycobot-go`. A per-model `models.go` entry wires the interface to the concrete `*mycobot.MechArm270` (etc.) and an embedded URDF. Viam-side IK against the URDF drives `MoveToPosition`; `SendAngles` drives `MoveToJointPositions`. `IsMoving` uses a local `atomic.Bool` to sidestep the firmware's query-stall-during-motion quirk.

**Tech Stack:** Go 1.25, `go.viam.com/rdk` (arm, referenceframe, motionplan, spatialmath, resource, module), `github.com/hipsterbrown/mycobot-go`, `viam` CLI ≥ 0.122.0, `//go:embed` for URDF bundling.

**Spec:** `/Users/nick.hehr/src/mycobot-go/dev/plans/2026-04-20-viam-arm-module-design.md`

**Working directory convention:** Every `cd`, file path, and command in this plan is relative to the user's home on this machine. The new module lives at `/Users/nick.hehr/src/viam-mycobot/`. The source library (`mycobot-go`) is at `/Users/nick.hehr/src/mycobot-go/`. The plan doc itself lives in the mycobot-go repo under `dev/plans/`; commits to the new `viam-mycobot` repo happen in that repo's own history.

---

## File Structure (new repo)

Target layout after all tasks:

```
~/src/viam-mycobot/
├── .github/workflows/deploy.yml    # from generator, pruned
├── cmd/module/main.go              # module.ModularMain(arm.API, models from models.go)
├── mycobot.go                      # Arm struct implementing arm.Arm — shared across models
├── mycobot_test.go                 # unit tests using fake client + FK round-trip
├── config.go                       # Config struct + Validate
├── config_test.go
├── models.go                       # per-model resource.Registration entries
├── kinematics.go                   # //go:embed URDFs, referenceframe.Model loader
├── kinematics/
│   └── mecharm270.urdf             # pruned (no meshes); FK-validated
├── go.mod
├── go.sum
├── Makefile                        # generator default
├── meta.json                       # declares hipsterbrown:mycobot:mecharm270
├── build.sh
├── setup.sh
├── README.md
└── .gitignore
```

**Responsibility per file:**
- `config.go` — JSON schema + validation; no runtime behavior.
- `kinematics.go` — URDF bytes + a `loadModel(name)` helper returning `referenceframe.Model`.
- `models.go` — one `init()` that registers each model triple, closing over its factory + URDF bytes.
- `mycobot.go` — the single `Arm` implementation shared by every model. Takes a `client` interface + a `referenceframe.Model` at construction.
- `cmd/module/main.go` — generated entry point; minimal.

---

## Conventions

**TDD rhythm per task:**

1. Write failing test.
2. Run test, confirm it fails for the expected reason.
3. Write minimal implementation.
4. Run test, confirm pass.
5. `go vet ./...` + `go test ./...` (full suite must stay green).
6. Commit.

**Commit message style:** match `mycobot-go` conventions — conventional commits (`feat(scope): …`, `test(scope): …`, `refactor(scope): …`).

**Module dependency on mycobot-go:** use a `replace` directive in `go.mod` pointing at `../mycobot-go` until mycobot-go publishes a tagged version. This keeps the two repos in lockstep without requiring the library to tag for every iteration.

---

## Tasks

### Task 1: Scaffold the new repo

**Files:**
- Create: `~/src/viam-mycobot/` (entire tree from generator)

- [ ] **Step 1: Verify Viam CLI is ≥ 0.122.0 and user is logged in**

Run:
```bash
viam version
viam whoami 2>&1 | head -3
```

Expected: version prints a `0.122.x` or later line; `whoami` prints the logged-in user (or "not logged in", in which case run `viam login`).

- [ ] **Step 2: Generate the scaffold into a scratch directory**

Run:
```bash
mkdir -p ~/src/_scratch-viam-gen && cd ~/src/_scratch-viam-gen && \
viam module generate \
  --name=mycobot \
  --language=go \
  --visibility=public \
  --public-namespace=hipsterbrown \
  --resource-subtype=arm \
  --model-name=mecharm270
```

Expected: creates `~/src/_scratch-viam-gen/mycobot/` with `cmd/module/main.go`, `mecharm270.go`, `go.mod`, `meta.json`, `Makefile`, `build.sh`, `setup.sh`, `.github/workflows/deploy.yml`, `README.md`.

- [ ] **Step 3: Move the scaffold to the final repo path**

Run:
```bash
mv ~/src/_scratch-viam-gen/mycobot ~/src/viam-mycobot && \
rmdir ~/src/_scratch-viam-gen
```

Expected: `~/src/viam-mycobot/` now holds the scaffold.

- [ ] **Step 4: Initialize git**

Run:
```bash
cd ~/src/viam-mycobot && \
git init && \
git add . && \
git commit -m "chore: scaffold viam-mycobot via viam module generate"
```

Expected: initial commit with all generated files.

- [ ] **Step 5: Add `.gitignore`**

Create `~/src/viam-mycobot/.gitignore`:

```
module.tar.gz
/bin/
*.test
.DS_Store
.claude/
```

Run:
```bash
cd ~/src/viam-mycobot && git add .gitignore && git commit -m "chore: add .gitignore"
```

---

### Task 2: Pin dependencies and wire mycobot-go

**Files:**
- Modify: `~/src/viam-mycobot/go.mod`

- [ ] **Step 1: Confirm Go version and add mycobot-go as a dependency with a local replace**

Run:
```bash
cd ~/src/viam-mycobot && \
go mod edit -go=1.25 && \
go mod edit -require=github.com/hipsterbrown/mycobot-go@v0.0.0-unpublished && \
go mod edit -replace=github.com/hipsterbrown/mycobot-go=../mycobot-go && \
go mod tidy
```

Expected: `go.mod` now has a `require` line for mycobot-go, a `replace` line to `../mycobot-go`, and `go.sum` is regenerated. `go mod tidy` completes without errors. (`go mod tidy` pulls the real version resolution from the local path, so the `v0.0.0-unpublished` placeholder gets rewritten to a pseudo-version.)

- [ ] **Step 2: Smoke-test the import**

Create a temporary `~/src/viam-mycobot/smoke_test.go`:

```go
package mycobot

import (
	"testing"

	"github.com/hipsterbrown/mycobot-go"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestMycobotImport(t *testing.T) {
	// Compile-time check: ensure the constructor exists and accepts Option.
	_ = func() *mycobot.MechArm270 { return mycobot.NewMechArm270("/dev/null") }
	_ = types.ModelMechArm270
}
```

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestMycobotImport -v ./...
```

Expected: PASS. If FAIL with "package mycobot declared in multiple files" from the generated `mecharm270.go`, that's addressed in Task 3.

- [ ] **Step 3: Remove the smoke test and commit the go.mod changes**

Run:
```bash
rm ~/src/viam-mycobot/smoke_test.go && \
cd ~/src/viam-mycobot && \
git add go.mod go.sum && \
git commit -m "chore: depend on local mycobot-go via replace directive"
```

---

### Task 3: Replace generated stub with skeleton files

The generator produces a single `mecharm270.go` containing a stub arm implementation. We'll delete it and create empty skeleton files for each responsibility, then fill them in subsequent tasks.

**Files:**
- Delete: `~/src/viam-mycobot/mecharm270.go`
- Create: `~/src/viam-mycobot/config.go`
- Create: `~/src/viam-mycobot/kinematics.go`
- Create: `~/src/viam-mycobot/models.go`
- Create: `~/src/viam-mycobot/mycobot.go`

- [ ] **Step 1: Delete the stub**

Run:
```bash
rm ~/src/viam-mycobot/mecharm270.go
```

- [ ] **Step 2: Create empty skeletons**

Create `~/src/viam-mycobot/config.go`:

```go
// Package mycobot provides a Viam module that exposes Elephant Robotics MyCobot-family
// arms as Viam arm.Arm components via the github.com/hipsterbrown/mycobot-go library.
package mycobot
```

Create `~/src/viam-mycobot/kinematics.go`:

```go
package mycobot
```

Create `~/src/viam-mycobot/models.go`:

```go
package mycobot
```

Create `~/src/viam-mycobot/mycobot.go`:

```go
package mycobot
```

- [ ] **Step 3: Verify the package still compiles**

Run:
```bash
cd ~/src/viam-mycobot && go build ./...
```

Expected: no output (success).

- [ ] **Step 4: Replace `cmd/module/main.go` with a placeholder that will be finalized in Task 14**

`module.ModularMain` takes `...resource.APIModel`, not an `API` — each model must be explicitly listed. Since we haven't registered any models yet, write a placeholder main that compiles but calls `ModularMain` with an empty list for now. We'll revisit in Task 14 after `MechArm270Model` exists.

Replace `~/src/viam-mycobot/cmd/module/main.go`:

```go
// Package main is the viam-mycobot module entry point.
package main

import (
	"go.viam.com/rdk/module"

	// Blank import so the mycobot package's init() runs and registers models.
	_ "github.com/hipsterbrown/viam-mycobot"
	// Blank import to ensure the global motion planner is registered. Without
	// this, motionplan.GetGlobal() panics the first time MoveToPosition runs.
	_ "go.viam.com/rdk/motionplan/armplanning"
)

func main() {
	// Models will be added in Task 14.
	module.ModularMain()
}
```

- [ ] **Step 5: Confirm `go build ./...` still succeeds and commit**

Run:
```bash
cd ~/src/viam-mycobot && go build ./... && \
git add -A && \
git commit -m "refactor: split generator stub into empty skeleton files"
```

Expected: build succeeds. Commit created.

---

### Task 4: Config struct + validation helper (TDD)

Viam's `resource.Registration[T, C]` takes a `ConfigT` whose `Validate(path) ([]string, []string, error)` method is called by the runtime. The shared `Config` can't know per-model supported baud rates, so the pattern is:

- `Config` holds shared fields and an unexported `validate(supportedBaud []int) error` helper — but no `Validate` method directly, so the runtime won't call it for the shared type.
- Each model defines a small wrapper type (e.g. `MechArm270Config struct { Config }`) that implements `Validate(path)` by calling `Config.validate` with its model-specific supported bauds. Registration uses the wrapper.

This keeps validation logic DRY while letting the runtime call a standard method.

**Files:**
- Modify: `~/src/viam-mycobot/config.go`
- Create: `~/src/viam-mycobot/config_test.go`

- [ ] **Step 1: Write failing tests for Config.validate (the internal helper)**

Create `~/src/viam-mycobot/config_test.go`:

```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidate_RequiresSerialPort(t *testing.T) {
	err := (&Config{}).validate([]int{115200, 1000000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "serial_port")
}

func TestConfigValidate_RejectsUnsupportedBaud(t *testing.T) {
	err := (&Config{SerialPort: "/dev/ttyUSB0", BaudRate: 9600}).validate([]int{115200, 1000000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "baud_rate")
}

func TestConfigValidate_AllowsZeroBaud(t *testing.T) {
	err := (&Config{SerialPort: "/dev/ttyUSB0"}).validate([]int{115200, 1000000})
	require.NoError(t, err)
}

func TestConfigValidate_RejectsBadTimeout(t *testing.T) {
	err := (&Config{SerialPort: "/dev/ttyUSB0", DefaultTimeout: "not-a-duration"}).validate([]int{115200})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default_timeout")
}

func TestConfigValidate_AcceptsValidTimeout(t *testing.T) {
	err := (&Config{SerialPort: "/dev/ttyUSB0", DefaultTimeout: "2s"}).validate([]int{115200})
	require.NoError(t, err)
}

func TestConfigValidate_HappyPath(t *testing.T) {
	err := (&Config{SerialPort: "/dev/ttyUSB0", BaudRate: 115200, UseCRC: true, DefaultTimeout: "1s"}).
		validate([]int{115200, 1000000})
	require.NoError(t, err)
}
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestConfigValidate -v
```

Expected: FAIL with `undefined: Config` (or similar — the Config type doesn't exist yet).

- [ ] **Step 3: Implement Config + validate helper**

Replace `~/src/viam-mycobot/config.go`:

```go
// Package mycobot provides a Viam module that exposes Elephant Robotics MyCobot-family
// arms as Viam arm.Arm components via the github.com/hipsterbrown/mycobot-go library.
package mycobot

import (
	"fmt"
	"time"
)

// Config is the shared attribute map for every mycobot arm model. Per-model
// wrappers in models.go implement resource.ConfigValidator by calling validate
// with their model-specific supported baud rates.
type Config struct {
	SerialPort     string `json:"serial_port"`
	BaudRate       int    `json:"baud_rate,omitempty"`
	UseCRC         bool   `json:"use_crc,omitempty"`
	DefaultTimeout string `json:"default_timeout,omitempty"`
}

// validate is the shared validator used by every per-model wrapper's Validate.
// It is unexported so the runtime does not treat Config directly as a ConfigValidator.
func (c *Config) validate(supportedBaud []int) error {
	if c.SerialPort == "" {
		return fmt.Errorf("serial_port is required")
	}
	if c.BaudRate != 0 {
		ok := false
		for _, b := range supportedBaud {
			if b == c.BaudRate {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("baud_rate %d not in supported set %v", c.BaudRate, supportedBaud)
		}
	}
	if c.DefaultTimeout != "" {
		if _, err := time.ParseDuration(c.DefaultTimeout); err != nil {
			return fmt.Errorf("default_timeout %q: %w", c.DefaultTimeout, err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run the tests and confirm pass**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestConfigValidate -v
```

Expected: all 6 tests PASS.

- [ ] **Step 5: Commit**

Run:
```bash
cd ~/src/viam-mycobot && \
git add config.go config_test.go && \
git commit -m "feat(config): add Config with per-model baud validation"
```

---

### Task 5: Define the narrow client interface + fake client

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Create: `~/src/viam-mycobot/fake_client_test.go`

- [ ] **Step 1: Define the client interface**

Append to `~/src/viam-mycobot/mycobot.go`:

```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/types"
)

// client is the narrow surface of mycobot-go used by this module. Models inject
// a concrete implementation (e.g. *mycobot.MechArm270) at registration time.
// Tests inject a fake.
type client interface {
	Open(ctx context.Context) error
	Close() error
	PowerOn(ctx context.Context) error
	Stop(ctx context.Context) error
	SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error
	GetAngles(ctx context.Context) (types.Angles, error)
	IsInPosition(ctx context.Context, data []float64, flag types.PositionFlag) (bool, error)

	// DoCommand passthroughs:
	SetColor(ctx context.Context, r, g, b byte) error
	SetPinMode(ctx context.Context, pin int, mode types.PinMode) error
	SetDigitalOutput(ctx context.Context, pin int, signal types.PinSignal) error
	GetDigitalInput(ctx context.Context, pin int) (types.PinSignal, error)
	JogAngle(ctx context.Context, joint types.JointID, direction types.Direction, speed types.Speed) error
	JogCoord(ctx context.Context, axis types.CoordAxis, direction types.Direction, speed types.Speed) error
	JogStop(ctx context.Context) error
	ReleaseServo(ctx context.Context, joint types.JointID) error
	FocusServo(ctx context.Context, joint types.JointID) error
}
```

- [ ] **Step 2: Verify `*mycobot.MechArm270` satisfies the interface**

Append a compile-time assertion to `~/src/viam-mycobot/mycobot.go`:

```go
// Compile-time check that the real mycobot-go client satisfies our narrow interface.
var _ client = (*mycobotgo.MechArm270)(nil)
```

Add the import at the top (grouped with the existing ones):

```go
	mycobotgo "github.com/hipsterbrown/mycobot-go"
```

Run:
```bash
cd ~/src/viam-mycobot && go build ./...
```

Expected: success. If it fails with a "does not satisfy client" error, surface which method is missing; the mycobot-go README (mycobot-go/README.md) lists the full method set.

- [ ] **Step 3: Create a fake client for tests**

Create `~/src/viam-mycobot/fake_client_test.go`:

```go
package mycobot

import (
	"context"
	"sync"

	"github.com/hipsterbrown/mycobot-go/types"
)

// fakeClient implements client for unit tests. It records calls and returns
// scripted values. All methods are thread-safe.
type fakeClient struct {
	mu sync.Mutex

	// Call log — each entry is an opcode name plus optional args, stored for assertions.
	calls []string

	// Scripted returns:
	angles        types.Angles
	anglesErr     error
	inPosition    bool
	inPositionErr error
	digitalInput  types.PinSignal
	sendAnglesErr error
	openErr       error

	// For verifying conversion: last SendAngles payload.
	lastSendAngles types.Angles
	lastSendSpeed  types.Speed
}

func (f *fakeClient) log(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, s)
}

func (f *fakeClient) Open(ctx context.Context) error  { f.log("Open"); return f.openErr }
func (f *fakeClient) Close() error                     { f.log("Close"); return nil }
func (f *fakeClient) PowerOn(ctx context.Context) error { f.log("PowerOn"); return nil }
func (f *fakeClient) Stop(ctx context.Context) error   { f.log("Stop"); return nil }

func (f *fakeClient) SendAngles(ctx context.Context, a types.Angles, s types.Speed) error {
	f.mu.Lock()
	f.lastSendAngles = append(types.Angles(nil), a...)
	f.lastSendSpeed = s
	f.mu.Unlock()
	f.log("SendAngles")
	return f.sendAnglesErr
}

func (f *fakeClient) GetAngles(ctx context.Context) (types.Angles, error) {
	f.log("GetAngles")
	return f.angles, f.anglesErr
}

func (f *fakeClient) IsInPosition(ctx context.Context, data []float64, flag types.PositionFlag) (bool, error) {
	f.log("IsInPosition")
	return f.inPosition, f.inPositionErr
}

func (f *fakeClient) SetColor(ctx context.Context, r, g, b byte) error { f.log("SetColor"); return nil }
func (f *fakeClient) SetPinMode(ctx context.Context, pin int, m types.PinMode) error {
	f.log("SetPinMode")
	return nil
}
func (f *fakeClient) SetDigitalOutput(ctx context.Context, pin int, s types.PinSignal) error {
	f.log("SetDigitalOutput")
	return nil
}
func (f *fakeClient) GetDigitalInput(ctx context.Context, pin int) (types.PinSignal, error) {
	f.log("GetDigitalInput")
	return f.digitalInput, nil
}
func (f *fakeClient) JogAngle(ctx context.Context, j types.JointID, d types.Direction, s types.Speed) error {
	f.log("JogAngle")
	return nil
}
func (f *fakeClient) JogCoord(ctx context.Context, a types.CoordAxis, d types.Direction, s types.Speed) error {
	f.log("JogCoord")
	return nil
}
func (f *fakeClient) JogStop(ctx context.Context) error          { f.log("JogStop"); return nil }
func (f *fakeClient) ReleaseServo(ctx context.Context, j types.JointID) error {
	f.log("ReleaseServo")
	return nil
}
func (f *fakeClient) FocusServo(ctx context.Context, j types.JointID) error {
	f.log("FocusServo")
	return nil
}
```

- [ ] **Step 4: Build and commit**

Run:
```bash
cd ~/src/viam-mycobot && go build ./... && go test ./... && \
git add mycobot.go fake_client_test.go && \
git commit -m "feat: add narrow client interface and test fake"
```

Expected: build + tests green.

---

### Task 6: Kinematics loader scaffolding (no URDF yet)

**Files:**
- Modify: `~/src/viam-mycobot/kinematics.go`
- Create: `~/src/viam-mycobot/kinematics_test.go`
- Create: `~/src/viam-mycobot/kinematics/.gitkeep`

- [ ] **Step 1: Add a placeholder URDF directory**

Run:
```bash
mkdir -p ~/src/viam-mycobot/kinematics && \
touch ~/src/viam-mycobot/kinematics/.gitkeep
```

- [ ] **Step 2: Write a failing test for the loader**

Create `~/src/viam-mycobot/kinematics_test.go`:

```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A tiny valid 1-joint URDF used only to exercise the loader.
const testURDF = `<?xml version="1.0"?>
<robot name="test-arm">
  <link name="base_link"/>
  <link name="link1"/>
  <joint name="joint1" type="revolute">
    <parent link="base_link"/>
    <child link="link1"/>
    <origin xyz="0 0 0" rpy="0 0 0"/>
    <axis xyz="0 0 1"/>
    <limit lower="-2.87" upper="2.87" effort="1" velocity="1"/>
  </joint>
</robot>`

func TestLoadModel_FromBytes(t *testing.T) {
	model, err := loadModel([]byte(testURDF), "test-arm")
	require.NoError(t, err)
	require.NotNil(t, model)
	assert.Equal(t, 1, len(model.DoF()))
}

func TestLoadModel_RejectsGarbage(t *testing.T) {
	_, err := loadModel([]byte("not xml"), "bogus")
	require.Error(t, err)
}
```

- [ ] **Step 3: Run the tests and confirm failure**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestLoadModel -v
```

Expected: FAIL with `undefined: loadModel`.

- [ ] **Step 4: Implement the loader**

Replace `~/src/viam-mycobot/kinematics.go`:

```go
package mycobot

import (
	"fmt"

	"go.viam.com/rdk/referenceframe"
)

// loadModel parses URDF bytes into a referenceframe.Model. Meshes are not
// loaded — pass a nil mesh map. Collision geometry must be primitive-only in
// the URDF (box/cylinder/sphere), which is enforced by how we author the URDF
// files in the kinematics/ directory (see spec, "Kinematics" section).
func loadModel(urdfBytes []byte, modelName string) (referenceframe.Model, error) {
	mc, err := referenceframe.UnmarshalModelXML(urdfBytes, modelName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("parse urdf: %w", err)
	}
	m, err := mc.ParseConfig(modelName)
	if err != nil {
		return nil, fmt.Errorf("build model: %w", err)
	}
	return m, nil
}
```

- [ ] **Step 5: Run tests and commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestLoadModel -v && \
git add kinematics.go kinematics_test.go kinematics/.gitkeep && \
git commit -m "feat(kinematics): add URDF bytes → referenceframe.Model loader"
```

Expected: both tests PASS.

---

### Task 7: Source and prune MechArm 270 URDF

**Files:**
- Create: `~/src/viam-mycobot/kinematics/mecharm270.urdf`
- Modify: `~/src/viam-mycobot/kinematics.go`
- Modify: `~/src/viam-mycobot/kinematics_test.go`

- [ ] **Step 1: Fetch upstream URDF**

Elephant Robotics publishes URDFs at `https://github.com/elephantrobotics/mycobot_ros`. Locate the MechArm 270 URDF (typically under `mycobot_description/urdf/mecharm/mecharm_270.urdf` or similar — the exact path may vary by ROS distro branch).

Run:
```bash
curl -sL -o /tmp/mecharm270-upstream.urdf \
  "https://raw.githubusercontent.com/elephantrobotics/mycobot_ros/noetic/mycobot_description/urdf/mecharm/mecharm_270.urdf"
head -30 /tmp/mecharm270-upstream.urdf
```

Expected: an XML URDF prints.

**If that URL 404s or the file is unsuitable**, try these in order before giving up:

1. `git clone https://github.com/elephantrobotics/mycobot_ros && rg -i 'mecharm' --type xml mycobot_ros` to locate any MechArm URDF on disk.
2. Check the `mycobot_ros2` sibling repo for a ROS 2 URDF.
3. If neither yields a usable URDF, author a 6-DOF URDF by hand from the MechArm 270 DH parameters published in Elephant Robotics' product documentation. Minimum contents: six revolute joints named `joint1`–`joint6`, each with correct `<origin>`, `<axis>`, and `<limit>` matching `mycobot-go`'s `modelConfigs` (±165° for J1–J5, ±175° for J6), plus a bounding-box `<collision>` per link. Reach ≈ 270mm from base to TCP; use ~50-80mm per link as a starting approximation.

If hand-authoring is required, split that into its own commit before moving to Step 2.

- [ ] **Step 2: Prune the URDF**

Copy `/tmp/mecharm270-upstream.urdf` to `~/src/viam-mycobot/kinematics/mecharm270.urdf` and edit:

1. Delete all `<visual>` elements.
2. In each `<collision>` element, replace `<mesh filename="…"/>` with a primitive: `<box size="0.08 0.08 0.12"/>` for each link (approximate from physical dimensions — refine later). An acceptable first pass is one box per link with size matching the link's bounding volume.
3. Drop `<material>` definitions (they only matter for `<visual>`).
4. Drop `<gazebo>` / `<transmission>` blocks — ROS-only, Viam ignores them.
5. Drop any `<link name="world"/>` / world-joint — Viam provides the world frame.
6. Verify joint `<limit>` values: `lower`/`upper` in radians. MechArm 270 `mycobot-go` caps:
   - Joints 1–5: ±165° = ±2.88 rad
   - Joint 6: ±175° = ±3.05 rad
   If upstream differs, rewrite to match these values (mycobot-go's firmware-reported limits are canonical).
7. Give the `<robot>` element `name="mecharm270"`.

- [ ] **Step 3: Embed the URDF and expose a `mecharm270Model` constructor**

Append to `~/src/viam-mycobot/kinematics.go`:

```go
import _ "embed"

//go:embed kinematics/mecharm270.urdf
var mecharm270URDF []byte

// mecharm270Model returns a fresh referenceframe.Model for the MechArm 270.
// Safe to call per-instance — the bytes are shared but Model values are not.
func mecharm270Model(name string) (referenceframe.Model, error) {
	return loadModel(mecharm270URDF, name)
}
```

(Consolidate the `import` block so `_ "embed"` sits with the other imports, not in a second block.)

- [ ] **Step 4: Add an FK round-trip test**

Append to `~/src/viam-mycobot/kinematics_test.go`:

```go
import (
	"math"

	"go.viam.com/rdk/referenceframe"
)

func TestMechArm270_FKRoundTrip(t *testing.T) {
	m, err := mecharm270Model("mecharm270")
	require.NoError(t, err)
	require.Equal(t, 6, len(m.DoF()))

	cases := [][]referenceframe.Input{
		{0, 0, 0, 0, 0, 0},
		{math.Pi / 4, -math.Pi / 6, math.Pi / 3, 0, math.Pi / 8, -math.Pi / 2},
	}
	for _, inputs := range cases {
		pose, err := m.Transform(inputs)
		require.NoError(t, err)
		require.NotNil(t, pose)
		// Sanity: TCP should be inside a 1m sphere (MechArm 270 has ~270mm reach).
		p := pose.Point()
		mag := math.Sqrt(p.X*p.X + p.Y*p.Y + p.Z*p.Z)
		assert.Less(t, mag, 1000.0, "FK out of envelope for %v", inputs)
	}
}
```

- [ ] **Step 5: Run tests and commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run 'TestLoadModel|TestMechArm270_FKRoundTrip' -v && \
git add kinematics/mecharm270.urdf kinematics.go kinematics_test.go && \
git rm --cached kinematics/.gitkeep 2>/dev/null; true
git add kinematics/.gitkeep 2>/dev/null; true
git commit -m "feat(kinematics): embed pruned MechArm 270 URDF + FK round-trip test"
```

Expected: both FK tests PASS. If `Transform` returns a NaN pose or the magnitude is way out of envelope, the URDF joint axes/origins are misaligned — iterate on the pruning.

---

### Task 8: Arm struct + simple methods (Name, Close, JointPositions, Stop, IsMoving)

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Modify: `~/src/viam-mycobot/mycobot_test.go` (create if absent)

- [ ] **Step 1: Write failing tests**

Create `~/src/viam-mycobot/mycobot_test.go`:

```go
package mycobot

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hipsterbrown/mycobot-go/types"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func newTestArm(t *testing.T, fc *fakeClient) *Arm {
	t.Helper()
	model, err := mecharm270Model("mecharm270-test")
	require.NoError(t, err)

	return &Arm{
		Named:         resource.Name{Name: "test-arm"}.AsNamed(),
		logger:        logging.NewTestLogger(t),
		client:        fc,
		model:         model,
		jointCount:    6,
		jointLimitsDeg: []float64Limit{
			{Min: -165, Max: 165}, {Min: -165, Max: 165}, {Min: -165, Max: 165},
			{Min: -165, Max: 165}, {Min: -165, Max: 165}, {Min: -175, Max: 175},
		},
		defaultSpeed: types.SpeedMedium,
	}
}

func TestArm_IsMoving_DefaultsFalse(t *testing.T) {
	a := newTestArm(t, &fakeClient{})
	moving, err := a.IsMoving(context.Background())
	require.NoError(t, err)
	assert.False(t, moving)
}

func TestArm_JointPositions_ConvertsDegToRad(t *testing.T) {
	fc := &fakeClient{angles: types.Angles{90, 0, -45, 0, 0, 0}}
	a := newTestArm(t, fc)
	positions, err := a.JointPositions(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, positions, 6)
	// referenceframe.Input is a type alias for float64 (radians for revolute joints).
	assert.InDelta(t, math.Pi/2, float64(positions[0]), 1e-6)
	assert.InDelta(t, 0.0, float64(positions[1]), 1e-6)
	assert.InDelta(t, -math.Pi/4, float64(positions[2]), 1e-6)
}

func TestArm_Stop_ForwardsToClient(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	a.moving.Store(true)
	require.NoError(t, a.Stop(context.Background(), nil))
	assert.Contains(t, fc.calls, "Stop")
	assert.False(t, a.moving.Load())
}

func TestArm_Close_CallsClientClose(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	require.NoError(t, a.Close(context.Background()))
	assert.Contains(t, fc.calls, "Close")
}
```

- [ ] **Step 2: Run tests and confirm they fail**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm -v
```

Expected: FAIL — `undefined: Arm`, `float64Limit`, etc.

- [ ] **Step 3: Implement the Arm struct + simple methods**

Replace the non-interface portion of `~/src/viam-mycobot/mycobot.go` (keep the `client` interface and the compile-time assertion). Full file:

```go
package mycobot

import (
	"context"
	_ "embed"
	"math"
	"sync"
	"sync/atomic"

	mycobotgo "github.com/hipsterbrown/mycobot-go"
	"github.com/hipsterbrown/mycobot-go/types"

	commonpb "go.viam.com/api/common/v1"
	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

// client is defined in this file (unchanged).
// ... (keep the interface + var _ client = ... assertion from Task 5)

type float64Limit struct{ Min, Max float64 }

// Arm is the shared arm.Arm implementation used by every mycobot model.
// We implement Close ourselves — don't embed resource.TriviallyCloseable.
type Arm struct {
	resource.Named

	logger         logging.Logger
	client         client
	model          referenceframe.Model
	jointCount     int
	jointLimitsDeg []float64Limit

	mu             sync.Mutex
	currentCfg     Config
	currentFactory func(port string, opts ...mycobotgo.Option) client
	defaultSpeed   types.Speed

	moving       atomic.Bool
	cancelPoller context.CancelFunc // cancels the most recent completion poller
	pollerWG     sync.WaitGroup
}

// Ensure we satisfy arm.Arm at compile time.
var _ arm.Arm = (*Arm)(nil)

// IsMoving returns the cached move-in-flight flag. It NEVER queries the firmware,
// because the MechArm firmware stalls query replies while the arm is moving (see
// mycobot-go docs, firmware-quirks memory).
func (a *Arm) IsMoving(_ context.Context) (bool, error) {
	return a.moving.Load(), nil
}

// JointPositions returns the current joint positions in radians.
func (a *Arm) JointPositions(ctx context.Context, _ map[string]interface{}) ([]referenceframe.Input, error) {
	angles, err := a.client.GetAngles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]referenceframe.Input, len(angles))
	for i, ang := range angles {
		// Input is a float64 alias; revolute joints are in radians.
		out[i] = float64(ang) * math.Pi / 180.0
	}
	return out, nil
}

// CurrentInputs is the framesystem.InputEnabled version of JointPositions.
func (a *Arm) CurrentInputs(ctx context.Context) ([]referenceframe.Input, error) {
	return a.JointPositions(ctx, nil)
}

// Stop halts motion and clears the in-flight-move flag + any polling goroutine.
func (a *Arm) Stop(ctx context.Context, _ map[string]interface{}) error {
	a.mu.Lock()
	if a.cancelPoller != nil {
		a.cancelPoller()
		a.cancelPoller = nil
	}
	a.mu.Unlock()

	err := a.client.Stop(ctx)
	a.moving.Store(false)
	a.pollerWG.Wait()
	return err
}

// Close releases the serial connection.
func (a *Arm) Close(ctx context.Context) error {
	_ = a.Stop(ctx, nil)
	return a.client.Close()
}

// Kinematics returns the cached referenceframe.Model.
func (a *Arm) Kinematics(_ context.Context) (referenceframe.Model, error) {
	return a.model, nil
}

// Geometries returns the geometries at the current joint pose.
func (a *Arm) Geometries(ctx context.Context, _ map[string]interface{}) ([]spatialmath.Geometry, error) {
	inputs, err := a.CurrentInputs(ctx)
	if err != nil {
		return nil, err
	}
	gif, err := a.model.Geometries(inputs)
	if err != nil {
		return nil, err
	}
	return gif.Geometries(), nil
}

// Get3DModels returns an empty map — we do not ship meshes.
func (a *Arm) Get3DModels(_ context.Context, _ map[string]interface{}) (map[string]*commonpb.Mesh, error) {
	return map[string]*commonpb.Mesh{}, nil
}

// EndPosition computes forward kinematics on the current joint inputs.
func (a *Arm) EndPosition(ctx context.Context, _ map[string]interface{}) (spatialmath.Pose, error) {
	inputs, err := a.CurrentInputs(ctx)
	if err != nil {
		return nil, err
	}
	return a.model.Transform(inputs)
}
```

(The movement methods and Reconfigure/DoCommand come in later tasks; leave them unimplemented for now but add stubs so `var _ arm.Arm = (*Arm)(nil)` compiles:)

```go
// Movement — stubs filled in by later tasks.
func (a *Arm) MoveToJointPositions(ctx context.Context, positions []referenceframe.Input, extra map[string]interface{}) error {
	panic("not implemented — see Task 9")
}
func (a *Arm) MoveToPosition(ctx context.Context, pose spatialmath.Pose, extra map[string]interface{}) error {
	panic("not implemented — see Task 10")
}
func (a *Arm) MoveThroughJointPositions(ctx context.Context, positions [][]referenceframe.Input, _ *arm.MoveOptions, _ map[string]any) error {
	panic("not implemented — see Task 10")
}
func (a *Arm) GoToInputs(ctx context.Context, inputs ...[]referenceframe.Input) error {
	panic("not implemented — see Task 10")
}
func (a *Arm) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	panic("not implemented — see Task 12")
}
func (a *Arm) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	panic("not implemented — see Task 13")
}
```

- [ ] **Step 4: Run tests, confirm pass, commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm -v && \
git add mycobot.go mycobot_test.go && \
git commit -m "feat(arm): add Arm struct with read-only + Stop/Close methods"
```

Expected: all 4 `TestArm_*` tests PASS.

---

### Task 9: MoveToJointPositions with joint-limit validation + completion polling

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Modify: `~/src/viam-mycobot/mycobot_test.go`

- [ ] **Step 1: Write failing tests**

Append to `~/src/viam-mycobot/mycobot_test.go`:

```go
func TestArm_MoveToJointPositions_RadToDeg(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	inputs := []referenceframe.Input{math.Pi / 2, 0, -math.Pi / 4, 0, 0, 0}
	require.NoError(t, a.MoveToJointPositions(context.Background(), inputs, nil))
	require.Len(t, fc.lastSendAngles, 6)
	assert.InDelta(t, 90.0, float64(fc.lastSendAngles[0]), 1e-3)
	assert.InDelta(t, -45.0, float64(fc.lastSendAngles[2]), 1e-3)
}

func TestArm_MoveToJointPositions_RejectsOutOfLimit(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	// J6 is capped at ±175° = ±3.05 rad; 3.5 rad is out of range.
	inputs := []referenceframe.Input{0, 0, 0, 0, 0, 3.5}
	err := a.MoveToJointPositions(context.Background(), inputs, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "joint")
	assert.NotContains(t, fc.calls, "SendAngles")
}

func TestArm_MoveToJointPositions_SpeedFromExtra(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	inputs := make([]referenceframe.Input, 6)
	err := a.MoveToJointPositions(context.Background(), inputs, map[string]interface{}{"speed": 20})
	require.NoError(t, err)
	assert.Equal(t, types.Speed(20), fc.lastSendSpeed)
}

func TestArm_MoveToJointPositions_ClearsIsMovingWhenSettled(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	inputs := make([]referenceframe.Input, 6)
	require.NoError(t, a.MoveToJointPositions(context.Background(), inputs, nil))
	moving, _ := a.IsMoving(context.Background())
	assert.False(t, moving, "IsMoving should be false after IsInPosition returns true")
}
```

- [ ] **Step 2: Run tests, confirm failure**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_MoveToJointPositions -v
```

Expected: FAIL with panic from the Task 8 stub.

- [ ] **Step 3: Implement MoveToJointPositions + helpers**

Replace the `MoveToJointPositions` stub in `mycobot.go`:

```go
// pollInterval is how often we check IsInPosition during a move.
const pollInterval = 100 * time.Millisecond

// MoveToJointPositions moves the arm in joint space. Blocks until the firmware
// reports the target position is reached, ctx is cancelled, or Stop is called.
func (a *Arm) MoveToJointPositions(ctx context.Context, positions []referenceframe.Input, extra map[string]interface{}) error {
	if len(positions) != a.jointCount {
		return fmt.Errorf("expected %d joint positions, got %d", a.jointCount, len(positions))
	}

	angles := make(types.Angles, a.jointCount)
	targetDeg := make([]float64, a.jointCount)
	for i, p := range positions {
		// Input is a float64 alias; revolute inputs are radians.
		deg := float64(p) * 180.0 / math.Pi
		lim := a.jointLimitsDeg[i]
		if deg < lim.Min || deg > lim.Max {
			return fmt.Errorf("joint %d: %.2f° out of range [%.2f, %.2f]", i+1, deg, lim.Min, lim.Max)
		}
		angles[i] = types.Angle(deg)
		targetDeg[i] = deg
	}

	speed := resolveSpeed(extra, a.defaultSpeed)

	a.mu.Lock()
	if a.cancelPoller != nil {
		a.cancelPoller() // cancel any still-running poller
	}
	pollCtx, cancel := context.WithCancel(context.Background())
	a.cancelPoller = cancel
	a.mu.Unlock()

	a.moving.Store(true)
	if err := a.client.SendAngles(ctx, angles, speed); err != nil {
		a.moving.Store(false)
		cancel()
		return err
	}

	// Block until IsInPosition reports true, our poll ctx is cancelled, or caller's ctx ends.
	return a.awaitCompletion(ctx, pollCtx, targetDeg, types.PositionAngles)
}

func (a *Arm) awaitCompletion(callerCtx, pollCtx context.Context, target []float64, flag types.PositionFlag) error {
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	defer func() {
		a.moving.Store(false)
		a.mu.Lock()
		a.cancelPoller = nil
		a.mu.Unlock()
	}()

	for {
		select {
		case <-callerCtx.Done():
			_ = a.client.Stop(context.Background())
			return callerCtx.Err()
		case <-pollCtx.Done():
			return nil // Stop was called; treat as success so callers don't error
		case <-t.C:
			ok, err := a.client.IsInPosition(callerCtx, target, flag)
			if err != nil {
				// Query stalls during motion are expected; don't bail.
				continue
			}
			if ok {
				return nil
			}
		}
	}
}

// resolveSpeed reads a numeric "speed" key from extra and clamps to 0-100.
// Returns defaultSpeed if absent or invalid.
func resolveSpeed(extra map[string]interface{}, defaultSpeed types.Speed) types.Speed {
	raw, ok := extra["speed"]
	if !ok {
		return defaultSpeed
	}
	var v float64
	switch n := raw.(type) {
	case int:
		v = float64(n)
	case int64:
		v = float64(n)
	case float64:
		v = n
	default:
		return defaultSpeed
	}
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return types.Speed(v)
}
```

Add imports `fmt` and `time` if not already present.

- [ ] **Step 4: Run the tests and confirm pass**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_MoveToJointPositions -v
```

Expected: all 4 tests PASS. If `ClearsIsMovingWhenSettled` hangs, check that the poll ticker fires fast enough — lower `pollInterval` during testing or inject it.

- [ ] **Step 5: Commit**

Run:
```bash
cd ~/src/viam-mycobot && go test ./... && \
git add mycobot.go mycobot_test.go && \
git commit -m "feat(arm): implement MoveToJointPositions with limit check + completion polling"
```

---

### Task 10: MoveToPosition (Viam IK) + MoveThroughJointPositions + GoToInputs

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Modify: `~/src/viam-mycobot/mycobot_test.go`

- [ ] **Step 1: Write failing test for MoveToPosition**

Append to `mycobot_test.go`:

```go
import "go.viam.com/rdk/spatialmath"

func TestArm_MoveToPosition_RunsIKAndSendsAngles(t *testing.T) {
	fc := &fakeClient{inPosition: true, angles: types.Angles{0, 0, 0, 0, 0, 0}}
	a := newTestArm(t, fc)

	// Pick a pose inside the MechArm 270's reachable envelope.
	// 150mm forward, 50mm up, zero rotation relative to world.
	target := spatialmath.NewPoseFromPoint(r3.Vector{X: 150, Y: 0, Z: 50})

	err := a.MoveToPosition(context.Background(), target, nil)
	require.NoError(t, err)
	assert.Contains(t, fc.calls, "SendAngles", "IK result must be sent to the arm")
}

func TestArm_MoveThroughJointPositions_CallsEach(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	waypoints := [][]referenceframe.Input{
		{0, 0, 0, 0, 0, 0},
		{0.1, 0, 0, 0, 0, 0},
	}
	require.NoError(t, a.MoveThroughJointPositions(context.Background(), waypoints, nil, nil))
	// Two waypoints → two SendAngles calls.
	count := 0
	for _, c := range fc.calls {
		if c == "SendAngles" {
			count++
		}
	}
	assert.Equal(t, 2, count)
}

func TestArm_GoToInputs_DelegatesToMoveThrough(t *testing.T) {
	fc := &fakeClient{inPosition: true}
	a := newTestArm(t, fc)
	waypoints := []referenceframe.Input{0.05, 0, 0, 0, 0, 0}
	require.NoError(t, a.GoToInputs(context.Background(), waypoints))
	assert.Contains(t, fc.calls, "SendAngles")
}
```

Add import: `"github.com/golang/geo/r3"` (used by spatialmath.NewPoseFromPoint).

- [ ] **Step 2: Run tests, confirm failure**

Run:
```bash
cd ~/src/viam-mycobot && go test -run 'TestArm_MoveToPosition|TestArm_MoveThrough|TestArm_GoTo' -v
```

Expected: FAIL — panic from stub implementations.

- [ ] **Step 3: Implement MoveToPosition via Viam motion planner, plus the other methods**

Replace the relevant stubs in `mycobot.go`:

```go
import (
	// add:
	"go.viam.com/rdk/motionplan"
)

// MoveToPosition runs Viam IK against the URDF and drives to the resulting joints.
func (a *Arm) MoveToPosition(ctx context.Context, pose spatialmath.Pose, extra map[string]interface{}) error {
	current, err := a.CurrentInputs(ctx)
	if err != nil {
		return err
	}
	plan, err := motionplan.GetGlobal().PlanFrameMotion(ctx, a.logger, pose, a.model, current, nil, nil)
	if err != nil {
		return fmt.Errorf("motion planning failed: %w", err)
	}
	if len(plan) == 0 {
		return fmt.Errorf("motion planner returned empty plan")
	}
	// Use the final step of the plan as the target.
	return a.MoveToJointPositions(ctx, plan[len(plan)-1], extra)
}

// MoveThroughJointPositions executes waypoints sequentially. No blending —
// each waypoint is a full MoveToJointPositions (blocking on completion).
func (a *Arm) MoveThroughJointPositions(ctx context.Context, positions [][]referenceframe.Input, _ *arm.MoveOptions, extra map[string]any) error {
	for i, goal := range positions {
		if err := a.MoveToJointPositions(ctx, goal, extra); err != nil {
			return fmt.Errorf("waypoint %d: %w", i, err)
		}
	}
	return nil
}

// GoToInputs is the framesystem.InputEnabled variant.
func (a *Arm) GoToInputs(ctx context.Context, inputs ...[]referenceframe.Input) error {
	return a.MoveThroughJointPositions(ctx, inputs, nil, nil)
}
```

- [ ] **Step 4: Run tests and commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm -v && \
git add mycobot.go mycobot_test.go && \
git commit -m "feat(arm): implement MoveToPosition via Viam IK + multi-waypoint moves"
```

Expected: all `TestArm_*` tests PASS. If `MoveToPosition` errors out with "planner returned empty plan", verify the test pose is inside the MechArm 270 reach envelope (the arm's reach is ~270mm).

---

### Task 11: Reconfigure with transport-key comparison

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Modify: `~/src/viam-mycobot/mycobot_test.go`

- [ ] **Step 1: Write failing tests**

Append to `mycobot_test.go`:

```go
func TestArm_Reconfigure_NoOpWhenTransportUnchanged(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	a.currentCfg = Config{SerialPort: "/dev/ttyUSB0", BaudRate: 115200}
	a.currentFactory = func(p string, _ ...mycobotgo.Option) client { return fc }

	// Build a fresh resource.Config with the same attributes.
	newCfg := resource.Config{
		Name:                "arm1",
		ConvertedAttributes: &Config{SerialPort: "/dev/ttyUSB0", BaudRate: 115200},
	}
	require.NoError(t, a.Reconfigure(context.Background(), nil, newCfg))
	assert.NotContains(t, fc.calls, "Close", "serial params unchanged → should NOT rebuild client")
}

func TestArm_Reconfigure_RebuildsWhenPortChanges(t *testing.T) {
	oldClient := &fakeClient{}
	newClient := &fakeClient{}
	a := newTestArm(t, oldClient)
	a.currentCfg = Config{SerialPort: "/dev/ttyUSB0"}
	a.currentFactory = func(p string, _ ...mycobotgo.Option) client { return newClient }

	newCfg := resource.Config{
		Name:                "arm1",
		ConvertedAttributes: &Config{SerialPort: "/dev/ttyUSB1"},
	}
	require.NoError(t, a.Reconfigure(context.Background(), nil, newCfg))
	assert.Contains(t, oldClient.calls, "Close", "port changed → old client must be closed")
	assert.Contains(t, newClient.calls, "Open", "port changed → new client must be opened")
}
```

- [ ] **Step 2: Run tests, confirm failure**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_Reconfigure -v
```

Expected: FAIL — panic from stub.

- [ ] **Step 3: Implement Reconfigure**

Replace the stub:

```go
// transportKey derives a comparable value from Config fields that affect
// the serial transport. Reconfigure rebuilds the client if and only if this
// value changes between old and new configs.
func transportKey(c Config) string {
	return fmt.Sprintf("%s|%d|%t|%s", c.SerialPort, c.BaudRate, c.UseCRC, c.DefaultTimeout)
}

func (a *Arm) Reconfigure(ctx context.Context, _ resource.Dependencies, conf resource.Config) error {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}
	if a.currentFactory == nil {
		return fmt.Errorf("Arm was not constructed with a factory — cannot reconfigure")
	}

	a.mu.Lock()
	sameTransport := transportKey(a.currentCfg) == transportKey(*newConf)
	a.mu.Unlock()

	if sameTransport {
		a.mu.Lock()
		a.currentCfg = *newConf
		a.mu.Unlock()
		return nil
	}

	// Close old client, build new one.
	if a.client != nil {
		_ = a.client.Close()
	}

	opts := buildOptions(*newConf)
	newC := a.currentFactory(newConf.SerialPort, opts...)
	if err := newC.Open(ctx); err != nil {
		return fmt.Errorf("open %s: %w", newConf.SerialPort, err)
	}

	a.mu.Lock()
	a.client = newC
	a.currentCfg = *newConf
	a.mu.Unlock()
	return nil
}

// buildOptions translates Config fields into mycobot-go Options.
func buildOptions(c Config) []mycobotgo.Option {
	var opts []mycobotgo.Option
	if c.BaudRate != 0 {
		opts = append(opts, mycobotgo.WithBaudRate(c.BaudRate))
	}
	if c.UseCRC {
		opts = append(opts, mycobotgo.WithCRC())
	}
	if c.DefaultTimeout != "" {
		if d, err := time.ParseDuration(c.DefaultTimeout); err == nil {
			opts = append(opts, mycobotgo.WithDefaultTimeout(d))
		}
	}
	return opts
}
```

- [ ] **Step 4: Run tests and commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_Reconfigure -v && \
git add mycobot.go mycobot_test.go && \
git commit -m "feat(arm): implement Reconfigure with transport-key comparison"
```

Expected: both tests PASS.

---

### Task 12: DoCommand passthroughs

**Files:**
- Modify: `~/src/viam-mycobot/mycobot.go`
- Modify: `~/src/viam-mycobot/mycobot_test.go`

- [ ] **Step 1: Write failing tests for DoCommand**

Append to `mycobot_test.go`:

```go
func TestArm_DoCommand_SetColor(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	_, err := a.DoCommand(context.Background(), map[string]interface{}{
		"command": "set_color",
		"r":       255, "g": 0, "b": 0,
	})
	require.NoError(t, err)
	assert.Contains(t, fc.calls, "SetColor")
}

func TestArm_DoCommand_SetDefaultSpeed(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	_, err := a.DoCommand(context.Background(), map[string]interface{}{
		"command": "set_default_speed", "speed": 30,
	})
	require.NoError(t, err)
	assert.Equal(t, types.Speed(30), a.defaultSpeed)
}

func TestArm_DoCommand_UnknownCommand(t *testing.T) {
	a := newTestArm(t, &fakeClient{})
	_, err := a.DoCommand(context.Background(), map[string]interface{}{"command": "nope"})
	require.Error(t, err)
}

func TestArm_DoCommand_JogAngle(t *testing.T) {
	fc := &fakeClient{}
	a := newTestArm(t, fc)
	_, err := a.DoCommand(context.Background(), map[string]interface{}{
		"command":   "jog_angle",
		"joint":     1,
		"direction": 1, // DirPositive
		"speed":     20,
	})
	require.NoError(t, err)
	assert.Contains(t, fc.calls, "JogAngle")
}
```

- [ ] **Step 2: Run tests, confirm failure**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_DoCommand -v
```

Expected: FAIL — panic from stub.

- [ ] **Step 3: Implement DoCommand**

Replace the stub with a dispatcher:

```go
func (a *Arm) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	name, _ := cmd["command"].(string)
	switch name {
	case "set_color":
		r := toByte(cmd["r"])
		g := toByte(cmd["g"])
		b := toByte(cmd["b"])
		return nil, a.client.SetColor(ctx, r, g, b)

	case "set_default_speed":
		s := resolveSpeed(cmd, a.defaultSpeed)
		a.mu.Lock()
		a.defaultSpeed = s
		a.mu.Unlock()
		return map[string]interface{}{"speed": int(s)}, nil

	case "set_pin_mode":
		return nil, a.client.SetPinMode(ctx, toInt(cmd["pin"]), types.PinMode(toInt(cmd["mode"])))

	case "set_digital_output":
		return nil, a.client.SetDigitalOutput(ctx, toInt(cmd["pin"]), types.PinSignal(toInt(cmd["signal"])))

	case "get_digital_input":
		sig, err := a.client.GetDigitalInput(ctx, toInt(cmd["pin"]))
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"signal": int(sig)}, nil

	case "jog_angle":
		return nil, a.client.JogAngle(ctx,
			types.JointID(toInt(cmd["joint"])),
			types.Direction(toInt(cmd["direction"])),
			resolveSpeed(cmd, a.defaultSpeed),
		)

	case "jog_coord":
		return nil, a.client.JogCoord(ctx,
			types.CoordAxis(toInt(cmd["axis"])),
			types.Direction(toInt(cmd["direction"])),
			resolveSpeed(cmd, a.defaultSpeed),
		)

	case "jog_stop":
		return nil, a.client.JogStop(ctx)

	case "release_servo":
		return nil, a.client.ReleaseServo(ctx, types.JointID(toInt(cmd["joint"])))

	case "focus_servo":
		return nil, a.client.FocusServo(ctx, types.JointID(toInt(cmd["joint"])))

	default:
		return nil, fmt.Errorf("unknown command %q", name)
	}
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func toByte(v interface{}) byte {
	n := toInt(v)
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	return byte(n)
}
```

- [ ] **Step 4: Run tests and commit**

Run:
```bash
cd ~/src/viam-mycobot && go test -run TestArm_DoCommand -v && \
git add mycobot.go mycobot_test.go && \
git commit -m "feat(arm): implement DoCommand passthroughs"
```

Expected: all 4 tests PASS.

---

### Task 13: Wire models.go with resource.RegisterComponent for MechArm 270

**Files:**
- Modify: `~/src/viam-mycobot/models.go`
- Modify: `~/src/viam-mycobot/mycobot.go` (add `newArm` constructor)

- [ ] **Step 1: Add a shared constructor in mycobot.go**

Append to `mycobot.go`:

```go
// newArm is the shared constructor used by every model registration. The
// per-model Registration calls resource.NativeConfig on its wrapper config
// type and passes the unwrapped shared Config into this function.
func newArm(
	ctx context.Context,
	resName resource.Name,
	cfg Config,
	logger logging.Logger,
	factory func(port string, opts ...mycobotgo.Option) client,
	urdfBytes []byte,
	jointLimits []float64Limit,
) (arm.Arm, error) {
	model, err := loadModel(urdfBytes, resName.ShortName())
	if err != nil {
		return nil, fmt.Errorf("load kinematics: %w", err)
	}

	c := factory(cfg.SerialPort, buildOptions(cfg)...)
	if err := c.Open(ctx); err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.SerialPort, err)
	}
	if err := c.PowerOn(ctx); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("power on: %w", err)
	}

	a := &Arm{
		Named:          resName.AsNamed(),
		logger:         logger,
		client:         c,
		model:          model,
		jointCount:     len(jointLimits),
		jointLimitsDeg: jointLimits,
		currentCfg:     cfg,
		currentFactory: factory,
		defaultSpeed:   types.SpeedMedium,
	}
	return a, nil
}
```

- [ ] **Step 2: Implement models.go with a per-model Config wrapper**

Replace `~/src/viam-mycobot/models.go`:

```go
package mycobot

import (
	"context"

	mycobotgo "github.com/hipsterbrown/mycobot-go"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

// Namespace for all models exposed by this module.
var namespace = resource.NewModelFamily("hipsterbrown", "mycobot")

// MechArm270Model is the registered model triple for the MechArm 270.
var MechArm270Model = namespace.WithModel("mecharm270")

var mechArm270Limits = []float64Limit{
	{Min: -165, Max: 165}, {Min: -165, Max: 165}, {Min: -165, Max: 165},
	{Min: -165, Max: 165}, {Min: -165, Max: 165}, {Min: -175, Max: 175},
}

var mechArm270SupportedBaud = []int{115200, 1000000}

// MechArm270Config is the per-model wrapper that satisfies the runtime's
// ConfigValidator interface by delegating to Config.validate with the
// MechArm 270's supported baud rates.
type MechArm270Config struct {
	Config
}

// Validate implements resource.ConfigValidator.
func (c *MechArm270Config) Validate(path string) ([]string, []string, error) {
	if err := c.Config.validate(mechArm270SupportedBaud); err != nil {
		return nil, nil, err
	}
	return nil, nil, nil
}

func init() {
	resource.RegisterComponent(
		arm.API,
		MechArm270Model,
		resource.Registration[arm.Arm, *MechArm270Config]{
			Constructor: func(
				ctx context.Context,
				_ resource.Dependencies,
				conf resource.Config,
				logger logging.Logger,
			) (arm.Arm, error) {
				wrapper, err := resource.NativeConfig[*MechArm270Config](conf)
				if err != nil {
					return nil, err
				}
				factory := func(port string, opts ...mycobotgo.Option) client {
					return mycobotgo.NewMechArm270(port, opts...)
				}
				return newArm(ctx, conf.ResourceName(), wrapper.Config, logger, factory, mecharm270URDF, mechArm270Limits)
			},
		},
	)
}
```

Why this shape:
- `resource.Registration[T, C]` has no `Validator` field — runtime calls `Validate(path)` on `C` directly. We give each per-model wrapper type a `Validate` that closes over its supported bauds.
- `AttributeMapConverter` is intentionally omitted. The runtime uses reflection to decode `resource.AttributeMap` into the concrete ConfigT (`*MechArm270Config`) based on JSON tags on the embedded `Config`.
- Embedding `Config` in `MechArm270Config` means the JSON attributes in robot configs stay flat — `serial_port`, `baud_rate`, etc. — regardless of which model is being configured.

- [ ] **Step 3: Build and confirm no compile errors**

Run:
```bash
cd ~/src/viam-mycobot && go build ./...
```

Expected: success. If `resource.NewModelFamily` or any other identifier is different in the installed RDK, consult `~/src/rdk/resource/` for the current signature.

- [ ] **Step 4: Write a registration-smoke test**

Append to `mycobot_test.go`:

```go
func TestModelRegistration_MechArm270(t *testing.T) {
	reg, ok := resource.LookupRegistration(arm.API, MechArm270Model)
	require.True(t, ok, "MechArm270 model must be registered with arm.API")
	require.NotNil(t, reg.Constructor)
}

func TestMechArm270Config_Validate(t *testing.T) {
	c := &MechArm270Config{Config: Config{SerialPort: "/dev/ttyUSB0", BaudRate: 115200}}
	_, _, err := c.Validate("path")
	require.NoError(t, err)

	c = &MechArm270Config{Config: Config{SerialPort: "/dev/ttyUSB0", BaudRate: 9600}}
	_, _, err = c.Validate("path")
	require.Error(t, err)
}
```

Imports to ensure are in `mycobot_test.go`:
- `"go.viam.com/rdk/components/arm"`
- `"go.viam.com/rdk/resource"`

Run:
```bash
cd ~/src/viam-mycobot && go test -run 'TestModelRegistration|TestMechArm270Config' -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:
```bash
cd ~/src/viam-mycobot && go test ./... && \
git add models.go mycobot.go mycobot_test.go && \
git commit -m "feat(models): register hipsterbrown:mycobot:mecharm270 with arm.API"
```

---

### Task 14: Wire cmd/module/main.go and meta.json

**Files:**
- Modify: `~/src/viam-mycobot/cmd/module/main.go`
- Modify: `~/src/viam-mycobot/meta.json`

- [ ] **Step 1: Replace main.go with the final form**

`module.ModularMain` takes `...resource.APIModel`. Every registered model must be listed explicitly — the blank import alone is not enough.

Replace `~/src/viam-mycobot/cmd/module/main.go`:

```go
// Package main is the viam-mycobot module entry point.
package main

import (
	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"

	mycobot "github.com/hipsterbrown/viam-mycobot"

	// Pulls in the global motion planner so MoveToPosition's IK path works.
	_ "go.viam.com/rdk/motionplan/armplanning"
)

func main() {
	module.ModularMain(
		resource.APIModel{API: arm.API, Model: mycobot.MechArm270Model},
		// Add one resource.APIModel per future model here.
	)
}
```

Note: the `mycobot` import is by name (not blank), because `main.go` references `mycobot.MechArm270Model`. Its `init()` still registers the component with the global resource registry, which `ModularMain` reads from internally.

- [ ] **Step 2: Update meta.json**

Read the current `~/src/viam-mycobot/meta.json` generated by the CLI. Ensure it contains (at minimum):

```json
{
  "module_id": "hipsterbrown:mycobot",
  "visibility": "public",
  "url": "https://github.com/hipsterbrown/viam-mycobot",
  "description": "Viam arm module for Elephant Robotics MyCobot-family robots (MechArm 270 and more).",
  "models": [
    {
      "api": "rdk:component:arm",
      "model": "hipsterbrown:mycobot:mecharm270",
      "markdown_link": "README.md#mecharm270"
    }
  ],
  "build": {
    "setup": "./setup.sh",
    "build": "make module.tar.gz",
    "path": "module.tar.gz",
    "arch": ["linux/amd64", "linux/arm64", "darwin/arm64"]
  },
  "entrypoint": "bin/viam-mycobot"
}
```

Check against the generator's output — keep any fields it added (like `first_run`) and only adjust the `models`, `url`, and `description` fields. The `entrypoint` must match what the Makefile produces.

- [ ] **Step 3: Verify the local build works**

Run:
```bash
cd ~/src/viam-mycobot && \
make module.tar.gz
```

Expected: the Makefile builds a binary at `bin/viam-mycobot` (or similar) and packages `module.tar.gz`. If the Makefile references a different entrypoint path, align `meta.json`'s `entrypoint` accordingly.

- [ ] **Step 4: Commit**

Run:
```bash
cd ~/src/viam-mycobot && \
git add cmd/module/main.go meta.json && \
git commit -m "feat(module): wire module entrypoint and declare mecharm270 in meta.json"
```

---

### Task 15: README and final verification

**Files:**
- Modify: `~/src/viam-mycobot/README.md`

- [ ] **Step 1: Write a short README**

Replace `~/src/viam-mycobot/README.md` with:

````markdown
# viam-mycobot

Viam arm module for [Elephant Robotics](https://www.elephantrobotics.com/) MyCobot-family robots. Exposes these arms as Viam `arm.Arm` components with URDF-based kinematics for motion planning.

## Supported models

| Model triple | Hardware | Status |
|---|---|---|
| `hipsterbrown:mycobot:mecharm270` | MechArm 270 | v0 |

Additional MyCobot variants will be added as [`mycobot-go`](https://github.com/hipsterbrown/mycobot-go) gains support.

## Configuration

```json
{
  "name": "my-arm",
  "api": "rdk:component:arm",
  "model": "hipsterbrown:mycobot:mecharm270",
  "attributes": {
    "serial_port": "/dev/ttyUSB0",
    "baud_rate": 115200,
    "use_crc": false,
    "default_timeout": "1s"
  }
}
```

| Field | Type | Default | Notes |
|---|---|---|---|
| `serial_port` | string | required | Path to the USB serial device. |
| `baud_rate` | int | `115200` | Must be in the model's supported set (`115200` or `1000000` for MechArm 270). |
| `use_crc` | bool | `false` | Set true for firmware builds that require CRC framing. |
| `default_timeout` | duration string | `"1s"` | Fallback per-command read timeout when ctx has no deadline. |

## DoCommand

Non-arm-API features are exposed via `DoCommand`:

| `command` | Extra keys | Description |
|---|---|---|
| `set_color` | `r`, `g`, `b` | Set Atom LED RGB (0-255). |
| `set_default_speed` | `speed` | Set the module's default move speed (0-100). |
| `set_pin_mode` | `pin`, `mode` | Configure a digital pin. |
| `set_digital_output` | `pin`, `signal` | Drive a digital output high or low. |
| `get_digital_input` | `pin` | Read a digital input. |
| `jog_angle` | `joint`, `direction`, `speed` | Incremental joint movement. |
| `jog_coord` | `axis`, `direction`, `speed` | Incremental cartesian movement. |
| `jog_stop` | — | Stop a JOG move. |
| `release_servo` | `joint` | Free a servo for manual movement. |
| `focus_servo` | `joint` | Re-engage a servo. |

All `MoveTo*` calls also honor a `"speed"` key in the `extra` map (0-100, clamped).

## Development

```bash
# Run tests
go test ./...

# Build the module artifact
make module.tar.gz

# Local reload against a running viam-server
viam module reload-local --part-id <part-id>
```

See `dev/plans/` for design and implementation history.
````

- [ ] **Step 2: Full verification — go vet, tests, build**

Run:
```bash
cd ~/src/viam-mycobot && \
go vet ./... && \
go test ./... && \
go build -o /tmp/viam-mycobot-smoke ./cmd/module && \
file /tmp/viam-mycobot-smoke
```

Expected: vet silent, all tests PASS, build produces a Mach-O 64-bit executable.

- [ ] **Step 3: Commit**

Run:
```bash
cd ~/src/viam-mycobot && \
git add README.md && \
git commit -m "docs: add README with configuration and DoCommand reference"
```

- [ ] **Step 4: Confirm the plan is complete**

Run:
```bash
cd ~/src/viam-mycobot && \
git log --oneline
```

Expected: a clean series of ~15 commits, each one TDD increment or scaffolding step. The module builds, all tests pass, and `hipsterbrown:mycobot:mecharm270` is registered for `rdk:component:arm`.

---

## Success Criteria

The plan is done when ALL of these hold:

1. `cd ~/src/viam-mycobot && go test ./...` exits 0 with every test passing.
2. `cd ~/src/viam-mycobot && go vet ./...` is silent.
3. `cd ~/src/viam-mycobot && make module.tar.gz` succeeds.
4. `git log` in `~/src/viam-mycobot/` shows a clean commit sequence where each commit builds standalone.
5. `meta.json` declares `hipsterbrown:mycobot:mecharm270` with `api=rdk:component:arm`.
6. `kinematics/mecharm270.urdf` has no `<mesh>` references and joint limits match `mycobot-go`'s `modelConfigs`.
7. The `Arm` struct's `IsMoving` never issues a wire query (verified by inspection).
8. `MoveToPosition` routes through Viam's motion planner (never through firmware `SendCoords`).

## Adding future models (post-v0)

Each new MyCobot / MyPalletizer model follows a three-change pattern:

1. Add `kinematics/<model>.urdf` (pruned) and a `//go:embed` block + `<model>Model` constructor in `kinematics.go`.
2. Add a `<Model>Model` variable, a limits slice, a supported-baud slice, and a `resource.RegisterComponent` block in `models.go`.
3. Add a `models` entry in `meta.json`.

No changes to `mycobot.go` are required unless the new model has a different joint count or firmware quirk not covered by the existing `client` interface.
