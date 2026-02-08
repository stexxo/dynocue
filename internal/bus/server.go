package bus

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Server struct {
	server *server.Server
	opts   *server.Options
	mu     sync.Mutex
	active bool
}

// NewServer creates a new Server instance with production-ready defaults.
// It does not start the server yet.
func NewServer(port int) *Server {
	opts := &server.Options{
		DontListen:    true,
		NoSigs:        true,            // We handle signals in our main app logic
		MaxPayload:    1024 * 1024 * 8, // 8MB limit
		PingInterval:  20 * time.Second,
		MaxPingsOut:   3,
		WriteDeadline: 10 * time.Second,
	}

	return &Server{
		opts: opts,
	}
}

// Start launches the embedded NATS server. It is thread-safe and
// prevents multiple starts.
func (b *Server) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.active {
		return errors.New("nats server is already running")
	}

	ns, err := server.NewServer(b.opts)
	if err != nil {
		return fmt.Errorf("nats initialization failed: %w", err)
	}

	// Start server in its own goroutine
	go ns.Start()

	// Wait for server to be operational
	if !ns.ReadyForConnections(5 * time.Second) {
		return errors.New("nats server failed to become ready")
	}

	b.server = ns
	b.active = true
	return nil
}

// GetInProcessConn creates a high-performance in-memory connection.
// Hardened with auto-reconnect logic and error handling.
func (b *Server) GetInProcessConn() (*nats.Conn, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.active {
		return nil, errors.New("cannot connect: bus is not started")
	}

	// Production connection options
	nc, err := nats.Connect("",
		nats.InProcessServer(b.server),
		nats.Name("Internal-Subsystem"),
		nats.MaxReconnects(-1), // Keep trying if internal pipe glitches
		nats.ReconnectWait(1*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			fmt.Printf("Subsystem disconnected: %v\n", err)
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create in-process connection: %w", err)
	}

	return nc, nil
}

// Stop gracefully shuts down the server and cleans up resources.
func (b *Server) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.server != nil {
		b.server.Shutdown()
		b.server.WaitForShutdown()
		b.active = false
	}
}
