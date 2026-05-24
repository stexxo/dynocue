// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/audio/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockObjectStore struct {
	jetstream.ObjectStore
	mock.Mock
}

func (m *mockObjectStore) Put(ctx context.Context, meta jetstream.ObjectMeta, reader io.Reader) (*jetstream.ObjectInfo, error) {
	data, _ := io.ReadAll(reader)
	args := m.Called(ctx, meta, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jetstream.ObjectInfo), args.Error(1)
}

func (m *mockObjectStore) Get(ctx context.Context, name string, opts ...jetstream.GetObjectOpt) (jetstream.ObjectResult, error) {
	args := m.Called(ctx, name, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jetstream.ObjectResult), args.Error(1)
}

func (m *mockObjectStore) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

type mockObjectResult struct {
	reader io.Reader
}

func (m *mockObjectResult) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockObjectResult) Close() error {
	return nil
}

func (m *mockObjectResult) Info() (*jetstream.ObjectInfo, error) {
	return &jetstream.ObjectInfo{}, nil
}

func (m *mockObjectResult) Error() error {
	return nil
}

func setup(t *testing.T) (*model.AudioModel, *AudioAPI, *mockObjectStore, *mockObjectStore) {
	am, err := model.NewAudioModel()
	require.NoError(t, err)

	mockOS := new(mockObjectStore)
	mockTempOS := new(mockObjectStore)
	pm := system.NewPersistenceManagerForTest("audio", nil, mockOS, mockTempOS, logging.NewNoopLogger())

	api := NewAudioAPI(am, pm, nil, logging.NewNoopLogger())
	return am, api, mockOS, mockTempOS
}

func TestCreateAudioFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, api, mockOS, mockTempOS := setup(t)

		req := &UploadAudioFileRequest{
			LocationInTempBucket: "temp/file.wav",
		}

		mockTempOS.On("Get", mock.Anything, "temp/file.wav", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader([]byte("audio data"))}, nil)
		mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
			return strings.HasPrefix(meta.Name, "audio/files/")
		}), []byte("audio data")).Return(&jetstream.ObjectInfo{}, nil)
		mockTempOS.On("Delete", mock.Anything, "temp/file.wav").Return(nil)

		resp, err := api.CreateAudioFile("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.FileId)

		mockOS.AssertExpectations(t)
		mockTempOS.AssertExpectations(t)
	})

	t.Run("Persistence Error", func(t *testing.T) {
		_, api, _, mockTempOS := setup(t)

		req := &UploadAudioFileRequest{
			LocationInTempBucket: "temp/file.wav",
		}

		mockTempOS.On("Get", mock.Anything, "temp/file.wav", mock.Anything).Return(nil, fmt.Errorf("disk error"))

		resp, err := api.CreateAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to copy file to object store")
	})
}

func TestReplaceAudioFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		am, api, mockOS, mockTempOS := setup(t)

		fileId := "test-file-id"
		fileKey := "files/test-file-id"
		err := am.AddFile(fileId, fileKey)
		require.NoError(t, err)

		req := &ReplaceAudioFileRequest{
			FileId:               fileId,
			LocationInTempBucket: "temp/new-file.wav",
		}

		mockTempOS.On("Get", mock.Anything, "temp/new-file.wav", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader([]byte("new data"))}, nil)
		mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
			return meta.Name == "audio/files/test-file-id"
		}), []byte("new data")).Return(&jetstream.ObjectInfo{}, nil)
		mockTempOS.On("Delete", mock.Anything, "temp/new-file.wav").Return(nil)

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		mockOS.AssertExpectations(t)
		mockTempOS.AssertExpectations(t)
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, api, _, _ := setup(t)

		req := &ReplaceAudioFileRequest{
			FileId:               "non-existent",
			LocationInTempBucket: "temp/new-file.wav",
		}

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get file from database")
	})

	t.Run("Persistence Error", func(t *testing.T) {
		am, api, _, mockTempOS := setup(t)

		fileId := "test-file-id"
		am.AddFile(fileId, "files/test-file-id")

		req := &ReplaceAudioFileRequest{
			FileId:               fileId,
			LocationInTempBucket: "temp/new-file.wav",
		}

		mockTempOS.On("Get", mock.Anything, "temp/new-file.wav", mock.Anything).Return(nil, fmt.Errorf("disk error"))

		resp, err := api.ReplaceAudioFile("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to copy file from temp location")
	})
}

func TestEnumerateAudioFiles(t *testing.T) {
	am, api, _, _ := setup(t)

	am.AddFile("id1", "key1")
	am.AddFile("id2", "key2")

	resp, err := api.EnumerateAudioFiles("sub", &EnumerateAudioFilesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Files, 2)
}

func TestGetAudioFile(t *testing.T) {
	am, api, _, _ := setup(t)

	am.AddFile("id1", "key1")

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

func TestDeleteAudioFile(t *testing.T) {
	am, api, _, _ := setup(t)

	am.AddFile("id1", "key1")

	resp, err := api.DeleteAudioFile("sub", &DeleteAudioFileRequest{FileId: "id1"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify deleted from model
	_, err = am.GetFile("id1")
	assert.Error(t, err)
}
