package data

import (
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"gitlab.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
	berrors "go.etcd.io/bbolt/errors"
)

const KeyMetadata = "metadata"

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

var ErrNoBucket = errors.New("bucket path does not exist")

func GetSubBucket(tx *bbolt.Bucket, keys ...[]byte) (*bbolt.Bucket, error) {
	if len(keys) == 0 {
		return tx, nil
	}

	b := tx.Bucket(keys[0])
	if b == nil {
		return nil, ErrNoBucket
	}

	return GetSubBucket(b, keys[1:]...)
}

func GetBucket(tx *bbolt.Tx, keys ...[]byte) (*bbolt.Bucket, error) {
	if len(keys) == 0 {
		return nil, ErrNoBucket
	}
	b := tx.Bucket(keys[0])
	if b == nil {
		return nil, ErrNoBucket
	}
	return GetSubBucket(b, keys[1:]...)
}

func GetKey[T any](b *bbolt.Bucket, v *T, key string) error {
	val := b.Get([]byte(key))
	if val == nil {
		return berrors.ErrBucketNotFound
	}
	return msgpack.Unmarshal(val, v)
}

func PutKey[T any](b *bbolt.Bucket, v T, key string) error {
	md, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), md)
}

func UpdateEntry[T any](b *bbolt.Bucket, entryKey, fieldKey, newValue string) (T, error) {
	var md T
	if err := GetKey(b, &md, entryKey); err != nil {
		return md, err
	}

	if err := utils.SetFieldByTag(&md, "msgpack", fieldKey, newValue); err != nil {
		return md, err
	}

	if err := PutKey(b, md, entryKey); err != nil {
		return md, err
	}

	return md, nil
}

// EnumerateBucketsForKey iterates over all sub-buckets and returns a slice of their entries for the provided key
func EnumerateBucketsForKey[T any](b *bbolt.Bucket, key string) ([]T, error) {
	var list []T
	err := b.ForEachBucket(func(k []byte) error {
		sb := b.Bucket(k)
		var md T
		if err := GetKey(sb, &md, key); err != nil {
			if errors.Is(err, berrors.ErrBucketNotFound) {
				return nil
			}
			return err
		}
		list = append(list, md)
		return nil
	})
	return list, err
}

// MoveBucket copies a bucket to a new numeric key, updates its number in metadata,
// and deletes the old bucket.
func MoveBucket[T any](parent *bbolt.Bucket, oldNum, newNum float64, updateNum func(*T, float64)) (T, error) {
	var outMetadata T
	oldKey := utils.Float64ToBytes(oldNum)
	newKey := utils.Float64ToBytes(newNum)

	sb := parent.Bucket(oldKey)
	if sb == nil {
		return outMetadata, ErrNoBucket
	}

	newSb, err := parent.CreateBucket(newKey)
	if err != nil {
		return outMetadata, err
	}

	if err := CopyBucket(sb, newSb); err != nil {
		return outMetadata, err
	}

	if err := parent.DeleteBucket(oldKey); err != nil {
		return outMetadata, fmt.Errorf("failed to delete old bucket: %w", err)
	}

	return outMetadata, nil
}
