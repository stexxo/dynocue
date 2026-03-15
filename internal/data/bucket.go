package data

import (
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"gitlab.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
)

var ErrNoBucket = errors.New("bucket path does not exist")
var ErrKeyNotFound = errors.New("key not found")

func GetKey(b *bbolt.Bucket, v any, key []byte) error {
	val := b.Get(key)
	if val == nil {
		return ErrKeyNotFound
	}
	return msgpack.Unmarshal(val, v)
}

func PutKey(b *bbolt.Bucket, v any, key []byte) error {
	md, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put(key, md)
}

func GetBucketFromRoot(tx *bbolt.Tx, readOnly bool, rootKey BucketKey, path ...BucketKey) (*bbolt.Bucket, error) {
	var b *bbolt.Bucket
	var err error
	if !readOnly && rootKey.CreateIfNotExists {
		b, err = tx.CreateBucketIfNotExists(rootKey.Key)
		if err != nil {
			return nil, fmt.Errorf("could not create root bucket: %w", err)
		}
	} else {
		b = tx.Bucket(rootKey.Key)
		if b == nil {
			return nil, fmt.Errorf("%w: could not find the root bucket %s", ErrNoBucket, rootKey.Key)
		}
	}
	return GetBucket(b, readOnly, path...)
}

func GetBucket(bucket *bbolt.Bucket, readOnly bool, keys ...BucketKey) (*bbolt.Bucket, error) {
	if len(keys) == 0 {
		return bucket, nil
	}

	var b *bbolt.Bucket
	var err error
	k := keys[0]
	if !readOnly && k.CreateIfNotExists {
		b, err = bucket.CreateBucketIfNotExists(k.Key)
	} else {
		b = bucket.Bucket(k.Key)
	}

	if b == nil {
		return nil, fmt.Errorf("could not find or create bucket %s: %w", k.Key, errors.Join(err, ErrNoBucket))
	}

	return GetBucket(b, readOnly, keys[1:]...)
}

func UpdateAttributeInKeyValuePair[T any](b *bbolt.Bucket, dbKey []byte, attrKey, newValue string) (*T, error) {
	out := new(T)
	if err := GetKey(b, out, dbKey); err != nil {
		return out, err
	}
	if err := utils.SetFieldByTag(out, "msgpack", attrKey, newValue); err != nil {
		return out, err
	}
	if err := PutKey(b, out, dbKey); err != nil {
		return out, err
	}
	return out, nil
}

func EnumerateKeysFromSubBuckets[T comparable, E any](b *bbolt.Bucket, key []byte, bucketKeyFn func([]byte) T) (map[T]*E, error) {
	vals := make(map[T]*E)
	err := b.ForEachBucket(func(k []byte) error {
		sb := b.Bucket(k)
		var md E
		if err := GetKey(sb, &md, key); err != nil {
			if errors.Is(err, ErrKeyNotFound) {
				return fmt.Errorf("could not find key %s", key)
			}
			return err
		}
		vals[bucketKeyFn(k)] = &md
		return nil
	})
	return vals, err
}

type BucketKey struct {
	Key               []byte
	CreateIfNotExists bool
}

func NewFloatBucketKey(key float64, createIfNotExists bool) BucketKey {
	return BucketKey{
		Key:               utils.Float64ToBytes(key),
		CreateIfNotExists: createIfNotExists,
	}
}

func NewStringBucketKey(key string, createIfNotExists bool) BucketKey {
	return BucketKey{
		Key:               []byte(key),
		CreateIfNotExists: createIfNotExists,
	}
}

func NewBucketKey(key []byte, createIfNotExists bool) BucketKey {
	return BucketKey{
		Key:               key,
		CreateIfNotExists: createIfNotExists,
	}
}

type KeyValuePair struct {
	Key   []byte
	Value interface{}
}

type NewResourceBootstrap struct {
	InitialValues []KeyValuePair
	Buckets       []BucketKey
}

func AddIncrementedSubBucket(tx *bbolt.Tx, rootBucket BucketKey, path []BucketKey, key float64, bootstrap NewResourceBootstrap) (*bbolt.Bucket, float64, error) {
	b, err := GetBucketFromRoot(tx, false, rootBucket, path...)
	if err != nil {
		return nil, 0, err
	}

	if key == 0 {
		key = NextBucketWholeNumber(b)
	}

	b, err = b.CreateBucket(utils.Float64ToBytes(key))
	if err != nil {
		return nil, 0, err
	}

	for _, kv := range bootstrap.InitialValues {
		err = PutKey(b, kv.Value, kv.Key)
		if err != nil {
			return nil, 0, err
		}
	}

	for _, kv := range bootstrap.Buckets {
		_, err := b.CreateBucket(kv.Key)
		if err != nil {
			return nil, 0, err
		}
	}

	return b, key, nil
}

func DeleteBucketByPath(tx *bbolt.Tx, key []byte, rootKey BucketKey, path ...BucketKey) error {
	b, err := GetBucketFromRoot(tx, false, rootKey, path...)
	if err != nil {
		return err
	}

	bucketToDelete := b.Bucket(key)
	if bucketToDelete == nil {
		return nil
	}

	return b.DeleteBucket(key)
}

// RenameSubBucket copies a bucket to a new numeric key, updates its number in metadata,
// and deletes the old bucket.
func RenameSubBucket(parent *bbolt.Bucket, oldNum, newNum float64) error {
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

var ErrBucketExists = errors.New("bucket already exists")
