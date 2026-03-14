package data

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/bbolt"
)

func TestCopyBucket(t *testing.T) {
	dbPath := "test_copy_bucket.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	require.NoError(t, err)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	err = db.Update(func(tx *bbolt.Tx) error {
		src, err := tx.CreateBucket([]byte("src"))
		require.NoError(t, err)

		// Add some keys
		err = src.Put([]byte("key1"), []byte("value1"))
		require.NoError(t, err)
		err = src.Put([]byte("key2"), []byte("value2"))
		require.NoError(t, err)

		// Add a sub-bucket
		sub, err := src.CreateBucket([]byte("sub"))
		require.NoError(t, err)
		err = sub.Put([]byte("subKey1"), []byte("subValue1"))
		require.NoError(t, err)

		dst, err := tx.CreateBucket([]byte("dst"))
		require.NoError(t, err)

		// Perform copy
		err = CopyBucket(src, dst)
		require.NoError(t, err)

		// Verify dst
		assert.Equal(t, []byte("value1"), dst.Get([]byte("key1")))
		assert.Equal(t, []byte("value2"), dst.Get([]byte("key2")))

		subDst := dst.Bucket([]byte("sub"))
		require.NotNil(t, subDst)
		assert.Equal(t, []byte("subValue1"), subDst.Get([]byte("subKey1")))

		return nil
	})
	require.NoError(t, err)
}
