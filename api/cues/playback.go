// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

const (
	CueStatusActive   = "PLAYING"
	CueStatusInactive = "INACTIVE"
)

const (
	RequestPlaybackExecute = "request.playback.execute"

	EventPlaybackState = "request.cuelist.playback.cue.state"
	EventSelectedCue   = "request.cuelist.playback.cue.selected"
)

type ExecuteCueInput struct {
	CueListNumber float64
	CueNumber     float64
	Silent        bool
}

type ExecuteCueOutput struct{}

type PlaybackStateEvent struct {
	CueListNumber float64
	CueNumber     float64
	Status        string
}

type PlaybackSelectedCueEvent struct {
	CueListNumber float64
	CueNumber     float64
}
