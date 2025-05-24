// file: mini/selector/strategy.go
package selector

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/rskv-p/mini/registry"
)

// ----------------------------------------------------
// Node selection strategies
// ----------------------------------------------------

var (
	ErrNoAvailableNodes = errors.New("selector: no available nodes")
)

// Random returns a Next function that selects a random node.
func Random(services []*registry.Service) Next {
	nodes := collectNodes(services)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, ErrNoAvailableNodes
		}
		return nodes[r.Intn(len(nodes))], nil
	}
}

// RoundRobin returns a Next function that cycles through nodes in order.
func RoundRobin(services []*registry.Service) Next {
	nodes := collectNodes(services)
	var (
		mu  sync.Mutex
		idx int
	)

	return func() (*registry.Node, error) {
		mu.Lock()
		defer mu.Unlock()

		if len(nodes) == 0 {
			return nil, ErrNoAvailableNodes
		}
		node := nodes[idx%len(nodes)]
		idx++
		return node, nil
	}
}

// First returns a Next function that always selects the first node.
func First(services []*registry.Service) Next {
	nodes := collectNodes(services)
	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, ErrNoAvailableNodes
		}
		return nodes[0], nil
	}
}

// ----------------------------------------------------
// Helper to collect nodes
// ----------------------------------------------------

// collectNodes flattens all service nodes into a single slice.
// Skips nil node entries for safety.
func collectNodes(services []*registry.Service) []*registry.Node {
	var nodes []*registry.Node
	for _, svc := range services {
		for _, n := range svc.Nodes {
			if n != nil {
				nodes = append(nodes, n)
			}
		}
	}
	return nodes
}
