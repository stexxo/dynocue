// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func testServer() (*server.Server, *nats.Conn) {
	s, _ := server.NewServer(&server.Options{
		DontListen: true,
		JetStream:  true,
	})
	s.Start()
	nc, _ := nats.Connect("", nats.InProcessServer(s))
	return s, nc
}

func setup(t *testing.T) (*model.AudioModel, *AudioAPI, jetstream.ObjectStore, jetstream.ObjectStore, *messaging.Messenger) {
	s, nc := testServer()
	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})

	am, err := model.NewAudioModel()
	require.NoError(t, err)

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	ctx := context.Background()
	os, err := js.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{Bucket: "audio-files"})
	require.NoError(t, err)

	tempOS, err := js.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{Bucket: "audio-temp"})
	require.NoError(t, err)

	messenger := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn: nc,
		Js:   js,
	})

	pm := system.NewPersistenceManagerForTest("audio", nil, os, tempOS, logging.NewNoopLogger())

	api, err := NewAudioAPI(am, pm, messenger, logging.NewNoopLogger())
	require.NoError(t, err)
	return am, api, os, tempOS, messenger
}

func TestCreateAudioFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api, os, tempOS, _ := setup(t)

		ctx := context.Background()
		tempKey := "temp/file.wav"
		_, err := tempOS.Put(ctx, jetstream.ObjectMeta{Name: tempKey}, bytes.NewReader([]byte("audio data")))
		require.NoError(t, err)

		req := &UploadAudioFileRequest{
			LocationInTempBucket: tempKey,
		}

		resp, err := api.CreateAudioFile("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.FileId)
		assert.Equal(t, uint(1), resp.Number)

		// Verify file is in the real object store
		res, err := os.Get(ctx, fmt.Sprintf("audio/files/%s", resp.FileId))
		assert.NoError(t, err)
		data, _ := io.ReadAll(res)
		assert.Equal(t, []byte("audio data"), data)

		// Verify temp file is deleted
		_, err = tempOS.Get(ctx, tempKey)
		assert.ErrorIs(t, err, jetstream.ErrObjectNotFound)
	})

	t.Run("Persistence Error", func(t *testing.T) {
		_, api, _, _, _ := setup(t)

		req := &UploadAudioFileRequest{
			LocationInTempBucket: "non-existent",
		}

		resp, err := api.CreateAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to copy file to object store")
	})

	t.Run("Number Exists", func(t *testing.T) {
		am, api, _, tempOS, _ := setup(t)

		_, err := am.AddFile("existing-file-id", "files/existing-file-id", 5)
		require.NoError(t, err)

		ctx := context.Background()
		tempKey := "temp/conflicting-file.wav"
		_, err = tempOS.Put(ctx, jetstream.ObjectMeta{Name: tempKey}, bytes.NewReader([]byte("audio data")))
		require.NoError(t, err)

		resp, err := api.CreateAudioFile("sub", &UploadAudioFileRequest{
			LocationInTempBucket: tempKey,
			Number:               5,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, model.ErrNumberExists)
	})
}

func TestReplaceAudioFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		am, api, os, tempOS, _ := setup(t)

		fileId := "test-file-id"
		fileKey := "files/test-file-id"
		_, err := am.AddFile(fileId, fileKey, 0)
		require.NoError(t, err)

		ctx := context.Background()
		tempKey := "temp/new-file.wav"
		_, err = tempOS.Put(ctx, jetstream.ObjectMeta{Name: tempKey}, bytes.NewReader([]byte("new data")))
		require.NoError(t, err)

		req := &ReplaceAudioFileRequest{
			FileId:               fileId,
			LocationInTempBucket: tempKey,
		}

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Verify file is updated in the real object store
		res, err := os.Get(ctx, "audio/files/test-file-id")
		assert.NoError(t, err)
		data, _ := io.ReadAll(res)
		assert.Equal(t, []byte("new data"), data)

		// Verify temp file is deleted
		_, err = tempOS.Get(ctx, tempKey)
		assert.ErrorIs(t, err, jetstream.ErrObjectNotFound)
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, api, _, _, _ := setup(t)

		req := &ReplaceAudioFileRequest{
			FileId:               "non-existent",
			LocationInTempBucket: "temp/new-file.wav",
		}

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, model.ErrFileNotFound)
	})

	t.Run("Persistence Error", func(t *testing.T) {
		am, api, _, _, _ := setup(t)

		fileId := "test-file-id"
		_, _ = am.AddFile(fileId, "files/test-file-id", 0)

		req := &ReplaceAudioFileRequest{
			FileId:               fileId,
			LocationInTempBucket: "non-existent",
		}

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to copy file from temp location")
	})
}

