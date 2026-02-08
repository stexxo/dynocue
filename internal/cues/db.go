package cues

import (
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func CreateCueList(db *badger.DB, number float64, id string) error {
	return db.Update(func(txn *badger.Txn) error {
		return errors.Join(txn.Set([]byte(fmt.Sprintf("cuelist:number:%f", number)), []byte(id)), txn.Commit())
	})
}

func GetCueListIds(db *badger.DB) ([]string, error) {
	var cueLists []string
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek([]byte("cuelist:number")); it.ValidForPrefix([]byte("cuelist:")); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				cueLists = append(cueLists, string(val))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cueLists, nil
}
