// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
)

type Manager struct {
	mu             sync.RWMutex
	connected      bool
	remote         bool
	client         *Client
	core           *core.DynoCue
	logger         logging.Logger
	onNewClientFns []func(*Client) error
}

func NewClientManager(logger logging.Logger) *Manager {
	return &Manager{logger: logger}
}

func (cm *Manager) Connected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connected
}

func (cm *Manager) Remote() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.remote
}

func (cm *Manager) WithClient(fn func(*Client) error) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.client == nil {
		return errors.New("client not connected")
	}
	return fn(cm.client)
}

func (cm *Manager) Core() (*core.DynoCue, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.core == nil {
		return nil, errors.New("not connected or not local")
	}
	return cm.core, nil
}

func (cm *Manager) ConnectLocal(core *core.DynoCue) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	conn, err := core.GetInProcessConn("local-client")
	if err != nil {
		return err
	}

	c := NewClient(conn, cm.logger)

	cm.client = c
	cm.core = core

	for _, fn := range cm.onNewClientFns {
		if err := fn(cm.client); err != nil {
			return err
		}
	}

	cm.connected = true
	cm.remote = false
	return nil
}

func (cm *Manager) ConnectRemote(addr string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	conn, err := nats.Connect(addr, nats.MaxReconnects(-1), nats.ReconnectWait(1*time.Second))
	if err != nil {
		return err
	}
	c := NewClient(conn, cm.logger)
	cm.client = c
	cm.core = nil

	for _, fn := range cm.onNewClientFns {
		if err := fn(cm.client); err != nil {
			return err
		}
	}

	cm.connected = true
	cm.remote = true
	return nil
}

func (cm *Manager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if !cm.connected {
		return nil
	}
	cm.client = nil
	cm.core = nil
	cm.connected = false
	cm.remote = false
	return nil
}

func (cm *Manager) OnNewClient(fn func(*Client) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onNewClientFns = append(cm.onNewClientFns, fn)
}