func TestEnumerateAudioFiles(t *testing.T) {
	am, api, _, _, _ := setup(t)

	_, _ = am.AddFile("id1", "key1", 0)
	_, _ = am.AddFile("id2", "key2", 0)

	resp, err := api.EnumerateAudioFiles("sub", &EnumerateAudioFilesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Files, 2)
}

func TestGetAudioFile(t *testing.T) {
	am, api, _, _, _ := setup(t)

	_, _ = am.AddFile("id1", "key1", 0)

	t.Run("Success", func(t *testing.T) {
		resp, err := api.GetAudioFile("sub", &GetAudioFileRequest{FileId: "id1"})
		assert.NoError(t, err)
		assert.Equal(t, "id1", resp.File.FileId)
	})

	t.Run("Not Found", func(t *testing.T) {
		resp, err := api.GetAudioFile("sub", &GetAudioFileRequest{FileId: "id-none"})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestUpdateAudioFileAttributes(t *testing.T) {
	am, api, _, _, _ := setup(t)

	_, err := am.AddFile("id1", "key1", 0)
	require.NoError(t, err)

	resp, err := api.UpdateAudioFileAttributes("sub", &UpdateAudioFileAttributesRequest{
		FileId: "id1",
		Field:  "label",
		Value:  "Intro",
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	file, err := am.GetFile("id1")
	require.NoError(t, err)
	assert.Equal(t, "Intro", file.Label)

	resp, err = api.UpdateAudioFileAttributes("sub", &UpdateAudioFileAttributesRequest{
		FileId: "id1",
		Field:  "number",
		Value:  uint(12),
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	file, err = am.GetFile("id1")
	require.NoError(t, err)
	assert.Equal(t, uint(12), file.Number)
}

func TestDeleteAudioFile(t *testing.T) {
	am, api, _, _, _ := setup(t)

	_, _ = am.AddFile("id1", "key1", 0)

	resp, err := api.DeleteAudioFile("sub", &DeleteAudioFileRequest{FileId: "id1"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify deleted from model
	_, err = am.GetFile("id1")
	assert.Error(t, err)
}

func TestFileEvents(t *testing.T) {
	am, _, _, _, messenger := setup(t)

	subCreated := make(chan *FileChangeEvent, 1)
	subDeleted := make(chan *FileChangeEvent, 1)

	_, err := messenger.Conn().Subscribe(FileCreatedEventSubject, func(m *nats.Msg) {
		var ev FileChangeEvent
		msgpack.Unmarshal(m.Data, &ev)
		subCreated <- &ev
	})
	require.NoError(t, err)

	_, err = messenger.Conn().Subscribe(DeleteFileEventSubject, func(m *nats.Msg) {
		var ev FileChangeEvent
		msgpack.Unmarshal(m.Data, &ev)
		subDeleted <- &ev
	})
	require.NoError(t, err)

	// Trigger created event
	_, err = am.AddFile("test-id", "test-key", 0)
	require.NoError(t, err)

	select {
	case ev := <-subCreated:
		assert.Equal(t, "test-id", ev.FileId)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for created event")
	}

	// Trigger deleted event
	err = am.DeleteFile("test-id")
	require.NoError(t, err)

	select {
	case ev := <-subDeleted:
		assert.Equal(t, "test-id", ev.FileId)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for deleted event")
	}
}
