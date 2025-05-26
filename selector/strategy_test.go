// file: mini/selector/strategy_test.go
package selector_test

import (
	"testing"

	"github.com/rskv-p/mini/registry"
	"github.com/rskv-p/mini/selector"
	"github.com/stretchr/testify/assert"
)

func makeServices(ids ...string) []*registry.Service {
	nodes := make([]*registry.Node, 0, len(ids))
	for _, id := range ids {
		nodes = append(nodes, &registry.Node{ID: id})
	}
	return []*registry.Service{
		{Name: "svc", Nodes: nodes},
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	s := makeServices("a", "b", "c")
	strategy := selector.RoundRobin(s)

	n1, err := strategy()
	assert.NoError(t, err)
	assert.Equal(t, "a", n1.ID)

	n2, _ := strategy()
	n3, _ := strategy()
	n4, _ := strategy()

	assert.Equal(t, "b", n2.ID)
	assert.Equal(t, "c", n3.ID)
	assert.Equal(t, "a", n4.ID)
}

func TestFirstStrategy(t *testing.T) {
	s := makeServices("first", "second")
	strategy := selector.First(s)

	n, err := strategy()
	assert.NoError(t, err)
	assert.Equal(t, "first", n.ID)
}

func TestRandomStrategy(t *testing.T) {
	s := makeServices("r1", "r2", "r3")
	strategy := selector.Random(s)

	n, err := strategy()
	assert.NoError(t, err)
	assert.NotEmpty(t, n.ID)
}

func TestNamedStrategyHelpers(t *testing.T) {
	name, fn := selector.NamedStrategy("sticky", selector.First)
	assert.Equal(t, "sticky", name)
	assert.NotNil(t, fn)

	opt := selector.StrategyWithName("custom", selector.RoundRobin)
	opts := selector.WithDefaults()
	opt(&opts)

	assert.Equal(t, "custom", opts.StrategyName)
	assert.NotNil(t, opts.Strategy)
}

func TestStrategyErrorWhenEmpty(t *testing.T) {
	// All strategies should return ErrNoAvailableNodes when no nodes
	empty := []*registry.Service{}

	r := selector.RoundRobin(empty)
	_, err := r()
	assert.ErrorIs(t, err, selector.ErrNoAvailableNodes)

	f := selector.First(empty)
	_, err = f()
	assert.ErrorIs(t, err, selector.ErrNoAvailableNodes)

	n := selector.Random(empty)
	_, err = n()
	assert.ErrorIs(t, err, selector.ErrNoAvailableNodes)
}

func TestCollectNodesSafety(t *testing.T) {
	var nilNode *registry.Node
	var nilSvc *registry.Service

	input := []*registry.Service{
		{Name: "ok", Nodes: []*registry.Node{
			{ID: "x"}, nilNode,
		}},
		nilSvc,
	}
	nodes := selector.TestableCollectNodes(input) // using test-only exported alias

	assert.Len(t, nodes, 1)
	assert.Equal(t, "x", nodes[0].ID)
}
