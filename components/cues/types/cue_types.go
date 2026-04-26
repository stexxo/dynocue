// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/stexxo/dynocue/util"
)

type Cue struct {
	Metadata CueMetadata                  `msgpack:"metadata" json:"metadata"`
	Actions  *util.NumberedSlice[*Action] `msgpack:"actions" json:"actions"`
}

func NewCue(cueListId string, number float64) *Cue {
	return &Cue{
		Metadata: CueMetadata{
			CueListId: cueListId,
			CueId:     uuid.NewString(),
			Number:    number,
		},
		Actions: util.NewNumberedSlice[*Action](),
	}
}

type CueMetadata struct {
	CueListId   string        `msgpack:"cueListId" json:"cueListId"`
	CueId       string        `msgpack:"cueId" json:"cueId"`
	Number      float64       `msgpack:"number" json:"number"`
	Label       string        `msgpack:"label" json:"label"`
	Description string        `msgpack:"description" json:"description"`
	Delay       time.Duration `msgpack:"delay" json:"delay"`
	Follow      time.Duration `msgpack:"follow" json:"follow"`
}

func (c *Cue) Num() float64 {
	return c.Metadata.Number
}

func (c *Cue) SetNum(n float64) {
	c.Metadata.Number = n
}
