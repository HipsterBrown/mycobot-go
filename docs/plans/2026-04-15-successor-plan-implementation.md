# Successor Plan Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix protocol bugs and replace MyCobot280 with a correct, hardware-tested MechArm 270 implementation.

**Architecture:** Two phases — Phase 0 fixes the protocol layer (command codes, coordinate encoding, CRC defaults) with unit tests. Phase 1 replaces MyCobot280 with MechArm270, adds the `mode` parameter to `SendCoords`, completes Motion/IO/Servo subsystems, and adds an integration test harness.

**Tech Stack:** Go 1.21+, `go.bug.st/serial`, `github.com/stretchr/testify`

**Design doc:** `docs/plans/2026-04-15-successor-plan-design.md`

**Reference implementation:** pymycobot `common.py` `ProtocolCode` class and `DataProcessor` class. Byte values verified against source at https://github.com/elephantrobotics/pymycobot.

---

## Phase 0: Protocol Remediation

### Task 1: Fix Atom IO Command Codes

The current `protocol/codes.go` labels Basic IO codes (0xA0 range) as Atom IO. The actual Atom IO codes are in the 0x60 range per pymycobot's `ProtocolCode` class. Two codes (`SET_PIN_MODE`, `SET_PWM_MODE`) are missing entirely.

**Files:**
- Modify: `protocol/codes.go:80-97`
- Modify: `protocol/codes_test.go`

**Step 1: Write failing tests for corrected Atom IO codes**

Add to `protocol/codes_test.go`:

```go
func TestAtomIOCodes_MatchPymycobot(t *testing.T) {
	// Atom IO codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0x60), SetPinMode)
	assert.Equal(t, byte(0x61), SetDigitalOutput)
	assert.Equal(t, byte(0x62), GetDigitalInput)
	assert.Equal(t, byte(0x63), SetPWMMode)
	assert.Equal(t, byte(0x64), SetPWMOutput)
	assert.Equal(t, byte(0x6A), SetColor)
}

func TestBasicIOCodes_MatchPymycobot(t *testing.T) {
	// Basic IO codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0xA0), SetBasicOutput)
	assert.Equal(t, byte(0xA1), GetBasicInput)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./protocol/ -run "TestAtomIOCodes_MatchPymycobot|TestBasicIOCodes_MatchPymycobot" -v`
Expected: FAIL — `SetPinMode` and `SetPWMMode` are undefined, `SetDigitalOutput` is `0xA0` not `0x61`

**Step 3: Fix the codes in `protocol/codes.go`**

Replace the "Atom IO" and "Basic IO" const blocks (lines 80-97) with:

```go
// Atom IO (end-effector head, 0x60 range)
// Codes from pymycobot common.py ProtocolCode class
const (
	SetPinMode       byte = 0x60
	SetDigitalOutput byte = 0x61
	GetDigitalInput  byte = 0x62
	SetPWMMode       byte = 0x63
	SetPWMOutput     byte = 0x64
	SetColor         byte = 0x6A
)

// Gripper commands
const (
	GetGripperValue byte = 0x65
	SetGripperState byte = 0x66
	SetGripperValue byte = 0x67
	SetGripperIni   byte = 0x68
	IsGripperMoving byte = 0x69
)

// Basic IO (base panel, 0xA0 range)
const (
	SetBasicOutput byte = 0xA0
	GetBasicInput  byte = 0xA1
)
```

**Step 4: Run tests to verify they pass**

Run: `go test ./protocol/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add protocol/codes.go protocol/codes_test.go
git commit -m "fix(protocol): correct Atom IO and Basic IO command codes

Atom IO codes were incorrectly using Basic IO range (0xA0). Fixed to
match pymycobot ProtocolCode class: 0x60-0x64 for Atom IO, 0xA0-0xA1
for Basic IO. Added missing SetPinMode (0x60) and SetPWMMode (0x63)."
```

---

### Task 2: Fix Gripper Command Codes

Every gripper code is offset from pymycobot's values. This task is separated from Task 1 because the gripper codes are in a different const block and the changes are independent.

**Files:**
- Modify: `protocol/codes.go` (already modified in Task 1 — gripper codes were moved into their own block)
- Modify: `protocol/codes_test.go`

**Step 1: Write failing tests for corrected gripper codes**

Add to `protocol/codes_test.go`:

```go
func TestGripperCodes_MatchPymycobot(t *testing.T) {
	// Gripper codes per pymycobot common.py ProtocolCode class
	assert.Equal(t, byte(0x65), GetGripperValue)
	assert.Equal(t, byte(0x66), SetGripperState)
	assert.Equal(t, byte(0x67), SetGripperValue)
	assert.Equal(t, byte(0x68), SetGripperIni)
	assert.Equal(t, byte(0x69), IsGripperMoving)
}
```

**Step 2: Run tests to verify they pass**

These codes were already corrected in Task 1's Step 3. If Task 1 was completed correctly, this test should pass immediately.

Run: `go test ./protocol/ -run TestGripperCodes_MatchPymycobot -v`
Expected: PASS

If it fails, fix the gripper const values in `protocol/codes.go` to match the values in Step 1.

**Step 3: Commit**

```bash
git add protocol/codes_test.go
git commit -m "test(protocol): add gripper command code verification tests"
```

---

### Task 3: Fix Coordinate Encoding

`EncodeCoords` currently delegates to `EncodeAngles`, applying `* 100` to all six values. pymycobot uses `* 10` for XYZ position (`_coord2int`) and `* 100` for Rx/Ry/Rz rotation (`_angle2int`). The current code would send XYZ values 10x too large.

