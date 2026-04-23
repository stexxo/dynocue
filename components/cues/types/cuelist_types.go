// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

import (
	"github.com/stexxo/dynocue/util"
)

type CueingModel struct {
	CueLists *util.NumberedSlice[*CueList] `msgpack:"cueLists" json:"cueLists"`
}

func NewCueingModel() *CueingModel {
	return &CueingModel{
		CueLists: util.NewNumberedSlice[*CueList](),
	}
}

type CueList struct {
	Metadata CueListMetadata          `msgpack:"metadata" json:"metadata"`
	Cues     util.NumberedSlice[*Cue] `msgpack:"cues" json:"cues"`
}

type CueListMetadata struct {
	Number      float64 `msgpack:"number" json:"number"`
	Label       string  `msgpack:"label" json:"label"`
	CueListType string  `msgpack:"cueListType" json:"cueListType"`
}

func (c *CueList) Num() float64 {
	return c.Metadata.Number
}

func (c *CueList) SetNum(number float64) {
	c.Metadata.Number = number
}
