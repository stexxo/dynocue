package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/subsystems"
	"golang.org/x/sync/errgroup"
)

// AppManager orchestrates the lifecycle and dependency injection for all system show.
type AppManager struct {
	bus        *bus.Server
	subsystems map[string]subsystems.Subsystem
	mu         sync.RWMutex
	isStarted  bool
}

// NewAppManager initializes a new manager and the primary system EventBus.
func NewAppManager(defaultShowPath string, port int) *AppManager {
	server := bus.NewServer(port)
	return &AppManager{
		bus:        server,
		subsystems: make(map[string]subsystems.Subsystem),
	}
}

// Register adds a subsystem to the registry and performs the required dependency injection.
func (m *AppManager) Register(sub subsystems.Subsystem) error {
	if sub == nil {
		return errors.New("attempted to register a nil subsystem")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.subsystems[sub.Name()] = sub
	slog.Info("AppManager: Subsystem registered", "name", sub.Name())
	return nil
}

// Start triggers the sequential startup of all registered subsystems.
// If any subsystem returns an error, the startup sequence is aborted.
func (m *AppManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isStarted {
		return nil
	}

	// Start Bus
	err := m.bus.Start()
	if err != nil {
		return err
	}

	defer func() {
		if !m.isStarted {
			m.bus.Stop()
		}
	}()

	execgroup := errgroup.Group{}
	for name, sub := range m.subsystems {
		execgroup.Go(func() error {
			conn, err := m.bus.GetInProcessConn()
			if err != nil {
				return err
			}
			if err := sub.Start(conn); err != nil {
				return fmt.Errorf("AppManager: failed to start subsystem [%s]: %w", name, err)
			}
			slog.Debug(sub.Name() + " started")
			return nil
		})
	}

	err = execgroup.Wait()
	if err != nil {
		return err
	}

	m.isStarted = true
	slog.Info("AppManager: All subsystems active")
	return nil
}

// Stop executes a graceful shutdown for all registered subsystems.
func (m *AppManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isStarted {
		return nil
	}

	execgroup := errgroup.Group{}
	for name, sub := range m.subsystems {
		execgroup.Go(func() error {
			if err := sub.Stop(); err != nil {
				// We log the error but continue stopping other subsystems to ensure cleanup.
				slog.Error("AppManager: Error during subsystem shutdown",
					"name", name,
					"error", err,
				)
				return err
			}
			return nil
		})
	}

	err := execgroup.Wait()
	if err != nil {
		return err
	}

	m.bus.Stop()

	m.isStarted = false
	return nil
}
