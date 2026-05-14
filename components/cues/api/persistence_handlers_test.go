package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/components/cues/model"
	"github.com/stexxo/dynocue/components/system"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
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
	return args.Get(0).(*jetstream.ObjectInfo), args.Error(1)
}

func (m *mockObjectStore) Get(ctx context.Context, name string, opts ...jetstream.GetObjectOpt) (jetstream.ObjectResult, error) {
	args := m.Called(ctx, name, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jetstream.ObjectResult), args.Error(1)
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

func TestSaveModel(t *testing.T) {
	m, _ := model.NewCueingModel()
	mockOS := new(mockObjectStore)

	pm := system.NewPersistenceManagerForTest("cueing", nil, mockOS, logging.NewNoopLogger())

	api := &CueingApi{
		model:       m,
		persistence: pm,
	}

	// Expect Put to be called for each persistent table: cuelists, cues, actions
	mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
		return meta.Name == "cueing/cuelists"
	}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)

	mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
		return meta.Name == "cueing/cues"
	}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)

	mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
		return meta.Name == "cueing/actions"
	}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)

	resp, err := api.SaveModel("", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	mockOS.AssertExpectations(t)
}

func TestLoadModel(t *testing.T) {
	m, _ := model.NewCueingModel()
	mockOS := new(mockObjectStore)

	pm := system.NewPersistenceManagerForTest("cueing", nil, mockOS, logging.NewNoopLogger())

	api := &CueingApi{
		model:       m,
		persistence: pm,
	}

	// Expect Get to be called for each persistent table
	var emptyBuf bytes.Buffer
	gw := gzip.NewWriter(&emptyBuf)
	gw.Close()
	emptyGzip := emptyBuf.Bytes()

	mockOS.On("Get", mock.Anything, "cueing/cuelists", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)
	mockOS.On("Get", mock.Anything, "cueing/cues", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)
	mockOS.On("Get", mock.Anything, "cueing/actions", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)

	resp, err := api.LoadModel("", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	mockOS.AssertExpectations(t)
}

func TestRegisterPersistenceApis(t *testing.T) {
	s, nc := testServer()
	defer s.Shutdown()
	defer nc.Close()

	m, _ := model.NewCueingModel()
	messenger := messaging.NewMessenger(&messaging.MessengerCfg{
		Conn: nc,
	})

	mockOS := new(mockObjectStore)
	pm := system.NewPersistenceManagerForTest("cueing", nil, mockOS, logging.NewNoopLogger())

	_, err := NewCueingApi(m, pm, messenger, logging.NewNoopLogger())
	require.NoError(t, err)

	t.Run("Save Registration", func(t *testing.T) {
		mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
			return meta.Name == "cueing/cuelists"
		}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)
		mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
			return meta.Name == "cueing/cues"
		}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)
		mockOS.On("Put", mock.Anything, mock.MatchedBy(func(meta jetstream.ObjectMeta) bool {
			return meta.Name == "cueing/actions"
		}), mock.Anything).Return(&jetstream.ObjectInfo{}, nil)

		resp, err := messaging.Request[string](messenger, SaveRequestSubject, "")
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})

	t.Run("Load Registration", func(t *testing.T) {
		var emptyBuf bytes.Buffer
		gw := gzip.NewWriter(&emptyBuf)
		gw.Close()
		emptyGzip := emptyBuf.Bytes()
		// Mock for all 3 tables
		mockOS.On("Get", mock.Anything, "cueing/cuelists", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)
		mockOS.On("Get", mock.Anything, "cueing/cues", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)
		mockOS.On("Get", mock.Anything, "cueing/actions", mock.Anything).Return(&mockObjectResult{reader: bytes.NewReader(emptyGzip)}, nil)

		resp, err := messaging.Request[string](messenger, LoadRequestSubject, "")
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})
}
