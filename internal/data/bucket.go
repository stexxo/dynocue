package data

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

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

func PutKey[T any](b *bbolt.Bucket, v *T, key string) error {
	md, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), md)
}

func UpdateEntry[T any](b *bbolt.Bucket, entryKey, fieldKey, newValue string) (*T, error) {
	out := new(T)
	if err := GetKey(b, out, entryKey); err != nil {
		return out, err
	}

	if err := utils.SetFieldByTag(out, "msgpack", fieldKey, newValue); err != nil {
		return out, err
	}

	if err := PutKey(b, out, entryKey); err != nil {
		return out, err
	}

	return out, nil
}

// EnumerateBucketsForKey iterates over all sub-buckets and returns a slice of their entries for the provided key
func EnumerateBucketsForKey[T comparable, E any](b *bbolt.Bucket, key string, keyFn func([]byte) T) (map[T]*E, error) {
	vals := make(map[T]*E)
	err := b.ForEachBucket(func(k []byte) error {
		sb := b.Bucket(k)
		var md E
		if err := GetKey(sb, &md, key); err != nil {
			if errors.Is(err, berrors.ErrBucketNotFound) {
				return nil
			}
			return err
		}
		vals[keyFn(k)] = &md
		return nil
	})
	return vals, err
}

// MoveBucket copies a bucket to a new numeric key, updates its number in metadata,
// and deletes the old bucket.
func MoveBucket(parent *bbolt.Bucket, oldNum, newNum float64) error {
	oldKey := utils.Float64ToBytes(oldNum)
	newKey := utils.Float64ToBytes(newNum)

	ob := parent.Bucket(oldKey)
	if ob == nil {
		return ErrNoBucket
	}

	nb := parent.Bucket(newKey)
	if nb != nil {
		return ErrBucketExists
	}

	newSb, err := parent.CreateBucket(newKey)
	if err != nil {
		return err
	}

	if err := CopyBucket(ob, newSb); err != nil {
		return err
	}

	if err := parent.DeleteBucket(oldKey); err != nil {
		return fmt.Errorf("failed to delete old bucket: %w", err)
	}

	return nil
}

var ErrBucketExists = errors.New("bucket already exists")

func AddResource[T any](bucket *bbolt.Bucket, number float64, metdataKey string, metadata *T) (*bbolt.Bucket, float64, error) {
	if number == 0 {
		number = NextBucketWholeNumber(bucket)
	}

	b := bucket.Bucket(utils.Float64ToBytes(number))
	if b != nil {
		return nil, 0, ErrBucketExists
	}

	slog.Debug("bucket does not yet exist, creating new resource " + strconv.FormatFloat(number, 'f', -1, 64))

	sb, err := bucket.CreateBucket(utils.Float64ToBytes(number))
	if err != nil {
		return nil, 0, err
	}

	err = PutKey[T](sb, metadata, metdataKey)
	if err != nil {
		return nil, 0, err
	}

	return sb, number, nil
}
