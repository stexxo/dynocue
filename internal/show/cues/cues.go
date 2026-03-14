package cues

import (
	"errors"
	"fmt"
	"log/slog"

	apibus "gitlab.com/stexxo/dynocue/api/bus"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/data"
	"gitlab.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

const (
	bucketCues = "cues"
)

type cueMetadata struct {
	Number float64 `msgpack:"number"`
	Label  string  `msgpack:"label"`
}

// NewCue creates a new cue within a specific cue list.
func (c *CueSystem) NewCue(sub string, in apicues.CreateCueInput) (*apibus.MessageResponse[apicues.CreateCueOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.CreateCueOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var outNum float64
	var outMetadata cueMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		cueBucket, err := data.GetBucket(tx, []byte(bucketCueLists), utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))

		outNum = in.Number
		if outNum == 0 {
			outNum = data.NextBucketWholeNumber(cueBucket)
		}

		nb, err := cueBucket.CreateBucket(utils.Float64ToBytes(outNum))
		if err != nil {
			return err
		}

		outMetadata = cueMetadata{Number: outNum}
		return data.PutKey(nb, outMetadata, "metadata")
	})

	if err != nil {
		if errors.Is(err, data.ErrNoBucket) {
			return &apibus.MessageResponse[apicues.CreateCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.CueListNumber),
				},
			}, nil
		}
		if errors.Is(err, berrors.ErrBucketExists) {
			return &apibus.MessageResponse[apicues.CreateCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceConflict,
					ErrorMessage: fmt.Sprintf("cue %f already exists in cuelist %f", outNum, in.CueListNumber),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCue, apicues.NewCueEvent{
		CueListNumber: in.CueListNumber,
		Number:        outMetadata.Number,
		Label:         outMetadata.Label,
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
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.UpdateCueMetadataOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var outMetadata cueMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		clb, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		nb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		var errUpdate error
		outMetadata, errUpdate = data.UpdateEntry[cueMetadata](nb, "metadata", in.Key, in.Value)
		return errUpdate
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.UpdateCueMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cue %f not found in cuelist %f", in.Number, in.CueListNumber),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCue, apicues.UpdateCueMetadataEvent{
		CueListNumber: in.CueListNumber,
		Number:        outMetadata.Number,
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
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.GetCueMetadataOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var md cueMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		clb := tx.Bucket([]byte(bucketCueLists))
		if clb == nil {
			return berrors.ErrBucketNotFound
		}

		nb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues), utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		return data.GetKey(nb, &md, "metadata")
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.GetCueMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cue %f not found in cuelist %f", in.Number, in.CueListNumber),
				},
			}, nil
		}
		return nil, err
	}

	return &apibus.MessageResponse[apicues.GetCueMetadataOutput]{
		ResponseValue: &apicues.GetCueMetadataOutput{
			CueListNumber: in.CueListNumber,
			Number:        md.Number,
			Label:         md.Label,
		},
	}, nil
}

// EnumerateCue returns a list of all cues within a specific cue list.
func (c *CueSystem) EnumerateCue(sub string, in apicues.EnumerateCueInput) (*apibus.MessageResponse[apicues.EnumerateCueOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.EnumerateCueOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var cues []struct {
		Number float64 `json:"number" msgpack:"number"`
		Label  string  `json:"label" msgpack:"label"`
	}
	err := c.db.View(func(tx *bbolt.Tx) error {
		clb := tx.Bucket([]byte(bucketCueLists))
		if clb == nil {
			return berrors.ErrBucketNotFound
		}

		cb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			if errors.Is(err, berrors.ErrBucketNotFound) {
				// if cuelist exists but cues bucket doesn't, it's just empty
				if clb.Bucket(utils.Float64ToBytes(in.CueListNumber)) != nil {
					return nil
				}
			}
			return err
		}

		list, err := data.EnumerateBucketsForKey[cueMetadata](cb, "metadata")
		if err != nil {
			return err
		}

		for _, md := range list {
			cues = append(cues, struct {
				Number float64 `json:"number" msgpack:"number"`
				Label  string  `json:"label" msgpack:"label"`
			}{
				Number: md.Number,
				Label:  md.Label,
			})
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.EnumerateCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.CueListNumber),
				},
			}, nil
		}
		return nil, err
	}

	return &apibus.MessageResponse[apicues.EnumerateCueOutput]{
		ResponseValue: &apicues.EnumerateCueOutput{
			CueListNumber: in.CueListNumber,
			Cues:          cues,
		},
	}, nil
}

// DeleteCue removes a specific cue from a cue list.
func (c *CueSystem) DeleteCue(sub string, in apicues.DeleteCueInput) (*apibus.MessageResponse[apicues.DeleteCueOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.DeleteCueOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		clb, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		cb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		key := utils.Float64ToBytes(in.Number)
		return cb.DeleteBucket(key)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.DeleteCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cue %f not found in cuelist %f", in.Number, in.CueListNumber),
				},
			}, nil
		}
		return nil, err
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
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.MoveCueOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	if in.OriginalNumber == in.NewNumber {
		return &apibus.MessageResponse[apicues.MoveCueOutput]{
			ResponseValue: &apicues.MoveCueOutput{
				CueListNumber: in.CueListNumber,
				NewNumber:     in.NewNumber,
			},
		}, nil
	}

	var outMetadata cueMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		clb, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		cb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.CueListNumber), []byte(bucketCues))
		if err != nil {
			return err
		}

		var errMove error
		outMetadata, errMove = data.MoveBucket(cb, in.OriginalNumber, in.NewNumber, func(md *cueMetadata, num float64) {
			md.Number = num
		})
		return errMove
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.MoveCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cue %f not found in cuelist %f", in.OriginalNumber, in.CueListNumber),
				},
			}, nil
		}
		if errors.Is(err, berrors.ErrBucketExists) {
			return &apibus.MessageResponse[apicues.MoveCueOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceConflict,
					ErrorMessage: fmt.Sprintf("cue %f already exists in cuelist %f", in.NewNumber, in.CueListNumber),
				},
			}, nil
		}
		return nil, err
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
		Number:        outMetadata.Number,
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
