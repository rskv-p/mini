package transport_test

import (
	"context"
	"testing"
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/transport"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	// prepare raw message without trace_id/context_id
	msg := codec.NewMessage("request")
	msg.Set("someKey", "someValue")
	data, err := codec.Marshal(msg)
	assert.NoError(t, err)

	called := false

	// wrap handler with trace middleware
	handler := transport.TraceMiddleware()(func(ctx context.Context, subject string, data []byte) error {
		called = true
		// ensure fields injected
		decoded := codec.NewMessage("")
		err := codec.Unmarshal(data, decoded)
		assert.NoError(t, err)

		traceID := decoded.GetString("trace_id")
		ctxID := decoded.GetContextID()

		assert.NotEmpty(t, traceID)
		assert.NotEmpty(t, ctxID)
		assert.Equal(t, "someValue", decoded.GetString("someKey"))
		return nil
	})

	err = handler(context.Background(), "test.trace", data)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestLoggerMiddleware(t *testing.T) {
	// prepare dummy message
	msg := codec.NewMessage("request")
	msg.Set("key", "value")
	data, err := codec.Marshal(msg)
	assert.NoError(t, err)

	// wrap handler with logger middleware
	called := false
	handler := transport.LoggerMiddleware()(func(ctx context.Context, subject string, data []byte) error {
		called = true
		time.Sleep(10 * time.Millisecond) // simulate delay
		return nil
	})

	err = handler(context.Background(), "test.logger", data)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestLoggerMiddleware_Error(t *testing.T) {
	// prepare dummy message
	msg := codec.NewMessage("request")
	data, _ := codec.Marshal(msg)

	// simulate failing handler
	handler := transport.LoggerMiddleware()(func(ctx context.Context, subject string, data []byte) error {
		return assert.AnError
	})

	err := handler(context.Background(), "test.logger.err", data)
	assert.Equal(t, assert.AnError, err)
}
