package messaging

import (
	"errors"
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

func TestRequestReply(t *testing.T) {
	t.Parallel()
	s, nc := testServer()
	defer nc.Close()
	defer s.Shutdown()

	t.Run("Request-Reply Success", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		err := Reply(messenger, false, "test.request", func(s string, req string) (*string, error) {
			assert.Equal(t, "test.request", s)
			assert.Equal(t, "ping", req)
			resp := "pong"
			return &resp, nil
		})
		assert.NoError(t, err)

		resp, err := Request[string](messenger, "test.request", "ping")
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "pong", *resp.Response)
		assert.Empty(t, resp.Error)
	})

	t.Run("Request-Reply with validation success", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		type ReqResp struct {
			Data string `validate:"required"`
		}

		err := Reply(messenger, true, "test.validate", func(s string, req ReqResp) (*ReqResp, error) {
			return &ReqResp{Data: "Received: " + req.Data}, nil
		})
		assert.NoError(t, err)

		resp, err := Request[ReqResp](messenger, "test.validate", ReqResp{Data: "hello"})
		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "Received: hello", resp.Response.Data)
	})

	t.Run("Request-Reply with validation failure", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		type ReqResp struct {
			Data string `validate:"required"`
		}

		err := Reply(messenger, true, "test.validate.fail", func(s string, req ReqResp) (*ReqResp, error) {
			res := ReqResp{Data: "should not reach"}
			return &res, nil
		})
		assert.NoError(t, err)

		resp, err := Request[ReqResp](messenger, "test.validate.fail", ReqResp{Data: ""})
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Error, "Request body in invalid.")
	})

	t.Run("Request-Reply with FriendlyError including internal error", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		internalErr := errors.New("database connection failed")
		err := Reply(messenger, false, "test.friendly.internal", func(s string, req string) (*string, error) {
			return nil, &FriendlyError{
				Err:         internalErr,
				FriendlyErr: "Could not save data.",
			}
		})
		assert.NoError(t, err)

		resp, err := Request[string](messenger, "test.friendly.internal", "ping")
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "Could not save data.", resp.Error)
	})

	t.Run("Request-Reply with unexpected error", func(t *testing.T) {
		messenger := NewMessenger(&MessengerCfg{Conn: nc})

		err := Reply(messenger, false, "test.unexpected", func(s string, req string) (*string, error) {
			return nil, errors.New("something went wrong")
		})
		assert.NoError(t, err)

		resp, err := Request[string](messenger, "test.unexpected", "ping")
		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "Encountered unexpected error while processing request.", resp.Error)
	})

	t.Run("FriendlyError formatting", func(t *testing.T) {
		fe := &FriendlyError{
			Err:         errors.New("internal"),
			FriendlyErr: "friendly",
		}
		assert.Equal(t, "friendly: internal", fe.Error())
		assert.Equal(t, "internal", fe.Unwrap().Error())

		fe2 := &FriendlyError{
			FriendlyErr: "only friendly",
		}
		assert.Equal(t, "only friendly", fe2.Error())
	})
}
