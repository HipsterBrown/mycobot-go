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

// IsConnected returns true if robot is connected
func (b *base) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// Open establishes serial connection and starts command loop
func (b *base) Open(ctx context.Context) error {
	conn, err := openSerial(b.port, &serial.Mode{
		BaudRate: b.baudrate,
	})
	if err != nil {
		return fmt.Errorf("failed to open serial port: %w", err)
	}

	b.conn = conn
	b.cmdChan = make(chan *command, 32)
	b.closeChan = make(chan struct{})

	b.mu.Lock()
	b.connected = true
	b.mu.Unlock()

	b.wg.Add(1)
	go b.commandLoop()

	return nil
}

// Close shuts down the command loop gracefully
func (b *base) Close() error {
	b.closeOnce.Do(func() {
		close(b.closeChan)
	})
	b.wg.Wait()
	return nil
}

// commandLoop runs in dedicated goroutine, owns serial connection
func (b *base) commandLoop() {
	defer b.wg.Done()
	defer b.conn.Close()
	defer func() {
		b.mu.Lock()
		b.connected = false
		b.mu.Unlock()
	}()

	for {
		select {
		case <-b.closeChan:
			return

		case cmd := <-b.cmdChan:
			// Check if command context is already cancelled
			if err := cmd.ctx.Err(); err != nil {
				cmd.response <- &response{err: err}
				continue
			}

			// Set CRC mode from base config
			cmd.request.UseCRC = b.useCRC

			// Encode and write command
			data, err := cmd.request.Encode()
			if err != nil {
				cmd.response <- &response{err: fmt.Errorf("encode failed: %w", err)}
				continue
			}

			// Drop stale bytes left by a prior fire-and-forget command so
			// they can't masquerade as the reply we're about to read.
			_ = b.conn.ResetInputBuffer()

			if _, err := b.conn.Write(data); err != nil {
				cmd.response <- &response{err: fmt.Errorf("write failed: %w", err)}
				continue
			}

			// pymycobot's has_reply=False commands: firmware sends nothing back.
			if !cmd.request.HasReply {
				cmd.response <- &response{data: nil, err: nil}
				continue
			}

			respData, err := b.readResponse(cmd.ctx, cmd.request.Code)
			cmd.response <- &response{data: respData, err: err}
		}
	}
}

// readResponse reads until it can decode a frame whose code matches
// expectedCode. Stray leading bytes are skipped; frames with a different code
// (stale data from a prior command) are discarded and scanning continues.
func (b *base) readResponse(ctx context.Context, expectedCode byte) ([]byte, error) {
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

	if err := b.conn.SetReadTimeout(timeout); err != nil {
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	buf := make([]byte, 0, 256)
	tmp := make([]byte, 64)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		n, err := b.conn.Read(tmp)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("serial read: %w", err)
		}
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}

		data, consumed, ok := extractMatchingFrame(buf, expectedCode, b.useCRC)
		if ok {
			return data, nil
		}
		if consumed > 0 {
			buf = buf[consumed:]
		}
	}

	return nil, fmt.Errorf("read timeout after %d bytes", len(buf))
}

// extractMatchingFrame looks for the next complete frame in buf whose code
// byte equals expectedCode. It returns the payload and the number of bytes
// to discard from buf; ok is true only on a full match. When ok is false but
// consumed > 0, the caller should drop those leading bytes and keep reading.
func extractMatchingFrame(buf []byte, expectedCode byte, useCRC bool) (data []byte, consumed int, ok bool) {
	start := bytes.Index(buf, headerSentinel)
	if start < 0 {
		// Keep the trailing byte: it might be the first FE of the next frame.
		if len(buf) > 1 {
			return nil, len(buf) - 1, false
		}
		return nil, 0, false
	}
	if len(buf)-start < minFrameLen {
		return nil, start, false
	}
	// Mismatched code = stale/foreign frame. Step past FE FE and rescan.
	if buf[start+3] != expectedCode {
		return nil, start + 2, false
	}

	resp, frameSize, err := protocol.Decode(buf[start:], useCRC)
	if err != nil {
		// Usually just incomplete; wait for more bytes.
		return nil, start, false
	}
	return resp.Data, start + frameSize, true
}

