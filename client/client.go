package client

import (
	"github.com/nats-io/nats.go"
	"github.com/stexxo/dynocue/core/logging"
	"github.com/stexxo/dynocue/core/messaging"
)

type Client struct {
	messenger *messaging.Messenger
}

func NewClient(conn *nats.Conn, logger logging.Logger) *Client {
	return &Client{
		messenger: messaging.NewMessenger(&messaging.MessengerCfg{
			Conn:   conn,
			Logger: logger,
		}),
	}
}
