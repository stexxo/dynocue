package model

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

func (m *CueingModel) CreateCue(cueListId string, number uint) (string, uint, error) {
	cue := types.Cue{
		CueListId: cueListId,
		CueId:     uuid.NewString(),
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		// Check that CueList Exists
		item, err := txn.First(TableCueLists, IndexId, cueListId)
		if err != nil {
			return err
		}
		if item == nil {
			return ErrCueListNotFound
		}

		num, err := getNextNumber[types.Cue](txn, number, TableCues, IndexNumberPrefix, []any{cueListId}, IndexNumber, []any{cueListId, number}, func(t *types.Cue) uint {
			return t.Number
		})
		if err != nil {
			return err
		}
		cue.Number = num
		if err := txn.Insert(TableCues, &cue); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", 0, err
	}

	return cue.CueId, cue.Number, nil
}
