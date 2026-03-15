package cues

import (
	"errors"
	"log/slog"
	"slices"

	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/data"
	"gitlab.com/stexxo/dynocue/internal/utils"
	apibus "gitlab.com/stexxo/dynocue/pkg/bus"
	"go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

type cueMetadata struct {
	Label string `msgpack:"label"`
}

// NewCue creates a new cue within a specific cue list.
func (c *CueSystem) NewCue(sub string, in apicues.CreateCueInput) (*apibus.MessageResponse[apicues.CreateCueOutput], error) {
	md := &cueMetadata{}
	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		_, key, err := data.AddIncrementedSubBucket(
			tx,
			BucketCueListKey,
			[]data.BucketKey{data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey},
			in.Number,
			data.NewResourceBootstrap{
				InitialValues: []data.KeyValuePair{{Key: KeyMetadata, Value: md}},
				Buckets:       []data.BucketKey{BucketActionsKey},
			},
		)
		outNum = key
		return err
	})
	if err != nil {
		slog.Error("failed to create new cue", "err", err.Error(), "number", outNum, "cueList", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.CreateCueOutput](nil, apibus.NewMessageError("failed to create cue")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCue, apicues.NewCueEvent{
		CueListNumber: in.CueListNumber,
		Number:        outNum,
		Label:         md.Label,
	}); err != nil {
		slog.Error("failed to publish change event for new cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.CreateCueOutput]{
		ResponseValue: &apicues.CreateCueOutput{
			CueListNumber: in.CueListNumber,
			Number:        outNum,
		},
	}, nil
}

// UpdateCueMetadata updates the metadata fields of an existing cue.
func (c *CueSystem) UpdateCueMetadata(sub string, in apicues.UpdateCueMetadataInput) (*apibus.MessageResponse[apicues.UpdateCueMetadataOutput], error) {
	var outMetadata *cueMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.Number, false))
		if err != nil {
			return err
		}
		outMetadata, err = data.UpdateAttributeInKeyValuePair[cueMetadata](b, KeyMetadata, in.Key, in.Value)
		return err
	})
	if err != nil {
		slog.Error("failed to update cue metdata", "err", err.Error(), "cuelist", in.Number, "cue", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.UpdateCueMetadataOutput](nil, apibus.NewMessageError("failed to update cue")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCue, apicues.UpdateCueMetadataEvent{
		CueListNumber: in.CueListNumber,
		Number:        in.Number,
		Label:         outMetadata.Label,
	}); err != nil {
		slog.Error("failed to publish change event for update cue metadata", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateCueMetadataOutput]{
		ResponseValue: &apicues.UpdateCueMetadataOutput{},
	}, nil
}

// GetCueMetadata retrieves the metadata for a specific cue.
func (c *CueSystem) GetCueMetadata(sub string, in apicues.GetCueMetadataInput) (*apibus.MessageResponse[apicues.GetCueMetadataOutput], error) {
	var md *cueMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.Number, false), BucketCuesKey, data.NewFloatBucketKey(in.Number, false))
		if err != nil {
			return err
		}
		md = new(cueMetadata)
		return data.GetKey(b, md, KeyMetadata)
	})

	if err != nil {
		slog.Error("failed to get cue metadata", "err", err.Error(), "cuelist", in.Number, "cue", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.GetCueMetadataOutput](nil, apibus.NewMessageError("failed to retrieve cue metadata")), nil
	}

	return &apibus.MessageResponse[apicues.GetCueMetadataOutput]{
		ResponseValue: &apicues.GetCueMetadataOutput{
			CueListNumber: in.CueListNumber,
			Number:        in.Number,
			Label:         md.Label,
		},
	}, nil
}

