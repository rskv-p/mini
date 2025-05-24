// file: arc/service/selector/selector.go
package selector

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rskv-p/mini/service/registry"
)

// ----------------------------------------------------
// Selector interfaces and types
// ----------------------------------------------------

var _ ISelector = (*Selector)(nil)

// ISelector selects a node ID or object for a service.
type ISelector interface {
	Init() error
	Select(service string, filters ...SelectorFilter) (string, error)
	SelectNode(service string, filters ...SelectorFilter) (*registry.Node, error)
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
		return n.Metadata[key] == value
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
	sOpts := Options{}
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

	cacheTTL time.Duration
	mu       sync.RWMutex
	cache    map[string]cachedServices
}

// Init ensures registry and strategy are set.
func (s *Selector) Init() error {
	if s.registry == nil {
		return errors.New("selector: registry is nil")
	}
	if s.opts.Strategy == nil {
		s.opts.Strategy = RoundRobin
	}
	if s.cacheTTL == 0 {
		s.cacheTTL = 2 * time.Second
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

// matchesAllFilters checks if node passes all filters.
func matchesAllFilters(n *registry.Node, filters []SelectorFilter) bool {
	for _, f := range filters {
		if !f(n) {
			return false
		}
	}
	return true
}

// getCachedServices returns services from cache or queries the registry.
func (s *Selector) getCachedServices(service string) ([]*registry.Service, error) {
	s.mu.RLock()
	entry, ok := s.cache[service]
	s.mu.RUnlock()

	if ok && time.Since(entry.timestamp) <= s.cacheTTL {
		return entry.services, nil
	}

	// not cached or expired
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
