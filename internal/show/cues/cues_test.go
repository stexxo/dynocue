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

func TestNewCue(t *testing.T) {
	// Start an in-process NATS server
	ns, err := bus.NewBus()
	require.NoError(t, err)
	defer ns.Shutdown()

	conn, err := bus.GetInProcessConn(ns)
	require.NoError(t, err)
	defer conn.Close()

	// Create a temporary database
	dbPath := "test_cues_new.db"
	db, err := bbolt.Open(dbPath, 0600, nil)
	require.NoError(t, err)
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	cs, err := NewCues(conn, db)
	require.NoError(t, err)

	// Create a CueList first
	clNum := float64(10)
	_, err = cs.NewCueList(apicues.RequestCreateCueList, apicues.CreateCueListInput{Number: clNum})
	require.NoError(t, err)

	t.Run("Create cue and verify event", func(t *testing.T) {
		eventChan := make(chan apicues.NewCueEvent, 1)
		sub, err := apibus.Subscribe[apicues.NewCueEvent](conn, apicues.EventNewCue, func(s string, msg *apicues.NewCueEvent) {
			assert.Equal(t, apicues.EventNewCue, s)
			eventChan <- *msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		cueNum := float64(1)
		res, err := cs.NewCue(apicues.RequestCreateCue, apicues.CreateCueInput{
			CueListNumber: clNum,
			Number:        cueNum,
		})
		require.NoError(t, err)
		require.Nil(t, res.MessageError)
		assert.Equal(t, cueNum, res.ResponseValue.Number)

		select {
		case event := <-eventChan:
			assert.Equal(t, clNum, event.CueListNumber)
			assert.Equal(t, cueNum, event.Number)
			assert.Equal(t, "", event.Label)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for NewCueEvent")
		}

		// Verify bucket structure
		err = db.View(func(tx *bbolt.Tx) error {
			clb := tx.Bucket([]byte("cuelists"))
			require.NotNil(t, clb)
			sb := clb.Bucket(utils.Float64ToBytes(clNum))
			require.NotNil(t, sb)
			cb := sb.Bucket([]byte("cues"))
			require.NotNil(t, cb)
			nb := cb.Bucket(utils.Float64ToBytes(cueNum))
			require.NotNil(t, nb)
			v := nb.Get([]byte("metadata"))
			require.NotNil(t, v)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("Create cue with number 0 (next whole number)", func(t *testing.T) {
		// Existing cue is 1. Next should be 2.
		res, err := cs.NewCue(apicues.RequestCreateCue, apicues.CreateCueInput{
			CueListNumber: clNum,
			Number:        0,
		})
		require.NoError(t, err)
		require.Nil(t, res.MessageError)
		assert.Equal(t, float64(2), res.ResponseValue.Number)
	})

	t.Run("Create cue in non-existent cuelist", func(t *testing.T) {
		res, err := cs.NewCue(apicues.RequestCreateCue, apicues.CreateCueInput{
			CueListNumber: 999,
			Number:        1,
		})
		require.NoError(t, err)
		assert.NotNil(t, res.MessageError)
		assert.Equal(t, apibus.CodeResourceNotFound, res.MessageError.Code)
	})

	t.Run("Create duplicate cue", func(t *testing.T) {
		res, err := cs.NewCue(apicues.RequestCreateCue, apicues.CreateCueInput{
			CueListNumber: clNum,
			Number:        1,
		})
		require.NoError(t, err)
		assert.NotNil(t, res.MessageError)
		assert.Equal(t, apibus.CodeResourceConflict, res.MessageError.Code)
	})

	t.Run("Update, Get, Enumerate, and Delete Cue", func(t *testing.T) {
		cueNum := float64(1)

		// Subscribe to update events
		updateChan := make(chan apicues.UpdateCueMetadataEvent, 1)
		sub, err := apibus.Subscribe[apicues.UpdateCueMetadataEvent](conn, apicues.EventUpdateCue, func(s string, msg *apicues.UpdateCueMetadataEvent) {
			updateChan <- *msg
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()

		// Update Label
		updateRes, err := cs.UpdateCueMetadata(apicues.RequestUpdateCueMetadata, apicues.UpdateCueMetadataInput{
			CueListNumber: clNum,
			Number:        cueNum,
			Key:           "label",
			Value:         "My New Cue",
		})
		require.NoError(t, err)
		require.Nil(t, updateRes.MessageError)

		// Verify update event
		select {
		case event := <-updateChan:
			assert.Equal(t, clNum, event.CueListNumber)
			assert.Equal(t, cueNum, event.Number)
			assert.Equal(t, "My New Cue", event.Label)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for UpdateCueMetadataEvent")
		}

		// Get Metadata
		getRes, err := cs.GetCueMetadata(apicues.RequestGetCueMetadata, apicues.GetCueMetadataInput{
			CueListNumber: clNum,
			Number:        cueNum,
		})
		require.NoError(t, err)
		require.Nil(t, getRes.MessageError)
		assert.Equal(t, "My New Cue", getRes.ResponseValue.Label)

		// Enumerate Cues
		enumRes, err := cs.EnumerateCue(apicues.RequestEnumerateCue, apicues.EnumerateCueInput{
			CueListNumber: clNum,
		})
		require.NoError(t, err)
		require.Nil(t, enumRes.MessageError)
		assert.Len(t, enumRes.ResponseValue.Cues, 2) // We have 1 and 2

		// Move Cue
		deleteChan := make(chan apicues.DeleteCueEvent, 1)
		newChan := make(chan apicues.NewCueEvent, 1)
		subDel, _ := apibus.Subscribe[apicues.DeleteCueEvent](conn, apicues.EventDeleteCue, func(s string, msg *apicues.DeleteCueEvent) {
			deleteChan <- *msg
		})
		subNew, _ := apibus.Subscribe[apicues.NewCueEvent](conn, apicues.EventNewCue, func(s string, msg *apicues.NewCueEvent) {
			newChan <- *msg
		})
		defer subDel.Unsubscribe()
		defer subNew.Unsubscribe()

		moveRes, err := cs.MoveCue(apicues.RequestMoveCue, apicues.MoveCueInput{
			CueListNumber:  clNum,
			OriginalNumber: cueNum,
			NewNumber:      3,
		})
		require.NoError(t, err)
		require.Nil(t, moveRes.MessageError)
		assert.Equal(t, float64(3), moveRes.ResponseValue.NewNumber)

		// Verify move events
		select {
		case event := <-deleteChan:
			assert.Equal(t, cueNum, event.Number)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for DeleteCueEvent during move")
		}
		select {
		case event := <-newChan:
			assert.Equal(t, float64(3), event.Number)
			assert.Equal(t, "My New Cue", event.Label)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for NewCueEvent during move")
		}

		// Delete Cue
		delRes, err := cs.DeleteCue(apicues.RequestDeleteCue, apicues.DeleteCueInput{
			CueListNumber: clNum,
			Number:        3,
		})
		require.NoError(t, err)
		require.Nil(t, delRes.MessageError)

		// Verify deletion
		enumRes, err = cs.EnumerateCue(apicues.RequestEnumerateCue, apicues.EnumerateCueInput{
			CueListNumber: clNum,
		})
		require.NoError(t, err)
		assert.Len(t, enumRes.ResponseValue.Cues, 1) // Only 2 remains
	})
}
