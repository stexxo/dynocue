package cues

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/vmihailenco/msgpack/v5"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/data"
	"gitlab.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

func (c *CueSystem) NewCueList(sub string, in apicues.NewCueListInput) (*bus.MessageResponse[apicues.NewCueListOutput], error) {
	var outNum float64
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("cuelists"))
		if err != nil {
			return fmt.Errorf("failed to create cuelists bucket: %w", err)
		}

		outNum = in.Number
		if outNum == 0 {
			outNum = data.NextBucketWholeNumber(b)
		}

		if _, err = b.CreateBucket(utils.Float64ToBytes(outNum)); err != nil {
			if errors.Is(err, berrors.ErrBucketExists) {
				return err
			}
			return fmt.Errorf("failed to create sub-bucket %f: %w", outNum, err)
		}

		md, err := msgpack.Marshal(apicues.NewCueListOutput{Number: outNum})
		if err != nil {
			return err
		}

		return b.Put([]byte("metadata"), md)
	})

	if err != nil {
		if errors.Is(err, berrors.ErrBucketExists) {
			return &bus.MessageResponse[apicues.NewCueListOutput]{
				MessageError: &bus.MessageError{
					Code:         bus.ConflictCode,
					ErrorMessage: fmt.Sprintf("cuelist %f already exists", outNum),
				},
			}, nil
		}
		return nil, err
	}

	if err = bus.Publish(c.conn, apicues.EventNewCueList, apicues.NewCueListEvent{Number: outNum}); err != nil {
		slog.Error("failed to publish change event for new cuelist", slog.String("err", err.Error()))
	}

	return &bus.MessageResponse[apicues.NewCueListOutput]{
		ResponseValue: &apicues.NewCueListOutput{Number: outNum},
	}, nil
}
