// file: mini/transport/io_test.go
package transport

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/stretchr/testify/assert"
)

type mockIConn struct {
	failPublish bool
	failRequest bool
}

func (m *mockIConn) Publish(subject string, data []byte) error {
	if m.failPublish {
		return errors.New("publish error")
	}
	return nil
}

func (m *mockIConn) Request(subject string, data []byte, timeout time.Duration) (codec.IMessage, error) {
	if m.failRequest {
		return nil, errors.New("request error")
	}
	msg := codec.NewMessage("")
	msg.Set("ok", true)
	return msg, nil
}

func (m *mockIConn) Subscribe(string, MsgHandler) (*Subscription, error) { return nil, nil }
func (m *mockIConn) SubscribeOnce(string, MsgHandler, time.Duration) (*Subscription, error) {
	return nil, nil
}
func (m *mockIConn) Close()            {}
func (m *mockIConn) IsConnected() bool { return true }
func (m *mockIConn) Ping() error       { return nil }
func (m *mockIConn) ListTopics() ([]string, error) {
	return []string{"a", "b"}, nil
}

func TestRequestWithContext_Success(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{}
	tr.opts.Timeout = time.Second

	called := false
	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)

	err := tr.RequestWithContext(context.Background(), "service.test", data, func(m codec.IMessage) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRequestWithContext_Failure(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{failRequest: true}
	tr.opts.Timeout = time.Second

	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)
	err := tr.RequestWithContext(context.Background(), "service.test", data, nil)
	assert.Error(t, err)
}

func TestPublish_Success(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{}
	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)

	err := tr.Publish("topic.test", data)
	assert.NoError(t, err)
}

func TestPublish_Error(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{failPublish: true}
	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)

	err := tr.Publish("topic.test", data)
	assert.Error(t, err)
}

func TestRespond(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{}
	msg := codec.NewMessage("")
	err := tr.Respond("reply.to", msg)
	assert.NoError(t, err)
}

func TestBroadcast(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{}
	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)
	err := tr.Broadcast([]string{"a", "b"}, data)
	assert.NoError(t, err)
}

func TestBroadcast_Error(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{failPublish: true}
	msg := codec.NewMessage("")
	data, _ := codec.Marshal(msg)
	err := tr.Broadcast([]string{"a"}, data)
	assert.Error(t, err)
}

func TestBroadcastf(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{}
	err := tr.Broadcastf("prefix.%s", "1", "2")
	assert.NoError(t, err)
}

func TestBroadcastf_Error(t *testing.T) {
	tr := New()
	tr.conn = &mockIConn{failPublish: true}
	err := tr.Broadcastf("prefix.%s", "1")
	assert.Error(t, err)
}

func Test_wrapChain(t *testing.T) {
	tr := New()
	called := false

	tr.Use(func(next TransportHandler) TransportHandler {
		return func(ctx context.Context, subject string, data []byte) error {
			called = true
			return next(ctx, subject, data)
		}
	})

	fn := func(subject string, data []byte) error { return nil }
	wrapped := tr.wrapChain(fn)
	err := wrapped("topic", []byte(`{}`))

	assert.NoError(t, err)
	assert.True(t, called)
}
