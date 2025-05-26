// file: mini/router/router_options_test.go
package router_test

import (
	"context"
	"testing"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/logger"
	"github.com/rskv-p/mini/router"
	"github.com/stretchr/testify/assert"
)

// testContextKey is a private type to avoid context key collisions (SA1029).
type testContextKey struct{}

func TestOptionsDefaults(t *testing.T) {
	opts := router.WithDefaults()

	assert.Equal(t, "default", opts.Name)
	assert.Nil(t, opts.NotFound)
	assert.Nil(t, opts.OnError)
	assert.Nil(t, opts.ContextDecorator)
	assert.Nil(t, opts.Logger)
	assert.Nil(t, opts.Wrappers)
}

func TestOptionBuilders(t *testing.T) {
	var calledNotFound bool
	var calledErrorHook bool
	var calledContextDecorator bool
	var capturedCtx context.Context

	mockLogger := logger.NewLogger("test", "debug")

	opts := router.WithDefaults()

	// Apply all Option builders
	options := []router.Option{
		router.Name("custom"),
		router.OnNotFound(func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			calledNotFound = true
			return &router.Error{StatusCode: 404, Message: "not found"}
		}),
		router.OnErrorHook(func(ctx context.Context, msg codec.IMessage, err *router.Error) {
			calledErrorHook = true
		}),
		router.WithLogger(mockLogger),
		router.WithContextDecorator(func(ctx context.Context, msg codec.IMessage) context.Context {
			calledContextDecorator = true
			capturedCtx = context.WithValue(ctx, testContextKey{}, true)
			return capturedCtx
		}),
		router.UseMiddleware(
			func(next router.Handler) router.Handler {
				return func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
					return next(ctx, msg, replyTo)
				}
			},
		),
	}

	for _, o := range options {
		o(&opts)
	}

	assert.Equal(t, "custom", opts.Name)
	assert.NotNil(t, opts.NotFound)
	assert.NotNil(t, opts.OnError)
	assert.NotNil(t, opts.Logger)
	assert.Len(t, opts.Wrappers, 1)

	// Test NotFound
	err := opts.NotFound(context.Background(), codec.NewMessage("request"), "reply.to")
	assert.True(t, calledNotFound)
	assert.Equal(t, 404, err.StatusCode)

	// Test OnError
	opts.OnError(context.Background(), codec.NewMessage("request"), &router.Error{StatusCode: 500})
	assert.True(t, calledErrorHook)

	// Test ContextDecorator
	ctx := opts.ContextDecorator(context.Background(), codec.NewMessage("request"))
	assert.True(t, calledContextDecorator)
	assert.Equal(t, true, ctx.Value(testContextKey{}))
}
