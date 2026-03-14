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

// GetSubBucket traverses nested buckets and returns the leaf bucket or an error if any part of the path is missing.
func GetSubBucket(b *bbolt.Bucket, keys ...[]byte) (*bbolt.Bucket, error) {
	curr := b
	for _, k := range keys {
		curr = curr.Bucket(k)
		if curr == nil {
			return nil, berrors.ErrBucketNotFound
		}
	}
	return curr, nil
}

// GetMetadata unmarshals msgpack metadata from a bucket.
func GetMetadata[T any](b *bbolt.Bucket, v *T) error {
	val := b.Get([]byte(KeyMetadata))
	if val == nil {
		return berrors.ErrBucketNotFound
	}
	return msgpack.Unmarshal(val, v)
}

// PutMetadata marshals and saves msgpack metadata to a bucket.
func PutMetadata[T any](b *bbolt.Bucket, v T) error {
	md, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(KeyMetadata), md)
}

// UpdateMetadataField updates a single field in the metadata struct and saves it.
func UpdateMetadataField[T any](b *bbolt.Bucket, fieldKey, newValue string) (T, error) {
	var md T
	if err := GetMetadata(b, &md); err != nil {
		return md, err
	}

	if err := utils.SetFieldByTag(&md, "msgpack", fieldKey, newValue); err != nil {
		return md, err
	}

	if err := PutMetadata(b, md); err != nil {
		return md, err
	}

	return md, nil
}

// EnumerateMetadata iterates over all sub-buckets and unmarshals their metadata.
func EnumerateMetadata[T any](b *bbolt.Bucket) ([]T, error) {
	var list []T
	err := b.ForEachBucket(func(k []byte) error {
		sb := b.Bucket(k)
		var md T
		if err := GetMetadata(sb, &md); err != nil {
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

// MoveBucket moves a bucket to a new numeric key and updates its number in metadata.
func MoveBucket[T any](parent *bbolt.Bucket, oldNum, newNum float64, updateNum func(*T, float64)) (T, error) {
	var outMetadata T
	oldKey := utils.Float64ToBytes(oldNum)
	newKey := utils.Float64ToBytes(newNum)

	sb := parent.Bucket(oldKey)
	if sb == nil {
		return outMetadata, berrors.ErrBucketNotFound
	}

	newSb, err := parent.CreateBucket(newKey)
	if err != nil {
		return outMetadata, err
	}

	if err := CopyBucket(sb, newSb); err != nil {
		return outMetadata, err
	}

	if err := GetMetadata(newSb, &outMetadata); err == nil {
		updateNum(&outMetadata, newNum)
		if err := PutMetadata(newSb, outMetadata); err != nil {
			return outMetadata, err
		}
	} else if !errors.Is(err, berrors.ErrBucketNotFound) {
		return outMetadata, err
	}

	if err := parent.DeleteBucket(oldKey); err != nil {
		return outMetadata, fmt.Errorf("failed to delete old bucket: %w", err)
	}

	return outMetadata, nil
}
