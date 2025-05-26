// file: mini/transport/transport_options.go
package transport

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rskv-p/mini/logger"
)

// ----------------------------------------------------
// Retry policy and metrics interface
// ----------------------------------------------------

// RetryPolicy defines the retry strategy per subject.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

// IMetrics allows collecting transport-level metrics.
type IMetrics interface {
	IncCounter(name string)
	AddLatency(name string, ms int64)
}

// ----------------------------------------------------
// Transport options and configuration
// ----------------------------------------------------

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

// Option is a function that applies a configuration change.
type Option func(*Options)

// ----------------------------------------------------
// Option constructors
// ----------------------------------------------------

// Subject sets the subscription subject (default: "default").
func Subject(sub string) Option {
	return func(o *Options) {
		o.Subject = sub
	}
}

// Addrs sets the NSQ server addresses.
func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

// Timeout sets request timeout.
func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// WithDebug enables verbose logging.
func WithDebug() Option {
	return func(o *Options) {
		o.Debug = true
	}
}

// EnableReconnect allows reconnecting on disconnect.
func EnableReconnect() Option {
	return func(o *Options) {
		o.AutoReconnect = true
	}
}

// WithMetrics sets the metrics collector.
func WithMetrics(m IMetrics) Option {
	return func(o *Options) {
		o.Metrics = m
	}
}

// WithLogger sets the internal logger.
func WithLogger(l logger.ILogger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithHooks registers on-retry and on-failure callbacks.
func WithHooks(onRetry func(string, int, error), onFailure func(string, error)) Option {
	return func(o *Options) {
		o.OnRetry = onRetry
		o.OnFailure = onFailure
	}
}

// WithRetryPolicy sets a retry policy for a specific subject.
func WithRetryPolicy(subject string, policy RetryPolicy) Option {
	return func(o *Options) {
		if o.RetryPolicies == nil {
			o.RetryPolicies = make(map[string]RetryPolicy)
		}
		o.RetryPolicies[subject] = policy
	}
}

// WithRetry sets the default retry policy for all subjects.
func WithRetry(attempts int, delay time.Duration) Option {
	return func(o *Options) {
		o.RetryPolicies["*"] = RetryPolicy{
			MaxAttempts: attempts,
			Delay:       delay,
		}
	}
}

// WithDeadLetterHandler handles final delivery failures.
func WithDeadLetterHandler(handler func(subject string, data []byte, err error)) Option {
	return func(o *Options) {
		o.DeadLetterHandler = handler
	}
}

// ----------------------------------------------------
// Defaults and env-based config
// ----------------------------------------------------

// WithDefaults returns a safe default configuration.
func WithDefaults() Options {
	return Options{
		Subject:       "default",
		Timeout:       DefaultRequestTimeout,
		Debug:         false,
		AutoReconnect: false,
		RetryPolicies: make(map[string]RetryPolicy),
	}
}

// FromEnv loads configuration from standard env vars.
// Uses prefix: SRV_BUS_*
func FromEnv() Option {
	return WithEnvPrefix("SRV_BUS_")
}

// WithEnvPrefix loads env config using the given prefix.
// Supports: {PREFIX}_ADDR, _SUBJECT, _TIMEOUT, _DEBUG
func WithEnvPrefix(prefix string) Option {
	return func(o *Options) {
		get := func(suffix string) string {
			return os.Getenv(prefix + strings.ToUpper(suffix))
		}

		if addr := get("addr"); addr != "" {
			o.Addrs = []string{addr}
		}
		if sub := get("subject"); sub != "" {
			o.Subject = sub
		}
		if timeout := get("timeout"); timeout != "" {
			if t, err := strconv.Atoi(timeout); err == nil {
				o.Timeout = time.Duration(t) * time.Millisecond
			}
		}
		if dbg := get("debug"); dbg == "1" || strings.ToLower(dbg) == "true" {
			o.Debug = true
		}
	}
}
