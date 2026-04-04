package messaging

import (
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stretchr/testify/assert"
)

func testServer() (*server.Server, *nats.Conn) {
	s, _ := server.NewServer(&server.Options{DontListen: true})
	s.Start()
	nc, _ := nats.Connect("", nats.InProcessServer(s))
	return s, nc
}

func TestNewMessenger(t *testing.T) {
	t.Parallel()
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	t.Run("NewMessenger creates a new Messenger instance using provided configuration", func(t *testing.T) {
		cfg := &MessengerCfg{
			Conn:      nc,
			Validator: validator.New(),
			Logger:    logging.NewNoopLogger(),
		}

		messenger := NewMessenger(cfg)
		assert.NotNil(t, messenger)
		assert.NotNil(t, messenger.subscriptions)
		assert.Same(t, cfg.Conn, messenger.conn)
		assert.Same(t, cfg.Validator, messenger.validator)
		assert.Same(t, cfg.Logger, messenger.logger)
	})

	t.Run("NewMessenger creates a new Messenger instance and uses defaults for config options not specified", func(t *testing.T) {
		cfg := &MessengerCfg{
			Conn: nc,
		}
		messenger := NewMessenger(cfg)
		assert.NotNil(t, messenger)
		assert.Same(t, cfg.Conn, messenger.conn)
		assert.NotNil(t, messenger.validator)
		assert.NotNil(t, messenger.logger)
		assert.NotNil(t, messenger.subscriptions)
	})
}

func TestPublishSubscribe(t *testing.T) {
	t.Parallel()
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	t.Run("Subscribe no validation", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		wg := sync.WaitGroup{}
		wg.Add(1)
		err := Subscribe[string](messenger, false, "test.test", func(s string, s2 *string) {
			assert.Equal(t, "test.test", s)
			assert.Equal(t, "value", *s2)
			wg.Done()
		})
		assert.NoError(t, err)

		err = Publish(messenger, "test.test", "value")
		assert.NoError(t, err)

		wg.Wait()

		subs, ok := messenger.GetSubscriptions("test.test")
		assert.True(t, ok)
		assert.Len(t, subs, 1)
	})

	t.Run("Subscribe with validation", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		wg := sync.WaitGroup{}
		wg.Add(1)
		err := Subscribe[struct {
			Test string `validate:"required"`
		}](messenger, true, "test.test", func(s string, s2 *struct {
			Test string `validate:"required"`
		}) {
			assert.Equal(t, "test.test", s)
			assert.Equal(t, "value", s2.Test)
			wg.Done()
		})
		assert.NoError(t, err)

		err = Publish(messenger, "test.test", &struct {
			Test string `validate:"required"`
		}{Test: "value"})
		assert.NoError(t, err)

		wg.Wait()

		subs, ok := messenger.GetSubscriptions("test.test")
		assert.True(t, ok)
		assert.Len(t, subs, 1)
	})
}
