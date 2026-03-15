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
	bucketCues    = "cues"
	bucketActions = "actions"
)

type cueMetadata struct {
	Label string `msgpack:"label"`
}

// NewCue creates a new cue within a specific cue list.
func (c *CueSystem) NewCue(sub string, in apicues.CreateCueInput) (*apibus.MessageResponse[apicues.CreateCueOutput], error) {
	md := &cueMetadata{}
	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		b, outNum, err = data.AddResource(b, in.Number, "metadata", md)
		if err != nil {
			return err
		}

		_, err = b.CreateBucket([]byte(bucketActions))
		if err != nil {
			return err
		}
		return err
	})

	if err != nil {
		errCode := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to create cue due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrBucketExists) {
			errCode = apibus.CodeResourceConflict
			errMessage = fmt.Sprintf("cue %f already exists", outNum)
		}
		return apibus.NewMessageResponse[apicues.CreateCueOutput](nil, apibus.NewMessageError(errCode, errMessage)), nil
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
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		outMetadata, err = data.UpdateEntry[cueMetadata](b, "metadata", in.Key, in.Value)
		return err
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to update cuelist metadata due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("cuelist %f not found", in.Number)
		}
		return apibus.NewMessageResponse[apicues.UpdateCueMetadataOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		return data.GetKey(b, md, "metadata")
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to update cuelist metadata due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("cuelist %f not found", in.Number)
		}
		return apibus.NewMessageResponse[apicues.GetCueMetadataOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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
	var values map[float64]*cueMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		values, err = data.EnumerateBucketsForKey[float64, cueMetadata](b, keyMetadata, func(bytes []byte) float64 {
			k, err := utils.BytesToFloat64(bytes)
			if err != nil {
				return 0
			}
			return k
		})
		return err
	})
	if err != nil {
		return apibus.NewMessageResponse[apicues.EnumerateCueOutput](nil, apibus.NewMessageError(apibus.CodeInternalError, fmt.Sprintf("failed to enumerate cues, %s", err.Error()))), nil
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
		cb, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		key := utils.Float64ToBytes(in.Number)
		return cb.DeleteBucket(key)
	})

	if err != nil && !errors.Is(err, berrors.ErrBucketNotFound) {
		return apibus.NewMessageResponse[apicues.DeleteCueOutput](nil, apibus.NewMessageError(apibus.CodeInternalError, fmt.Sprintf("failed to delete cue, %s", err.Error()))), nil
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

	var outMetadata *cueMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		err = data.MoveBucket(b, in.OriginalNumber, in.NewNumber)
		if err != nil {
			return err
		}

		b, err = data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues), utils.Float64ToBytes(in.NewNumber))
		if err != nil {
			return err
		}

		outMetadata = new(cueMetadata)
		err = data.GetKey(b, outMetadata, "metadata")
		return err
	})

	if err != nil {
		code := apibus.CodeInternalError
		errMessage := fmt.Sprintf("failed to move cue due to internal problem %s", err.Error())
		if errors.Is(err, data.ErrNoBucket) {
			code = apibus.CodeResourceNotFound
			errMessage = fmt.Sprintf("could not find bucket for original cue number %d", in.OriginalNumber)
		}

		if errors.Is(err, data.ErrBucketExists) {
			code = apibus.CodeResourceConflict
			errMessage = fmt.Sprintf("destination cue number %d already exists", in.OriginalNumber)
		}

		return apibus.NewMessageResponse[apicues.MoveCueOutput](nil, apibus.NewMessageError(code, errMessage)), nil
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
