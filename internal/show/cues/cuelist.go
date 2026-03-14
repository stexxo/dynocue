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
	bucketCueLists = "cuelists"
	keyMetadata    = "metadata"
)

type cueListMetadata struct {
	Number   float64 `msgpack:"number"`
	Label    string  `msgpack:"label"`
	ListType string  `msgpack:"listType"`
}

func (c *CueSystem) NewCueList(sub string, in apicues.CreateCueListInput) (*apibus.MessageResponse[apicues.CreateCueListOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.CreateCueListOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var outNum float64
	var outMetadata cueListMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return fmt.Errorf("failed to create %s bucket: %w", bucketCueLists, err)
		}

		outNum = in.Number
		if outNum == 0 {
			outNum = data.NextBucketWholeNumber(b)
		}

		sb, err := b.CreateBucket(utils.Float64ToBytes(outNum))
		if err != nil {
			return err
		}

		outMetadata = cueListMetadata{Number: outNum}
		return data.PutMetadata(sb, outMetadata)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketExists) {
			return &apibus.MessageResponse[apicues.CreateCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceConflict,
					ErrorMessage: fmt.Sprintf("cuelist %f already exists", outNum),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{
		Number:   outMetadata.Number,
		Label:    outMetadata.Label,
		ListType: outMetadata.ListType,
	}); err != nil {
		slog.Error("failed to publish change event for new cuelist", slog.String("err", err.Error()))
	}
	slog.Debug("new cuelist event published successfully")

	return &apibus.MessageResponse[apicues.CreateCueListOutput]{
		ResponseValue: &apicues.CreateCueListOutput{Number: outNum},
	}, nil
}

func (c *CueSystem) UpdateCueListMetadata(sub string, in apicues.UpdateCueListMetadataInput) (*apibus.MessageResponse[apicues.UpdateCueListMetadataOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var outMetadata cueListMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		clb, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		sb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		var errUpdate error
		outMetadata, errUpdate = data.UpdateMetadataField[cueListMetadata](sb, in.Key, in.Value)
		return errUpdate
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.Number),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventUpdateCueList, apicues.UpdateCueListMetadataEvent{
		Number:   outMetadata.Number,
		Label:    outMetadata.Label,
		ListType: outMetadata.ListType,
	}); err != nil {
		slog.Error("failed to publish change event for update cuelist metadata", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
		ResponseValue: &apicues.UpdateCueListMetadataOutput{},
	}, nil
}

func (c *CueSystem) GetCueListMetadata(sub string, in apicues.GetCueListMetadataInput) (*apibus.MessageResponse[apicues.GetCueListMetadataOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.GetCueListMetadataOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var md cueListMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		clb := tx.Bucket([]byte(bucketCueLists))
		if clb == nil {
			return berrors.ErrBucketNotFound
		}

		sb, err := data.GetSubBucket(clb, utils.Float64ToBytes(in.Number))
		if err != nil {
			return err
		}

		return data.GetMetadata(sb, &md)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.GetCueListMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.Number),
				},
			}, nil
		}
		return nil, err
	}

	return &apibus.MessageResponse[apicues.GetCueListMetadataOutput]{
		ResponseValue: &apicues.GetCueListMetadataOutput{
			Number:   md.Number,
			Label:    md.Label,
			ListType: md.ListType,
		},
	}, nil
}

func (c *CueSystem) EnumerateCueList(sub string, in apicues.EnumerateCueListInput) (*apibus.MessageResponse[apicues.EnumerateCueListOutput], error) {
	var cueLists []struct {
		Number   float64 `json:"number" msgpack:"number"`
		Label    string  `json:"label" msgpack:"label"`
		ListType string  `json:"listType" msgpack:"listType"`
	}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketCueLists))
		if b == nil {
			return nil
		}

		list, err := data.EnumerateMetadata[cueListMetadata](b)
		if err != nil {
			return err
		}
		for _, md := range list {
			cueLists = append(cueLists, struct {
				Number   float64 `json:"number" msgpack:"number"`
				Label    string  `json:"label" msgpack:"label"`
				ListType string  `json:"listType" msgpack:"listType"`
			}{
				Number:   md.Number,
				Label:    md.Label,
				ListType: md.ListType,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &apibus.MessageResponse[apicues.EnumerateCueListOutput]{
		ResponseValue: &apicues.EnumerateCueListOutput{CueLists: cueLists},
	}, nil
}

func (c *CueSystem) DeleteCueList(sub string, in apicues.DeleteCueListInput) (*apibus.MessageResponse[apicues.DeleteCueListOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		key := utils.Float64ToBytes(in.Number)
		return b.DeleteBucket(key)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.Number),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{Number: in.Number}); err != nil {
		slog.Error("failed to publish change event for delete cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
		ResponseValue: &apicues.DeleteCueListOutput{},
	}, nil
}

func (c *CueSystem) MoveCueList(sub string, in apicues.MoveCueListInput) (*apibus.MessageResponse[apicues.MoveCueListOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.MoveCueListOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.CodePayloadValidationFailure,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	if in.OriginalNumber == in.NewNumber {
		return &apibus.MessageResponse[apicues.MoveCueListOutput]{
			ResponseValue: &apicues.MoveCueListOutput{NewNumber: in.NewNumber},
		}, nil
	}

	var outMetadata cueListMetadata
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		var errMove error
		outMetadata, errMove = data.MoveBucket(b, in.OriginalNumber, in.NewNumber, func(md *cueListMetadata, num float64) {
			md.Number = num
		})
		return errMove
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.MoveCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceNotFound,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.OriginalNumber),
				},
			}, nil
		}
		if errors.Is(err, berrors.ErrBucketExists) {
			return &apibus.MessageResponse[apicues.MoveCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.CodeResourceConflict,
					ErrorMessage: fmt.Sprintf("cuelist %f already exists", in.NewNumber),
				},
			}, nil
		}
		return nil, err
	}

	// Emit events: delete old, create new
	if err = apibus.Publish(c.conn, apicues.EventDeleteCueList, apicues.DeleteCueListEvent{
		Number: in.OriginalNumber,
	}); err != nil {
		slog.Error("failed to publish delete event for move cuelist", slog.String("err", err.Error()))
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{
		Number:   outMetadata.Number,
		Label:    outMetadata.Label,
		ListType: outMetadata.ListType,
	}); err != nil {
		slog.Error("failed to publish create event for move cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.MoveCueListOutput]{
		ResponseValue: &apicues.MoveCueListOutput{NewNumber: in.NewNumber},
	}, nil
}
