// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

const (
	CueListTypeSequential = "SEQUENTIAL"
)

type CueList struct {
	CueListId   string `msgpack:"cueListId" json:"cueListId"`
	Number      uint   `msgpack:"number" json:"number"`
	Label       string `msgpack:"label" json:"label"`
	CueListType string `msgpack:"cueListType" json:"cueListType"`
	WrapAtEnd   bool   `msgpack:"wrapAtEnd" json:"wrapAtEnd"`
}
