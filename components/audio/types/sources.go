// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package types

type AudioFile struct {
	FileId string `msgpack:"fileId" json:"fileId"`
	Key    string `msgpack:"key" json:"key"`
}

type AudioSource struct {
	SourceId string `msgpack:"sourceId" json:"sourceId"`
	FileId   string `msgpack:"fileId" json:"fileId"`
	Label    string `msgpack:"sourceLabel" json:"sourceLabel"`
}
