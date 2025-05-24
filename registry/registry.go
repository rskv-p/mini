// file: mini/registry/registry.go
package registry

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"
)

var _ IRegistry = (*Registry)(nil)

// IRegistry defines the interface for service registration and discovery.
type IRegistry interface {
	Init() error
	Register(*Service) error
	Deregister(*Service) error
	GetService(string) ([]*Service, error)
	ListServices() ([]*Service, error)

	TotalServices() int
	TotalNodes(string) int
	Dump() map[string][]string

	AddWatcher(Watcher)
	RemoveWatcher(Watcher)

	UpdateTTL(time.Duration)
	Close()
}

// Watcher gets notified on registry changes.
type Watcher interface {
	Notify(*Service)
}

// Service represents a registered service and its nodes.
type Service struct {
	Name    string           `json:"name"`
	Nodes   []*Node          `json:"nodes"`
	nodeMap map[string]*Node `json:"-"`
}

// Node represents an instance of a service.
type Node struct {
	ID       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	LastSeen time.Time         `json:"-"`
}

// Registry is an in-memory implementation of IRegistry.
type Registry struct {
	mu        sync.RWMutex
	services  map[string]*Service
	watchers  map[Watcher]struct{}
	ttl       time.Duration
	stopPurge chan struct{}
	once      sync.Once
}

// NewRegistry creates a new in-memory registry instance.
func NewRegistry() *Registry {
	r := &Registry{
		services:  make(map[string]*Service),
		watchers:  make(map[Watcher]struct{}),
		ttl:       30 * time.Second,
		stopPurge: make(chan struct{}),
	}
	go r.startJanitor()
	return r
}

func (r *Registry) Init() error {
	return nil
}

func (r *Registry) Register(s *Service) error {
	if s == nil || len(s.Nodes) == 0 {
		return errors.New("registry: at least one node is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.services[s.Name]
	if !ok {
		svc := &Service{
			Name:    s.Name,
			Nodes:   nil,
			nodeMap: make(map[string]*Node),
		}
		for _, n := range s.Nodes {
			n.LastSeen = time.Now()
			svc.Nodes = append(svc.Nodes, n)
			svc.nodeMap[n.ID] = n
		}
		r.services[s.Name] = svc
		r.notifyWatchers(svc)
		return nil
	}

	changed := false
	for _, n := range s.Nodes {
		if existing.nodeMap == nil {
			existing.nodeMap = make(map[string]*Node)
		}
		if old, ok := existing.nodeMap[n.ID]; ok {
			old.LastSeen = time.Now()
		} else {
			n.LastSeen = time.Now()
			existing.Nodes = append(existing.Nodes, n)
			existing.nodeMap[n.ID] = n
			changed = true
		}
	}
	if changed {
		r.notifyWatchers(existing)
	}
	return nil
}

func (r *Registry) Deregister(s *Service) error {
	if s == nil || len(s.Nodes) == 0 {
		return errors.New("registry: at least one node is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.services[s.Name]
	if !ok {
		return nil
	}

	nodeID := s.Nodes[0].ID
	delete(existing.nodeMap, nodeID)

	var remaining []*Node
	for _, n := range existing.Nodes {
		if n.ID != nodeID {
			remaining = append(remaining, n)
		}
	}
	existing.Nodes = remaining

	if len(existing.Nodes) == 0 {
		delete(r.services, s.Name)
	}
	r.notifyWatchers(existing)
	return nil
}

func (r *Registry) GetService(name string) ([]*Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	svc, ok := r.services[name]
	if !ok {
		return nil, errors.New("registry: service not found")
	}
	return []*Service{cloneService(svc)}, nil
}

func (r *Registry) ListServices() ([]*Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*Service
	for _, svc := range r.services {
		list = append(list, cloneService(svc))
	}
	return list, nil
}

// ----------------------------------------------------
// Utilities
// ----------------------------------------------------

func (r *Registry) TotalServices() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}

func (r *Registry) TotalNodes(service string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	svc, ok := r.services[service]
	if !ok {
		return 0
	}
	return len(svc.Nodes)
}

func (r *Registry) Dump() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string][]string)
	for name, svc := range r.services {
		for _, n := range svc.Nodes {
			out[name] = append(out[name], n.ID)
		}
	}
	// Sort for stable output
	for _, list := range out {
		sort.Strings(list)
	}
	return out
}

// ----------------------------------------------------
// Watchers
// ----------------------------------------------------

func (r *Registry) AddWatcher(w Watcher) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.watchers[w] = struct{}{}
}

func (r *Registry) RemoveWatcher(w Watcher) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.watchers, w)
}

func (r *Registry) notifyWatchers(svc *Service) {
	for w := range r.watchers {
		go func(w Watcher) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("[registry] watcher panic: %v — removing", err)
					r.RemoveWatcher(w)
				}
			}()
			w.Notify(cloneService(svc))
		}(w)
	}
}

// ----------------------------------------------------
// TTL Cleanup
// ----------------------------------------------------

func (r *Registry) startJanitor() {
	r.once.Do(func() {
		ticker := time.NewTicker(r.ttl)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					r.purgeExpired()
				case <-r.stopPurge:
					return
				}
			}
		}()
	})
}

func (r *Registry) purgeExpired() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for name, svc := range r.services {
		var active []*Node
		for _, n := range svc.Nodes {
			if now.Sub(n.LastSeen) <= r.ttl {
				active = append(active, n)
			} else {
				delete(svc.nodeMap, n.ID)
			}
		}
		svc.Nodes = active
		if len(active) == 0 {
			delete(r.services, name)
			log.Printf("[registry] auto-removed stale service: %s", name)
		}
	}
}

// UpdateTTL changes the TTL duration for service expiration.
func (r *Registry) UpdateTTL(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ttl = d
}

// Close stops the janitor routine.
func (r *Registry) Close() {
	close(r.stopPurge)
}

// ----------------------------------------------------
// Internal helpers
// ----------------------------------------------------

func cloneService(src *Service) *Service {
	out := &Service{
		Name:    src.Name,
		Nodes:   make([]*Node, len(src.Nodes)),
		nodeMap: make(map[string]*Node, len(src.Nodes)),
	}
	for i, n := range src.Nodes {
		cpy := *n
		out.Nodes[i] = &cpy
		out.nodeMap[cpy.ID] = &cpy
	}
	return out
}
