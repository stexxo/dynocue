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

const (
	bucketCueLists = "cuelists"
	keyMetadata    = "metadata"
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
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return fmt.Errorf("failed to create %s bucket: %w", bucketCueLists, err)
		}
		b, outNum, err = data.AddResource(b, in.Number, "metadata", md)
		if err != nil {
			return err
		}

		_, err = b.CreateBucket([]byte(bucketCues))
		return err
	})

	if err != nil {
		errCode := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to create cuelist due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrBucketExists) {
			errCode = apibus.CodeResourceConflict
			errMessage = fmt.Sprintf("cuelist %f already exists", outNum)
		}
		return apibus.NewMessageResponse[apicues.CreateCueListOutput](nil, apibus.NewMessageError(errCode, errMessage)), nil
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
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		outMetadata, err = data.UpdateEntry[cueListMetadata](b, "metadata", in.Key, in.Value)
		return err
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to update cuelist metadata due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("cuelist %f not found", in.Number)
		}
		return apibus.NewMessageResponse[apicues.UpdateCueListMetadataOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}
		return data.GetKey(b, &md, "metadata")
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to update cuelist metadata due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("cuelist %f not found", in.Number)
		}
		return apibus.NewMessageResponse[apicues.GetCueListMetadataOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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

	var values map[float64]*cueListMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucket(tx, []byte(bucketCueLists))
		if err != nil {
			return err
		}

		values, err = data.EnumerateBucketsForKey[float64, cueListMetadata](b, keyMetadata, func(bytes []byte) float64 {
			k, err := utils.BytesToFloat64(bytes)
			if err != nil {
				return 0
			}
			return k
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return apibus.NewMessageResponse[apicues.EnumerateCueListOutput](nil, apibus.NewMessageError(apibus.CodeInternalError, fmt.Sprintf("failed to enumerate cuelists, %s", err.Error()))), nil
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
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		key := utils.Float64ToBytes(in.Number)
		return b.DeleteBucket(key)
	})

	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		return apibus.NewMessageResponse[apicues.DeleteCueListOutput](nil, apibus.NewMessageError(apibus.CodeInternalError, fmt.Sprintf("failed to delete cuelists, %s", err.Error()))), nil
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

	var outMetadata *cueListMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}
		err = data.MoveBucket(b, in.OriginalNumber, in.NewNumber)
		if err != nil {
			return err
		}

		b, err = data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.NewNumber))
		if err != nil {
			return err
		}

		outMetadata = new(cueListMetadata)
		err = data.GetKey(b, outMetadata, "metadata")
		return err
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to update cuelist metadata due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("could not find bucket for original cuelist number %d", in.OriginalNumber)
		}

		if errors.Is(err, data.ErrBucketExists) {
			code = apibus.CodeResourceConflict
			errMessage = fmt.Sprintf("destination cuelist number %d already exists", in.OriginalNumber)
		}

		return apibus.NewMessageResponse[apicues.MoveCueListOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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
