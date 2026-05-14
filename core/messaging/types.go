// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package messaging

import (
	"fmt"
)

type ResponseEnvelope[T any] struct {
	Success  bool   `json:"success,omitempty" msgpack:"success,omitzero"`
	Response *T     `json:"response,omitempty" msgpack:"response,omitzero"`
	Error    string `json:"error,omitempty" msgpack:"error,omitzero"`
}

type SubscriptionHandler[T any] func(string, T)

type ReplyHandler[T any, E any] func(string, T) (*E, error)

type FriendlyError struct {
	Err         error
	FriendlyErr string
}

func (e *FriendlyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.FriendlyErr, e.Err)
	}
	return e.FriendlyErr
}

func (e *FriendlyError) Unwrap() error {
	return e.Err
}
