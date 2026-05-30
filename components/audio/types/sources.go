// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

type AudioFile struct {
	FileId string `msgpack:"fileId" json:"fileId"`
	Number uint   `msgpack:"number" json:"number"`
	Key    string `msgpack:"key" json:"key"`
	Label  string `msgpack:"label" json:"label"`
}

type AudioSource struct {
	SourceId string `msgpack:"sourceId" json:"sourceId"`
	Number   uint   `msgpack:"number" json:"number"`
	FileId   string `msgpack:"fileId" json:"fileId"`
	Label    string `msgpack:"sourceLabel" json:"sourceLabel"`
}
