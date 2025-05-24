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

// Old: Handler func(*codec.Message, string) *router.Error — удалено
// Используем router.Handler напрямую через router.Router

type HandlerWrapper = router.HandlerWrapper

// ----------------------------------------------------
// Retry and Hooks
// ----------------------------------------------------

type RetryConfig struct {
	Count    int
	Interval time.Duration
}

type Hooks struct {
	OnStart    func()
	OnStop     func()
	OnError    func(error)
	OnMessage  func(codec.IMessage)
	OnShutdown func()
	OnInit     func()
}

// ----------------------------------------------------
// Service options structure
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

// Option applies configuration to Options.
type Option func(*Options)

// ----------------------------------------------------
// Default Options builder
// ----------------------------------------------------

func newOptions(opts ...Option) Options {
	opt := Options{
		Context: context.NewContext(),
		Retry: RetryConfig{
			Count:    3,
			Interval: 100 * time.Millisecond,
		},
	}

	for _, o := range opts {
		o(&opt)
	}

	// Apply defaults if not set
	if opt.Registry == nil {
		opt.Registry = registry.NewRegistry()
	}
	if opt.Router == nil {
		opt.Router = router.NewRouter()
	}
	if opt.Transport == nil {
		opt.Transport = transport.New()
	}
	if opt.Selector == nil {
		opt.Selector = selector.NewSelector(opt.Registry, selector.SetStrategy(selector.RoundRobin))
	}
	if opt.Logger == nil {
		opt.Logger = logger.NewLogger("service", "info")
	}

	return opt
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

// Clone creates a deep copy of Options.
func (o *Options) Clone() Options {
	copy := *o
	copy.HdlrWrappers = append([]HandlerWrapper{}, o.HdlrWrappers...)
	copy.Retry = o.Retry
	copy.Hooks = o.Hooks
	return copy
}

// Validate checks for missing critical dependencies.
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

// ErrMissing formats a config missing error.
func ErrMissing(name string) error {
	return &MissingDependencyError{name}
}

type MissingDependencyError struct {
	Dependency string
}

func (e *MissingDependencyError) Error() string {
	return "missing dependency: " + e.Dependency
}
