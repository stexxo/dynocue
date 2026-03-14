package cues

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apibus "gitlab.com/stexxo/dynocue/api/bus"
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
		sub, err := apibus.Subscribe[apicues.NewCueListEvent](conn, apicues.EventNewCueList, func(s string, msg *apicues.NewCueListEvent) {
			assert.Equal(t, apicues.EventNewCueList, s)
			eventChan <- *msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 1})
		require.NoError(t, err)
		assert.Equal(t, float64(1), res.ResponseValue.Number)

		select {
		case event := <-eventChan:
			assert.Equal(t, float64(1), event.Number)
			assert.Equal(t, "", event.Label)
			assert.Equal(t, "", event.ListType)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for NewCueListEvent")
		}

		// Verify metadata was stored correctly
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("cuelists"))
			sb := b.Bucket(utils.Float64ToBytes(1))
			require.NotNil(t, sb)
			v := sb.Get([]byte("metadata"))
			require.NotNil(t, v)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Update, Get, Enumerate, and Delete", func(t *testing.T) {
		// Subscribe to update events
		updateChan := make(chan apicues.UpdateCueListMetadataEvent, 2)
		sub, err := apibus.Subscribe[apicues.UpdateCueListMetadataEvent](conn, apicues.EventUpdateCueList, func(s string, msg *apicues.UpdateCueListMetadataEvent) {
			updateChan <- *msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Update Label
		updateRes, err := cs.UpdateCueListMetadata("request.cuelist.metadata.update.label", apicues.UpdateCueListMetadataInput{
			Number: 1,
			Key:    "label",
			Value:  "Test Label",
		})
		require.NoError(t, err)
		require.Nil(t, updateRes.MessageError)
		assert.NotNil(t, updateRes.ResponseValue)

		// Verify update event
		select {
		case event := <-updateChan:
			assert.Equal(t, float64(1), event.Number)
			assert.Equal(t, "Test Label", event.Label)
			assert.Equal(t, "", event.ListType)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for UpdateCueListMetadataEvent")
		}

		// Update ListType
		updateRes, err = cs.UpdateCueListMetadata("request.cuelist.metadata.update.listType", apicues.UpdateCueListMetadataInput{
			Number: 1,
			Key:    "listType",
			Value:  "Main",
		})
		require.NoError(t, err)
		require.Nil(t, updateRes.MessageError)
		assert.NotNil(t, updateRes.ResponseValue)

		// Verify update event for ListType
		select {
		case event := <-updateChan:
			assert.Equal(t, float64(1), event.Number)
			assert.Equal(t, "Test Label", event.Label)
			assert.Equal(t, "Main", event.ListType)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for UpdateCueListMetadataEvent (ListType)")
		}

		// Get
		getRes, err := cs.GetCueListMetadata(apicues.RequestGetCueListMetadata, apicues.GetCueListMetadataInput{Number: 1})
		require.NoError(t, err)
		assert.Equal(t, float64(1), getRes.ResponseValue.Number)
		assert.Equal(t, "Test Label", getRes.ResponseValue.Label)
		assert.Equal(t, "Main", getRes.ResponseValue.ListType)

		// Verify update events
		// (We should probably have a more robust way to test events, but for now we just verify the database)

		// Enumerate
		enumRes, err := cs.EnumerateCueList(apicues.RequestEnumerateCueList, apicues.EnumerateCueListInput{})
		require.NoError(t, err)
		assert.Len(t, enumRes.ResponseValue.CueLists, 1)
		assert.Equal(t, float64(1), enumRes.ResponseValue.CueLists[0].Number)

		// Delete
		deleteChan := make(chan apicues.DeleteCueListEvent, 1)
		dsub, err := apibus.Subscribe[apicues.DeleteCueListEvent](conn, apicues.EventDeleteCueList, func(s string, msg *apicues.DeleteCueListEvent) {
			deleteChan <- *msg
		})
		require.NoError(t, err)
		defer dsub.Unsubscribe()

		deleteRes, err := cs.DeleteCueList(apicues.RequestDeleteCueList, apicues.DeleteCueListInput{Number: 1})
		require.NoError(t, err)
		assert.NotNil(t, deleteRes.ResponseValue)

		// Verify delete event
		select {
		case event := <-deleteChan:
			assert.Equal(t, float64(1), event.Number)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for DeleteCueListEvent")
		}

		// Verify deleted
		getRes, err = cs.GetCueListMetadata(apicues.RequestGetCueListMetadata, apicues.GetCueListMetadataInput{Number: 1})
		require.NoError(t, err)
		assert.NotNil(t, getRes.MessageError)
		assert.Equal(t, apibus.CodeResourceNotFound, getRes.MessageError.Code)
	})

	t.Run("Create next cuelist (0 provided)", func(t *testing.T) {
		// We deleted 1. Next should be 1 again (since max is 0).
		// Wait, if we delete 1, NextBucketWholeNumber will see no buckets and return 1.
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 0})
		require.NoError(t, err)
		assert.NotNil(t, res.ResponseValue)
		assert.Equal(t, float64(1), res.ResponseValue.Number)

		// Verify bucket exists
		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("cuelists"))
			sb := b.Bucket(utils.Float64ToBytes(1))
			require.NotNil(t, sb)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Create another specific number", func(t *testing.T) {
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 10})
		require.NoError(t, err)
		assert.Equal(t, float64(10), res.ResponseValue.Number)

		// Create a non-whole number list
		res, err = cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 10.5})
		require.NoError(t, err)
		assert.Equal(t, float64(10.5), res.ResponseValue.Number)

		// Next 0 should be 11, not 11.5
		res, err = cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 0})
		require.NoError(t, err)
		assert.Equal(t, float64(11), res.ResponseValue.Number)
	})

	t.Run("Sub-bucket already exists", func(t *testing.T) {
		// 10 already exists, should return CodeResourceConflict
		res, err := cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: 10})
		require.NoError(t, err)
		require.NotNil(t, res.MessageError)
		assert.Equal(t, apibus.CodeResourceConflict, res.MessageError.Code)
		assert.Contains(t, res.MessageError.ErrorMessage, "already exists")
	})

	t.Run("Validation failure", func(t *testing.T) {
		// Subscribe to update events
		updateChan := make(chan apicues.UpdateCueListMetadataEvent, 1)
		sub, err := apibus.Subscribe[apicues.UpdateCueListMetadataEvent](conn, apicues.EventUpdateCueList, func(s string, msg *apicues.UpdateCueListMetadataEvent) {
			updateChan <- *msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Update with empty value (should now be allowed)
		res, err := cs.UpdateCueListMetadata(apicues.RequestUpdateCueListMetadata+".label", apicues.UpdateCueListMetadataInput{
			Number: 1,
			Key:    "label",
			Value:  "",
		})
		require.NoError(t, err)
		require.Nil(t, res.MessageError)
		assert.NotNil(t, res.ResponseValue)

		// Verify event also has empty label
		select {
		case event := <-updateChan:
			assert.Equal(t, float64(1), event.Number)
			assert.Equal(t, "", event.Label)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for UpdateCueListMetadataEvent (empty value)")
		}

		// Get with invalid number (gt=0)
		resGet, err := cs.GetCueListMetadata(apicues.RequestGetCueListMetadata, apicues.GetCueListMetadataInput{Number: -1})
		require.NoError(t, err)
		require.NotNil(t, resGet.MessageError)
		assert.Equal(t, apibus.CodePayloadValidationFailure, resGet.MessageError.Code)
		assert.Contains(t, resGet.MessageError.ErrorMessage, "validation failed")
	})
}
