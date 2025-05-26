// file: service/service_test.go
package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/logger"
	"github.com/rskv-p/mini/router"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------
// Mocks for logger and service
// ----------------------------------------------------

type testLogger struct {
	fields map[string]any
}

func (l *testLogger) Debug(msg string, args ...any) {}
func (l *testLogger) Info(msg string, args ...any)  {}
func (l *testLogger) Warn(msg string, args ...any)  {}
func (l *testLogger) Error(msg string, args ...any) {}
func (l *testLogger) WithContext(contextID string) logger.ILogger {
	return l
}
func (l *testLogger) With(key string, value any) logger.LoggerEntry {
	if l.fields == nil {
		l.fields = make(map[string]any)
	}
	l.fields[key] = value
	return &testLoggerEntry{fields: l.fields}
}
func (l *testLogger) SetLevel(level string) {}
func (l *testLogger) Clone() logger.ILogger { return l }

type testLoggerEntry struct {
	fields map[string]any
}

func (e *testLoggerEntry) Debug(msg string, args ...any) {}
func (e *testLoggerEntry) Info(msg string, args ...any)  {}
func (e *testLoggerEntry) Warn(msg string, args ...any)  {}
func (e *testLoggerEntry) Error(msg string, args ...any) {}
func (e *testLoggerEntry) With(key string, value any) logger.LoggerEntry {
	if e.fields == nil {
		e.fields = make(map[string]any)
	}
	e.fields[key] = value
	return e
}
func (e *testLoggerEntry) Clone() logger.LoggerEntry {
	return e
}

var _ logger.ILogger = (*testLogger)(nil)
var _ logger.LoggerEntry = (*testLoggerEntry)(nil)

// ----------------------------------------------------
// Test service with in-memory response recorder
// ----------------------------------------------------

type testService struct {
	Service
	Responses []codec.IMessage
}

func newTestService() *testService {
	return &testService{
		Service: Service{
			logger:  &testLogger{},
			actions: make(map[string]actionInfo),
		},
	}
}

// Append response to internal buffer (no-op reply)
func (s *testService) Respond(msg codec.IMessage, replyTo string) error {
	s.Responses = append(s.Responses, msg)
	return nil
}

// Replaces prepareHandler with mock Respond
func (s *testService) prepareHandler(fn ActionFunc) func(ctx context.Context, raw codec.IMessage, replyTo string) *router.Error {
	return func(ctx context.Context, raw codec.IMessage, replyTo string) *router.Error {
		ctxID := raw.GetContextID()
		if ctxID != "" {
			ctx = context.WithValue(ctx, ContextIDKey, ctxID)
		}
		node := raw.GetNode()
		body := raw.GetBodyMap()

		info, ok := s.actions[node]
		if ok && len(info.schema) > 0 {
			for _, f := range info.schema {
				v, exists := body[f.Name]
				if f.Required && (!exists || v == nil || isEmpty(v)) {
					msg := fmt.Sprintf("missing required field: %s", f.Name)
					s.logger.WithContext(ctxID).Warn(msg)

					resp := codec.NewJsonResponse(ctxID, 400)
					resp.SetError(errors.New(msg))
					_ = s.Respond(resp, replyTo)
					return &router.Error{StatusCode: 400, Message: msg}
				}
			}
		}

		wrapped := chainMiddlewares(fn, s.middlewares...)

		defer func() {
			if r := recover(); r != nil {
				s.logger.WithContext(ctxID).Error("panic in action: %v", r)
				resp := codec.NewJsonResponse(ctxID, 500)
				resp.SetError(fmt.Errorf("internal error"))
				_ = s.Respond(resp, replyTo)
			}
		}()

		result, err := wrapped(ctx, body)
		status := 200
		if err != nil {
			status = 500
		}
		resp := codec.NewJsonResponse(ctxID, status)

		if err != nil {
			s.logger.WithContext(ctxID).Error("action error: %v", err)
			resp.SetError(err)
			_ = s.Respond(resp, replyTo)
			return &router.Error{StatusCode: status, Message: err.Error()}
		}

		resp.SetResult(result)
		_ = s.Respond(resp, replyTo)
		return nil
	}
}

// ----------------------------------------------------
// Unit tests
// ----------------------------------------------------

func TestRegisterAndCallAction(t *testing.T) {
	s := newTestService()
	called := false

	s.RegisterAction("test.echo", nil, func(ctx context.Context, input map[string]any) (any, error) {
		called = true
		return input["msg"], nil
	})

	assert.Contains(t, s.ListActions(), "test.echo")
	assert.NotEmpty(t, s.GetSchemas())

	msg := codec.NewMessage("")
	msg.SetContextID("ctx-123")
	msg.Set("msg", "hi")
	msg.SetNode("test.echo")

	h := s.prepareHandler(s.actions["test.echo"].handler)
	err := h(context.Background(), msg, "")
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestMiddlewareExecution(t *testing.T) {
	s := newTestService()
	order := []string{}

	s.Use(func(next ActionFunc) ActionFunc {
		return func(ctx context.Context, in map[string]any) (any, error) {
			order = append(order, "mw1")
			return next(ctx, in)
		}
	})
	s.Use(func(next ActionFunc) ActionFunc {
		return func(ctx context.Context, in map[string]any) (any, error) {
			order = append(order, "mw2")
			return next(ctx, in)
		}
	})

	s.RegisterAction("test.order", nil, func(ctx context.Context, input map[string]any) (any, error) {
		order = append(order, "handler")
		return nil, nil
	})

	msg := codec.NewMessage("")
	msg.SetContextID("ctx-mw")
	msg.SetNode("test.order")

	err := s.prepareHandler(s.actions["test.order"].handler)(context.Background(), msg, "")
	assert.Nil(t, err)
	assert.Equal(t, []string{"mw1", "mw2", "handler"}, order)
}

func TestValidationFailure(t *testing.T) {
	s := newTestService()

	s.RegisterAction("test.required", []InputSchemaField{
		{Name: "foo", Type: "string", Required: true},
	}, func(ctx context.Context, in map[string]any) (any, error) {
		return "ok", nil
	})

	msg := codec.NewMessage("")
	msg.SetContextID("ctx-val")
	msg.SetNode("test.required")

	err := s.prepareHandler(s.actions["test.required"].handler)(context.Background(), msg, "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Message, "missing required field")
}

func TestActionSchemaOpenAPI(t *testing.T) {
	s := newTestService()

	s.RegisterAction("schema.test", []InputSchemaField{
		{Name: "foo", Type: "string", Required: true},
		{Name: "bar", Type: "int", Required: false},
	}, func(ctx context.Context, in map[string]any) (any, error) {
		return "ok", nil
	})

	schemas := s.GetOpenAPISchemas()
	assert.Contains(t, schemas, "schema.test")
	obj := schemas["schema.test"].(map[string]any)
	assert.Equal(t, "object", obj["type"])
	assert.Contains(t, obj["properties"].(map[string]any), "foo")
	assert.Contains(t, obj["required"], "foo")
}

func TestIsEmpty(t *testing.T) {
	assert.True(t, isEmpty(nil))
	assert.True(t, isEmpty(""))
	assert.True(t, isEmpty([]byte{}))
	assert.False(t, isEmpty("non-empty"))
	assert.False(t, isEmpty([]byte("x")))
}
