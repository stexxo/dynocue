package internal

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/stexxo/dynocue/dynod/internal/bus"
	"gitlab.com/stexxo/dynocue/dynod/internal/show"
)

type MockSubsystem struct {
	mock.Mock
}

func (m *MockSubsystem) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSubsystem) Start(client *bus.Client, sm *show.Show) error {
	args := m.Called(client, sm)
	return args.Error(0)
}

func (m *MockSubsystem) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewAppManager(t *testing.T) {
	t.Parallel()
	mgr := NewAppManager(rand.Intn(65535))
	assert.NotNil(t, mgr)
}

func TestAppManager_Register(t *testing.T) {
	t.Parallel()

	t.Run("successfully register a subsystem", func(t *testing.T) {
		t.Parallel()

		mgr := NewAppManager(rand.Intn(65535))
		assert.NotNil(t, mgr)

		sub := &MockSubsystem{}
		sub.On("Name").Return("sub")

		err := mgr.Register(sub)
		assert.NoError(t, err)
	})

	t.Run("fails to register a nil subsystem", func(t *testing.T) {
		t.Parallel()

		mgr := NewAppManager(rand.Intn(65535))
		assert.NotNil(t, mgr)
		err := mgr.Register(nil)
		assert.Error(t, err)
	})
}

func TestAppManager_StartStop(t *testing.T) {
	t.Parallel()

	t.Run("Fails to start a subsystem", func(t *testing.T) {
		t.Parallel()

		mgr := NewAppManager(rand.Intn(65535))
		assert.NotNil(t, mgr)
		sub := &MockSubsystem{}
		sub.On("Name").Return("sub")
		sub.On("Start", mock.AnythingOfType("*bus.Client"), mock.AnythingOfType("*show.Show")).Return(errors.New("error"))
		err := mgr.Register(sub)
		assert.NoError(t, err)
		err = mgr.Start()
		assert.Error(t, err)
	})

	t.Run("starts & stops successfully", func(t *testing.T) {
		t.Parallel()

		mgr := NewAppManager(rand.Intn(65535))
		assert.NotNil(t, mgr)
		sub := &MockSubsystem{}
		sub.On("Name").Return("sub")
		sub.On("Start", mock.AnythingOfType("*bus.Client"), mock.AnythingOfType("*show.Show")).Return(nil)
		sub.On("Stop").Return(nil)

		err := mgr.Register(sub)
		assert.NoError(t, err)

		err = mgr.Start()
		assert.NoError(t, err)

		err = mgr.Stop()
		assert.NoError(t, err)
	})

	t.Run("fails to stop a subsystem", func(t *testing.T) {
		t.Parallel()

		mgr := NewAppManager(rand.Intn(65535))
		assert.NotNil(t, mgr)
		sub := &MockSubsystem{}
		sub.On("Name").Return("sub")
		sub.On("Start", mock.AnythingOfType("*bus.Client"), mock.AnythingOfType("*show.Show")).Return(nil)
		sub.On("Stop").Return(errors.New("error"))

		err := mgr.Register(sub)
		assert.NoError(t, err)

		err = mgr.Start()
		assert.NoError(t, err)

		err = mgr.Stop()
		assert.Error(t, err)
	})
}
