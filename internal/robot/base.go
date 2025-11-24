package robot

import (
	"context"
	"sync"

	"go.bug.st/serial"
	"github.com/yourusername/mycobot-go/protocol"
)

// Base provides common robot functionality
type Base struct {
	port      string
	baudrate  int
	useCRC    bool
	connected bool

	// Serial connection (owned by command loop goroutine)
	conn *serial.Port

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
