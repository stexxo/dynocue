package data

import (
	"errors"

	"github.com/tidwall/buntdb"
)

func GetConfiguredVersion(db *buntdb.DB, schemaName string) (string, bool, error) {
	val, err := GetValue(db, "setup:"+schemaName)
	if errors.Is(err, buntdb.ErrNotFound) {
		return "", false, nil
	}

	if err != nil {
		return "", false, err
	}

	return val, true, nil
}

func GetValue(db *buntdb.DB, key string) (string, error) {
	var value string
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		value = val
		return nil
	})
	return value, err
}
