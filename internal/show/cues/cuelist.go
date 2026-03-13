package cues

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
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

func (c *CueSystem) NewCueList(sub string, in apicues.NewCueListInput) (*apibus.MessageResponse[apicues.NewCueListOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.NewCueListOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.ValidationErrorCode,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var outNum float64
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
			if errors.Is(err, berrors.ErrBucketExists) {
				return err
			}
			return fmt.Errorf("failed to create sub-bucket %f: %w", outNum, err)
		}

		md, err := msgpack.Marshal(cueListMetadata{Number: outNum})
		if err != nil {
			return err
		}

		return sb.Put([]byte(keyMetadata), md)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketExists) {
			return &apibus.MessageResponse[apicues.NewCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.ConflictCode,
					ErrorMessage: fmt.Sprintf("cuelist %f already exists", outNum),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{Number: outNum}); err != nil {
		slog.Error("failed to publish change event for new cuelist", slog.String("err", err.Error()))
	}

	return &apibus.MessageResponse[apicues.NewCueListOutput]{
		ResponseValue: &apicues.NewCueListOutput{Number: outNum},
	}, nil
}

func (c *CueSystem) UpdateCueListMetadata(sub string, in apicues.UpdateCueListMetadataInput) (*apibus.MessageResponse[apicues.UpdateCueListMetadataOutput], error) {
	if err := apibus.Validate(in); err != nil {
		return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
			MessageError: &apibus.MessageError{
				Code:         apibus.ValidationErrorCode,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	parts := strings.Split(sub, ".")
	field := parts[len(parts)-1]
	eventSub := fmt.Sprintf("%s.%s", apicues.EventUpdateCueList, field)

	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketCueLists))
		if err != nil {
			return err
		}

		sb := b.Bucket(utils.Float64ToBytes(in.Number))
		if sb == nil {
			return berrors.ErrBucketNotFound
		}

		val := sb.Get([]byte(keyMetadata))
		if val == nil {
			return berrors.ErrBucketNotFound
		}

		var md cueListMetadata
		if err := msgpack.Unmarshal(val, &md); err != nil {
			return err
		}

		parts := strings.Split(sub, ".")
		f := parts[len(parts)-1]

		if f == "*" {
			return errors.New("cannot update wildcard subject")
		}

		if err := utils.SetFieldByTag(&md, "msgpack", f, in.Value); err != nil {
			return err
		}

		mdBytes, err := msgpack.Marshal(md)
		if err != nil {
			return err
		}

		return sb.Put([]byte(keyMetadata), mdBytes)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.UpdateCueListMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.NotFoundCode,
					ErrorMessage: fmt.Sprintf("cuelist %f not found", in.Number),
				},
			}, nil
		}
		return nil, err
	}

	if err = apibus.Publish(c.conn, eventSub, apicues.UpdateCueListMetadataEvent{
		Number: in.Number,
		Value:  in.Value,
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
				Code:         apibus.ValidationErrorCode,
				ErrorMessage: fmt.Sprintf("validation failed: %s", err.Error()),
			},
		}, nil
	}

	var md cueListMetadata
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketCueLists))
		if b == nil {
			return berrors.ErrBucketNotFound
		}

		sb := b.Bucket(utils.Float64ToBytes(in.Number))
		if sb == nil {
			return berrors.ErrBucketNotFound
		}

		val := sb.Get([]byte(keyMetadata))
		if val == nil {
			return berrors.ErrBucketNotFound
		}

		return msgpack.Unmarshal(val, &md)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.GetCueListMetadataOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.NotFoundCode,
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
		Number   float64 `msgpack:"number"`
		Label    string  `msgpack:"label"`
		ListType string  `msgpack:"listType"`
	}
	err := c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketCueLists))
		if b == nil {
			return nil
		}

		return b.ForEachBucket(func(k []byte) error {
			sb := b.Bucket(k)
			v := sb.Get([]byte(keyMetadata))
			if v == nil {
				return nil
			}
			var md cueListMetadata
			if err := msgpack.Unmarshal(v, &md); err != nil {
				return err
			}
			cueLists = append(cueLists, struct {
				Number   float64 `msgpack:"number"`
				Label    string  `msgpack:"label"`
				ListType string  `msgpack:"listType"`
			}{
				Number:   md.Number,
				Label:    md.Label,
				ListType: md.ListType,
			})
			return nil
		})
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
				Code:         apibus.ValidationErrorCode,
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
		if b.Bucket(key) == nil {
			return berrors.ErrBucketNotFound
		}

		return b.DeleteBucket(key)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketNotFound) {
			return &apibus.MessageResponse[apicues.DeleteCueListOutput]{
				MessageError: &apibus.MessageError{
					Code:         apibus.NotFoundCode,
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
