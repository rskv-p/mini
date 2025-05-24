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

var (
	OnPanic func(service, function string, recovered any)
	log     logger.ILogger = logger.NewLogger("recover", "warn")
)

// SetLogger allows injecting a custom logger instance.
func SetLogger(l logger.ILogger) {
	log = l
}

// ----------------------------------------------------
// Panic recovery functions
// ----------------------------------------------------

// RecoverWithContext captures and logs a panic with metadata and optional data.
func RecoverWithContext(service, function string, data any) error {
	if r := recover(); r != nil {
		stack := string(debug.Stack())

		log.With(tagService, service).
			With(tagFunction, function).
			Error("panic: %v", r)

		if data != nil {
			log.With(tagContext, fmt.Sprintf("%+v", data)).Error("panic context")
		}

		log.Error("stacktrace:\n%s", stack)

		if OnPanic != nil {
			OnPanic(service, function, r)
		}

		return fmt.Errorf("panic recovered in %s.%s: %v", service, function, r)
	}
	return nil
}

// RecoverExplicit logs a known recovered panic with metadata and context.
func RecoverExplicit(service, function string, recovered any, data any) {
	if recovered == nil {
		return
	}

	stack := string(debug.Stack())

	log.With(tagService, service).
		With(tagFunction, function).
		Error("panic: %v", recovered)

	if data != nil {
		log.With(tagContext, fmt.Sprintf("%+v", data)).Error("panic context")
	}

	log.Error("stacktrace:\n%s", stack)

	if OnPanic != nil {
		OnPanic(service, function, recovered)
	}
}

// Safe runs the given function safely, recovering and logging any panic with label.
func Safe(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.With(tagLabel, label).Error("panic: %v", r)
			log.Error("stacktrace:\n%s", stack)
			if OnPanic != nil {
				OnPanic("Safe", label, r)
			}
		}
	}()
	fn()
}

// RecoverFunc is like Safe but returns error on panic.
func RecoverFunc(label string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.With(tagLabel, label).Error("panic: %v", r)
			log.Error("stacktrace:\n%s", stack)
			if OnPanic != nil {
				OnPanic("RecoverFunc", label, r)
			}
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn()
}

// ----------------------------------------------------
// Handler wrapper
// ----------------------------------------------------

// RecoverHandler wraps a router.Handler with panic recovery.
func RecoverHandler(service, function string, next router.Handler) router.Handler {
	return func(ctx context.Context, msg codec.IMessage, replyTo string) *router.Error {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				log.With(tagService, service).
					With(tagFunction, function).
					Error("panic: %v", r)
				log.With(tagContext, fmt.Sprintf("%+v", msg)).Error("panic context")
				log.Error("stacktrace:\n%s", stack)

				if OnPanic != nil {
					OnPanic(service, function, r)
				}
			}
		}()
		return next(ctx, msg, replyTo)
	}
}

// ----------------------------------------------------
// Universal context function wrapper
// ----------------------------------------------------

// RecoverableFunc is a context-aware function that may panic.
type RecoverableFunc func(ctx context.Context) error

// WrapRecover wraps a context-aware function with panic protection.
func WrapRecover(service, function string, f RecoverableFunc) RecoverableFunc {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())

				log.With(tagService, service).
					With(tagFunction, function).
					Error("panic: %v", r)
				log.Error("stacktrace:\n%s", stack)

				if OnPanic != nil {
					OnPanic(service, function, r)
				}

				err = fmt.Errorf("panic recovered in %s.%s: %v", service, function, r)
			}
		}()
		return f(ctx)
	}
}
