package srv

import (
	"time"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_api"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Service Struct
//---------------------

// Service represents the main service that holds modules.
type Service struct {
	Name    string
	Modules []typ.IModule
}

//---------------------
// Service Lifecycle
//---------------------

// New creates a new service, adding the built-in modules automatically.
func New(name string) *Service {
	s := &Service{
		Name:    name,
		Modules: []typ.IModule{},
	}

	// Register built-in modules
	mod.RegisterBuiltinModules()

	// Add built-in modules to the service's modules list
	s.AddModule(&m_api.APIActionModule{})

	return s
}

// Start initializes all modules and registers their actions.
func (s *Service) Start() error {
	// Initialize each module and register their actions
	for _, mod := range s.Modules {
		if err := mod.Init(); err != nil {
			return err
		}
		for _, a := range mod.Actions() {
			m_api.RegisterAction(a.Name, a.Func, m_api.APIOption{
				Public: a.Public,
				Doc:    "", // Documentation can be added later
			})
		}
	}
	s.registerSystemEndpoints() // Register default system endpoints
	return nil
}

// Stop gracefully stops all modules.
func (s *Service) Stop() error {
	// Stop each module in the service
	for _, mod := range s.Modules {
		if err := mod.Stop(); err != nil {
			return err
		}
	}
	return nil
}

//---------------------
// Default System Verbs
//---------------------

// registerSystemEndpoints registers default system endpoints for the service.
func (s *Service) registerSystemEndpoints() {
	// Ping endpoint, used to check service health
	act.Register("system.ping", func(a typ.IAction) any {
		return map[string]any{
			"pong":  true,
			"ts":    time.Now().Format(time.RFC3339),
			"svc":   s.Name,
			"alive": true,
		}
	})

	// Info endpoint, provides information about the service
	act.Register("system.info", func(a typ.IAction) any {
		modules := []string{}
		for _, m := range s.Modules {
			modules = append(modules, m.Name())
		}
		return map[string]any{
			"name":    s.Name,
			"modules": modules,
			"ts":      time.Now().Format(time.RFC3339),
		}
	})

	// Stats endpoint, provides service statistics
	act.Register("system.stats", func(a typ.IAction) any {
		return map[string]any{
			"actions": len(act.Handlers),
			"modules": len(s.Modules),
			"ts":      time.Now().Format(time.RFC3339),
		}
	})

	// Docs endpoint, provides documentation for the service's actions
	act.Register("system.docs", func(a typ.IAction) any {
		docs := map[string]any{}
		for name, _ := range act.Handlers {
			apiEntry, ok := m_api.Get(name)
			docs[name] = map[string]any{
				"handler": name,
				"public":  ok && apiEntry.Options.Public,
				"doc":     apiEntry.Options.Doc,
				"tags":    apiEntry.Options.Tags,
			}
		}
		return docs
	})
}

//---------------------
// Module Management
//---------------------

// AddModule adds a new module to the service.
func (s *Service) AddModule(mod typ.IModule) {
	s.Modules = append(s.Modules, mod)
}

// RegisterBuiltInModule registers a specific built-in module with initialization logic.
func RegisterBuiltInModule(modName string, initFunc func() error) {
	module := &mod.Module{
		ModName: modName,
		OnInit:  initFunc,
	}

	// Dynamically register the module
	mod.Register(module)
}
