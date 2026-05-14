// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"
	"github.com/hashicorp/go-memdb"
	"github.com/stexxo/dynocue/db"
)

var ErrNumberExists = errors.New("number already exists")

func getNextNumber[T any](txn *memdb.Txn, num uint, table string, getNextIndex string, getNextIndexKeys []any, numberIndex string, numberIndexKeys []any, getNumFn func(*T) uint) (uint, error) {
	if num == 0 {
		last, err := db.GetLastTxn[T](txn, table, getNextIndex, getNextIndexKeys...)
		if errors.Is(err, db.ErrItemNotFound) {
			return 1, nil
		}
		if err != nil {
			return 0, err
		}
		return getNumFn(last) + 1, nil
	} else {
		existing, err := txn.First(table, numberIndex, numberIndexKeys...)
		if err != nil {
			return 0, err
		}
		if existing != nil {
			return 0, ErrNumberExists
		}
		return num, nil
	}
}