// EnumerateCue returns a list of all cues within a specific cue list.
func (c *CueSystem) EnumerateCue(sub string, in apicues.EnumerateCueInput) (*apibus.MessageResponse[apicues.EnumerateCueOutput], error) {
	values := map[float64]*cueMetadata{}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
		if errors.Is(err, data.ErrNoBucket) {
			return nil
		}
		if err != nil {
			return err
		}
		values, err = data.EnumerateKeysFromSubBuckets[float64, cueMetadata](b, KeyMetadata, func(bytes []byte) float64 {
			k, err := utils.BytesToFloat64(bytes)
			if err != nil {
				return 0
			}
			return k
		})
		return err
	})
	if err != nil {
		slog.Error("failed to enumerate cues", "err", err.Error(), "cuelist", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.EnumerateCueOutput](nil, apibus.NewMessageError("failed to enumerate cues")), nil
	}

	out := make([]apicues.GetCueMetadataOutput, 0, len(values))
	for k, v := range values {
		i, _ := slices.BinarySearchFunc(out, k, func(a apicues.GetCueMetadataOutput, b float64) int {
			if a.Number < b {
				return -1
			} else if a.Number > b {
				return 1
			}
			return 0
		})
		out = slices.Insert(out, i, apicues.GetCueMetadataOutput{
			CueListNumber: in.CueListNumber,
			Number:        k,
			Label:         v.Label,
		})
	}

	return &apibus.MessageResponse[apicues.EnumerateCueOutput]{
		ResponseValue: &apicues.EnumerateCueOutput{
			CueListNumber: in.CueListNumber,
			Cues:          out,
		},
	}, nil
}

// DeleteCue removes a specific cue from a cue list.
func (c *CueSystem) DeleteCue(sub string, in apicues.DeleteCueInput) (*apibus.MessageResponse[apicues.DeleteCueOutput], error) {

	err := c.db.Update(func(tx *bbolt.Tx) error {
		return data.DeleteBucketByPath(tx, utils.Float64ToBytes(in.Number), BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
	})

	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		slog.Error("failed to delete cue", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.Number)
		return apibus.NewMessageResponse[apicues.DeleteCueOutput](nil, apibus.NewMessageError("failed to delete cue")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteCue, apicues.DeleteCueEvent{
		CueListNumber: in.CueListNumber,
		Number:        in.Number,
	}); err != nil {
		slog.Error("failed to publish change event for delete cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteCueOutput]{
		ResponseValue: &apicues.DeleteCueOutput{},
	}, nil
}

// MoveCue changes the number of an existing cue.
func (c *CueSystem) MoveCue(sub string, in apicues.MoveCueInput) (*apibus.MessageResponse[apicues.MoveCueOutput], error) {
	if in.OriginalNumber == in.NewNumber {
		return &apibus.MessageResponse[apicues.MoveCueOutput]{
			ResponseValue: &apicues.MoveCueOutput{
				CueListNumber: in.CueListNumber,
				NewNumber:     in.NewNumber,
			},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
		if err != nil {
			return err
		}

		err = data.RenameSubBucket(b, in.OriginalNumber, in.NewNumber)
		return err
	})

	if err != nil {
		slog.Error("failed to move cue", "err", err.Error(), "cuelist", in.CueListNumber, "cue", in.OriginalNumber, "newNumber", in.NewNumber)
		return apibus.NewMessageResponse[apicues.MoveCueOutput](nil, apibus.NewMessageError("failed to move cue")), nil
	}

	outMetadata := &cueMetadata{}
	err = c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.NewNumber, false))
		if err != nil {
			return err
		}

		err = data.GetKey(b, outMetadata, KeyMetadata)
		return err
	})
	if err != nil {
		slog.Error("failed to get data about cue after move", "err", err.Error(), "cuelist", in.OriginalNumber, "newNumber", in.NewNumber)
		return apibus.NewMessageResponse[apicues.MoveCueOutput](nil, apibus.NewMessageError("failed to get data about cue after move")), nil
	}

	// Emit events: delete old, create new
	if err = apibus.Publish(c.conn, apicues.EventDeleteCue, apicues.DeleteCueEvent{
		CueListNumber: in.CueListNumber,
		Number:        in.OriginalNumber,
	}); err != nil {
		slog.Error("failed to publish delete event for move cue", slog.String("err", err.Error()))
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCue, apicues.NewCueEvent{
		CueListNumber: in.CueListNumber,
		Number:        in.NewNumber,
		Label:         outMetadata.Label,
	}); err != nil {
		slog.Error("failed to publish create event for move cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveCueOutput]{
		ResponseValue: &apicues.MoveCueOutput{
			CueListNumber: in.CueListNumber,
			NewNumber:     in.NewNumber,
		},
	}, nil
}
