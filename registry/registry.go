// file: mini/registry/registry.go
package registry

import (
	"sync"
	"time"
)

var _ IRegistry = (*Registry)(nil)

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

type Watcher interface {
	Notify(*Service)
}

type Service struct {
	Name    string           `json:"name"`
	Nodes   []*Node          `json:"nodes"`
	nodeMap map[string]*Node `json:"-"`
}

type Node struct {
	ID       string            `json:"id"`
	Metadata map[string]string `json:"metadata,omitempty"`
	LastSeen time.Time         `json:"-"`
}

type Registry struct {
	mu        sync.RWMutex
	services  map[string]*Service
	watchers  map[Watcher]struct{}
	ttl       time.Duration
	stopPurge chan struct{}
	once      sync.Once
}

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