// SendCommand queues a command and waits for response
func (b *base) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	if !b.IsConnected() {
		return nil, ErrNotConnected
	}

	responseChan := make(chan *response, 1)

	select {
	case b.cmdChan <- &command{
		ctx:      ctx,
		request:  cmd,
		response: responseChan,
	}:
		// Command queued successfully

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-b.closeChan:
		return nil, ErrRobotClosed
	}

	// Wait for response
	select {
	case resp := <-responseChan:
		return resp.data, resp.err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// PowerOn powers on all servos
func (b *base) PowerOn(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOn}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// PowerOff powers off all servos
func (b *base) PowerOff(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.PowerOff}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// IsPowerOn returns true if robot is powered on
func (b *base) IsPowerOn(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPowerOn, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// SendAngles sends joint angles to the robot
func (b *base) SendAngles(ctx context.Context, angles types.Angles, speed types.Speed) error {
	if err := angles.Validate(b.config.JointCount, b.config.Model); err != nil {
		return err
	}
	if err := speed.Validate(); err != nil {
		return err
	}

	data := protocol.EncodeAngles(angles.ToFloat64())
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code:     protocol.SendAngles,
		Data:     data,
		HasReply: true,
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetAngles retrieves current joint angles
func (b *base) GetAngles(ctx context.Context) (types.Angles, error) {
	cmd := protocol.Command{Code: protocol.GetAngles, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
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
func (b *base) SendCoords(ctx context.Context, coord types.Coord, speed types.Speed, mode types.CoordMode) error {
	if err := speed.Validate(); err != nil {
		return err
	}
	if err := mode.Validate(); err != nil {
		return err
	}

	data := protocol.EncodeCoords(coord.X, coord.Y, coord.Z, coord.Rx, coord.Ry, coord.Rz)
	data = append(data, byte(speed), byte(mode))

	// pymycobot sends SEND_COORDS with has_reply=False when a mode byte is
	// present (see generate.py send_coords). We always include mode, so this
	// is a fire-and-forget command.
	cmd := protocol.Command{
		Code: protocol.SendCoords,
		Data: data,
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetCoords retrieves current coordinate position
func (b *base) GetCoords(ctx context.Context) (types.Coord, error) {
	cmd := protocol.Command{Code: protocol.GetCoords, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
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
func (b *base) IsMoving(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsMoving, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// JogAngle performs incremental joint movement
func (b *base) JogAngle(ctx context.Context, joint types.JointID, direction types.Direction, speed types.Speed) error {
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

	_, err := b.SendCommand(ctx, cmd)
	return err
}

// JogCoord performs incremental coordinate movement
func (b *base) JogCoord(ctx context.Context, axis types.CoordAxis, direction types.Direction, speed types.Speed) error {
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

	_, err := b.SendCommand(ctx, cmd)
	return err
}

// JogStop stops JOG movement
func (b *base) JogStop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.JogStop}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// Pause pauses current movement
func (b *base) Pause(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Pause}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// Resume resumes paused movement
func (b *base) Resume(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Resume}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// Stop stops all movement
func (b *base) Stop(ctx context.Context) error {
	cmd := protocol.Command{Code: protocol.Stop}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// IsPaused returns true if robot is paused
func (b *base) IsPaused(ctx context.Context) (bool, error) {
	cmd := protocol.Command{Code: protocol.IsPaused, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}

	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// SendAngle moves a single joint to the specified angle
func (b *base) SendAngle(ctx context.Context, joint types.JointID, angle types.Angle, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(int(angle*100))...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code:     protocol.SendAngle,
		Data:     data,
		HasReply: true,
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// SendCoord moves a single coordinate axis to the specified value
func (b *base) SendCoord(ctx context.Context, axis types.CoordAxis, value float64, speed types.Speed) error {
	if err := speed.Validate(); err != nil {
		return err
	}

	// XYZ axes (0-2) use * 10, rotation axes (3-5) use * 100
	var encoded int
	if axis <= types.AxisZ {
		encoded = int(value * 10)
	} else {
		encoded = int(value * 100)
	}

	data := []byte{byte(axis + 1)}
	data = append(data, protocol.EncodeInt16(encoded)...)
	data = append(data, byte(speed))

	cmd := protocol.Command{
		Code:     protocol.SendCoord,
		Data:     data,
		HasReply: true,
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// IsInPosition checks if the robot is at the target position.
// Use PositionAngles to check joint angles or PositionCoords to check coordinates.
func (b *base) IsInPosition(ctx context.Context, data []float64, flag types.PositionFlag) (bool, error) {
	var encoded []byte
	if flag == types.PositionAngles {
		encoded = protocol.EncodeAngles(data)
	} else {
		if len(data) != 6 {
			return false, fmt.Errorf("expected 6 coordinate values, got %d", len(data))
		}
		encoded = protocol.EncodeCoords(data[0], data[1], data[2], data[3], data[4], data[5])
	}
	encoded = append(encoded, byte(flag))

	cmd := protocol.Command{
		Code:     protocol.IsInPosition,
		Data:     encoded,
		HasReply: true,
	}
	resp, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(resp) > 0 {
		return resp[0] == 1, nil
	}
	return false, nil
}

// SetPinMode configures a pin as input, output, or input with pullup
func (b *base) SetPinMode(ctx context.Context, pin int, mode types.PinMode) error {
	cmd := protocol.Command{
		Code: protocol.SetPinMode,
		Data: []byte{byte(pin), byte(mode)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// SetDigitalOutput sets a digital pin high or low
func (b *base) SetDigitalOutput(ctx context.Context, pin int, signal types.PinSignal) error {
	cmd := protocol.Command{
		Code: protocol.SetDigitalOutput,
		Data: []byte{byte(pin), byte(signal)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetDigitalInput reads the state of a digital input pin
func (b *base) GetDigitalInput(ctx context.Context, pin int) (types.PinSignal, error) {
	cmd := protocol.Command{
		Code:     protocol.GetDigitalInput,
		Data:     []byte{byte(pin)},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return types.SignalLow, err
	}
	if len(data) > 0 {
		return types.PinSignal(data[0]), nil
	}
	return types.SignalLow, nil
}

// SetPWMMode configures a pin for PWM output
func (b *base) SetPWMMode(ctx context.Context, pin int) error {
	cmd := protocol.Command{
		Code: protocol.SetPWMMode,
		Data: []byte{byte(pin)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// SetPWMOutput sets PWM frequency and duty cycle on a channel
func (b *base) SetPWMOutput(ctx context.Context, channel int, freq int, dutyCycle int) error {
	if dutyCycle < 0 || dutyCycle > 256 {
		return fmt.Errorf("duty cycle %d out of range [0, 256]", dutyCycle)
	}

	cmd := protocol.Command{
		Code: protocol.SetPWMOutput,
		Data: []byte{byte(channel), byte(freq >> 8), byte(freq & 0xFF), byte(dutyCycle)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// SetColor sets the RGB color of the Atom LED
func (b *base) SetColor(ctx context.Context, red, green, blue byte) error {
	_, err := b.SendCommand(ctx, protocol.Command{
		Code: protocol.SetColor,
		Data: []byte{red, green, blue},
	})
	return err
}

// ReleaseServo powers off a single servo, allowing free movement
func (b *base) ReleaseServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.ReleaseServo,
		Data: []byte{byte(joint)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// FocusServo powers on a single servo
func (b *base) FocusServo(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.FocusServo,
		Data: []byte{byte(joint)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// IsServoEnabled checks if a specific servo is powered on
func (b *base) IsServoEnabled(ctx context.Context, joint types.JointID) (bool, error) {
	cmd := protocol.Command{
		Code:     protocol.IsServoEnable,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return false, err
	}
	if len(data) > 0 {
		return data[0] == 1, nil
	}
	return false, nil
}

// GetEncoder reads the encoder value for a single joint (0-4096)
func (b *base) GetEncoder(ctx context.Context, joint types.JointID) (int, error) {
	cmd := protocol.Command{
		Code:     protocol.GetEncoder,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		return int(binary.BigEndian.Uint16(data[:2])), nil
	}
	return 0, nil
}

// SetEncoder sets the encoder value for a single joint (0-4096)
func (b *base) SetEncoder(ctx context.Context, joint types.JointID, value int) error {
	data := []byte{byte(joint)}
	data = append(data, protocol.EncodeInt16(value)...)

	cmd := protocol.Command{
		Code: protocol.SetEncoder,
		Data: data,
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetEncoders reads all encoder values
func (b *base) GetEncoders(ctx context.Context) ([]int, error) {
	cmd := protocol.Command{Code: protocol.GetEncoders, HasReply: true}
	data, err := b.SendCommand(ctx, cmd)
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
func (b *base) SetEncoders(ctx context.Context, encoders []int, speed types.Speed) error {
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
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetServoData reads a servo parameter
func (b *base) GetServoData(ctx context.Context, joint types.JointID, dataID byte) (int, error) {
	cmd := protocol.Command{
		Code:     protocol.GetServoData,
		Data:     []byte{byte(joint), dataID},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 1 {
		return int(data[0]), nil
	}
	return 0, nil
}

// SetServoData writes a servo parameter
func (b *base) SetServoData(ctx context.Context, joint types.JointID, dataID byte, value int) error {
	cmd := protocol.Command{
		Code: protocol.SetServoData,
		Data: []byte{byte(joint), dataID, byte(value)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// SetServoCalibration sets the current position as angle zero for a joint.
// This writes to non-volatile memory on the servo.
func (b *base) SetServoCalibration(ctx context.Context, joint types.JointID) error {
	cmd := protocol.Command{
		Code: protocol.SetServoCalibration,
		Data: []byte{byte(joint)},
	}
	_, err := b.SendCommand(ctx, cmd)
	return err
}

// GetJointMin reads the minimum angle limit for a joint from firmware
func (b *base) GetJointMin(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code:     protocol.GetJointMinAngle,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
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
func (b *base) GetJointMax(ctx context.Context, joint types.JointID) (float64, error) {
	cmd := protocol.Command{
		Code:     protocol.GetJointMaxAngle,
		Data:     []byte{byte(joint)},
		HasReply: true,
	}
	data, err := b.SendCommand(ctx, cmd)
	if err != nil {
		return 0, err
	}
	if len(data) >= 2 {
		value := int16(binary.BigEndian.Uint16(data[:2]))
		return float64(value) / 100.0, nil
	}
	return 0, nil
}
