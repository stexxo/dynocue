package main

import (
	"path"
	"strings"

	"github.com/nats-io/nats.go"
	"gitlab.com/stexxo/dynocue/internal/subsystems"
)

type Commands struct {
	bus     *nats.Conn
	manager *subsystems.ShowManager
}

func (c *Commands) OpenLocalShow(filePath string) (string, bool) {
	if filePath == "" {
		return "", false
	}

	if !strings.HasSuffix(filePath, ".dynocue") {
		return "", false
	}

	mgr, err := subsystems.NewShowManager(filePath)
	if err != nil {
		return "", false
	}

	evBus, err := mgr.GetBusConnection()
	if err != nil {
		return "", false
	}

	c.CloseShow()
	c.manager = mgr
	c.bus = evBus

	return path.Base(filePath), true
}

func (c *Commands) CreateLocalShow(filePath string) (string, bool) {
	if filePath == "" {
		return "", false
	}

	if !strings.HasSuffix(filePath, ".dynocue") {
		filePath += ".dynocue"
	}

	mgr, err := subsystems.NewShowManager(filePath)
	if err != nil {
		return "", false
	}

	evBus, err := mgr.GetBusConnection()
	if err != nil {
		return "", false
	}

	c.CloseShow()
	c.manager = mgr
	c.bus = evBus

	return path.Base(filePath), true
}

func (c *Commands) CloseShow() bool {
	if c.bus == nil {
		c.bus.Close()
		c.bus = nil
	}

	if c.manager != nil {
		c.manager.Stop()
		c.manager = nil
	}

	return true
}
