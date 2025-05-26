// file: mini/registry/registry_test.go
package registry_test

import (
	"testing"
	"time"

	"github.com/rskv-p/mini/registry"
	"github.com/stretchr/testify/assert"
)

type mockWatcher struct {
	notified []*registry.Service
}

func (w *mockWatcher) Notify(s *registry.Service) {
	w.notified = append(w.notified, s)
}

func TestRegistryLifecycle(t *testing.T) {
	r := registry.NewRegistry()
	defer r.Close()

	assert.NoError(t, r.Init())

	// Register watcher
	w := &mockWatcher{}
	r.AddWatcher(w)

	// Register a service
	svc := &registry.Service{
		Name: "demo",
		Nodes: []*registry.Node{
			{ID: "node1"},
		},
	}
	err := r.Register(svc)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.TotalServices())
	assert.Equal(t, 1, r.TotalNodes("demo"))

	// Dump check
	dump := r.Dump()
	assert.Contains(t, dump, "demo")
	assert.Equal(t, []string{"node1"}, dump["demo"])

	// Watcher should have been notified (eventually)
	assert.Eventually(t, func() bool {
		return len(w.notified) == 1
	}, time.Second, 10*time.Millisecond)
	assert.Equal(t, "demo", w.notified[0].Name)

	// Get service
	list, err := r.GetService("demo")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "demo", list[0].Name)

	// List all services
	all, err := r.ListServices()
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	// Deregister node
	err = r.Deregister(&registry.Service{
		Name:  "demo",
		Nodes: []*registry.Node{{ID: "node1"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, r.TotalNodes("demo"))
	assert.Equal(t, 0, r.TotalServices())

	// Remove watcher (cleanup)
	r.RemoveWatcher(w)
}

func TestRegistry_TTLExpiry(t *testing.T) {
	r := registry.NewRegistry()
	defer r.Close()

	r.UpdateTTL(10 * time.Millisecond)

	// Register a service
	svc := &registry.Service{
		Name: "ttltest",
		Nodes: []*registry.Node{
			{ID: "n1"},
		},
	}
	assert.NoError(t, r.Register(svc))
	assert.Equal(t, 1, r.TotalServices())

	// Wait until service expires via janitor
	assert.Eventually(t, func() bool {
		return r.TotalServices() == 0
	}, 150*time.Millisecond, 10*time.Millisecond)
}
