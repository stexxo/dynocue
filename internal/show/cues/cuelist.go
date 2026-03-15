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

type cueListMetadata struct {
	Label    string `msgpack:"label"`
	ListType string `msgpack:"listType"`
}

// NewCueList creates a new cue list with an initial number.
func (c *CueSystem) NewCueList(sub string, in apicues.CreateCueListInput) (*apibus.MessageResponse[apicues.CreateCueListOutput], error) {
	md := &cueListMetadata{}

	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		_, key, err := data.AddIncrementedSubBucket(
			tx,
			BucketCueListKey,
			nil,
			in.Number,
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
		Number:   outNum,
		Label:    md.Label,
		ListType: md.ListType,
	}); err != nil {
		slog.Error("failed to publish change event for new cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.CreateCueListOutput]{
		ResponseValue: &apicues.CreateCueListOutput{Number: outNum},
	}, nil
}

// UpdateCueListMetadata updates the metadata fields of an existing cue list.
func (c *CueSystem) UpdateCueListMetadata(sub string, in apicues.UpdateCueListMetadataInput) (*apibus.MessageResponse[apicues.UpdateCueListMetadataOutput], error) {
	var outMetadata *cueListMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey, data.NewFloatBucketKey(in.Number, false))
		if err != nil {
			return err
		}
		outMetadata, err = data.UpdateAttributeInKeyValuePair[cueListMetadata](b, KeyMetadata, in.Key, in.Value)
		return err
	})
	if err != nil {
		slog.Error("failed to update cuelist metdata", "err", err.Error(), "cuelist", in.Number)
		return apibus.NewMessageResponse[apicues.UpdateCueListMetadataOutput](nil, apibus.NewMessageError("failed to update cuelist")), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCueList, apicues.UpdateCueListMetadataEvent{
		Number:   in.Number,
		Label:    outMetadata.Label,
		ListType: outMetadata.ListType,
	}); err != nil {
		slog.Error("failed to publish change event for update cuelist metadata", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
		ResponseValue: &apicues.UpdateCueListMetadataOutput{},
	}, nil
}

// GetCueListMetadata retrieves the metadata for a specific cue list.
func (c *CueSystem) GetCueListMetadata(sub string, in apicues.GetCueListMetadataInput) (*apibus.MessageResponse[apicues.GetCueListMetadataOutput], error) {
	var md *cueListMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.Number, false))
		if err != nil {
			return err
		}
		return data.GetKey(b, &md, KeyMetadata)
	})

	if err != nil {
		slog.Error("failed to retrieve cuelist", "err", err.Error(), "cuelist", in.Number)
		return apibus.NewMessageResponse[apicues.GetCueListMetadataOutput](nil, apibus.NewMessageError("failed to retrieve cuelist")), nil
	}

	return &apibus.MessageResponse[apicues.GetCueListMetadataOutput]{
		ResponseValue: &apicues.GetCueListMetadataOutput{
			Number:   in.Number,
			Label:    md.Label,
			ListType: md.ListType,
		},
	}, nil
}

// EnumerateCueList returns a list of all existing cue lists.
func (c *CueSystem) EnumerateCueList(sub string, in apicues.EnumerateCueListInput) (*apibus.MessageResponse[apicues.EnumerateCueListOutput], error) {

	values := map[float64]*cueListMetadata{}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey)
		if errors.Is(err, data.ErrNoBucket) {
			return nil
		}
		if err != nil {
			return err
		}
		values, err = data.EnumerateKeysFromSubBuckets[float64, cueListMetadata](b, KeyMetadata, func(bytes []byte) float64 {
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

	out := make([]apicues.GetCueListMetadataOutput, 0, len(values))
	for k, v := range values {
		i, _ := slices.BinarySearchFunc(out, k, func(a apicues.GetCueListMetadataOutput, b float64) int {
			if a.Number < b {
				return -1
			} else if a.Number > b {
				return 1
			}
			return 0
		})
		out = slices.Insert(out, i, apicues.GetCueListMetadataOutput{
			Number:   k,
			Label:    v.Label,
			ListType: v.ListType,
		})
	}

	return &apibus.MessageResponse[apicues.EnumerateCueListOutput]{
		ResponseValue: &apicues.EnumerateCueListOutput{CueLists: out},
	}, nil
}

// DeleteCueList removes a cue list and all its associated cues.
func (c *CueSystem) DeleteCueList(sub string, in apicues.DeleteCueListInput) (*apibus.MessageResponse[apicues.DeleteCueListOutput], error) {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		return data.DeleteBucketByPath(tx, utils.Float64ToBytes(in.Number), BucketCueListKey)
	})
	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		slog.Error("failed to delete cuelist", "err", err.Error(), "cuelist", in.Number)
		return apibus.NewMessageResponse[apicues.DeleteCueListOutput](nil, apibus.NewMessageError(fmt.Sprintf("failed to delete cuelists, %s", err.Error()))), nil
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{Number: in.Number}); err != nil {
		slog.Error("failed to publish change event for delete cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
		ResponseValue: &apicues.DeleteCueListOutput{},
	}, nil
}

// MoveCueList changes the number of an existing cue list.
func (c *CueSystem) MoveCueList(sub string, in apicues.MoveCueListInput) (*apibus.MessageResponse[apicues.MoveCueListOutput], error) {
	if in.OriginalNumber == in.NewNumber {
		return &apibus.MessageResponse[apicues.MoveCueListOutput]{
			ResponseValue: &apicues.MoveCueListOutput{NewNumber: in.NewNumber},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, false, BucketCueListKey)
		if err != nil {
			return err
		}
		err = data.RenameSubBucket(b, in.OriginalNumber, in.NewNumber)
		return err
	})
	if err != nil {
		slog.Error("failed to move cue list", "err", err.Error(), "cuelist", in.OriginalNumber, "newNumber", in.NewNumber)
		return apibus.NewMessageResponse[apicues.MoveCueListOutput](nil, apibus.NewMessageError("failed to move cue lists")), nil
	}

	outMetadata := &cueListMetadata{}
	err = c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucketFromRoot(tx, true, BucketCueListKey, data.NewFloatBucketKey(in.NewNumber, false))
		if err != nil {
			return err
		}

		err = data.GetKey(b, outMetadata, KeyMetadata)
		return err
	})
	if err != nil {
		slog.Error("failed to get data about cue list after move", "err", err.Error(), "cuelist", in.OriginalNumber, "newNumber", in.NewNumber)
		return apibus.NewMessageResponse[apicues.MoveCueListOutput](nil, apibus.NewMessageError("failed to get data about cue list after move")), nil
	}

	// Emit events: delete old, create new
	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{
		Number: in.OriginalNumber,
	}); err != nil {
		slog.Error("failed to publish delete event for move cuelist", slog.String("err", err.Error()))
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{
		Number:   in.NewNumber,
		Label:    outMetadata.Label,
		ListType: outMetadata.ListType,
	}); err != nil {
		slog.Error("failed to publish create event for move cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveCueListOutput]{
		ResponseValue: &apicues.MoveCueListOutput{OriginalNumber: in.OriginalNumber, NewNumber: in.NewNumber},
	}, nil
}
