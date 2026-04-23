// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import "github.com/google/uuid"

type Cue struct {
	Metadata CueMetadata `msgpack:"metadata" json:"metadata"`
}

func NewCue(number float64) *Cue {
	return &Cue{
		Metadata: CueMetadata{
			Id:     uuid.NewString(),
			Number: number,
		},
	}
}

type CueMetadata struct {
	Id     string  `msgpack:"id" json:"id"`
	Number float64 `msgpack:"number" json:"number"`
	Label  string  `msgpack:"label" json:"label"`
}

func (c *Cue) Num() float64 {
	return c.Metadata.Number
}

func (c *Cue) SetNum(n float64) {
	c.Metadata.Number = n
}