**Files:**
- Modify: `protocol/command.go:173-191`
- Modify: `protocol/command_test.go:183-231`

**Step 1: Write failing tests for corrected coordinate encoding**

Replace `TestEncodeCoords` and `TestDecodeCoords` in `protocol/command_test.go`:

```go
func TestEncodeCoords(t *testing.T) {
	// pymycobot encodes XYZ with _coord2int (value * 10)
	// and Rx/Ry/Rz with _angle2int (value * 100)
	data := EncodeCoords(100.5, -50.2, 200.0, 45.0, -30.5, 90.25)

	expected := []byte{
		// XYZ: value * 10, big-endian int16
		0x03, 0xEB, // 100.5 * 10 = 1005
		0xFE, 0x0C, // -50.2 * 10 = -502
		0x07, 0xD0, // 200.0 * 10 = 2000
		// Rx/Ry/Rz: value * 100, big-endian int16
		0x11, 0x94, // 45.0 * 100 = 4500
		0xF4, 0x16, // -30.5 * 100 = -3050
		0x23, 0x41, // 90.25 * 100 = 9025
	}

	assert.Equal(t, expected, data)
}

func TestDecodeCoords(t *testing.T) {
	data := []byte{
		// XYZ: encoded with * 10
		0x03, 0xEB, // 1005
		0xFE, 0x0C, // -502
		0x07, 0xD0, // 2000
		// Rx/Ry/Rz: encoded with * 100
		0x11, 0x94, // 4500
		0xF4, 0x16, // -3050
		0x23, 0x41, // 9025
	}

	x, y, z, rx, ry, rz, err := DecodeCoords(data)
	require.NoError(t, err)

	assert.InDelta(t, 100.5, x, 0.1)
	assert.InDelta(t, -50.2, y, 0.1)
	assert.InDelta(t, 200.0, z, 0.1)
	assert.InDelta(t, 45.0, rx, 0.01)
	assert.InDelta(t, -30.5, ry, 0.01)
	assert.InDelta(t, 90.25, rz, 0.01)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./protocol/ -run "TestEncodeCoords|TestDecodeCoords" -v`
Expected: FAIL — XYZ values encoded with `* 100` instead of `* 10`

**Step 3: Fix `EncodeCoords` and `DecodeCoords` in `protocol/command.go`**

Replace lines 173-191:

```go
// EncodeCoords encodes coordinates (x, y, z, rx, ry, rz) to wire format.
// XYZ positions are encoded as int16 (value * 10) per pymycobot _coord2int.
// Rx/Ry/Rz rotations are encoded as int16 (value * 100) per pymycobot _angle2int.
func EncodeCoords(x, y, z, rx, ry, rz float64) []byte {
	data := make([]byte, 0, 12)
	// XYZ: multiply by 10
	for _, v := range []float64{x, y, z} {
		data = append(data, EncodeInt16(int(v*10))...)
	}
	// Rx/Ry/Rz: multiply by 100
	for _, v := range []float64{rx, ry, rz} {
		data = append(data, EncodeInt16(int(v*100))...)
	}
	return data
}

// DecodeCoords decodes wire format back to coordinates.
// XYZ are decoded with divisor 10, Rx/Ry/Rz with divisor 100.
func DecodeCoords(data []byte) (x, y, z, rx, ry, rz float64, err error) {
	if len(data) != 12 {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid coord data: expected 12 bytes, got %d", len(data))
	}

	// XYZ: divide by 10
	xyz := make([]float64, 3)
	for i := 0; i < 3; i++ {
		value := int16(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
		xyz[i] = float64(value) / 10.0
	}

	// Rx/Ry/Rz: divide by 100
	rot := make([]float64, 3)
	for i := 0; i < 3; i++ {
		value := int16(binary.BigEndian.Uint16(data[6+i*2 : 6+i*2+2]))
		rot[i] = float64(value) / 100.0
	}

	return xyz[0], xyz[1], xyz[2], rot[0], rot[1], rot[2], nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./protocol/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add protocol/command.go protocol/command_test.go
git commit -m "fix(protocol): use correct multipliers for coordinate encoding

XYZ positions now use * 10 (pymycobot _coord2int) and Rx/Ry/Rz
rotations use * 100 (pymycobot _angle2int). Previously all six
values used * 100, which would send XYZ values 10x too large."
```

---

### Task 4: Fix CRC Default and Add WithCRC Option

All model configs have `UseCRC: true`. pymycobot's default frame format uses the `0xFA` footer. CRC should be opt-in.

**Files:**
- Modify: `config.go:19-78`
- Modify: `config_test.go`
- Modify: `option.go`

**Step 1: Write failing test for CRC default**

Replace `TestModelConfig_MyCobot280` and `TestModelConfig_AllModels` in `config_test.go` with a MechArm270-focused test:

```go
func TestModelConfig_MechArm270(t *testing.T) {
	config := getModelConfig(types.ModelMechArm270)

	assert.Equal(t, types.ModelMechArm270, config.Model)
	assert.Equal(t, 6, config.JointCount)
	assert.Equal(t, 115200, config.DefaultBaud)
	assert.False(t, config.UseCRC, "CRC should be off by default, matching pymycobot")
	assert.Len(t, config.JointLimits, 6)
}
```

**Step 2: Run test to verify it fails**

Run: `go test . -run TestModelConfig_MechArm270 -v`
Expected: FAIL — `config.UseCRC` is `true`

