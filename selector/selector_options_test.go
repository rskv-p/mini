// file: mini/selector/selector_options_test.go
package selector_test

import (
	"testing"
	"time"

	"github.com/rskv-p/mini/registry"
	"github.com/rskv-p/mini/selector"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------------
// Default options
// ----------------------------------------------------

func TestWithDefaults(t *testing.T) {
	opts := selector.WithDefaults()

	assert.NotNil(t, opts.Strategy)
	assert.Equal(t, "round_robin", opts.StrategyName)
	assert.Equal(t, 2*time.Second, opts.CacheTTL)
}

// ----------------------------------------------------
// Custom strategy (unnamed)
// ----------------------------------------------------

func TestSetStrategy(t *testing.T) {
	customCalled := false

	custom := func(svcs []*registry.Service) selector.Next {
		customCalled = true
		return func() (*registry.Node, error) {
			return &registry.Node{ID: "custom"}, nil
		}
	}

	opts := selector.WithDefaults()
	selector.SetStrategy(custom)(&opts)

	// StrategyName remains unchanged
	assert.Equal(t, "round_robin", opts.StrategyName)

	next := opts.Strategy(nil)
	node, err := next()
	assert.NoError(t, err)
	assert.Equal(t, "custom", node.ID)
	assert.True(t, customCalled)
}

// ----------------------------------------------------
// Named strategy override
// ----------------------------------------------------

func TestSetStrategyNamed(t *testing.T) {
	namedCalled := false

	named := func(svcs []*registry.Service) selector.Next {
		namedCalled = true
		return func() (*registry.Node, error) {
			return &registry.Node{ID: "named"}, nil
		}
	}

	opts := selector.WithDefaults()
	selector.SetStrategyNamed("custom_named", named)(&opts)

	assert.Equal(t, "custom_named", opts.StrategyName)

	next := opts.Strategy(nil)
	node, err := next()
	assert.NoError(t, err)
	assert.Equal(t, "named", node.ID)
	assert.True(t, namedCalled)
}

// ----------------------------------------------------
// TTL override
// ----------------------------------------------------

func TestSetCacheTTL(t *testing.T) {
	opts := selector.WithDefaults()
	selector.SetCacheTTL(5 * time.Second)(&opts)

	assert.Equal(t, 5*time.Second, opts.CacheTTL)
}
