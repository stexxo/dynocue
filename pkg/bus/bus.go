// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package bus

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
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
	ErrorMessage string `msgpack:"errorMessage"`
}

func NewMessageError(errorMessage string) *MessageError {
	return &MessageError{
		ErrorMessage: errorMessage,
	}
}

// MessageResponse is a generic response wrapper that includes optional error information.
type MessageResponse[T any] struct {
	ResponseValue *T            `msgpack:"responseValue,omitzero"`
	MessageError  *MessageError `msgpack:"messageError,omitzero"`
}

func NewMessageResponse[T any](v *T, Err *MessageError) *MessageResponse[T] {
	return &MessageResponse[T]{
		ResponseValue: v,
		MessageError:  Err,
	}
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
	slog.Debug("publishing message to NATS", "subject", subject, "data", string(data))

	return nc.Publish(subject, data)
}

// SubscribeHandler is a function that processes a published message.
type SubscribeHandler[T any] func(string, T)

// Subscribe deserializes the incoming message using msgpack and calls the handler.
func Subscribe[T any](nc *nats.Conn, subject string, handler SubscribeHandler[*T]) (*nats.Subscription, error) {
	return nc.Subscribe(subject, func(m *nats.Msg) {
		msg := new(T)
		if err := msgpack.Unmarshal(m.Data, msg); err != nil {
			// In a real app, we might want to log this error or handle it differently.
			return
		}

		if err := Validate(*msg); err != nil {
			return
		}

		handler(m.Subject, msg)
	})
}

// Request validates and serializes the request using msgpack, sends it, and deserializes the response.
func Request[Req any, Res any](nc *nats.Conn, subject string, req Req) (*MessageResponse[Res], error) {
	data, err := msgpack.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	msg, err := nc.Request(subject, data, 100*time.Millisecond)
	if err != nil {
		return nil, err
	}

	msgRes := new(MessageResponse[Res])
	if err := msgpack.Unmarshal(msg.Data, msgRes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return msgRes, nil
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