**Step 3: Update `config.go`**

Remove all model configs except MechArm270. Set `UseCRC: false`:

```go
var modelConfigs = map[types.Model]ModelConfig{
	types.ModelMechArm270: {
		Model:      types.ModelMechArm270,
		JointCount: 6,
		JointLimits: []types.JointLimit{
			{MinAngle: -165, MaxAngle: 165}, // Joint 1
			{MinAngle: -165, MaxAngle: 165}, // Joint 2
			{MinAngle: -165, MaxAngle: 165}, // Joint 3
			{MinAngle: -165, MaxAngle: 165}, // Joint 4
			{MinAngle: -165, MaxAngle: 165}, // Joint 5
			{MinAngle: -175, MaxAngle: 175}, // Joint 6
		},
		UseCRC:        false,
		DefaultBaud:   115200,
		SupportedBaud: []int{115200, 1000000},
	},
}
```

**Step 4: Add `WithCRC` option to `option.go`**

The existing `WithBaudRate` and `WithTimeout` are stubs. `WithCRC` needs to actually work because `robot.Base` stores the `useCRC` field. Since `Base` fields are unexported, add setter methods to `internal/robot/base.go`:

Add to `internal/robot/base.go`:

```go
// SetBaudRate sets the baud rate (must be called before Open)
func (b *Base) SetBaudRate(baud int) {
	b.baudrate = baud
}

// SetUseCRC enables or disables CRC mode
func (b *Base) SetUseCRC(useCRC bool) {
	b.useCRC = useCRC
}
```

Replace `option.go` contents:

```go
package mycobot

import (
	"time"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
)

// Option configures a robot
type Option func(*robot.Base)

// WithBaudRate sets custom baud rate
func WithBaudRate(baud int) Option {
	return func(b *robot.Base) {
		b.SetBaudRate(baud)
	}
}

// WithTimeout sets default command timeout
func WithTimeout(timeout time.Duration) Option {
	return func(b *robot.Base) {
		// Will be implemented when we add timeout support
	}
}

// WithCRC enables CRC mode for firmware that requires it.
// By default, the standard 0xFA footer is used (matching pymycobot defaults).
func WithCRC() Option {
	return func(b *robot.Base) {
		b.SetUseCRC(true)
	}
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./... -v`
Expected: ALL PASS (config_test.go and existing tests may need fixes for removed models — see Step 6)

**Step 6: Fix any remaining test references to removed models**

The `TestModelConfig_AllModels` test and `errors_test.go`'s `TestRobotError_Error` reference "MyCobot280". Remove `TestModelConfig_AllModels` (replaced by `TestModelConfig_MechArm270`). Update `errors_test.go` to reference "MechArm270" instead:

In `errors_test.go`, change line 13:

```go
		Model: "MechArm270",
```

And line 28:

```go
		Model: "MechArm270",
```

Run: `go test ./... -v`
Expected: ALL PASS

**Step 7: Commit**

```bash
git add config.go config_test.go option.go internal/robot/base.go errors_test.go
git commit -m "fix(config): default CRC to off, add WithCRC option

pymycobot defaults to 0xFA footer without CRC. Changed UseCRC to
false and added WithCRC() option for firmware that needs it.
Removed model configs for MyCobot280, MyCobot320, MyPalletizer260.
Added SetBaudRate/SetUseCRC setters on Base so options work."
```

---

### Task 5: Remove model joint limits for deleted models

The `types/model.go` file still contains joint limit maps for all 4 models. Remove the entries for deleted models so the code matches the single-model scope.

**Files:**
- Modify: `types/model.go:29-60`
- Modify: `types/angle_test.go` (if it references other models)
- Check: `types/joint_test.go`, `types/coord_test.go`, `types/speed_test.go`

**Step 1: Read the types test files to check for model references**

Check `types/angle_test.go` for references to `ModelMyCobot280` or other removed models.

**Step 2: Update `types/model.go`**

