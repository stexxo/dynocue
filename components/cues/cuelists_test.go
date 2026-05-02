// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cues

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

type PersistenceRegistrationResponse struct {
	KeyValueStoreName string `msgpack:"keyValueStoreName" json:"keyValueStoreName"`
	ObjectStoreName   string `msgpack:"objectStoreName" json:"objectStoreName"`
}

func setupTestCueing(t *testing.T) (*Cueing, func()) {
	s, _ := server.NewServer(&server.Options{DontListen: true, JetStream: true})
	s.Start()
	nc, _ := nats.Connect("", nats.InProcessServer(s))

	js, _ := jetstream.New(nc)

	logger := logging.NewNoopLogger()
	p := New(logger)

	// Manually initialize what's needed for SubsystemCore and Cueing
	p.SubsystemCore = core.NewSubsystemCore("cueing", logger, p.onStart)

	// Initialize the database manually
	err := p.initiateDatabase()
	require.NoError(t, err)

	// We need a messenger to satisfy p.Messenger() calls in CreateCueList
	_ = messaging.NewMessenger(&messaging.MessengerCfg{
		Conn:      nc,
		Logger:    logger,
		Validator: validator.New(),
		Js:        js,
	})

	// Use reflection to set the private messenger field in SubsystemCore
	// Actually, we can just use p.Start(nc) if we mock the persistence responder
	// but it's easier to just inject the messenger if possible.
	// Since we are in the same package 'cues', we might still not see 'messenger' in SubsystemCore because it's in package 'core'.
	// Let's see if we can use a simpler approach.
	// SubsystemCore.Start sets the messenger. If we don't want to call onStart, we have a problem.

	// Alternative: call p.Start(nc) but make sure RegisterWithPersistence doesn't block.
	// We can do this by providing a responder for RegisterWithPersistence.
	go func() {
		// subject: "request.system.persistence.register"
		nc.Subscribe("request.system.persistence.register", func(m *nats.Msg) {
			// Create buckets first
			ctx := context.Background()
			js.CreateKeyValue(ctx, jetstream.KeyValueConfig{Bucket: "test_kv"})
			js.CreateObjectStore(ctx, jetstream.ObjectStoreConfig{Bucket: "test_obj"})

			// We must respond with a msgpack encoded ResponseEnvelope[PersistenceRegistrationResponse]
			resp := messaging.ResponseEnvelope[PersistenceRegistrationResponse]{
				Success: true,
				Response: &PersistenceRegistrationResponse{
					KeyValueStoreName: "test_kv",
					ObjectStoreName:   "test_obj",
				},
			}
			data, _ := msgpack.Marshal(resp)
			m.Respond(data)
		})
	}()

	err = p.Start(nc)
	require.NoError(t, err)

	cleanup := func() {
		p.SubsystemCore.Stop()
		nc.Close()
		s.Shutdown()
	}

	return p, cleanup
}

func TestCreateCueList(t *testing.T) {
	p, cleanup := setupTestCueing(t)
	defer cleanup()

	t.Run("Create cue list with automatic numbering starting at 1", func(t *testing.T) {
		// Create a fresh Cueing for this test to ensure it's empty
		p2, cleanup2 := setupTestCueing(t)
		defer cleanup2()

		req := &CreateCueListRequest{
			Number:      0,
			CueListType: "SEQUENTIAL",
		}
		resp, err := p2.CreateCueList("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint(1), resp.Number)
	})

	t.Run("Create cue list with specific number", func(t *testing.T) {
		req := &CreateCueListRequest{
			Number:      10,
			CueListType: "SEQUENTIAL",
		}
		resp, err := p.CreateCueList("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint(10), resp.Number)
		assert.NotEmpty(t, resp.Id)
	})

	t.Run("Create cue list with automatic numbering", func(t *testing.T) {
		// We already have 10 from previous test
		req := &CreateCueListRequest{
			Number:      0,
			CueListType: "SEQUENTIAL",
		}
		resp, err := p.CreateCueList("sub", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, uint(11), resp.Number)
	})

	t.Run("Create cue list with duplicate number fails", func(t *testing.T) {
		req := &CreateCueListRequest{
			Number:      10,
			CueListType: "SEQUENTIAL",
		}
		resp, err := p.CreateCueList("sub", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), CueListNumberExists)
	})
}

func TestEnumerateCueLists(t *testing.T) {
	p, cleanup := setupTestCueing(t)
	defer cleanup()

	t.Run("Enumerate empty cue lists", func(t *testing.T) {
		resp, err := p.EnumerateCueLists("sub", &EnumerateCueListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.CueLists)
	})

	t.Run("Enumerate multiple cue lists", func(t *testing.T) {
		// Create some cue lists
		_, _ = p.CreateCueList("sub", &CreateCueListRequest{Number: 1, CueListType: "SEQUENTIAL"})
		_, _ = p.CreateCueList("sub", &CreateCueListRequest{Number: 3, CueListType: "SEQUENTIAL"})
		_, _ = p.CreateCueList("sub", &CreateCueListRequest{Number: 2, CueListType: "SEQUENTIAL"})

		resp, err := p.EnumerateCueLists("sub", &EnumerateCueListsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.CueLists, 3)

		// Check order
		assert.Equal(t, uint(1), resp.CueLists[0].Number)
		assert.Equal(t, uint(2), resp.CueLists[1].Number)
		assert.Equal(t, uint(3), resp.CueLists[2].Number)
	})
}
