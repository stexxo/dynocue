// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

const ExecuteCueRequestSubject = "request.cueing.execute"
const ExecuteCueResponseSubject = "response.cueing.execute"

type ExecuteCueRequest struct {
	CueListId string `msgpack:"cueListId"`
	CueId     string `msgpack:"cueId"`
}
