// file: mini/router/router_test.go
package router_test

import (
	"context"
	"testing"

	"github.com/rskv-p/mini/constant"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/router"
	"github.com/stretchr/testify/assert"
)

func newTestMessage(node string, body map[string]any) codec.IMessage {
	msg := codec.NewMessage("request")
	msg.SetNode(node)
	msg.SetContextID("test-ctx")
	msg.SetBody(body)
	return msg
}

var ErrCalled *router.Error

func TestRouterLifecycle(t *testing.T) {
	r := router.NewRouter(
		router.Name("test"),
		router.OnErrorHook(func(_ context.Context, _ codec.IMessage, err *router.Error) {
			ErrCalled = err
		}),
	)

	assert.NoError(t, r.Init())

	called := false
	r.Add(&router.Node{
		ID: "test.action",
		Handler: func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			called = true
			return nil
		},
	})

	msg := newTestMessage("test.action", nil)
	h, err := r.Dispatch(msg)
	assert.NoError(t, err)
	h(context.Background(), msg, "")

	assert.True(t, called)
	assert.Len(t, r.Routes(), 1)
	assert.Equal(t, "test", r.GetOptions().Name)

	assert.NoError(t, r.Register())
	assert.NoError(t, r.Deregister())
	assert.Empty(t, r.Routes())
}

func TestValidation(t *testing.T) {
	called := false
	r := router.NewRouter()

	r.Add(&router.Node{
		ID: "val.action",
		Handler: func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			called = true
			return nil
		},
		ValidationRules: map[string][]string{
			"name": {"required", "min:3", "max:5"},
		},
		ValidationMessages: map[string]string{
			"name.required": "name is required",
			"name.min":      "too short",
			"name.max":      "too long",
		},
	})

	tests := []struct {
		body     map[string]any
		wantErr  string
		skipCall bool
	}{
		{body: map[string]any{}, wantErr: "name is required", skipCall: true},
		{body: map[string]any{"name": "ab"}, wantErr: "too short", skipCall: true},
		{body: map[string]any{"name": "abcdef"}, wantErr: "too long", skipCall: true},
		{body: map[string]any{"name": "test"}, wantErr: "", skipCall: false},
	}

	for _, tt := range tests {
		msg := newTestMessage("val.action", tt.body)
		h, err := r.Dispatch(msg)
		assert.NoError(t, err)
		errResp := h(context.Background(), msg, "")
		if tt.wantErr != "" {
			assert.NotNil(t, errResp)
			assert.Contains(t, errResp.Message, tt.wantErr)
		} else {
			assert.Nil(t, errResp)
		}
	}
	assert.True(t, called)
}

func TestWrapMiddleware(t *testing.T) {
	order := []string{}
	w1 := func(next router.Handler) router.Handler {
		return func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			order = append(order, "w1")
			return next(ctx, msg, replyTo)
		}
	}
	w2 := func(next router.Handler) router.Handler {
		return func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			order = append(order, "w2")
			return next(ctx, msg, replyTo)
		}
	}

	h := func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
		order = append(order, "handler")
		return nil
	}

	wrapped := router.Wrap(h, []router.HandlerWrapper{w1, w2})
	wrapped(context.Background(), newTestMessage("x", nil), "")
	assert.Equal(t, []string{"w1", "w2", "handler"}, order)
}

func TestErrorHook(t *testing.T) {
	called := false
	r := router.NewRouter(
		router.OnErrorHook(func(ctx context.Context, msg codec.IMessage, err *router.Error) {
			called = true
			assert.Equal(t, 500, err.StatusCode)
		}),
	)

	r.Add(&router.Node{
		ID: "error.handler",
		Handler: func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			return &router.Error{StatusCode: 500, Message: "boom"}
		},
	})

	msg := newTestMessage("error.handler", nil)
	h, err := r.Dispatch(msg)
	assert.NoError(t, err)
	resp := h(context.Background(), msg, "")
	assert.NotNil(t, resp)
	assert.Equal(t, 500, resp.StatusCode)
	assert.True(t, called)
}

func TestNotFound(t *testing.T) {
	r := router.NewRouter(
		router.OnNotFound(func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
			return &router.Error{StatusCode: 404, Message: "missing"}
		}),
	)

	msg := newTestMessage("unknown.route", nil)
	h, err := r.Dispatch(msg)
	assert.NoError(t, err)
	errResp := h(context.Background(), msg, "")
	assert.Equal(t, 404, errResp.StatusCode)
	assert.Equal(t, "missing", errResp.Message)
}

func TestDispatchErrors(t *testing.T) {
	r := router.NewRouter()

	_, err := r.Dispatch(nil)
	assert.Equal(t, constant.ErrEmptyMessage, err)

	msg := newTestMessage("", nil)
	_, err = r.Dispatch(msg)
	assert.Equal(t, constant.ErrInvalidPath, err)
}
