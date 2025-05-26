// file: mini/registry/internal_test.go
package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type panicWatcher struct{}

func (p *panicWatcher) Notify(_ *Service) {
	panic("boom")
}

func TestInternal_RegistryOps(t *testing.T) {
	r := NewRegistry()
	defer r.Close()

	svc := &Service{
		Name: "test",
		Nodes: []*Node{
			{ID: "n1"},
		},
	}

	// Register nil service
	err := r.Register(nil)
	assert.Error(t, err)

	// Register empty node list
	err = r.Register(&Service{Name: "bad", Nodes: nil})
	assert.Error(t, err)

	// Register real service
	err = r.Register(svc)
	assert.NoError(t, err)

	// Register duplicate — should not panic
	err = r.Register(svc)
	assert.NoError(t, err)

	// GetService
	services, err := r.GetService("test")
	assert.NoError(t, err)
	assert.Len(t, services, 1)

	// ListServices
	list, err := r.ListServices()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)

	// Dump
	dump := r.Dump()
	assert.Contains(t, dump, "test")
	assert.Equal(t, []string{"n1"}, dump["test"])

	// TotalServices & TotalNodes
	assert.Equal(t, 1, r.TotalServices())
	assert.Equal(t, 1, r.TotalNodes("test"))
	assert.Equal(t, 0, r.TotalNodes("unknown"))

	// Deregister: invalid
	assert.Error(t, r.Deregister(nil))
	assert.Error(t, r.Deregister(&Service{Name: "x", Nodes: nil}))

	// Deregister: remove node
	err = r.Deregister(&Service{
		Name:  "test",
		Nodes: []*Node{{ID: "n1"}},
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, r.TotalServices())
}

func TestInternal_Watchers(t *testing.T) {
	r := NewRegistry()
	defer r.Close()

	// Add panic watcher
	pw := &panicWatcher{}
	r.AddWatcher(pw)

	// Should not panic due to watcher recovery
	_ = r.Register(&Service{
		Name: "panic",
		Nodes: []*Node{
			{ID: "p1"},
		},
	})

	// Remove watcher
	r.RemoveWatcher(pw)
}

func TestInternal_JanitorTTL(t *testing.T) {
	r := NewRegistry()
	r.UpdateTTL(10 * time.Millisecond)

	_ = r.Register(&Service{
		Name: "expire",
		Nodes: []*Node{
			{ID: "exp1"},
		},
	})

	// Confirm registered
	assert.Equal(t, 1, r.TotalServices())

	time.Sleep(50 * time.Millisecond)

	assert.Eventually(t, func() bool {
		return r.TotalServices() == 0
	}, 100*time.Millisecond, 10*time.Millisecond)

	r.Close()
}

func TestInternal_cloneService(t *testing.T) {
	orig := &Service{
		Name: "svc",
		Nodes: []*Node{
			{ID: "x", Metadata: map[string]string{"k": "v"}},
		},
		nodeMap: map[string]*Node{
			"x": {ID: "x"},
		},
	}
	clone := cloneService(orig)

	assert.Equal(t, "svc", clone.Name)
	assert.Equal(t, 1, len(clone.Nodes))
	assert.Equal(t, "x", clone.Nodes[0].ID)
	assert.Equal(t, "v", clone.Nodes[0].Metadata["k"])

	// Ensure deep copy (pointers not shared)
	clone.Nodes[0].Metadata["k"] = "modified"
	assert.Equal(t, "v", orig.Nodes[0].Metadata["k"])
}
