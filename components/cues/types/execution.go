// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import "time"

type CueExecution struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`

	Active       bool      `msgpack:"active" json:"active"`
	Selected     bool      `msgpack:"selected" json:"selected"`
	CueExecStart time.Time `msgpack:"cueExecStart" json:"cueExecStart"`

	DelayActive bool      `msgpack:"delayActive" json:"delayActive"`
	DelayStart  time.Time `msgpack:"delayStart" json:"delayStart"`

	FollowActive bool      `msgpack:"followActive" json:"followActive"`
	FollowStart  time.Time `msgpack:"followStart" json:"followStart"`
}

type ActionExecution struct {
	CueListId string `msgpack:"cueListId" json:"cueListId"`
	CueId     string `msgpack:"cueId" json:"cueId"`
	ActionId  string `msgpack:"actionId" json:"actionId"`

	ActionStarted time.Time `msgpack:"actionStarted" json:"actionStarted"`

	DelayActive  bool      `msgpack:"delayActive" json:"delayActive"`
	DelayStarted time.Time `msgpack:"delayStarted" json:"delayStarted"`
}
