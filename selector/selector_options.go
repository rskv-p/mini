// file: mini/selector/selector_options.go
package selector

import (
	"time"
)

// ----------------------------------------------------
// Selector options and functional setup
// ----------------------------------------------------

// Options configures node selection behavior.
type Options struct {
	Strategy     Strategy      // Node selection strategy function
	StrategyName string        // Human-readable name of strategy
	CacheTTL     time.Duration // TTL for cached service registry entries
}

// Option applies configuration changes to Options.
type Option func(*Options)

// SetStrategy sets the node selection strategy (e.g., RoundRobin, Random).
func SetStrategy(fn Strategy) Option {
	return func(o *Options) {
		o.Strategy = fn
	}
}

// SetStrategyNamed sets the strategy function and its name (for introspection/logging).
func SetStrategyNamed(name string, fn Strategy) Option {
	return func(o *Options) {
		o.Strategy = fn
		o.StrategyName = name
	}
}

// SetCacheTTL sets the cache time-to-live for registry entries.
func SetCacheTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.CacheTTL = ttl
	}
}

// WithDefaults returns safe default options.
func WithDefaults() Options {
	return Options{
		Strategy:     RoundRobin,
		StrategyName: "round_robin",
		CacheTTL:     2 * time.Second,
	}
}
