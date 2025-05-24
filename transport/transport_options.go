// file: arc/service/transport/transport_options.go
package transport

import (
	"os"
	"strconv"
	"time"

	"github.com/rskv-p/mini/service/logger"
)

// ----------------------------------------------------
// Retry policy and metrics interface
// ----------------------------------------------------

// RetryPolicy defines retry behavior for a subject.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

// IMetrics collects transport metrics.
type IMetrics interface {
	IncCounter(name string)
	AddLatency(name string, ms int64)
}

// ----------------------------------------------------
// Transport options and hooks
// ----------------------------------------------------

// Options configures transport behavior.
type Options struct {
	Subject           string
	Addrs             []string
	Timeout           time.Duration
	Debug             bool
	AutoReconnect     bool
	Metrics           IMetrics
	Logger            logger.ILogger
	OnRetry           func(subject string, attempt int, err error)
	OnFailure         func(subject string, err error)
	RetryPolicies     map[string]RetryPolicy
	DeadLetterHandler func(subject string, data []byte, err error)
}

// Option applies a configuration to Options.
type Option func(*Options)

// ----------------------------------------------------
// Option constructors
// ----------------------------------------------------

// Subject sets the transport subject.
func Subject(sub string) Option {
	return func(o *Options) {
		o.Subject = sub
	}
}

// Addrs sets NSQ server addresses.
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// Timeout sets the request timeout.
func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// WithDebug enables debug logging.
func WithDebug() Option {
	return func(o *Options) {
		o.Debug = true
	}
}

// EnableReconnect enables automatic reconnection.
func EnableReconnect() Option {
	return func(o *Options) {
		o.AutoReconnect = true
	}
}

// WithMetrics sets a custom metrics collector.
func WithMetrics(m IMetrics) Option {
	return func(o *Options) {
		o.Metrics = m
	}
}

// WithLogger sets a custom logger.
func WithLogger(l logger.ILogger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithHooks sets OnRetry and OnFailure callbacks.
func WithHooks(onRetry func(string, int, error), onFailure func(string, error)) Option {
	return func(o *Options) {
		o.OnRetry = onRetry
		o.OnFailure = onFailure
	}
}

// WithRetryPolicy sets retry policy for a specific subject.
func WithRetryPolicy(subject string, policy RetryPolicy) Option {
	return func(o *Options) {
		if o.RetryPolicies == nil {
			o.RetryPolicies = make(map[string]RetryPolicy)
		}
		o.RetryPolicies[subject] = policy
	}
}

// WithDeadLetterHandler sets the dead-letter handler.
func WithDeadLetterHandler(handler func(subject string, data []byte, err error)) Option {
	return func(o *Options) {
		o.DeadLetterHandler = handler
	}
}

// ----------------------------------------------------
// Defaults and environment loader
// ----------------------------------------------------

// WithDefaults returns default transport settings.
func WithDefaults() Options {
	return Options{
		Subject:       "default",
		Timeout:       DefaultRequestTimeout,
		Debug:         false,
		AutoReconnect: false,
		RetryPolicies: make(map[string]RetryPolicy),
	}
}

// FromEnv populates options from environment variables:
// SRV_BUS_ADDR, SRV_BUS_SUBJECT, SRV_BUS_TIMEOUT, SRV_BUS_DEBUG.
func FromEnv() Option {
	return func(o *Options) {
		if addr := os.Getenv("SRV_BUS_ADDR"); addr != "" {
			o.Addrs = []string{addr}
		}
		if sub := os.Getenv("SRV_BUS_SUBJECT"); sub != "" && o.Subject == "" {
			o.Subject = sub
		}
		if timeout := os.Getenv("SRV_BUS_TIMEOUT"); timeout != "" {
			if t, err := strconv.Atoi(timeout); err == nil {
				o.Timeout = time.Duration(t) * time.Millisecond
			}
		}
		if dbg := os.Getenv("SRV_BUS_DEBUG"); dbg == "1" || dbg == "true" {
			o.Debug = true
		}
	}
}
