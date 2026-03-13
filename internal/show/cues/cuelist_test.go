package cues

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apicues "gitlab.com/stexxo/dynocue/api/cues"
	"gitlab.com/stexxo/dynocue/internal/bus"
	"gitlab.com/stexxo/dynocue/internal/utils"
	"go.etcd.io/bbolt"
)

func TestNewCueList(t *testing.T) {
	// Start an in-process NATS server
	ns, err := bus.NewBus()
	require.NoError(t, err)
	defer ns.Shutdown()

	conn, err := bus.GetInProcessConn(ns)
	require.NoError(t, err)
	defer conn.Close()

	// Create a temporary database
	dbPath := "test_cues.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	require.NoError(t, err)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	cs, err := NewCues(conn, db)
	require.NoError(t, err)

	t.Run("Create cuelist and verify event", func(t *testing.T) {
		eventChan := make(chan apicues.NewCueListEvent, 1)
		sub, err := bus.Subscribe(conn, apicues.RequestCreateCueList+"*", func(s string, msg apicues.NewCueListEvent) {
			assert.Equal(t, apicues.EventNewCueList+"1", s)
			eventChan <- msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 1})
		require.NoError(t, err)
		assert.Equal(t, float64(1), res.ResponseValue.Number)

		select {
		case event := <-eventChan:
			assert.Equal(t, float64(1), event.Number)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for NewCueListEvent")
		}
	})

	t.Run("Create next cuelist (0 provided)", func(t *testing.T) {
		// We already have 1. Next should be 2.
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 0})
		require.NoError(t, err)
		assert.NotNil(t, res.ResponseValue)
		assert.Equal(t, float64(2), res.ResponseValue.Number)

		// Verify bucket exists
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("cuelists"))
			sb := b.Bucket(utils.Float64ToBytes(2))
			require.NotNil(t, sb)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Create another specific number", func(t *testing.T) {
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 10})
		require.NoError(t, err)
		assert.Equal(t, float64(10), res.ResponseValue.Number)

		// Create a non-whole number list
		res, err = cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 10.5})
		require.NoError(t, err)
		assert.Equal(t, float64(10.5), res.ResponseValue.Number)

		// Next 0 should be 11, not 11.5
		res, err = cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 0})
		require.NoError(t, err)
		assert.Equal(t, float64(11), res.ResponseValue.Number)
	})

	t.Run("Sub-bucket already exists", func(t *testing.T) {
		// 10 already exists, should return ConflictCode
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.NewCueListInput{Number: 10})
		require.NoError(t, err)
		require.NotNil(t, res.MessageError)
		assert.Equal(t, bus.ConflictCode, res.MessageError.Code)
		assert.Contains(t, res.MessageError.ErrorMessage, "already exists")
	})
}
