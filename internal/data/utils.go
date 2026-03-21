// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package data

import (
	"math"

	"github.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
)

func NextBucketWholeNumber(b *bbolt.Bucket) float64 {
	var maxNum float64
	_ = b.ForEachBucket(func(k []byte) error {
		if n, err := utils.BytesToFloat64(k); err == nil {
			if n > maxNum {
				maxNum = n
			}
		}
		return nil
	})

	return math.Floor(maxNum) + 1
}

func NextBucketKeyWholeNumber(b *bbolt.Bucket) float64 {
	var maxNum float64
	_ = b.ForEach(func(k, v []byte) error {
		if v == nil {
			return nil
		}
		if n, err := utils.BytesToFloat64(k); err == nil {
			if n > maxNum {
				maxNum = n
			}
		}
		return nil
	})

	return math.Floor(maxNum) + 1
}
