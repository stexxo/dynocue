package proto

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

// --- Errors and Constants ---

// MsgError is a custom error structure for the messaging system.
type MsgError struct {
	Code    int    `msgpack:"code"`
	Message string `msgpack:"message"`
}

// Well-known messaging error codes.
const (
	ErrCodeSystemError    = 1000 + iota // Internal system error
	ErrCodeInvalidPayload               // Invalid request payload or parameters
	ErrCodeNotFound                     // Requested resource or subject not found
	ErrCodeTimeout                      // Request timed out
)

// --- Types ---

// MessageResponse contains a generic response body and an optional error.
type MessageResponse[T any] struct {
	Body  *T
	Error *MsgError
}

// Handler processes an incoming message of type T.
type Handler[T any] func(T)

// RequestHandler processes an incoming request of type T, returning a MessageResponse and an error.
type RequestHandler[T any, R any] func(T) (MessageResponse[R], error)

// msgResponse is an internal wrapper for responses.
type msgResponse struct {
	Body  msgpack.RawMessage `msgpack:"body"`
	Error *MsgError          `msgpack:"error"`
}

// --- Messaging Helpers ---

// Publish serializes the body using MsgPack and publishes it to the subject.
func Publish(nc *nats.Conn, subject string, body any) error {
	data, err := msgpack.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	return nc.Publish(subject, data)
}

// Subscribe listens on a subject and processes incoming messages using the provided handler.
func Subscribe[T any](nc *nats.Conn, subject string, handler Handler[T]) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(msg *nats.Msg) {
		var body T
		if err := msgpack.Unmarshal(msg.Data, &body); err != nil {
			return
		}
		handler(body)
	})
}

// Request sends a request and returns the deserialized response body and any errors.
func Request[T any](nc *nats.Conn, subject string, requestBody any, timeout time.Duration) (*T, *MsgError, error) {
	data, err := msgpack.Marshal(requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	respMsg, err := nc.Request(subject, data, timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("nats request failed: %w", err)
	}

	var msgResp msgResponse
	if err := msgpack.Unmarshal(respMsg.Data, &msgResp); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var result *T
	if len(msgResp.Body) > 0 {
		var val T
		if err := msgpack.Unmarshal(msgResp.Body, &val); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal response body: %w", err)
		}
		result = &val
	}

	return result, msgResp.Error, nil
}

// Respond handles incoming requests by calling the provided handler and sending back a response.
func Respond[T any, R any](nc *nats.Conn, subject string, handler RequestHandler[*T, *R]) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(msg *nats.Msg) {
		request := new(T)
		if err := msgpack.Unmarshal(msg.Data, request); err != nil {
			msgResp := msgResponse{
				Error: &MsgError{Code: ErrCodeInvalidPayload, Message: "Invalid request payload: " + err.Error()},
			}
			finalData, _ := msgpack.Marshal(msgResp)
			_ = msg.Respond(finalData)
			return
		}

		resp, err := handler(request)

		var msgErr *MsgError
		var responseBody any

		if err != nil {
			msgErr = &MsgError{Code: ErrCodeSystemError, Message: err.Error()}
		} else {
			msgErr = resp.Error
			responseBody = resp.Body
		}

		var respData []byte
		if responseBody != nil {
			respData, err = msgpack.Marshal(responseBody)
			if err != nil {
				msgErr = &MsgError{Code: ErrCodeSystemError, Message: "System error during marshaling"}
				respData = nil
			}
		}

		msgResp := msgResponse{
			Body:  respData,
			Error: msgErr,
		}

		finalData, _ := msgpack.Marshal(msgResp)
		_ = msg.Respond(finalData)
	})
}

// --- Messenger ---

// Messenger manages NATS subscriptions.
type Messenger struct {
	nc   *nats.Conn
	subs map[string][]*nats.Subscription
}

// NewMessenger creates a new Messenger instance.
func NewMessenger(nc *nats.Conn) *Messenger {
	return &Messenger{
		nc:   nc,
		subs: make(map[string][]*nats.Subscription),
	}
}

// Subscriptions returns all active subscriptions managed by the Messenger.
func (m *Messenger) Subscriptions() []*nats.Subscription {
	var allSubs []*nats.Subscription
	for _, subs := range m.subs {
		allSubs = append(allSubs, subs...)
	}
	return allSubs
}

// Close unsubscribes all managed subscriptions.
func (m *Messenger) Close() error {
	var errs error
	for subject, subs := range m.subs {
		for _, sub := range subs {
			if err := sub.Unsubscribe(); err != nil {
				errs = fmt.Errorf("failed to unsubscribe from %s: %w", subject, err)
			}
		}
	}
	m.subs = nil
	return errs
}

// Handle registers a request handler and manages its subscription.
func Handle[T any, R any](m *Messenger, subject string, handler RequestHandler[*T, *R]) error {
	sub, err := Respond(m.nc, subject, handler)
	if err != nil {
		return err
	}
	m.subs[subject] = append(m.subs[subject], sub)
	return nil
}
