package recover_test

import (
	"context"
	"testing"

	"github.com/rskv-p/mini/codec"
	recoverpkg "github.com/rskv-p/mini/recover"
	"github.com/rskv-p/mini/router"
	"github.com/stretchr/testify/assert"
)

var panicHookTriggered bool
var panicCapturedService, panicCapturedFunc string
var panicCapturedValue any

func TestMain(m *testing.M) {
	recoverpkg.OnPanic = func(service, fn string, r any) {
		panicHookTriggered = true
		panicCapturedService = service
		panicCapturedFunc = fn
		panicCapturedValue = r
	}
	m.Run()
}

func TestRecoverWithContext(t *testing.T) {
	panicHookTriggered = false

	func() {
		defer recoverpkg.RecoverWithContext("svc", "fn", map[string]any{"meta": 1})
		panic("test1")
	}()
	assert.True(t, panicHookTriggered)
	assert.Equal(t, "svc", panicCapturedService)
	assert.Equal(t, "fn", panicCapturedFunc)
	assert.Equal(t, "test1", panicCapturedValue)
}

func TestRecoverExplicit(t *testing.T) {
	panicHookTriggered = false
	recoverpkg.RecoverExplicit("svc2", "fn2", "manual", map[string]any{"meta": true})

	assert.True(t, panicHookTriggered)
	assert.Equal(t, "svc2", panicCapturedService)
	assert.Equal(t, "fn2", panicCapturedFunc)
	assert.Equal(t, "manual", panicCapturedValue)
}

func TestSafe(t *testing.T) {
	panicHookTriggered = false
	recoverpkg.Safe("my-safe", func() {
		panic("in safe")
	})
	assert.True(t, panicHookTriggered)
	assert.Equal(t, "Safe", panicCapturedService)
	assert.Equal(t, "my-safe", panicCapturedFunc)
}

func TestRecoverFunc_NoPanic(t *testing.T) {
	err := recoverpkg.RecoverFunc("no-panic", func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestRecoverFunc_WithPanic(t *testing.T) {
	err := recoverpkg.RecoverFunc("with-panic", func() error {
		panic("boom")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic: boom")
}

func TestWrapRecover_NoPanic(t *testing.T) {
	fn := recoverpkg.WrapRecover("service", "fn", func(ctx context.Context) error {
		return nil
	})
	err := fn(context.Background())
	assert.NoError(t, err)
}

func TestWrapRecover_WithPanic(t *testing.T) {
	fn := recoverpkg.WrapRecover("svcX", "fnX", func(ctx context.Context) error {
		panic("ctx-panic")
	})
	err := fn(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered in svcX.fnX")
}

func TestRecoverHandler_NoPanic(t *testing.T) {
	called := false
	h := recoverpkg.RecoverHandler("svc", "fn", func(ctx context.Context, msg codec.IMessage, reply string) *router.Error {
		called = true
		return nil
	})
	err := h(context.Background(), codec.NewMessage(""), "reply.topic")
	assert.True(t, called)
	assert.Nil(t, err)
}

func TestRecoverHandler_WithPanic(t *testing.T) {
	called := false
	h := recoverpkg.RecoverHandler("svc", "fn", func(ctx context.Context, msg codec.IMessage, reply string) *router.Error {
		called = true
		panic("router boom")
	})
	_ = h(context.Background(), codec.NewMessage(""), "reply.topic")
	assert.True(t, called)
}
