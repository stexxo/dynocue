package data

import "go.etcd.io/bbolt"

// CopyBucket recursively copies all keys and sub-buckets from src to dst.
func CopyBucket(src, dst *bbolt.Bucket) error {
	// Copy key-value pairs
	if err := src.ForEach(func(k, v []byte) error {
		if v != nil {
			return dst.Put(k, v)
		}
		return nil
	}); err != nil {
		return err
	}

	// Copy sub-buckets recursively
	return src.ForEachBucket(func(k []byte) error {
		subSrc := src.Bucket(k)
		subDst, err := dst.CreateBucket(k)
		if err != nil {
			return err
		}
		return CopyBucket(subSrc, subDst)
	})
}
