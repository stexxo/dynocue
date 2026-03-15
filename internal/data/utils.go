package data

import (
	"math"

	"gitlab.com/stexxo/dynocue/internal/utils"
)

func NextBucketWholeNumber(b Bucket) float64 {
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
