package bus

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	ibus "gitlab.com/stexxo/dynocue/internal/bus"
)

type TestMsg struct {
	ID    int    `msgpack:"id"`
	Value string `msgpack:"value"`
}

func TestBusGeneric(t *testing.T) {
	// Start an in-process NATS server
	ns, err := ibus.NewBus()
	require.NoError(t, err)
	defer ns.Shutdown()

	nc, err := ibus.GetInProcessConn(ns)
	require.NoError(t, err)
	defer nc.Close()

	t.Run("Publish and Subscribe", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		var receivedMsg TestMsg
		subject := "test.pubsub"

		sub, err := Subscribe(nc, subject, func(s string, msg TestMsg) {
			assert.Equal(t, subject, s)
			receivedMsg = msg
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		sentMsg := TestMsg{ID: 1, Value: "hello"}
		err = Publish(nc, subject, sentMsg)
		require.NoError(t, err)

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, sentMsg, receivedMsg)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for message")
		}
	})

	t.Run("Request and Reply", func(t *testing.T) {
		subject := "test.request"

		sub, err := Reply(nc, subject, func(sub string, req TestMsg) (*MessageResponse[TestMsg], error) {
			assert.Equal(t, subject, sub)
			if req.ID == -2 {
				return nil, assert.AnError
			}
			if req.ID < 0 {
				return &MessageResponse[TestMsg]{
					MessageError: &MessageError{Code: 400, ErrorMessage: "invalid id"},
				}, nil
			}
			res := TestMsg{ID: req.ID, Value: "echo: " + req.Value}
			return &MessageResponse[TestMsg]{
				ResponseValue: &res,
			}, nil
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		t.Run("Success", func(t *testing.T) {
			reqMsg := TestMsg{ID: 2, Value: "world"}
			resMsg, err := Request[TestMsg, TestMsg](nc, subject, reqMsg, 1*time.Second)
			require.NoError(t, err)

			assert.Equal(t, reqMsg.ID, resMsg.ID)
			assert.Equal(t, "echo: world", resMsg.Value)
		})

		t.Run("Custom Error", func(t *testing.T) {
			reqMsg := TestMsg{ID: -1, Value: "bad"}
			_, err := Request[TestMsg, TestMsg](nc, subject, reqMsg, 1*time.Second)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "bus error [400]: invalid id")
		})

		t.Run("Go Error", func(t *testing.T) {
			reqMsg := TestMsg{ID: -2, Value: "worse"}
			_, err := Request[TestMsg, TestMsg](nc, subject, reqMsg, 1*time.Second)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "bus error [500]: assert.AnError")
		})

		t.Run("Invalid Payload", func(t *testing.T) {
			// Send raw invalid msgpack data
			invalidData := []byte{0x91} // msgpack fixarray with 1 element, but no element following it
			msg, err := nc.Request(subject, invalidData, 1*time.Second)
			require.NoError(t, err)

			var msgRes MessageResponse[TestMsg]
			err = msgpack.Unmarshal(msg.Data, &msgRes)
			require.NoError(t, err)

			require.NotNil(t, msgRes.MessageError)
			assert.Equal(t, InvalidPayloadCode, msgRes.MessageError.Code)
			assert.Contains(t, msgRes.MessageError.ErrorMessage, "failed to unmarshal request")
		})

		t.Run("Validation Error", func(t *testing.T) {
			type ValidatedMsg struct {
				ID int `msgpack:"id" validate:"gt=0"`
			}
			vSub, err := Reply(nc, "test.validation", func(s string, req ValidatedMsg) (*MessageResponse[ValidatedMsg], error) {
				return &MessageResponse[ValidatedMsg]{ResponseValue: &req}, nil
			})
			require.NoError(t, err)
			defer vSub.Unsubscribe()

			_, err = Request[ValidatedMsg, ValidatedMsg](nc, "test.validation", ValidatedMsg{ID: 0}, 1*time.Second)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "validation failed")
		})

		t.Run("Timeout", func(t *testing.T) {
			_, err := Request[TestMsg, TestMsg](nc, "non.existent.subject", TestMsg{ID: 1}, 10*time.Millisecond)
			require.Error(t, err)
		})

		t.Run("Invalid Response Payload", func(t *testing.T) {
			sub2, err := nc.Subscribe("test.invalid.res", func(m *nats.Msg) {
				_ = m.Respond([]byte{0x91}) // invalid msgpack
			})
			require.NoError(t, err)
			defer sub2.Unsubscribe()

			_, err = Request[TestMsg, TestMsg](nc, "test.invalid.res", TestMsg{ID: 1}, 1*time.Second)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to unmarshal response")
		})

		t.Run("Nil Response Value", func(t *testing.T) {
			sub2, err := Reply(nc, "test.nil.res", func(_ string, req TestMsg) (*MessageResponse[TestMsg], error) {
				return &MessageResponse[TestMsg]{}, nil
			})
			require.NoError(t, err)
			defer sub2.Unsubscribe()

			res, err := Request[TestMsg, TestMsg](nc, "test.nil.res", TestMsg{ID: 1}, 1*time.Second)
			require.NoError(t, err)
			assert.Equal(t, 0, res.ID)
		})
	})

	t.Run("Publish Error", func(t *testing.T) {
		closedNc, err := ibus.GetInProcessConn(ns)
		require.NoError(t, err)
		closedNc.Close()

		err = Publish(closedNc, "test", TestMsg{ID: 1})
		require.Error(t, err)
	})

	t.Run("Subscribe Unmarshal Error", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		// This is tricky because the handler isn't called if unmarshal fails.
		// We'll use a second subscription to know when the message has been processed.
		subject := "test.sub.unmarshal"
		_, err := Subscribe(nc, subject, func(_ string, msg TestMsg) {
			t.Error("handler should not be called")
		})
		require.NoError(t, err)

		// Send invalid data
		err = nc.Publish(subject, []byte{0x91})
		require.NoError(t, err)

		// Just wait a bit to ensure the message was processed (or ignored)
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Server Connection Errors", func(t *testing.T) {
		t.Run("GetInProcessConn - Stopped Server", func(t *testing.T) {
			ns2, err := ibus.NewBus()
			require.NoError(t, err)
			ns2.Shutdown()

			_, err = ibus.GetInProcessConn(ns2)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "bus is not started")
		})

		t.Run("GetInProcessConn - Failed Connection", func(t *testing.T) {
			// This is hard to trigger with in-process server, but we can try
			// by passing a nil server if the function doesn't check for it.
			// Actually the function checks s.Running().
			_, err = ibus.GetInProcessConn(nil)
			require.Error(t, err)
		})
	})

	t.Run("Marshaling Errors", func(t *testing.T) {
		// Use a type that cannot be marshaled by msgpack
		// Functions cannot be marshaled.
		type NotMarshalable struct {
			Fn func()
		}
		bad := NotMarshalable{Fn: func() {}}

		t.Run("Publish Marshal Error", func(t *testing.T) {
			err := Publish(nc, "test", bad)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to marshal message")
		})

		t.Run("Request Marshal Error", func(t *testing.T) {
			_, err := Request[any, any](nc, "test", bad, 1*time.Second)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to marshal request")
		})

		t.Run("Reply Marshal Error", func(t *testing.T) {
			sub2, err := Reply(nc, "test.marshal.reply", func(_ string, req int) (*MessageResponse[NotMarshalable], error) {
				// req should be 1
				if req != 1 {
					return nil, fmt.Errorf("unexpected req: %d", req)
				}
				return &MessageResponse[NotMarshalable]{
					ResponseValue: &bad,
				}, nil
			})
			require.NoError(t, err)
			defer sub2.Unsubscribe()

			data, _ := msgpack.Marshal(1)
			_, err = nc.Request("test.marshal.reply", data, 500*time.Millisecond)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "timeout")
		})

	})
}
