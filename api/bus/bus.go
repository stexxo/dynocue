package bus

import (
	"fmt"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

// Message bus error codes.
const (
	InvalidPayloadCode  = 400 // Invalid payload
	NotFoundCode        = 404 // Resource not found
	ConflictCode        = 409 // Resource conflict
	ValidationErrorCode = 422 // Validation failed
	InternalErrorCode   = 500 // Internal processing failure
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// Validate performs structural validation on the given object.
func Validate(obj any) error {
	if obj == nil {
		return nil
	}
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Struct || (rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct) {
		return validate.Struct(obj)
	}
	return nil
}

// MessageError contains error information for the message bus.
type MessageError struct {
	Code         int    `msgpack:"code"`
	ErrorMessage string `msgpack:"errorMessage"`
}

// MessageResponse is a generic response wrapper that includes optional error information.
type MessageResponse[T any] struct {
	ResponseValue *T            `msgpack:"responseValue,omitzero"`
	MessageError  *MessageError `msgpack:"messageError,omitzero"`
}

// Publish validates and serializes the message using msgpack and publishes it to the given subject.
func Publish[T any](nc *nats.Conn, subject string, msg T) error {
	if err := Validate(msg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	data, err := msgpack.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return nc.Publish(subject, data)
}

// SubscribeHandler is a function that processes a published message.
type SubscribeHandler[T any] func(string, T)

// Subscribe deserializes the incoming message using msgpack and calls the handler.
func Subscribe[T any](nc *nats.Conn, subject string, handler SubscribeHandler[T]) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(m *nats.Msg) {
		var msg T
		if err := msgpack.Unmarshal(m.Data, &msg); err != nil {
			// In a real app, we might want to log this error or handle it differently.
			return
		}
		handler(m.Subject, msg)
	})
}

// Request validates and serializes the request using msgpack, sends it, and deserializes the response.
func Request[Req any, Res any](nc *nats.Conn, subject string, req Req, timeout time.Duration) (Res, error) {
	var res Res
	if err := Validate(req); err != nil {
		return res, fmt.Errorf("validation failed: %w", err)
	}

	data, err := msgpack.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := nc.Request(subject, data, timeout)
	if err != nil {
		return res, err
	}

	var msgRes MessageResponse[Res]
	if err := msgpack.Unmarshal(msg.Data, &msgRes); err != nil {
		return res, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if msgRes.MessageError != nil {
		return res, fmt.Errorf("bus error [%d]: %s", msgRes.MessageError.Code, msgRes.MessageError.ErrorMessage)
	}

	if msgRes.ResponseValue != nil {
		return *msgRes.ResponseValue, nil
	}

	return res, nil
}

// ReplyHandler is a function that processes a request and returns a MessageResponse.
type ReplyHandler[Req any, Res any] func(string, Req) (*MessageResponse[Res], error)

// Reply listens for requests, deserializes the request, calls the handler,
// and serializes the response back to the requester.
func Reply[Req any, Res any](nc *nats.Conn, subject string, handler ReplyHandler[Req, Res]) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(m *nats.Msg) {
		var req Req
		if err := msgpack.Unmarshal(m.Data, &req); err != nil {
			msgRes := &MessageResponse[Res]{
				MessageError: &MessageError{
					Code:         InvalidPayloadCode,
					ErrorMessage: fmt.Sprintf("failed to unmarshal request: %s", err.Error()),
				},
			}

			data, marshalErr := msgpack.Marshal(msgRes)
			if marshalErr != nil {
				_ = m.Respond(nil)
				return
			}

			_ = m.Respond(data)
			return
		}

		if err := Validate(req); err != nil {
			msgRes := &MessageResponse[Res]{
				MessageError: &MessageError{
					Code:         ValidationErrorCode,
					ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
				},
			}

			data, marshalErr := msgpack.Marshal(msgRes)
			if marshalErr != nil {
				_ = m.Respond(nil)
				return
			}

			_ = m.Respond(data)
			return
		}

		msgRes, handlerErr := handler(m.Subject, req)

		if handlerErr != nil && msgRes == nil {
			msgRes = &MessageResponse[Res]{
				MessageError: &MessageError{
					Code:         InternalErrorCode,
					ErrorMessage: handlerErr.Error(),
				},
			}
		}

		if msgRes == nil {
			msgRes = &MessageResponse[Res]{}
		}

		data, err := msgpack.Marshal(msgRes)
		if err != nil {
			return
		}

		_ = m.Respond(data)
	})
}
