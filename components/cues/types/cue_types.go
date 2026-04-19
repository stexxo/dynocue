// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

type Cue struct {
	Metadata CueMetadata `msgpack:"metadata" json:"metadata"`
}

type CueMetadata struct {
	Number float64
	Label  string
}

func (c *Cue) Num() float64 {
	return c.Metadata.Number
}

func (c *Cue) SetNum(n float64) {
	c.Metadata.Number = n
}
