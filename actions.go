// file: mini/actions.go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/router"
)

type contextKey string

const ContextIDKey contextKey = "contextID"

func ContextIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value(ContextIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ----------------------------------------------------
// Action interface and types
// ----------------------------------------------------

type ActionFunc func(ctx context.Context, input map[string]any) (any, error)

type Middleware func(ActionFunc) ActionFunc

type InputSchemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type IAction interface {
	Name() string
	Schema() []InputSchemaField
	Handle(ctx context.Context, input map[string]any) (any, error)
}

type actionInfo struct {
	schema  []InputSchemaField
	handler ActionFunc
}

// ----------------------------------------------------
// Registration
// ----------------------------------------------------

func (s *Service) RegisterAction(name string, schema []InputSchemaField, fn ActionFunc) {
	if s.actions == nil {
		s.actions = make(map[string]actionInfo)
	}
	s.actions[name] = actionInfo{schema: schema, handler: fn}
}

func (s *Service) RegisterActions(actions ...IAction) {
	for _, a := range actions {
		s.RegisterAction(a.Name(), a.Schema(), a.Handle)
	}
}

func (s *Service) Use(mw Middleware) {
	s.middlewares = append(s.middlewares, mw)
}

func (s *Service) ListActions() []string {
	keys := make([]string, 0, len(s.actions))
	for k := range s.actions {
		keys = append(keys, k)
	}
	return keys
}

func (s *Service) GetSchemas() map[string][]InputSchemaField {
	out := make(map[string][]InputSchemaField, len(s.actions))
	for k, v := range s.actions {
		out[k] = v.schema
	}
	return out
}

func (s *Service) ActionSchema(name string) ([]InputSchemaField, bool) {
	info, ok := s.actions[name]
	return info.schema, ok
}

func (s *Service) GetOpenAPISchemas() map[string]any {
	schemas := make(map[string]any, len(s.actions))
	for name, info := range s.actions {
		required := []string{}
		properties := make(map[string]any)
		for _, f := range info.schema {
			properties[f.Name] = map[string]string{"type": f.Type}
			if f.Required {
				required = append(required, f.Name)
			}
		}
		schema := map[string]any{"type": "object", "properties": properties}
		if len(required) > 0 {
			schema["required"] = required
		}
		schemas[name] = schema
	}
	return schemas
}

// ----------------------------------------------------
// Middleware chaining
// ----------------------------------------------------

func chainMiddlewares(fn ActionFunc, mws ...Middleware) ActionFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		fn = mws[i](fn)
	}
	return fn
}

// ----------------------------------------------------
// prepareHandler — action → router.Handler
// ----------------------------------------------------

func (s *Service) prepareHandler(fn ActionFunc) router.Handler {
	return func(ctx context.Context, raw codec.IMessage, replyTo string) *router.Error {
		ctxID := raw.GetContextID()
		if ctxID != "" {
			ctx = context.WithValue(ctx, ContextIDKey, ctxID)
		}

		node := raw.GetNode()
		body := raw.GetBodyMap()

		// validate schema
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

		// apply middleware
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
// Utils
// ----------------------------------------------------

func isEmpty(v any) bool {
	switch x := v.(type) {
	case string:
		return x == ""
	case []byte:
		return len(x) == 0
	case nil:
		return true
	default:
		return false
	}
}
