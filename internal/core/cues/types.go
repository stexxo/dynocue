// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

import "gitlab.com/stexxo/dynocue/internal/data"

var (
	BucketCueListKey = data.NewStringBucketKey("cuelists", true)
	BucketCuesKey    = data.NewStringBucketKey("cues", true)
	BucketActionsKey = data.NewStringBucketKey("actions", true)

	KeyMetadata = []byte("metadata")
)

type CueListMetadataDbModel struct {
	Label    string `msgpack:"label"`
	ListType string `msgpack:"listType"`
}

type CueMetadataDbModel struct {
	Label string `msgpack:"label"`
}

type ActionDbModel struct {
	Label string `msgpack:"label"`
}
