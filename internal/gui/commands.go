package gui

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/wailsapp/wails/v3/pkg/application"
	apibus "gitlab.com/stexxo/dynocue/api/bus"
	"gitlab.com/stexxo/dynocue/internal/show"
)

type Commands struct {
	app *application.App

	show          *show.Show
	conn          *nats.Conn
	subscriptions []*nats.Subscription
}

func NewCommands() *Commands {
	return &Commands{}
}

func (c *Commands) SetApplication(app *application.App) {
	c.app = app
}

func (c *Commands) OpenShow(path string) (string, bool) {
	if path == "" {
		return "", false
	}

	if !strings.HasSuffix(path, ".dynocue") {
		path = path + ".dynocue"
	}

	if c.show != nil {
		c.show.Close()
	}
	s, err := show.NewShow(path)
	if err != nil {
		return "", false
	}
	c.show = s
	c.conn, err = s.GetConn()
	if err != nil {
		return "", false
	}

	err = c.SubscribeToAll()
	if err != nil {
		slog.Error("Failed to subscribe to all events", "error", err)
		return "", false
	}

	return path, true
}

func (c *Commands) CloseShow() {
	if c.show != nil {
		c.conn.Close()
		c.show.Close()
		c.show = nil
		c.conn = nil
	}
}

func makeRequest[T any, E any](c *Commands, subject string, input T) (*E, error) {
	if c.show == nil || c.conn == nil {
		return nil, errors.New("show closed")
	}

	res, err := apibus.Request[T, E](c.conn, subject, input)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errors.New("no response returned")
	}

	if res.MessageError != nil {
		return nil, fmt.Errorf("error code %d: %s", res.MessageError.Code, res.MessageError.ErrorMessage)
	}

	return res.ResponseValue, nil
}
