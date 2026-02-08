package bus

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
)

// Message decouples the application logic from the underlying NATS library.
type Message struct {
	Subject string
	Reply   string // Required for Request/Response patterns
	Data    []byte
}

// Handler represents the logic to execute when a subject receives a message.
type Handler func(m Message)

// Client manages active NATS subscriptions and provides messaging primitives.
type Client struct {
	name   string
	conn   *nats.Conn
	mu     sync.Mutex
	subs   []*nats.Subscription
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient initializes a live client. It requires an active NATS connection.
func NewClient(name string, nc *nats.Conn) (*Client, error) {
	if nc == nil {
		return nil, errors.New("nats connection cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		name:   name,
		conn:   nc,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// --- Subscription Methods ---

// Subscribe adds a standard 1:N broadcast subscription.
func (c *Client) Subscribe(subject string, handler Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, err := c.conn.Subscribe(subject, c.wrap(handler))
	if err != nil {
		return fmt.Errorf("subscribe failed: %w", err)
	}

	c.configureSub(sub)
	return nil
}

func (c *Client) TemporarySubscription(subject string, handler Handler) (func() error, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, err := c.conn.Subscribe(subject, c.wrap(handler))
	if err != nil {
		return nil, fmt.Errorf("subscribe failed: %w", err)
	}

	c.configureSub(sub)
	return func() error { return sub.Unsubscribe() }, nil
}

// QueueSubscribe adds a load-balanced 1:1 worker subscription.
func (c *Client) QueueSubscribe(subject, queue string, handler Handler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, err := c.conn.QueueSubscribe(subject, queue, c.wrap(handler))
	if err != nil {
		return fmt.Errorf("queue subscribe failed: %w", err)
	}

	c.configureSub(sub)
	return nil
}

// --- Publishing & RPC Methods ---

// Send publishes a fire-and-forget message.
func (c *Client) Send(m Message) error {
	if err := c.conn.Publish(m.Subject, m.Data); err != nil {
		return fmt.Errorf("send failed: %w", err)
	}
	return c.conn.Flush()
}

// SendHelper wraps Send and logs errors on failures. Returns a boolean indicating whether the send was successful
func (c *Client) SendHelper(subject string, data []byte) bool {
	err := c.Send(Message{Subject: subject, Data: data})
	if err != nil {
		slog.Error("failed to send message", "subject", subject, "error", err)
		return false
	}
	return true
}

// Request sends a message and waits for a response within the context's deadline.
func (c *Client) Request(ctx context.Context, m Message) (Message, error) {
	resp, err := c.conn.RequestWithContext(ctx, m.Subject, m.Data)
	if err != nil {
		return Message{}, fmt.Errorf("request failed: %w", err)
	}

	return Message{
		Subject: resp.Subject,
		Reply:   resp.Reply,
		Data:    resp.Data,
	}, nil
}

// RequestHelper wraps Request and logs errors on failures. Returns a boolean indicating whether the request was successful
func (c *Client) RequestHelper(subject string, data []byte) (Message, bool) {
	resp, err := c.Request(c.Context(), Message{Subject: subject, Data: data})
	if err != nil {
		slog.Error("failed to send request message", "subject", subject, "error", err)
		return Message{}, false
	}
	return resp, true
}

// Respond sends a reply back to a requester.
func (c *Client) Respond(reply string, data []byte) error {
	if reply == "" {
		return errors.New("cannot respond: message has no reply subject")
	}

	if err := c.conn.Publish(reply, data); err != nil {
		return fmt.Errorf("respond failed: %w", err)
	}

	return c.conn.Flush()
}

// RespondHelper wraps Respond and logs errors on failures. Returns a boolean indicating whether the response was successful
func (c *Client) RespondHelper(original Message, responseData []byte) bool {
	if original.Reply == "" {
		return false
	}
	if err := c.Respond(original.Reply, responseData); err != nil {
		slog.Error("failed to send response message", "subject", original.Reply, "error", err)
		return false
	}
	return true
}

// --- Internal Helpers ---

// wrap provides panic recovery and converts nats.Msg to the local Message type.
func (c *Client) wrap(h Handler) nats.MsgHandler {
	return func(m *nats.Msg) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in bus handler",
					slog.String("client", c.name),
					slog.String("subject", m.Subject),
					slog.Any("error", r),
				)
			}
		}()
		h(Message{
			Subject: m.Subject,
			Reply:   m.Reply,
			Data:    m.Data,
		})
	}
}

// configureSub applies buffering limits and tracks the subscription for cleanup.
func (c *Client) configureSub(sub *nats.Subscription) {
	// 1k messages or 16MB buffer before being considered a "Slow Consumer"
	_ = sub.SetPendingLimits(1000, 1024*1024*16)
	c.subs = append(c.subs, sub)
}

// Stop gracefully drains all subscriptions and cancels the internal context.
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cancel()

	var lastErr error
	for _, sub := range c.subs {
		if err := sub.Drain(); err != nil {
			lastErr = err
		}
	}

	c.subs = nil
	return lastErr
}

// Context returns the client's lifecycle context.
func (c *Client) Context() context.Context {
	return c.ctx
}
