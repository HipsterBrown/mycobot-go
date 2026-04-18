package mycobot

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"go.bug.st/serial"
	"github.com/hipsterbrown/mycobot-go/types"
)

func TestWithDefaultTimeout_setsField(t *testing.T) {
	b := newBase("/dev/null", getModelConfig(types.ModelMechArm270))
	WithDefaultTimeout(250 * time.Millisecond)(b)

	if b.defaultTimeout != 250*time.Millisecond {
		t.Fatalf("defaultTimeout = %v, want 250ms", b.defaultTimeout)
	}
}

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
func (f *fakePort) Drain() error              { return nil }
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

func TestBase_NoReplyCommandReturnsImmediately(t *testing.T) {
	arm, fake := openTestArm(t, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := arm.PowerOn(ctx)
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

func TestBase_ReplyCommandDecodes(t *testing.T) {
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

func TestBase_SkipsStaleFrameWithWrongCode(t *testing.T) {
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

func TestBase_SkipsGarbageBeforeHeader(t *testing.T) {
	reply := append(
		[]byte{0xAA, 0xBB, 0xCC},
		0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA,
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

func TestBase_WithDefaultTimeoutHonored(t *testing.T) {
	fake := newFakePort(nil)
	installFakePort(t, fake)

	arm := NewMechArm270("/dev/fake", WithDefaultTimeout(80*time.Millisecond))
	if err := arm.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer arm.Close()

	start := time.Now()
	_, err := arm.IsPowerOn(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed < 50*time.Millisecond || elapsed > 300*time.Millisecond {
		t.Errorf("elapsed = %v, expected ~80ms (±range)", elapsed)
	}
}

func TestBase_FlushesBeforeEveryWrite(t *testing.T) {
	reply := []byte{0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA}
	// Load only the first reply; we'll inject the second after the first command
	// returns so that the greedy 64-byte Read in readResponse doesn't consume
	// both frames in a single call and discard the second one.
	arm, fake := openTestArm(t, reply)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if _, err := arm.IsPowerOn(ctx); err != nil {
		t.Fatal(err)
	}

	// Inject the second reply now that the first command has been fully handled.
	fake.mu.Lock()
	fake.readBuf.Write(reply)
	fake.mu.Unlock()

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

func TestBase_AssemblesFrameAcrossReads(t *testing.T) {
	full := []byte{0xFE, 0xFE, 0x03, 0x12, 0x01, 0xFA}
	first := full[:3]
	second := full[3:]

	fake := newFakePort(first)
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
