// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

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

// NewCue creates a new cue within a specific cue list.
func (c *CueSystem) NewCue(sub string, in apicues.CreateCueInput) (*apibus.MessageResponse[apicues.CreateCueOutput], error) {
	md := &CueMetadataDbModel{}
	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		_, key, err := data.AddIncrementedSubBucket(
			tx,
			BucketCueListKey,
			[]data.BucketKey{data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey},
			in.CueNumber,
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
		Cue: apicues.Cue{
			CueNumber:   outNum,
			Label:       md.Label,
			Description: md.Description,
		},
	}); err != nil {
		slog.Error("failed to publish change event for new cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.CreateCueOutput]{
		ResponseValue: &apicues.CreateCueOutput{
			CueListNumber: in.CueListNumber,
			CueNumber:     outNum,
		},
	}, nil
}

// UpdateCue updates the metadata fields of an existing cue.
func (c *CueSystem) UpdateCue(sub string, in apicues.UpdateCueInput) (*apibus.MessageResponse[apicues.UpdateCueOutput], error) {
	var outMetadata *CueMetadataDbModel
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false))
		if err != nil {
			return err
		}
		outMetadata, err = data.UpdateAttributeInKeyValuePair[CueMetadataDbModel](b, KeyMetadata, in.Key, in.Value)
		return err
	})
	if err != nil {
		slog.Error("failed to update cue", "err", err.Error(), "cue", in.CueNumber, "cueList", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.UpdateCueOutput](nil, apibus.NewMessageError("failed to update cue")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCue, apicues.UpdateCueEvent{
		CueListNumber: in.CueListNumber,
		Cue: apicues.Cue{
			CueNumber:   in.CueNumber,
			Label:       outMetadata.Label,
			Description: outMetadata.Description,
		},
	}); err != nil {
		slog.Error("failed to publish change event for update cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateCueOutput]{
		ResponseValue: &apicues.UpdateCueOutput{},
	}, nil
}

// GetCue retrieves a specific cue.
func (c *CueSystem) GetCue(sub string, in apicues.GetCueInput) (*apibus.MessageResponse[apicues.GetCueOutput], error) {
	var md *CueMetadataDbModel
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.CueNumber, false))
		if err != nil {
			return err
		}
		md = new(CueMetadataDbModel)
		return data.GetKey(b, KeyMetadata, md)
	})

	if err != nil {
		slog.Error("failed to get cue", "err", err.Error(), "cue", in.CueNumber, "cueList", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.GetCueOutput](nil, apibus.NewMessageError("failed to retrieve cue")), nil
	}

	return &apibus.MessageResponse[apicues.GetCueOutput]{
		ResponseValue: &apicues.GetCueOutput{
			CueListNumber: in.CueListNumber,
			Cue: apicues.Cue{
				CueNumber:   in.CueNumber,
				Label:       md.Label,
				Description: md.Description,
			},
		},
	}, nil
}

// EnumerateCue returns a list of all cues within a specific cue list.
func (c *CueSystem) EnumerateCue(sub string, in apicues.EnumerateCueInput) (*apibus.MessageResponse[apicues.EnumerateCueOutput], error) {
	values := map[float64]*CueMetadataDbModel{}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
		if errors.Is(err, data.ErrNoBucket) {
			return nil
		}
		if err != nil {
			return err
		}
		values, err = data.EnumerateKeysFromSubBuckets[float64, CueMetadataDbModel](b, KeyMetadata, func(bytes []byte) float64 {
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

	out := make([]apicues.GetCueOutput, 0, len(values))
	for k, v := range values {
		i, _ := slices.BinarySearchFunc(out, k, func(a apicues.GetCueOutput, b float64) int {
			if a.Cue.CueNumber < b {
				return -1
			} else if a.Cue.CueNumber > b {
				return 1
			}
			return 0
		})
		out = slices.Insert(out, i, apicues.GetCueOutput{
			CueListNumber: in.CueListNumber,
			Cue: apicues.Cue{
				CueNumber:   k,
				Label:       v.Label,
				Description: v.Description,
			},
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
		return data.DeleteBucketByPath(tx, utils.Float64ToBytes(in.CueNumber), BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
	})

	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		slog.Error("failed to delete cue", "err", err.Error(), "cueList", in.CueListNumber, "cue", in.CueNumber)
		return apibus.NewMessageResponse[apicues.DeleteCueOutput](nil, apibus.NewMessageError("failed to delete cue")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteCue, apicues.DeleteCueEvent{
		CueListNumber: in.CueListNumber,
		CueNumber:     in.CueNumber,
	}); err != nil {
		slog.Error("failed to publish change event for delete cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteCueOutput]{
		ResponseValue: &apicues.DeleteCueOutput{},
	}, nil
}

// MoveCue changes the number of an existing cue.
func (c *CueSystem) MoveCue(sub string, in apicues.MoveCueInput) (*apibus.MessageResponse[apicues.MoveCueOutput], error) {
	if in.OriginalCueNumber == in.NewCueNumber {
		return &apibus.MessageResponse[apicues.MoveCueOutput]{
			ResponseValue: &apicues.MoveCueOutput{
				CueListNumber: in.CueListNumber,
				NewCueNumber:  in.NewCueNumber,
			},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey)
		if err != nil {
			return err
		}

		err = data.RenameSubBucket(b, in.OriginalCueNumber, in.NewCueNumber)
		return err
	})

	if err != nil {
		slog.Error("failed to move cue", "err", err.Error(), "cueList", in.CueListNumber, "cue", in.OriginalCueNumber, "newCueNumber", in.NewCueNumber)
		return apibus.NewMessageResponse[apicues.MoveCueOutput](nil, apibus.NewMessageError("failed to move cue")), nil
	}

	outMetadata := &CueMetadataDbModel{}
	err = c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false), BucketCuesKey, data.NewFloatBucketKey(in.NewCueNumber, false))
		if err != nil {
			return err
		}

		err = data.GetKey(b, KeyMetadata, outMetadata)
		return err
	})
	if err != nil {
		slog.Error("failed to get data about cue after move", "err", err.Error(), "cueList", in.CueListNumber, "originalCueNumber", in.OriginalCueNumber, "newCueNumber", in.NewCueNumber)
		return apibus.NewMessageResponse[apicues.MoveCueOutput](nil, apibus.NewMessageError("failed to get data about cue after move")), nil
	}

	// Emit events: delete old, create new
	if err = apibus.Publish(c.conn, apicues.EventDeleteCue, apicues.DeleteCueEvent{
		CueListNumber: in.CueListNumber,
		CueNumber:     in.OriginalCueNumber,
	}); err != nil {
		slog.Error("failed to publish delete event for move cue", slog.String("err", err.Error()))
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCue, apicues.NewCueEvent{
		CueListNumber: in.CueListNumber,
		Cue: apicues.Cue{
			CueNumber: in.NewCueNumber,
			Label:     outMetadata.Label,
		},
	}); err != nil {
		slog.Error("failed to publish create event for move cue", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveCueOutput]{
		ResponseValue: &apicues.MoveCueOutput{
			CueListNumber: in.CueListNumber,
			NewCueNumber:  in.NewCueNumber,
		},
	}, nil
}
