// file: arc/service/selector/selector_options.go
package selector

import "time"

// ----------------------------------------------------
// Selector options
// ----------------------------------------------------

// Options configures node selection behavior.
type Options struct {
	Strategy Strategy      // node selection strategy
	CacheTTL time.Duration // TTL for cached service data
}

// Option is a functional option for Selector Options.
type Option func(*Options)

// SetStrategy sets the node selection strategy (e.g., RoundRobin, Random).
func SetStrategy(fn Strategy) Option {
	return func(o *Options) {
		o.Strategy = fn
	}
}

// SetCacheTTL sets the cache time-to-live for service entries.
func SetCacheTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.CacheTTL = ttl
	}
}

// WithDefaults returns default options (RoundRobin + 2s cache).
func WithDefaults() Options {
	return Options{
		Strategy: RoundRobin,
		CacheTTL: 2 * time.Second,
	}
}
