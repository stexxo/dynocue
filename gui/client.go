package gui

import (
	"errors"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/client"
	"github.com/stexxo/dynocue/core"
	"github.com/stexxo/dynocue/core/logging"
)

type ClientManager struct {
	connected bool
	remote    bool
	client    *client.Client
	core      *core.DynoCue
	logger    logging.Logger
}

func NewClientManager(logger logging.Logger) *ClientManager {
	return &ClientManager{logger: logger}
}

func (cm *ClientManager) Connected() bool {
	return cm.connected
}

func (cm *ClientManager) Remote() bool {
	return cm.remote
}

func (cm *ClientManager) Client() (*client.Client, error) {
	if cm.client == nil {
		return nil, errors.New("client not connected")
	}
	return cm.client, nil
}

func (cm *ClientManager) Core() (*core.DynoCue, error) {
	if cm.core == nil {
		return nil, errors.New("not connected or not local")
	}
	return cm.core, nil
}

func (cm *ClientManager) ConnectLocal(core *core.DynoCue) error {
	conn, err := core.GetInProcessConn("local-client")
	if err != nil {
		return err
	}

	c := client.NewClient(conn, cm.logger)

	cm.client = c
	cm.core = core

	cm.connected = true
	cm.remote = false
	return nil
}

func (cm *ClientManager) ConnectRemote(addr string) error {
	conn, err := nats.Connect(addr, nats.MaxReconnects(-1), nats.ReconnectWait(1*time.Second))
	if err != nil {
		return err
	}
	c := client.NewClient(conn, cm.logger)
	cm.client = c
	cm.connected = true
	cm.remote = true
	return nil
}

func (cm *ClientManager) Disconnect() error {
	if !cm.connected {
		return nil
	}
	cm.client = nil
	cm.core = nil
	cm.connected = false
	cm.remote = false
	return nil
}
