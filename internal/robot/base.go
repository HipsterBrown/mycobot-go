package robot

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"go.bug.st/serial"
	"github.com/hipsterbrown/mycobot-go/internal/errors"
	"github.com/hipsterbrown/mycobot-go/protocol"
)

// headerSentinel is the two-byte frame header used to resync the stream.
var headerSentinel = []byte{protocol.Header, protocol.Header}

// Base provides common robot functionality
type Base struct {
	port      string
	baudrate  int
	useCRC    bool
	connected bool

	// Serial connection (owned by command loop goroutine)
	conn serial.Port

	// Command queue
	cmdChan   chan *command
	closeChan chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup

	mu sync.RWMutex // Protects connected flag
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

// NewBase creates a new base robot
func NewBase(port string, baudrate int, useCRC bool) *Base {
	return &Base{
		port:      port,
		baudrate:  baudrate,
		useCRC:    useCRC,
		connected: false,
	}
}

// SetBaudRate sets the baud rate (must be called before Open)
func (b *Base) SetBaudRate(baud int) {
	b.baudrate = baud
}

// SetUseCRC enables or disables CRC mode
func (b *Base) SetUseCRC(useCRC bool) {
	b.useCRC = useCRC
}

// IsConnected returns true if robot is connected
func (b *Base) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// Open establishes serial connection and starts command loop
func (b *Base) Open(ctx context.Context) error {
	conn, err := serial.Open(b.port, &serial.Mode{
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

// commandLoop runs in dedicated goroutine, owns serial connection
func (b *Base) commandLoop() {
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

// minFrameLen is header(2) + length(1) + code(1) + footer/crc(1).
const minFrameLen = 5

// readResponse reads until it can decode a frame whose code matches
// expectedCode. Stray leading bytes are skipped; frames with a different code
// (stale data from a prior command) are discarded and scanning continues.
func (b *Base) readResponse(ctx context.Context, expectedCode byte) ([]byte, error) {
	timeout := 1 * time.Second
	if d, ok := ctx.Deadline(); ok {
		timeout = time.Until(d)
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
func (b *Base) SendCommand(ctx context.Context, cmd protocol.Command) ([]byte, error) {
	if !b.IsConnected() {
		return nil, errors.ErrNotConnected
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
		return nil, errors.ErrRobotClosed
	}

	// Wait for response
	select {
	case resp := <-responseChan:
		return resp.data, resp.err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close shuts down the command loop gracefully
func (b *Base) Close() error {
	b.closeOnce.Do(func() {
		close(b.closeChan)
	})
	b.wg.Wait()
	return nil
}
