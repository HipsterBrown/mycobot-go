package robot

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"go.bug.st/serial"
	"github.com/yourusername/mycobot-go/internal/errors"
	"github.com/yourusername/mycobot-go/protocol"
)

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

			if _, err := b.conn.Write(data); err != nil {
				cmd.response <- &response{err: fmt.Errorf("write failed: %w", err)}
				continue
			}

			// Read response with context deadline
			respData, err := b.readResponse(cmd.ctx, cmd.request.Code)
			cmd.response <- &response{data: respData, err: err}
		}
	}
}

// readResponse reads and decodes response from serial port
func (b *Base) readResponse(ctx context.Context, expectedCode byte) ([]byte, error) {
	// Read response with timeout
	timeout := 1 * time.Second
	if d, ok := ctx.Deadline(); ok {
		timeout = time.Until(d)
	}

	// Set read timeout on the serial port
	if err := b.conn.SetReadTimeout(timeout); err != nil {
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	// Read until we have enough data for a minimal packet
	buf := make([]byte, 256)
	totalRead := 0

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		n, err := b.conn.Read(buf[totalRead:])
		if err != nil && err != io.EOF {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("read timeout")
			}
			continue
		}

		totalRead += n

		// Try to decode what we have
		if totalRead >= 5 {
			resp, err := protocol.Decode(buf[:totalRead], b.useCRC)
			if err == nil {
				return resp.Data, nil
			}
		}
	}

	return nil, fmt.Errorf("read timeout after %d bytes", totalRead)
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
