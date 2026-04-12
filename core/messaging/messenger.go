// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package messaging

import (
	"cmp"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/vmihailenco/msgpack/v5"
)

type MessengerCfg struct {
	Conn      *nats.Conn          // Required
	Js        jetstream.JetStream // Required
	Validator *validator.Validate // optional, uses default if not provided
	Logger    logging.Logger      // optional, noop if not provided
}

type Messenger struct {
	conn          *nats.Conn
	js            jetstream.JetStream
	subscriptions map[string][]*nats.Subscription
	validator     *validator.Validate
	logger        logging.Logger
}

func NewMessenger(cfg *MessengerCfg) *Messenger {
	return &Messenger{
		conn:          cfg.Conn,
		js:            cfg.Js,
		subscriptions: make(map[string][]*nats.Subscription),
		validator:     cmp.Or(cfg.Validator, validator.New()),
		logger:        cmp.Or[logging.Logger](cfg.Logger, logging.NewNoopLogger()),
	}
}

func (m *Messenger) GetSubscriptions(subject string) ([]*nats.Subscription, bool) {
	subs, ok := m.subscriptions[subject]
	return subs, ok
}

func (m *Messenger) JetStream() jetstream.JetStream {
	return m.js
}

func (m *Messenger) Conn() *nats.Conn {
	return m.conn
}

func Publish[T any](m *Messenger, sub string, msg T) error {
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message with msgpack, %w", err)
	}
	return m.conn.Publish(sub, data)
}

func Request[T any](m *Messenger, subject string, msg any) (*ResponseEnvelope[T], error) {
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message with msgpack, %w", err)
	}

	resp, err := m.conn.Request(subject, data, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("failed to request message, %w", err)
	}

	out := new(ResponseEnvelope[T])
	err = msgpack.Unmarshal(resp.Data, out)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response, %w", err)
	}

	return out, nil
}

func RequestRetry[T any](m *Messenger, subject string, msg any, retries int, delay time.Duration) (*ResponseEnvelope[T], error) {
	var allErr error
	for i := 0; i < retries; i++ {
		resp, err := Request[T](m, subject, msg)
		if err == nil {
			return resp, nil
		}
		allErr = errors.Join(allErr, err)
		m.logger.Debug("request failed, retrying", "attempt", i+1, "retries", retries, "subject", subject, "error", err.Error())
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("failed to request message after %d retries, %w", retries, allErr)
}

func Subscribe[T any](m *Messenger, structValidation bool, subject string, handler SubscriptionHandler[*T]) error {
	sub, err := m.conn.Subscribe(subject, func(msg *nats.Msg) {
		message := new(T)
		err := msgpack.Unmarshal(msg.Data, message)
		if err != nil {
			m.logger.Error("failed to unmarshal message with msgpack", "subject", subject, "error", err)
			return
		}

		if structValidation {
			err = m.validator.Struct(message)
			if err != nil {
				m.logger.Error("failed to validate message", "subject", subject, "error", err)
				return
			}
		}

		handler(subject, message)
	})

	if err != nil {
		return err
	}
	m.subscriptions[subject] = append(m.subscriptions[subject], sub)
	return nil
}

func Reply[Req any, Resp any](m *Messenger, structValidation bool, subject string, handler ReplyHandler[*Req, Resp]) error {
	sub, err := m.conn.Subscribe(subject, func(msg *nats.Msg) {
		// Build Response Envelope
		// Defer ensure response is always sent
		resp := new(ResponseEnvelope[Resp])
		defer func() {
			outBytes, err := msgpack.Marshal(resp)
			if err != nil {
				m.logger.Error("failed to marshal response with msgpack", "subject", subject, "error", err)
				return
			}

			err = msg.Respond(outBytes)
			if err != nil {
				m.logger.Error("failed to respond message with msgpack", "subject", subject, "error", err)
				return
			}
		}()

		// Parse Request
		req := new(Req)
		err := msgpack.Unmarshal(msg.Data, req)
		if err != nil {
			m.logger.Error("failed to unmarshal message with msgpack", "subject", subject, "error", err)
			resp.Success = false
			resp.Error = "Could not parse request body."
			return
		}

		// Validate Request If Configured
		if structValidation {
			err = m.validator.Struct(req)
			if err != nil {
				m.logger.Error("failed to validate message body", "subject", subject, "error", err)
				resp.Success = false
				resp.Error = "Request body in invalid."
				return
			}
		}

		// Execute Handler and Handle Error
		out, err := handler(subject, req)
		if err != nil {
			m.logger.Error("failed to handle request", "subject", subject, "error", err)
			resp.Success = false
			t, ok := errors.AsType[*FriendlyError](err)
			if !ok {
				resp.Error = "Encountered unexpected error while processing request."
			} else {
				resp.Error = t.FriendlyErr
			}
			return
		}

		// Set Response
		resp.Success = true
		resp.Response = out
	})
	if err != nil {
		m.logger.Error("failed to subscribe to subject", "subject", subject, "error", err)
		return err
	}

	m.subscriptions[subject] = append(m.subscriptions[subject], sub)
	return nil
}
