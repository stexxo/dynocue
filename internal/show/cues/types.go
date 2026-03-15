package cues

import "gitlab.com/stexxo/dynocue/internal/data"

var (
	BucketCueListKey = data.NewStringBucketKey("cuelists", true)
	BucketCuesKey    = data.NewStringBucketKey("cues", true)
	BucketActionsKey = data.NewStringBucketKey("actions", true)

	KeyMetadata = []byte("metadata")
)
