// file: arc/service/router/router_options.go
package router

import (
	"context"

	"github.com/rskv-p/mini/service/codec"
)

// ----------------------------------------------------
// Router options and hooks
// ----------------------------------------------------

// Options configures router behavior.
type Options struct {
	Name     string           // logical router name
	NotFound RejectHandler    // called when route is missing
	OnError  ErrorHook        // called after a handler returns an error
	Wrappers []HandlerWrapper // global middleware (applied outermost last)
}

// Option applies a configuration mutation to Options.
type Option func(*Options)

// RejectHandler is invoked when no route is found for a given message.
type RejectHandler func(ctx context.Context, msg codec.IMessage, replyTo string) *Error

// ErrorHook is called after a Handler returns an error.
type ErrorHook func(ctx context.Context, msg codec.IMessage, err *Error)

// ----------------------------------------------------
// Option builders
// ----------------------------------------------------

// Name sets a logical name for the router (for introspection/logs).
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

// OnErrorHook registers a hook triggered after handler error.
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

// WithDefaults returns a set of safe default options.
func WithDefaults() Options {
	return Options{
		Name:     "default",
		Wrappers: nil,
	}
}
