// file: mini/selector/selector_test.go
package selector_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rskv-p/mini/registry"
	"github.com/rskv-p/mini/selector"
	"github.com/stretchr/testify/assert"
)

func TestSelector_LifecycleAndSelect(t *testing.T) {
	mockReg := newMockRegistry()
	mockReg.services["svc.a"] = []*registry.Service{
		{
			Name: "svc.a",
			Nodes: []*registry.Node{
				{ID: "node-1", Metadata: map[string]string{"zone": "eu"}},
				{ID: "node-2", Metadata: map[string]string{"zone": "us"}},
			},
		},
	}

	sel := selector.NewSelector(mockReg)
	assert.NoError(t, sel.Init())

	// Select without filters
	id, err := sel.Select("svc.a")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Select with MatchMeta filter
	id, _ = sel.Select("svc.a", selector.MatchMeta("zone", "us"))
	assert.Equal(t, "node-2", id)

	// SelectNode with no match
	_, err = sel.SelectNode("svc.a", selector.MatchMeta("zone", "xx"))
	assert.Error(t, err)

	// Select on missing service
	_, err = sel.Select("svc.404")
	assert.Error(t, err)

	// Call again (now uses cache)
	id2, _ := sel.Select("svc.a", selector.MatchMeta("zone", "eu"))
	assert.Equal(t, "node-1", id2)
}

func TestSelector_InvalidateAndDump(t *testing.T) {
	mockReg := newMockRegistry()
	mockReg.services["svc.b"] = []*registry.Service{
		{Name: "svc.b", Nodes: []*registry.Node{{ID: "node-b"}}},
	}

	sel := selector.NewSelector(mockReg, selector.SetCacheTTL(1*time.Hour))
	_ = sel.Init()

	id, err := sel.Select("svc.b")
	assert.NoError(t, err)
	assert.Equal(t, "node-b", id)

	// Dump must show node
	dump := sel.DumpCache()
	assert.Contains(t, dump, "svc.b")
	assert.Contains(t, dump["svc.b"], "node-b")

	// Invalidate and check removed
	sel.Invalidate("svc.b")
	dump = sel.DumpCache()
	assert.NotContains(t, dump, "svc.b")
}

func TestSelector_CacheExpiry(t *testing.T) {
	mockReg := newMockRegistry()
	mockReg.services["svc.c"] = []*registry.Service{
		{Name: "svc.c", Nodes: []*registry.Node{{ID: "node-c"}}},
	}

	sel := selector.NewSelector(mockReg, selector.SetCacheTTL(10*time.Millisecond))
	_ = sel.Init()

	id, err := sel.Select("svc.c")
	assert.NoError(t, err)
	assert.Equal(t, "node-c", id)

	time.Sleep(20 * time.Millisecond)

	// Registry call should happen again
	mockReg.called = false
	_, _ = sel.Select("svc.c")
	assert.True(t, mockReg.called)
}

func TestSelector_NilRegistryError(t *testing.T) {
	sel := selector.NewSelector(nil)
	err := sel.Init()
	assert.Error(t, err)
}

func TestSelector_StrategyRoundRobin(t *testing.T) {
	nodes := []*registry.Node{{ID: "a"}, {ID: "b"}}
	svc := []*registry.Service{{Name: "svc", Nodes: nodes}}

	rr := selector.RoundRobin(svc)
	n1, _ := rr()
	n2, _ := rr()
	n3, _ := rr()
	assert.Equal(t, "a", n1.ID)
	assert.Equal(t, "b", n2.ID)
	assert.Equal(t, "a", n3.ID)
}

func TestSelector_StrategyRandom(t *testing.T) {
	nodes := []*registry.Node{{ID: "x"}, {ID: "y"}}
	svc := []*registry.Service{{Name: "svc", Nodes: nodes}}

	fn := selector.Random(svc)
	n, err := fn()
	assert.NoError(t, err)
	assert.Contains(t, []string{"x", "y"}, n.ID)
}

func TestSelector_StrategyFirst(t *testing.T) {
	nodes := []*registry.Node{{ID: "f1"}, {ID: "f2"}}
	svc := []*registry.Service{{Name: "svc", Nodes: nodes}}

	fn := selector.First(svc)
	n, err := fn()
	assert.NoError(t, err)
	assert.Equal(t, "f1", n.ID)
}

func TestSelector_MatchHelpers(t *testing.T) {
	node := &registry.Node{
		ID:       "test",
		Metadata: map[string]string{"region": "eu", "zone": "a"},
	}
	assert.True(t, selector.MatchMeta("region", "eu")(node))
	assert.False(t, selector.MatchMeta("zone", "b")(node))
	assert.True(t, selector.MatchID("test")(node))
	assert.False(t, selector.MatchID("other")(node))
}

// --------------------------
// Mock Registry for testing
// --------------------------

type mockRegistry struct {
	services map[string][]*registry.Service
	called   bool
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{services: make(map[string][]*registry.Service)}
}

func (m *mockRegistry) Init() error                          { return nil }
func (m *mockRegistry) Register(s *registry.Service) error   { return nil }
func (m *mockRegistry) Deregister(s *registry.Service) error { return nil }
func (m *mockRegistry) GetService(name string) ([]*registry.Service, error) {
	m.called = true
	svc, ok := m.services[name]
	if !ok {
		return nil, errors.New("not found")
	}
	return svc, nil
}
func (m *mockRegistry) ListServices() ([]*registry.Service, error) { return nil, nil }
func (m *mockRegistry) TotalServices() int                         { return 0 }
func (m *mockRegistry) TotalNodes(string) int                      { return 0 }
func (m *mockRegistry) Dump() map[string][]string                  { return nil }
func (m *mockRegistry) AddWatcher(registry.Watcher)                {}
func (m *mockRegistry) RemoveWatcher(registry.Watcher)             {}
func (m *mockRegistry) UpdateTTL(time.Duration)                    {}
func (m *mockRegistry) Close()                                     {}
