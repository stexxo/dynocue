package model

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/components/cues/types"
	"github.com/stexxo/dynocue/db"
)

var ErrCueListExists = errors.New("cue list with given number already exists")
var ErrCueListNotFound = errors.New("cue list not found")

func (m *CueingModel) CreateCueList(number uint, cueListType string) (string, uint, error) {
	cl := types.CueList{
		CueListId:   uuid.NewString(),
		Number:      number,
		CueListType: cueListType,
	}

	err := db.WithWrite(m.persistent, func(txn *memdb.Txn) error {
		if cl.Number == 0 {
			last, err := db.GetLastTxn[types.CueList](txn, TableCueLists, IndexNumber)
			if errors.Is(err, db.ErrItemNotFound) {
				cl.Number = 1
			} else if err != nil {
				return err
			} else {
				cl.Number = last.Number + 1
			}
		} else {
			existing, err := txn.First(TableCueLists, IndexNumber, cl.Number)
			if err != nil {
				return err
			}
			if existing != nil {
				return ErrCueListExists
			}
		}

		if err := txn.Insert(TableCueLists, &cl); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return cl.CueListId, cl.Number, nil
}

func (m *CueingModel) EnumerateCueLists() ([]types.CueList, error) {
	out, err := db.GetAllDb[types.CueList](m.persistent, TableCueLists, IndexNumber)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (m *CueingModel) GetCueListByNumber(number uint) (*types.CueList, error) {
	out, err := db.GetFirstDb[types.CueList](m.persistent, TableCueLists, IndexNumber, number)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, ErrCueListNotFound
	}
	return out, nil
}

func (m *CueingModel) GetCueListById(id string) (*types.CueList, error) {
	out, err := db.GetFirstDb[types.CueList](m.persistent, TableCueLists, IndexId, id)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, ErrCueListNotFound
	}
	return out, nil
}
