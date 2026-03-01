package msg

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

// Hooks define granular entry points for observability.
type Hooks struct {
	// OnDecodeError: Incoming bytes are malformed.
	OnDecodeError func(subject string, err error, rawData []byte)
	// OnHandlerError: The actor logic returned a non-nil error.
	OnHandlerError func(subject string, err error)
	// OnTransportError: Failed to send a response or publish a message.
	OnTransportError func(subject string, err error)
	// OnPanic: The actor crashed.
	OnPanic func(subject string, err any, stack []byte)
}

// The Result is the standard wire-format envelope for all Request-Reply interactions.
type Result[T any] struct {
	Data  T      `msgpack:"d,omitempty"`
	Error string `msgpack:"e,omitempty"`
}

type Messenger struct {
	nc    *nats.Conn
	hooks Hooks
}

// NewMessenger initializes the system and ensures all hooks are safe to call.
func NewMessenger(nc *nats.Conn, hooks Hooks) *Messenger {
	if hooks.OnDecodeError == nil {
		hooks.OnDecodeError = func(string, error, []byte) {}
	}
	if hooks.OnHandlerError == nil {
		hooks.OnHandlerError = func(string, error) {}
	}
	if hooks.OnTransportError == nil {
		hooks.OnTransportError = func(string, error) {}
	}
	if hooks.OnPanic == nil {
		hooks.OnPanic = func(string, any, []byte) {}
	}
	return &Messenger{nc: nc, hooks: hooks}
}

// --- PUB/SUB ---

func Bind[T any](m *Messenger, subject string, handler func(T, *nats.Msg)) (*nats.Subscription, error) {
	return m.nc.Subscribe(subject, func(msg *nats.Msg) {
		defer m.recoverPanic(msg.Subject)

		var payload T
		if err := msgpack.Unmarshal(msg.Data, &payload); err != nil {
			m.hooks.OnDecodeError(msg.Subject, err, msg.Data)
			return
		}

		handler(payload, msg)
	})
}

func (m *Messenger) Publish(subject string, v any) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		err = fmt.Errorf("marshal error: %w", err)
		m.hooks.OnTransportError(subject, err)
		return err
	}
	if err := m.nc.Publish(subject, data); err != nil {
		m.hooks.OnTransportError(subject, err)
		return err
	}
	return nil
}

// --- REQUEST/REPLY ---

func BindReply[TReq any, TRes any](m *Messenger, subject string, handler func(TReq, *nats.Msg) (TRes, error)) (*nats.Subscription, error) {
	return m.nc.Subscribe(subject, func(msg *nats.Msg) {
		// Recovery ensures we try to send an error back even on panic
		defer func() {
			if r := recover(); r != nil {
				m.hooks.OnPanic(msg.Subject, r, debug.Stack())
				m.sendRawError(msg, fmt.Sprintf("internal actor panic on %s", msg.Subject))
			}
		}()

		var req TReq
		if err := msgpack.Unmarshal(msg.Data, &req); err != nil {
			m.hooks.OnDecodeError(msg.Subject, err, msg.Data)
			m.sendRawError(msg, "protocol error: invalid request format")
			return
		}

		res, err := handler(req, msg)

		var result Result[TRes]
		if err != nil {
			m.hooks.OnHandlerError(msg.Subject, err)
			result.Error = err.Error()
		} else {
			result.Data = res
		}

		out, marshalErr := msgpack.Marshal(result)
		if marshalErr != nil {
			m.hooks.OnTransportError(msg.Subject, marshalErr)
			m.sendRawError(msg, "internal error: failed to encode response")
			return
		}

		if respErr := msg.Respond(out); respErr != nil {
			m.hooks.OnTransportError(msg.Subject, respErr)
		}
	})
}

func Request[TReq any, TRes any](m *Messenger, subject string, req TReq, timeout time.Duration) (TRes, error) {
	var final TRes

	data, err := msgpack.Marshal(req)
	if err != nil {
		return final, fmt.Errorf("request marshal error: %w", err)
	}

	msg, err := m.nc.Request(subject, data, timeout)
	if err != nil {
		// Note: Don't call OnTransportError here as the caller usually handles timeout/network errors
		return final, err
	}

	var res Result[TRes]
	if err := msgpack.Unmarshal(msg.Data, &res); err != nil {
		m.hooks.OnDecodeError(subject, err, msg.Data)
		return final, fmt.Errorf("response decode error: %w", err)
	}

	if res.Error != "" {
		return final, errors.New(res.Error)
	}

	return res.Data, nil
}

// --- HELPERS ---

func (m *Messenger) recoverPanic(subject string) {
	if r := recover(); r != nil {
		m.hooks.OnPanic(subject, r, debug.Stack())
	}
}

func (m *Messenger) sendRawError(msg *nats.Msg, errStr string) {
	res := Result[struct{}]{Error: errStr}
	out, _ := msgpack.Marshal(res)
	if err := msg.Respond(out); err != nil {
		m.hooks.OnTransportError(msg.Subject, err)
	}
}
