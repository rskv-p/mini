// file: mini/recover/recover.go
package recover

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/logger"
	"github.com/rskv-p/mini/router"
)

const (
	tagService  = "service"
	tagFunction = "function"
	tagContext  = "context"
	tagLabel    = "label"
)

// ----------------------------------------------------
// Global panic hook (optional)
// ----------------------------------------------------

var OnPanic func(service, function string, recovered any)
var log logger.ILogger = logger.NewLogger("recover", "warn")

// SetLogger allows injecting a custom logger instance (e.g. for tracing or testing).
func SetLogger(l logger.ILogger) {
	log = l
}

// ----------------------------------------------------
// Panic recovery functions
// ----------------------------------------------------

// RecoverWithContext captures and logs a panic with metadata and optional data.
func RecoverWithContext(service, function string, data any) {
	if r := recover(); r != nil {
		log.With(tagService, service).
			With(tagFunction, function).
			Error("panic: %v", r)

		if data != nil {
			log.With(tagContext, fmt.Sprintf("%+v", data)).Error("panic context")
		}

		log.Error("stacktrace:\n%s", string(debug.Stack()))

		if OnPanic != nil {
			OnPanic(service, function, r)
		}
	}
}

// RecoverExplicit logs a known recovered panic with metadata and context.
func RecoverExplicit(service, function string, recovered any, data any) {
	if recovered == nil {
		return
	}

	log.With(tagService, service).
		With(tagFunction, function).
		Error("panic: %v", recovered)

	if data != nil {
		log.With(tagContext, fmt.Sprintf("%+v", data)).Error("panic context")
	}

	log.Error("stacktrace:\n%s", string(debug.Stack()))

	if OnPanic != nil {
		OnPanic(service, function, recovered)
	}
}

// Safe runs the given function safely, recovering and logging any panic with label.
func Safe(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.With(tagLabel, label).Error("panic: %v", r)
			log.Error("stacktrace:\n%s", string(debug.Stack()))
			if OnPanic != nil {
				OnPanic("Safe", label, r)
			}
		}
	}()
	fn()
}

// ----------------------------------------------------
// Handler wrapper
// ----------------------------------------------------

// RecoverHandler wraps a router.Handler with panic recovery.
func RecoverHandler(service, function string, next router.Handler) router.Handler {
	return func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
		defer RecoverWithContext(service, function, msg)
		return next(ctx, msg, replyTo)
	}
}

// ----------------------------------------------------
// Universal wrapper
// ----------------------------------------------------

// RecoverableFunc is a context-aware function that may panic.
type RecoverableFunc func(ctx context.Context) error

// WrapRecover wraps a context-aware function with panic protection.
func WrapRecover(service, function string, f RecoverableFunc) RecoverableFunc {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				RecoverWithContext(service, function, nil)
				err = fmt.Errorf("panic recovered in %s.%s: %v", service, function, r)
			}
		}()
		return f(ctx)
	}
}
