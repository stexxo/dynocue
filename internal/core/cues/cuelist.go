// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package cues

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/data"
	"gitlab.com/stexxo/dynocue/internal/utils"
	apibus "gitlab.com/stexxo/dynocue/pkg/bus"
	"go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

// NewCueList creates a new cue list with an initial number.
func (c *CueSystem) NewCueList(sub string, in apicues.CreateCueListInput) (*apibus.MessageResponse[apicues.CreateCueListOutput], error) {
	md := &CueListMetadataDbModel{}

	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		_, key, err := data.AddIncrementedSubBucket(
			tx,
			BucketCueListKey,
			nil,
			in.CueListNumber,
			data.NewResourceBootstrap{
				InitialValues: []data.KeyValuePair{{Key: KeyMetadata, Value: md}},
				Buckets:       []data.BucketKey{BucketCuesKey},
			},
		)
		outNum = key
		return err
	})
	if err != nil {
		slog.Error("failed to create new cue list", "err", err.Error(), "number", outNum)
		return apibus.NewMessageResponse[apicues.CreateCueListOutput](nil, apibus.NewMessageError("failed to create cuelist")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{
		CueList: apicues.CueList{
			CueListNumber: outNum,
			Label:         md.Label,
			ListType:      md.ListType,
		},
	}); err != nil {
		slog.Error("failed to publish change event for new cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.CreateCueListOutput]{
		ResponseValue: &apicues.CreateCueListOutput{CueListNumber: outNum},
	}, nil
}

// UpdateCueListMetadata updates the metadata fields of an existing cue list.
func (c *CueSystem) UpdateCueList(sub string, in apicues.UpdateCueListInput) (*apibus.MessageResponse[apicues.UpdateCueListOutput], error) {
	var outMetadata *CueListMetadataDbModel
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false))
		if err != nil {
			return err
		}
		outMetadata, err = data.UpdateAttributeInKeyValuePair[CueListMetadataDbModel](b, KeyMetadata, in.Key, in.Value)
		return err
	})
	if err != nil {
		slog.Error("failed to update cuelist", "err", err.Error(), "cuelist", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.UpdateCueListOutput](nil, apibus.NewMessageError("failed to update cuelist")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCueList, apicues.UpdateCueListEvent{
		CueList: apicues.CueList{
			CueListNumber: in.CueListNumber,
			Label:         outMetadata.Label,
			ListType:      outMetadata.ListType,
		},
	}); err != nil {
		slog.Error("failed to publish change event for update cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateCueListOutput]{
		ResponseValue: &apicues.UpdateCueListOutput{},
	}, nil
}

// GetCueList retrieves a specific cue list.
func (c *CueSystem) GetCueList(sub string, in apicues.GetCueListInput) (*apibus.MessageResponse[apicues.GetCueListOutput], error) {
	var md *CueListMetadataDbModel
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.CueListNumber, false))
		if err != nil {
			return err
		}
		return data.GetKey(b, KeyMetadata, &md)
	})

	if err != nil {
		slog.Error("failed to retrieve cuelist", "err", err.Error(), "cuelist", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.GetCueListOutput](nil, apibus.NewMessageError("failed to retrieve cuelist")), nil
	}

	return &apibus.MessageResponse[apicues.GetCueListOutput]{
		ResponseValue: &apicues.GetCueListOutput{
			CueList: apicues.CueList{
				CueListNumber: in.CueListNumber,
				Label:         md.Label,
				ListType:      md.ListType,
			},
		},
	}, nil
}