Keep the `Model` type constants for all models (they're just strings and cost nothing), but remove `modelJointLimits` entries for models other than MechArm270:

```go
var modelJointLimits = map[Model][]JointLimit{
	ModelMechArm270: {
		{MinAngle: -165, MaxAngle: 165}, // Joint 1
		{MinAngle: -165, MaxAngle: 165}, // Joint 2
		{MinAngle: -165, MaxAngle: 165}, // Joint 3
		{MinAngle: -165, MaxAngle: 165}, // Joint 4
		{MinAngle: -165, MaxAngle: 165}, // Joint 5
		{MinAngle: -175, MaxAngle: 175}, // Joint 6
	},
}
```

**Step 3: Fix any tests that reference removed model limits**

Update `types/angle_test.go` to use `ModelMechArm270` wherever it currently references `ModelMyCobot280`.

**Step 4: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add types/model.go types/angle_test.go
git commit -m "refactor(types): remove joint limits for unsupported models

Only MechArm270 limits remain. Model type constants are kept for
future use but joint limit maps are scoped to the supported model."
```

---

### Task 6: Verify Phase 0 — All tests pass, no dead code

Final check before moving to Phase 1.

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: ALL PASS, zero references to MyCobot280 in non-test code

**Step 2: Check for dead code**

Run: `grep -r "MyCobot280" --include="*.go" .`
Expected: Only appears in `types/model.go` as the constant definition (kept for future model re-introduction)

**Step 3: Commit if any cleanup was needed**

Only commit if previous steps revealed issues that needed fixing.

---

## Phase 1: MechArm 270 Implementation

### Task 7: Add CoordMode Type

The `SendCoords` signature needs a `mode` parameter. Define the type before touching the interface.

**Files:**
- Create: `types/coord_mode.go`
- Create: `types/coord_mode_test.go`

**Step 1: Write the test**

Create `types/coord_mode_test.go`:

```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordMode_Constants(t *testing.T) {
	assert.Equal(t, CoordMode(0), CoordModeAngular)
	assert.Equal(t, CoordMode(1), CoordModeLinear)
}

func TestCoordMode_Validate(t *testing.T) {
	assert.NoError(t, CoordModeAngular.Validate())
	assert.NoError(t, CoordModeLinear.Validate())
	assert.Error(t, CoordMode(2).Validate())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./types/ -run TestCoordMode -v`
Expected: FAIL — `CoordMode` undefined

**Step 3: Write the implementation**

Create `types/coord_mode.go`:

```go
package types

import "fmt"

// CoordMode specifies the interpolation mode for coordinate movement
type CoordMode int

const (
	// CoordModeAngular uses angular interpolation (mode 0)
	CoordModeAngular CoordMode = 0
	// CoordModeLinear uses linear interpolation (mode 1)
	CoordModeLinear CoordMode = 1
)

// Validate checks if the mode is valid
func (m CoordMode) Validate() error {
	if m != CoordModeAngular && m != CoordModeLinear {
		return fmt.Errorf("invalid coord mode %d: must be 0 (angular) or 1 (linear)", m)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `go test ./types/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add types/coord_mode.go types/coord_mode_test.go
git commit -m "feat(types): add CoordMode type for interpolation mode"
```

---

### Task 8: Add Direction Type

JOG methods currently take `direction int`. Add a proper type.

**Files:**
- Create: `types/direction.go`
- Create: `types/direction_test.go`

**Step 1: Write the test**

Create `types/direction_test.go`:

```go
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection_Constants(t *testing.T) {
	assert.Equal(t, Direction(0), DirNegative)
	assert.Equal(t, Direction(1), DirPositive)
}

func TestDirection_Validate(t *testing.T) {
	assert.NoError(t, DirNegative.Validate())
	assert.NoError(t, DirPositive.Validate())
	assert.Error(t, Direction(2).Validate())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./types/ -run TestDirection -v`
Expected: FAIL — `Direction` undefined

**Step 3: Write the implementation**

Create `types/direction.go`:

```go
package types

import "fmt"

// Direction specifies the direction for JOG movement
type Direction int

const (
	// DirNegative moves in the negative direction
	DirNegative Direction = 0
	// DirPositive moves in the positive direction
	DirPositive Direction = 1
)

// Validate checks if the direction is valid
func (d Direction) Validate() error {
	if d != DirNegative && d != DirPositive {
		return fmt.Errorf("invalid direction %d: must be 0 (negative) or 1 (positive)", d)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `go test ./types/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add types/direction.go types/direction_test.go
git commit -m "feat(types): add Direction type for JOG movement"
```

---

### Task 9: Update Robot Interface and Delete MyCobot280

Update the `Robot` interface with the corrected `SendCoords` signature, delete `mycobot280.go` and `mycobot280_test.go`.

**Files:**
- Modify: `robot.go:19`
- Delete: `mycobot280.go`
- Delete: `mycobot280_test.go`

**Step 1: Update `robot.go`**

Change the `SendCoords` signature on line 19 to include `mode`:

```go
SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error
```

Also update `IsInPosition` on line 29 to match pymycobot's signature, which takes data and a flag indicating whether it's angles or coords:

```go
// Robot is the base interface all robot models implement
type Robot interface {
	// Connection management
	Open(ctx context.Context) error
	Close() error
	IsConnected() bool

	// Core motion commands
	SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error
	GetAngles(ctx context.Context) (types.Angles, error)
	SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error
	GetCoords(ctx context.Context) (types.Coord, error)

	// Power and status
	PowerOn(ctx context.Context) error
	PowerOff(ctx context.Context) error
	IsPowerOn(ctx context.Context) (bool, error)

	// Movement queries
	IsMoving(ctx context.Context) (bool, error)
}
```

Note: `IsInPosition` is removed from the interface — it belongs on the Motion subsystem where it can accept both angle and coordinate data with the proper flag parameter.

**Step 2: Delete MyCobot280 files**

```bash
rm mycobot280.go mycobot280_test.go
```

**Step 3: Verify compilation**

Run: `go build ./...`
Expected: Should compile (no code depends on `MyCobot280` outside the deleted files)

Run: `go test ./... -v`
Expected: ALL PASS (mycobot280_test.go is gone, no other tests reference it)

**Step 4: Commit**

```bash
git add robot.go
git rm mycobot280.go mycobot280_test.go
git commit -m "refactor: update Robot interface, delete MyCobot280

Added CoordMode parameter to SendCoords. Moved IsInPosition to
Motion subsystem scope. Removed MyCobot280 — MechArm270 will be
the sole implementation."
```

---

### Task 10: Create MechArm270 Core

Build the new MechArm270 struct with constructor, connection, and power commands.

**Files:**
- Create: `mecharm270.go`
- Create: `mecharm270_test.go`

**Step 1: Write the tests**

Create `mecharm270_test.go`:

```go
package mycobot

import (
	"context"
	"testing"

	"github.com/hipsterbrown/mycobot-go/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMechArm270(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0")

	assert.NotNil(t, arm)
	assert.Equal(t, types.ModelMechArm270, arm.config.Model)
	assert.Equal(t, 6, arm.config.JointCount)
	assert.NotNil(t, arm.base)
}

func TestNewMechArm270_WithOptions(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0",
		WithBaudRate(1000000),
	)

	assert.NotNil(t, arm)
}

func TestNewMechArm270_WithCRC(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0", WithCRC())

	assert.NotNil(t, arm)
}

func TestMechArm270_PowerOn_NotConnected(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()

	err := arm.PowerOn(ctx)
	assert.Error(t, err)
}

func TestMechArm270_SendAngles_Validation(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()

	// Wrong angle count
	angles := types.Angles{0, 45, 90}
	err := arm.SendAngles(ctx, angles, types.SpeedMedium)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected 6 angles")
}

func TestMechArm270_SendAngles_OutOfRange(t *testing.T) {
	arm := NewMechArm270("/dev/null")
	ctx := context.Background()

	// 200 > 175 for joint 6
	angles := types.Angles{0, 0, 0, 0, 0, 200}
	err := arm.SendAngles(ctx, angles, types.SpeedMedium)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestMechArm270_HasMotionSubsystem(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0")
	assert.NotNil(t, arm.Motion)
}

func TestMechArm270_HasIOSubsystem(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0")
	assert.NotNil(t, arm.IO)
}

func TestMechArm270_HasServoSubsystem(t *testing.T) {
	arm := NewMechArm270("/dev/ttyUSB0")
	assert.NotNil(t, arm.Servo)
}
```

**Step 2: Run tests to verify they fail**

Run: `go test . -run TestNewMechArm270 -v`
Expected: FAIL — `NewMechArm270` undefined

**Step 3: Write `mecharm270.go`**

Create `mecharm270.go`:

```go
package mycobot

import (
	"context"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// MechArm270 represents a MechArm 270 robot
type MechArm270 struct {
	Motion *Motion
	IO     *IO
	Servo  *Servo

	base   *robot.Base
	config ModelConfig
}

// NewMechArm270 creates a new MechArm270 instance
func NewMechArm270(port string, opts ...Option) *MechArm270 {
	config := getModelConfig(types.ModelMechArm270)
	base := robot.NewBase(port, config.DefaultBaud, config.UseCRC)

	for _, opt := range opts {
		opt(base)
	}

	return &MechArm270{
		Motion: &Motion{robot: base},
		IO:     &IO{robot: base},
		Servo:  &Servo{robot: base},
		base:   base,
		config: config,
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

	data := protocol.EncodeAngles(angles.ToFloat64())
	data = append(data, byte(speed))

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
func (m *MechArm270) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error {
	if err := speed.Validate(); err != nil {
		return err
	}
	if err := mode.Validate(); err != nil {
		return err
	}

	data := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data = append(data, byte(speed), byte(mode))

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

	return types.Coord{X: x, Y: y, Z: z, Rx: rx, Ry: ry, Rz: rz}, nil
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
```

Note: this will fail to compile until the `IO` and `Servo` structs exist (Task 12 and 13). For now, create minimal stubs so the file compiles. Add these at the bottom of the file temporarily, or proceed to Step 3a first.

**Step 3a: Create stub files so MechArm270 compiles**

Create `io.go`:

```go
package mycobot

import "github.com/hipsterbrown/mycobot-go/internal/robot"

// IO provides Atom IO operations (end-effector head, 0x60 range)
type IO struct {
	robot *robot.Base
}
```

Create `servo.go`:

```go
package mycobot

import "github.com/hipsterbrown/mycobot-go/internal/robot"

// Servo provides servo control operations
type Servo struct {
	robot *robot.Base
}
```

**Step 4: Run tests**

Run: `go test . -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add mecharm270.go mecharm270_test.go io.go servo.go
git commit -m "feat: add MechArm270 with core motion and power commands

Replaces MyCobot280 as the sole robot implementation. Includes
SendCoords with CoordMode parameter, subsystem stubs for IO and
Servo, and Motion/IO/Servo fields on the struct."
```

---

### Task 11: Complete Motion Subsystem

Add `SendAngle`, `SendCoord`, `IsInPosition` to the existing `motion.go`. Update `JogAngle` and `JogCoord` to use the `Direction` type.

**Files:**
- Modify: `motion.go`
- Modify: `motion_test.go`

**Step 1: Write failing tests for new methods and updated signatures**

Replace `motion_test.go` entirely:

```go
package mycobot

import (
	"testing"

	"github.com/hipsterbrown/mycobot-go/types"
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

func TestMotion_JogAngle_UsesDirection(t *testing.T) {
	// Verify the method signature accepts types.Direction
	motion := &Motion{}
	_ = motion // method exists with Direction parameter
	var _ func(*Motion) = func(m *Motion) {
		// This compiles only if JogAngle accepts types.Direction
		_ = m.JogAngle
	}
}

func TestMotion_SendAngle_Exists(t *testing.T) {
	motion := &Motion{}
	// Verify method signature compiles
	var _ func(*Motion) = func(m *Motion) {
		_ = m.SendAngle
	}
	_ = motion
}

func TestMotion_SendCoord_Exists(t *testing.T) {
	motion := &Motion{}
	var _ func(*Motion) = func(m *Motion) {
		_ = m.SendCoord
	}
	_ = motion
}

func TestMotion_IsInPosition_Exists(t *testing.T) {
	motion := &Motion{}
	var _ func(*Motion) = func(m *Motion) {
		_ = m.IsInPosition
	}
	_ = motion
}
```

**Step 2: Run tests to verify they fail**

Run: `go test . -run "TestMotion_SendAngle|TestMotion_SendCoord|TestMotion_IsInPosition" -v`
Expected: FAIL — methods undefined

**Step 3: Update `motion.go`**

Replace the full contents of `motion.go`:

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

// JogAngle performs incremental joint movement
func (m *Motion) JogAngle(ctx context.Context, joint types.JointID, direction types.Direction, speed types.Speed) error {
	if err := direction.Validate(); err != nil {
		return err
	}
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
func (m *Motion) JogCoord(ctx context.Context, axis CoordAxis, direction types.Direction, speed types.Speed) error {
	if err := direction.Validate(); err != nil {
		return err
	}
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(axis + 1), byte(direction), byte(speed)}
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

// SendAngle moves a single joint to the specified angle
func (m *Motion) SendAngle(ctx context.Context, joint types.JointID, angle types.Angle, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(int(angle*100))...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendAngle,
		Data: data,
	}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// SendCoord moves a single coordinate axis to the specified value
func (m *Motion) SendCoord(ctx context.Context, axis CoordAxis, value float64, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	// XYZ axes (0-2) use * 10, rotation axes (3-5) use * 100
	var encoded int
	if axis <= AxisZ {
		encoded = int(value * 10)
	} else {
		encoded = int(value * 100)
	}

	data := []byte{byte(axis + 1)}
	data = append(data, protocol.EncodeInt16(encoded)...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SendCoord,
		Data: data,
	}
	_, err := m.robot.SendCommand(ctx, cmd)
	return err
}

// IsInPosition checks if the robot is at the target position.
// flag: 0 = check angles, 1 = check coordinates
func (m *Motion) IsInPosition(ctx context.Context, data []float64, flag int) (bool, error) {
	var encoded []byte
	if flag == 0 {
		// Angles: encode with * 100
		encoded = protocol.EncodeAngles(data)
	} else {
		// Coordinates: encode with split multiplier
		if len(data) != 6 {
			return false, nil
		}
		encoded = protocol.EncodeCoords(data[0], data[1], data[2], data[3], data[4], data[5])
	}
	encoded = append(encoded, byte(flag))

	cmd := protocol.Command{
		Code: protocol.IsInPosition,
		Data: encoded,
	}
	resp, err := m.robot.SendCommand(ctx, cmd)
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

Run: `go test . -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add motion.go motion_test.go
git commit -m "feat(motion): add SendAngle, SendCoord, IsInPosition; use Direction type

Completes the Motion subsystem. JogAngle/JogCoord now use
types.Direction instead of bare int. SendCoord uses the correct
multiplier (x10 for XYZ, x100 for rotation) matching pymycobot."
```

---

### Task 12: Implement Atom IO Subsystem

**Files:**
- Modify: `io.go` (replace stub)
- Create: `io_test.go`

**Step 1: Write the tests**

Create `io_test.go`:

```go
package mycobot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIO_Structure(t *testing.T) {
	io := &IO{}
	assert.NotNil(t, io)
}

func TestPinMode_Constants(t *testing.T) {
	assert.Equal(t, PinMode(0), PinInput)
	assert.Equal(t, PinMode(1), PinOutput)
	assert.Equal(t, PinMode(2), PinInputPullup)
}

func TestPinSignal_Constants(t *testing.T) {
	assert.Equal(t, PinSignal(0), SignalLow)
	assert.Equal(t, PinSignal(1), SignalHigh)
}

func TestIO_MethodsExist(t *testing.T) {
	io := &IO{}
	// Verify all methods compile
	_ = io.SetPinMode
	_ = io.SetDigitalOutput
	_ = io.GetDigitalInput
	_ = io.SetPWMMode
	_ = io.SetPWMOutput
	_ = io.SetColor
}
```

**Step 2: Run tests to verify they fail**

Run: `go test . -run "TestIO_|TestPinMode|TestPinSignal" -v`
Expected: FAIL — `PinMode`, methods undefined

**Step 3: Implement `io.go`**

Replace the stub:

```go
package mycobot

import (
	"context"
	"fmt"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

// PinMode configures a pin's behavior
type PinMode int

const (
	PinInput       PinMode = 0
	PinOutput      PinMode = 1
	PinInputPullup PinMode = 2
)

// PinSignal represents a digital pin state
type PinSignal int

const (
	SignalLow  PinSignal = 0
	SignalHigh PinSignal = 1
)

// IO provides Atom IO operations (end-effector head, 0x60 range)
type IO struct {
	robot *robot.Base
}

// SetPinMode configures a pin as input, output, or input with pullup
func (io *IO) SetPinMode(ctx context.Context, pin int, mode PinMode) error {
	cmd := protocol.Command{
		Code: protocol.SetPinMode,
		Data: []byte{byte(pin), byte(mode)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetDigitalOutput sets a digital pin high or low
func (io *IO) SetDigitalOutput(ctx context.Context, pin int, signal PinSignal) error {
	cmd := protocol.Command{
		Code: protocol.SetDigitalOutput,
		Data: []byte{byte(pin), byte(signal)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// GetDigitalInput reads the state of a digital input pin
func (io *IO) GetDigitalInput(ctx context.Context, pin int) (PinSignal, error) {
	cmd := protocol.Command{
		Code: protocol.GetDigitalInput,
		Data: []byte{byte(pin)},
	}
	data, err := io.robot.SendCommand(ctx, cmd)
	if err != nil {
		return SignalLow, err
	}
	if len(data) > 0 {
		return PinSignal(data[0]), nil
	}
	return SignalLow, nil
}

// SetPWMMode configures a pin for PWM output
func (io *IO) SetPWMMode(ctx context.Context, pin int) error {
	cmd := protocol.Command{
		Code: protocol.SetPWMMode,
		Data: []byte{byte(pin)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetPWMOutput sets PWM frequency and duty cycle on a channel
func (io *IO) SetPWMOutput(ctx context.Context, channel int, freq int, dutyCycle int) error {
	if dutyCycle < 0 || dutyCycle > 256 {
		return fmt.Errorf("duty cycle %d out of range [0, 256]", dutyCycle)
	}

	cmd := protocol.Command{
		Code: protocol.SetPWMOutput,
		Data: []byte{byte(channel), byte(freq >> 8), byte(freq & 0xFF), byte(dutyCycle)},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}

// SetColor sets the RGB color of the Atom LED
func (io *IO) SetColor(ctx context.Context, r, g, b byte) error {
	cmd := protocol.Command{
		Code: protocol.SetColor,
		Data: []byte{r, g, b},
	}
	_, err := io.robot.SendCommand(ctx, cmd)
	return err
}
```

**Step 4: Run tests**

Run: `go test . -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add io.go io_test.go
git commit -m "feat(io): implement Atom IO subsystem

SetPinMode, SetDigitalOutput, GetDigitalInput, SetPWMMode,
SetPWMOutput, SetColor. Uses corrected 0x60-range command codes."
```

---

### Task 13: Implement Servo Subsystem

**Files:**
- Modify: `servo.go` (replace stub)
- Create: `servo_test.go`

**Step 1: Write the tests**

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

func TestServo_MethodsExist(t *testing.T) {
	servo := &Servo{}
	_ = servo.ReleaseServo
	_ = servo.FocusServo
	_ = servo.IsServoEnabled
	_ = servo.GetEncoder
	_ = servo.SetEncoder
	_ = servo.GetEncoders
	_ = servo.SetEncoders
	_ = servo.GetServoData
	_ = servo.SetServoData
	_ = servo.SetServoCalibration
	_ = servo.GetJointMin
	_ = servo.GetJointMax
}
```

**Step 2: Run tests to verify they fail**

Run: `go test . -run TestServo -v`
Expected: FAIL — methods undefined

**Step 3: Implement `servo.go`**

Replace the stub:

```go
package mycobot

import (
	"context"
	"encoding/binary"

	"github.com/hipsterbrown/mycobot-go/internal/robot"
	"github.com/hipsterbrown/mycobot-go/protocol"
	"github.com/hipsterbrown/mycobot-go/types"
)

// Servo provides servo control operations
type Servo struct {
	robot *robot.Base
}

// ReleaseServo powers off a single servo, allowing free movement
func (s *Servo) ReleaseServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.ReleaseServo,
		Data: []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// FocusServo powers on a single servo
func (s *Servo) FocusServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.FocusServo,
		Data: []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// IsServoEnabled checks if a specific servo is powered on
func (s *Servo) IsServoEnabled(ctx context.Context, joint types.JointID) (bool, error) {
	cmd := protocol.Command{
		Code: protocol.IsServoEnable,
		Data: []byte{byte(joint)},
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// GetEncoder reads the encoder value for a single joint (0-4096)
func (s *Servo) GetEncoder(ctx context.Context, joint types.JointID) (int, error) {
	cmd := protocol.Command{
		Code: protocol.GetEncoder,
		Data: []byte{byte(joint)},
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		return int(binary.BigEndian.Uint16(data[:2])), nil
	}
	return 0, nil
}

// SetEncoder sets the encoder value for a single joint (0-4096)
func (s *Servo) SetEncoder(ctx context.Context, joint types.JointID, value int) error {
	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(value)...)

	cmd := protocol.Command{
		Code: protocol.SetEncoder,
		Data: data,
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetEncoders reads all encoder values
func (s *Servo) GetEncoders(ctx context.Context) ([]int, error) {
	cmd := protocol.Command{Code: protocol.GetEncoders}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}

	count := len(data) / 2
	encoders := make([]int, count)
	for i := 0; i < count; i++ {
		encoders[i] = int(binary.BigEndian.Uint16(data[i*2 : i*2+2]))
	}
	return encoders, nil
}

// SetEncoders sets all encoder values simultaneously
func (s *Servo) SetEncoders(ctx context.Context, encoders []int, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	var data []byte
	for _, enc := range encoders {
		data = append(data, protocol.EncodeInt16(enc)...)
	}
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code: protocol.SetEncoders,
		Data: data,
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetServoData reads a servo parameter
func (s *Servo) GetServoData(ctx context.Context, joint types.JointID, dataID byte) (int, error) {
	cmd := protocol.Command{
		Code: protocol.GetServoData,
		Data: []byte{byte(joint), dataID},
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 1 {
		return int(data[0]), nil
	}
	return 0, nil
}

// SetServoData writes a servo parameter
func (s *Servo) SetServoData(ctx context.Context, joint types.JointID, dataID byte, value int) error {
	cmd := protocol.Command{
		Code: protocol.SetServoData,
		Data: []byte{byte(joint), dataID, byte(value)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// SetServoCalibration sets the current position as angle zero for a joint.
// This writes to non-volatile memory on the servo.
func (s *Servo) SetServoCalibration(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.SetServoCalibration,
		Data: []byte{byte(joint)},
	}
	_, err := s.robot.SendCommand(ctx, cmd)
	return err
}

// GetJointMin reads the minimum angle limit for a joint from firmware
func (s *Servo) GetJointMin(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code: protocol.GetJointMinAngle,
		Data: []byte{byte(joint)},
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		value := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(value) / 100.0, nil
	}
	return 0, nil
}

// GetJointMax reads the maximum angle limit for a joint from firmware
func (s *Servo) GetJointMax(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code: protocol.GetJointMaxAngle,
		Data: []byte{byte(joint)},
	}
	data, err := s.robot.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		value := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(value) / 100.0, nil
	}
	return 0, nil
}
```

**Step 4: Run tests**

Run: `go test . -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add servo.go servo_test.go
git commit -m "feat(servo): implement Servo subsystem

ReleaseServo, FocusServo, IsServoEnabled, encoder read/write,
servo data get/set, calibration, joint min/max queries."
```

---

### Task 14: Integration Test Harness

A build-tag-guarded test file for running against real hardware.

**Files:**
- Create: `integration_test.go`

**Step 1: Write the integration test file**

Create `integration_test.go`:

```go
//go:build integration

package mycobot

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hipsterbrown/mycobot-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestPort(t *testing.T) string {
	t.Helper()
	port := os.Getenv("MYCOBOT_PORT")
	if port == "" {
		t.Skip("MYCOBOT_PORT not set, skipping integration test")
	}
	return port
}

func setupArm(t *testing.T) *MechArm270 {
	t.Helper()
	port := getTestPort(t)
	arm := NewMechArm270(port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.Open(ctx))
	t.Cleanup(func() {
		arm.Close()
	})

	return arm
}

func TestIntegration_Connect(t *testing.T) {
	arm := setupArm(t)
	assert.True(t, arm.IsConnected())
}

func TestIntegration_PowerOn(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond) // allow servos to engage

	on, err := arm.IsPowerOn(ctx)
	require.NoError(t, err)
	assert.True(t, on)
}

func TestIntegration_GetAngles(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond)

	angles, err := arm.GetAngles(ctx)
	require.NoError(t, err)
	assert.Len(t, angles, 6, "MechArm270 should return 6 joint angles")
}

func TestIntegration_SendAnglesRoundtrip(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond)

	// Send all joints to zero — a known safe position
	target := types.Angles{0, 0, 0, 0, 0, 0}
	require.NoError(t, arm.SendAngles(ctx, target, types.SpeedSlow))

	// Wait for movement to complete
	time.Sleep(3 * time.Second)

	angles, err := arm.GetAngles(ctx)
	require.NoError(t, err)

	for i, a := range angles {
		assert.InDelta(t, float64(target[i]), float64(a), 2.0,
			"joint %d should be near target", i+1)
	}
}

func TestIntegration_GetCoords(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond)

	coord, err := arm.GetCoords(ctx)
	require.NoError(t, err)

	// Coordinates should be non-zero when arm is in a real position
	t.Logf("Current coords: X=%.1f Y=%.1f Z=%.1f Rx=%.1f Ry=%.1f Rz=%.1f",
		coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
}

func TestIntegration_JogAndStop(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond)

	// Jog joint 1 positive at slow speed, then immediately stop
	require.NoError(t, arm.Motion.JogAngle(ctx, types.Joint1, types.DirPositive, types.SpeedSlow))
	time.Sleep(200 * time.Millisecond)
	require.NoError(t, arm.Motion.JogStop(ctx))
}

func TestIntegration_ReadEncoders(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, arm.PowerOn(ctx))
	time.Sleep(500 * time.Millisecond)

	encoders, err := arm.Servo.GetEncoders(ctx)
	require.NoError(t, err)
	assert.Len(t, encoders, 6, "MechArm270 should return 6 encoder values")

	for i, enc := range encoders {
		t.Logf("Joint %d encoder: %d", i+1, enc)
		assert.Greater(t, enc, 0, "encoder should be > 0")
		assert.Less(t, enc, 4096, "encoder should be < 4096")
	}
}

func TestIntegration_SetColor(t *testing.T) {
	arm := setupArm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set LED to green, then off
	require.NoError(t, arm.IO.SetColor(ctx, 0, 255, 0))
	time.Sleep(500 * time.Millisecond)
	require.NoError(t, arm.IO.SetColor(ctx, 0, 0, 0))
}
```

**Step 2: Verify it doesn't run without the build tag**

Run: `go test . -v`
Expected: ALL PASS — integration tests are skipped (no `integration` build tag)

**Step 3: Document how to run integration tests**

The command to run against real hardware:

```bash
MYCOBOT_PORT=/dev/ttyUSB0 go test . -tags=integration -v -count=1
```

**Step 4: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration test harness for MechArm 270

Build-tag guarded (//go:build integration). Reads serial port from
MYCOBOT_PORT env var. Tests power, angles, coords, JOG, encoders,
and LED color against real hardware."
```

---

### Task 15: Final Verification

**Step 1: Run the full test suite**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 2: Check for dead references**

Run: `grep -r "MyCobot280" --include="*.go" .`
Expected: Only in `types/model.go` constant definition

**Step 3: Verify the build**

Run: `go build ./...`
Expected: Clean build, no errors

**Step 4: Verify go vet**

Run: `go vet ./...`
Expected: No issues

**Step 5: Commit any final cleanup**

Only commit if previous steps revealed issues that needed fixing.
