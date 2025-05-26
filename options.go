// file: mini/options.go
package service

import (
	"time"

	"github.com/rskv-p/mini/codec"
	"github.com/rskv-p/mini/context"
	"github.com/rskv-p/mini/logger"
	"github.com/rskv-p/mini/registry"
	"github.com/rskv-p/mini/router"
	"github.com/rskv-p/mini/selector"
	"github.com/rskv-p/mini/transport"
)

// ----------------------------------------------------
// Handler types and wrappers
// ----------------------------------------------------

type HandlerWrapper = router.HandlerWrapper

// ----------------------------------------------------
// Retry and lifecycle hooks
// ----------------------------------------------------

type RetryConfig struct {
	Count    int
	Interval time.Duration
}

type Hooks struct {
	OnInit     func()
	OnStart    func()
	OnStop     func()
	OnError    func(error)
	OnMessage  func(codec.IMessage)
	OnShutdown func()
}

// ----------------------------------------------------
// Options struct for IService
// ----------------------------------------------------

type Options struct {
	Transport transport.ITransport
	Registry  registry.IRegistry
	Router    router.IRouter
	Context   context.IContext
	Selector  selector.ISelector
	Logger    logger.ILogger
	Retry     RetryConfig
	Hooks     Hooks

	HdlrWrappers []HandlerWrapper
	Debug        bool
}

// Option defines a configuration function.
type Option func(*Options)

// ----------------------------------------------------
// Default builder
// ----------------------------------------------------

func newOptions(opts ...Option) Options {
	o := Options{
		Context: context.NewContext(),
		Retry: RetryConfig{
			Count:    3,
			Interval: 100 * time.Millisecond,
		},
	}

	for _, opt := range opts {
		opt(&o)
	}

	// Set safe defaults
	if o.Registry == nil {
		o.Registry = registry.NewRegistry()
	}
	if o.Router == nil {
		o.Router = router.NewRouter()
	}
	if o.Transport == nil {
		o.Transport = transport.New()
	}
	if o.Selector == nil {
		o.Selector = selector.NewSelector(o.Registry, selector.SetStrategy(selector.RoundRobin))
	}
	if o.Logger == nil {
		o.Logger = logger.NewLogger("service", "info")
	}

	return o
}

// ----------------------------------------------------
// Option constructors
// ----------------------------------------------------

func Transport(t transport.ITransport) Option {
	return func(o *Options) { o.Transport = t }
}

func Registry(r registry.IRegistry) Option {
	return func(o *Options) { o.Registry = r }
}

func Router(r router.IRouter) Option {
	return func(o *Options) { o.Router = r }
}

func Context(c context.IContext) Option {
	return func(o *Options) { o.Context = c }
}

func Selector(s selector.ISelector) Option {
	return func(o *Options) { o.Selector = s }
}

func Logger(l logger.ILogger) Option {
	return func(o *Options) { o.Logger = l }
}

func WithRetry(count int, interval time.Duration) Option {
	return func(o *Options) {
		o.Retry.Count = count
		o.Retry.Interval = interval
	}
}

func WithHooks(h Hooks) Option {
	return func(o *Options) { o.Hooks = h }
}

func WrapHandler(w ...HandlerWrapper) Option {
	return func(o *Options) {
		o.HdlrWrappers = append(o.HdlrWrappers, w...)
	}
}

func EnableDebug() Option {
	return func(o *Options) { o.Debug = true }
}

// ----------------------------------------------------
// Utility methods
// ----------------------------------------------------

func (o *Options) Clone() Options {
	c := *o
	c.HdlrWrappers = append([]HandlerWrapper{}, o.HdlrWrappers...)
	c.Retry = o.Retry
	c.Hooks = o.Hooks
	return c
}

func (o *Options) Validate() error {
	if o.Transport == nil {
		return ErrMissing("Transport")
	}
	if o.Registry == nil {
		return ErrMissing("Registry")
	}
	if o.Router == nil {
		return ErrMissing("Router")
	}
	if o.Selector == nil {
		return ErrMissing("Selector")
	}
	if o.Context == nil {
		return ErrMissing("Context")
	}
	return nil
}

// ----------------------------------------------------
// Error type for missing dependencies
// ----------------------------------------------------

func ErrMissing(name string) error {
	return &MissingDependencyError{name}
}

type MissingDependencyError struct {
	Dependency string
}

func (e *MissingDependencyError) Error() string {
	return "missing dependency: " + e.Dependency
}