// EnumerateCueList returns a list of all existing cue lists.
func (c *CueSystem) EnumerateCueList(sub string, in apicues.EnumerateCueListInput) (*apibus.MessageResponse[apicues.EnumerateCueListOutput], error) {

	values := map[float64]*CueListMetadataDbModel{}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey)
		if errors.Is(err, data.ErrNoBucket) {
			return nil
		}
		if err != nil {
			return err
		}
		values, err = data.EnumerateKeysFromSubBuckets[float64, CueListMetadataDbModel](b, KeyMetadata, func(bytes []byte) float64 {
			k, err := utils.BytesToFloat64(bytes)
			if err != nil {
				return 0
			}
			return k
		})
		return err
	})
	if err != nil {
		slog.Error("failed to enumerate cuelists", "err", err.Error())
		return apibus.NewMessageResponse[apicues.EnumerateCueListOutput](nil, apibus.NewMessageError("failed to enumerate cuelists")), nil
	}

	out := make([]apicues.GetCueListOutput, 0, len(values))
	for k, v := range values {
		i, _ := slices.BinarySearchFunc(out, k, func(a apicues.GetCueListOutput, b float64) int {
			if a.CueList.CueListNumber < b {
				return -1
			} else if a.CueList.CueListNumber > b {
				return 1
			}
			return 0
		})
		out = slices.Insert(out, i, apicues.GetCueListOutput{
			CueList: apicues.CueList{
				CueListNumber: k,
				Label:         v.Label,
				ListType:      v.ListType,
			},
		})
	}

	return &apibus.MessageResponse[apicues.EnumerateCueListOutput]{
		ResponseValue: &apicues.EnumerateCueListOutput{CueLists: out},
	}, nil
}

// DeleteCueList removes a cue list and all its associated cues.
func (c *CueSystem) DeleteCueList(sub string, in apicues.DeleteCueListInput) (*apibus.MessageResponse[apicues.DeleteCueListOutput], error) {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		return data.DeleteBucketByPath(tx, utils.Float64ToBytes(in.CueListNumber), BucketCueListKey)
	})
	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		slog.Error("failed to delete cuelist", "err", err.Error(), "cuelist", in.CueListNumber)
		return apibus.NewMessageResponse[apicues.DeleteCueListOutput](nil, apibus.NewMessageError(fmt.Sprintf("failed to delete cuelists, %s", err.Error()))), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{CueListNumber: in.CueListNumber}); err != nil {
		slog.Error("failed to publish change event for delete cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
		ResponseValue: &apicues.DeleteCueListOutput{},
	}, nil
}

// MoveCueList changes the number of an existing cue list.
func (c *CueSystem) MoveCueList(sub string, in apicues.MoveCueListInput) (*apibus.MessageResponse[apicues.MoveCueListOutput], error) {
	if in.OriginalCueListNumber == in.NewCueListNumber {
		return &apibus.MessageResponse[apicues.MoveCueListOutput]{
			ResponseValue: &apicues.MoveCueListOutput{NewCueListNumber: in.NewCueListNumber},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey)
		if err != nil {
			return err
		}
		err = data.RenameSubBucket(b, in.OriginalCueListNumber, in.NewCueListNumber)
		return err
	})
	if err != nil {
		slog.Error("failed to move cue list", "err", err.Error(), "cuelist", in.OriginalCueListNumber, "newNumber", in.NewCueListNumber)
		return apibus.NewMessageResponse[apicues.MoveCueListOutput](nil, apibus.NewMessageError("failed to move cue lists")), nil
	}

	outMetadata := &CueListMetadataDbModel{}
	err = c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.NewCueListNumber, false))
		if err != nil {
			return err
		}

		err = data.GetKey(b, KeyMetadata, outMetadata)
		return err
	})
	if err != nil {
		slog.Error("failed to get data about cue list after move", "err", err.Error(), "cuelist", in.OriginalCueListNumber, "newNumber", in.NewCueListNumber)
		return apibus.NewMessageResponse[apicues.MoveCueListOutput](nil, apibus.NewMessageError("failed to get data about cue list after move")), nil
	}

	// Emit events: delete old, create new
	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{
		CueListNumber: in.OriginalCueListNumber,
	}); err != nil {
		slog.Error("failed to publish delete event for move cuelist", slog.String("err", err.Error()))
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{
		CueList: apicues.CueList{
			CueListNumber: in.NewCueListNumber,
			Label:         outMetadata.Label,
			ListType:      outMetadata.ListType,
		},
	}); err != nil {
		slog.Error("failed to publish create event for move cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveCueListOutput]{
		ResponseValue: &apicues.MoveCueListOutput{OriginalCueListNumber: in.OriginalCueListNumber, NewCueListNumber: in.NewCueListNumber},
	}, nil
}
