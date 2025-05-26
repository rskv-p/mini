// file: mini/selector/selector.go
package selector

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rskv-p/mini/registry"
)

// For test coverage only (not exported normally)
var TestableCollectNodes = collectNodes

// ----------------------------------------------------
// Selector interfaces and types
// ----------------------------------------------------

var _ ISelector = (*Selector)(nil)

// ISelector selects a node ID or object for a service.
type ISelector interface {
	Init() error
	Select(service string, filters ...SelectorFilter) (string, error)
	SelectNode(service string, filters ...SelectorFilter) (*registry.Node, error)
	Invalidate(service string)
	DumpCache() map[string][]string
}

// Strategy defines how to pick nodes from services.
type Strategy func([]*registry.Service) Next

// Next returns the next selected node.
type Next func() (*registry.Node, error)

// SelectorFilter filters nodes by arbitrary criteria.
type SelectorFilter func(*registry.Node) bool

// ----------------------------------------------------
// Filter helpers
// ----------------------------------------------------

// MatchMeta returns a filter matching node metadata key/value.
func MatchMeta(key, value string) SelectorFilter {
	return func(n *registry.Node) bool {
		return n != nil && n.Metadata[key] == value
	}
}

func MatchID(id string) SelectorFilter {
	return func(n *registry.Node) bool {
		return n != nil && n.ID == id
	}
}

// ----------------------------------------------------
// Selector implementation
// ----------------------------------------------------

type cachedServices struct {
	services  []*registry.Service
	timestamp time.Time
}

// NewSelector creates a new Selector with given registry and options.
func NewSelector(reg registry.IRegistry, opts ...Option) ISelector {
	sOpts := WithDefaults()
	for _, opt := range opts {
		opt(&sOpts)
	}
	return &Selector{
		registry: reg,
		opts:     sOpts,
		cache:    make(map[string]cachedServices),
	}
}

// Selector holds registry reference and strategy options.
type Selector struct {
	registry registry.IRegistry
	opts     Options

	mu    sync.RWMutex
	cache map[string]cachedServices
}

// Init ensures registry and strategy are set.
func (s *Selector) Init() error {
	if s.registry == nil {
		return errors.New("selector: registry is nil")
	}
	if s.opts.Strategy == nil {
		s.opts.Strategy = RoundRobin
	}
	if s.opts.CacheTTL <= 0 {
		s.opts.CacheTTL = 2 * time.Second
	}
	return nil
}

// Select returns the ID of a selected node.
func (s *Selector) Select(service string, filters ...SelectorFilter) (string, error) {
	node, err := s.SelectNode(service, filters...)
	if err != nil {
		return "", err
	}
	return node.ID, nil
}

// SelectNode returns the selected node object.
func (s *Selector) SelectNode(service string, filters ...SelectorFilter) (*registry.Node, error) {
	services, err := s.getCachedServices(service)
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("selector: service %q not found", service)
	}

	// apply filters to nodes
	var filtered []*registry.Service
	for _, svc := range services {
		if svc == nil || len(svc.Nodes) == 0 {
			continue
		}
		var keep []*registry.Node
		for _, n := range svc.Nodes {
			if matchesAllFilters(n, filters) {
				keep = append(keep, n)
			}
		}
		if len(keep) > 0 {
			filtered = append(filtered, &registry.Service{Name: svc.Name, Nodes: keep})
		}
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("selector: no nodes matched filters for service %q", service)
	}

	// pick using strategy
	next := s.opts.Strategy(filtered)
	return next()
}

// Invalidate clears cached service entry.
func (s *Selector) Invalidate(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, service)
}

// DumpCache returns snapshot of cached node IDs per service.
func (s *Selector) DumpCache() map[string][]string {
	out := make(map[string][]string)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for name, entry := range s.cache {
		for _, svc := range entry.services {
			for _, n := range svc.Nodes {
				if n != nil {
					out[name] = append(out[name], n.ID)
				}
			}
		}
	}
	return out
}

// getCachedServices returns services from cache or queries the registry.
func (s *Selector) getCachedServices(service string) ([]*registry.Service, error) {
	s.mu.RLock()
	entry, ok := s.cache[service]
	s.mu.RUnlock()

	if ok && time.Since(entry.timestamp) <= s.opts.CacheTTL {
		return entry.services, nil
	}

	services, err := s.registry.GetService(service)
	if err != nil {
		return nil, fmt.Errorf("selector: registry error for %q: %w", service, err)
	}

	s.mu.Lock()
	s.cache[service] = cachedServices{
		services:  services,
		timestamp: time.Now(),
	}
	s.mu.Unlock()

	return services, nil
}

// matchesAllFilters checks if node passes all filters.
func matchesAllFilters(n *registry.Node, filters []SelectorFilter) bool {
	for _, f := range filters {
		if !f(n) {
			return false
		}
	}
	return true
}
