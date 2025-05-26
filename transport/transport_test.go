package transport

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------
// Mock connection (implements IConn)
// ----------------------------------------------------

type mockConn struct {
	subscribed map[string]bool
	mu         sync.Mutex
	closed     bool
}

func (m *mockConn) Publish(subject string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrDisconnected
	}
	return nil
}

func (m *mockConn) Request(subject string, data []byte, timeout time.Duration) (codec.IMessage, error) {
	return codec.NewMessage(""), nil
}

func (m *mockConn) Subscribe(subject string, handler MsgHandler) (*Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.subscribed == nil {
		m.subscribed = make(map[string]bool)
	}
	m.subscribed[subject] = true
	return &Subscription{topic: subject, channel: "test"}, nil
}

func (m *mockConn) SubscribeOnce(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error) {
	return m.Subscribe(subject, handler)
}

func (m *mockConn) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
}

func (m *mockConn) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.closed
}

func (m *mockConn) Ping() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrDisconnected
	}
	return nil
}

func (m *mockConn) ListTopics() ([]string, error) {
	return []string{"foo.a", "foo.b", "bar.x"}, nil
}

// ----------------------------------------------------
// Tests
// ----------------------------------------------------

func TestTransport_Lifecycle(t *testing.T) {
	tr := New()
	tr.conn = &mockConn{}

	assert.NoError(t, tr.Ping())
	assert.True(t, tr.IsConnected())
	assert.NoError(t, tr.Health())
	assert.NoError(t, tr.Close())
}

func TestTransport_Subscribe(t *testing.T) {
	tr := New()
	tr.conn = &mockConn{}
	tr.SetHandler(func([]byte) error { return nil })
	assert.NoError(t, tr.Subscribe())
}

func TestTransport_SubscribeTopic(t *testing.T) {
	tr := New()
	tr.conn = &mockConn{}
	assert.NoError(t, tr.SubscribeTopic("custom.topic", func([]byte) error { return nil }))
}

func TestTransport_SubscribePrefix(t *testing.T) {
	mock := &mockConn{}
	tr := New()
	tr.conn = mock

	tr.Use(func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			return next(ctx, subject, data)
		}
	})

	err := tr.SubscribePrefix("foo.", func([]byte) error { return nil })
	assert.NoError(t, err)

	mock.mu.Lock()
	defer mock.mu.Unlock()
	assert.True(t, mock.subscribed["foo.a"])
	assert.True(t, mock.subscribed["foo.b"])
	assert.False(t, mock.subscribed["bar.x"])
}

func TestTransport_Unsubscribe(t *testing.T) {
	tr := New()
	tr.conn = &mockConn{}
	tr.SetHandler(func([]byte) error { return nil })
	assert.NoError(t, tr.Subscribe())
	assert.NoError(t, tr.Unsubscribe())
}

func TestTransport_wrap(t *testing.T) {
	tr := New()
	var sequence []string

	tr.Use(func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			sequence = append(sequence, "mw1")
			return next(ctx, subject, data)
		}
	})
	tr.Use(func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			sequence = append(sequence, "mw2")
			return next(ctx, subject, data)
		}
	})

	handler := tr.wrap(func(ctx context.Context, subject string, data []byte) error {
		sequence = append(sequence, "handler")
		return nil
	})

	validData := []byte(`{"type":"test"}`)

	err := handler(context.Background(), "test.topic", validData)
	assert.NoError(t, err)
	assert.Equal(t, []string{"mw1", "mw2", "handler"}, sequence)
}

func TestTransport_HealthFailures(t *testing.T) {
	tr := New()
	mock := &mockConn{}
	mock.closed = true
	tr.conn = mock

	assert.Error(t, tr.Ping())
	assert.Error(t, tr.Health())
}
