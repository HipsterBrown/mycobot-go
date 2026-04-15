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
