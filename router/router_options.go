// file: mini/router/router_options.go
package router

import (
	"context"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/logger"
)

// ----------------------------------------------------
// Router options and hooks
// ----------------------------------------------------

// Options configures router behavior.
type Options struct {
	Name             string           // Logical router name (e.g. "auth/v1")
	NotFound         RejectHandler    // Called when route is missing
	OnError          ErrorHook        // Called after a handler returns an error
	Wrappers         []HandlerWrapper // Global middleware (outermost first)
	ContextDecorator ContextDecorator // Optional context enrichment
	Logger           logger.ILogger   // Optional structured logger
}

// Option applies a configuration mutation to Options.
type Option func(*Options)

// RejectHandler is invoked when no route is found for a given message.
type RejectHandler func(ctx context.Context, msg codec.IMessage, replyTo string) *Error

// ErrorHook is called after a Handler returns an error.
type ErrorHook func(ctx context.Context, msg codec.IMessage, err *Error)

// ContextDecorator allows enriching context before calling handler.
type ContextDecorator func(context.Context, codec.IMessage) context.Context

// ----------------------------------------------------
// Option builders
// ----------------------------------------------------

// Name sets a logical name for the router (used in logs/debug).
func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// OnNotFound sets the fallback handler when a route is not found.
func OnNotFound(h RejectHandler) Option {
	return func(o *Options) {
		o.NotFound = h
	}
}

// OnErrorHook registers a hook triggered after a handler returns error.
func OnErrorHook(h ErrorHook) Option {
	return func(o *Options) {
		o.OnError = h
	}
}

// UseMiddleware appends middleware to the global handler chain.
// Wrappers are applied in reverse order: outermost first.
func UseMiddleware(wrappers ...HandlerWrapper) Option {
	return func(o *Options) {
		o.Wrappers = append(o.Wrappers, wrappers...)
	}
}

// WithLogger sets a structured logger for the router.
func WithLogger(l logger.ILogger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithContextDecorator sets a function to mutate context before handler execution.
func WithContextDecorator(fn ContextDecorator) Option {
	return func(o *Options) {
		o.ContextDecorator = fn
	}
}

// WithDefaults returns safe default options.
func WithDefaults() Options {
	return Options{
		Name:     "default",
		Wrappers: nil,
	}
}
