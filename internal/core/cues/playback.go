// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

import (
	"sync"
)

type PlaybackManager struct {
	mu       sync.Mutex
	selected map[float64]float64
}

func NewPlaybackManager() *PlaybackManager {
	return &PlaybackManager{}
}
