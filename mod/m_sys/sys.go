package m_sys

import (
	"time"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// SystemModule
//---------------------

// SystemModule represents the system that provides system actions.
type SystemModule struct {
	Module    typ.IModule    // The module implementing IModule interface
	logClient typ.ILogClient // Client for logging
}

// Ensure SystemModule implements the IModule interface.
var _ typ.IModule = (*SystemModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (s *SystemModule) Name() string {
	return s.Module.Name()
}

// Stop stops the module (specific shutdown logic can be added).
func (s *SystemModule) Stop() error {
	// Log the stop action
	s.logClient.Info("Stopping System module", map[string]interface{}{"module": s.Name()})
	return nil
}

// Start starts the module (specific logic can be added).
func (s *SystemModule) Start() error {
	// Log the start action
	s.logClient.Info("Starting System module", map[string]interface{}{"module": s.Name()})
	return nil
}

// Init initializes the module (specific logic can be added).
func (s *SystemModule) Init() error {
	// Log the initialization
	s.logClient.Info("Initializing System module", map[string]interface{}{"module": s.Name()})
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions for the system module.
func (s *SystemModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
		{
			Name: "system.ping", // Ping action to check service status
			Func: func(a typ.IAction) any {
				// Log the ping action
				s.logClient.Info("Ping received", map[string]interface{}{"action": "system.ping"})
				// Return a simple response with timestamp and service status
				return map[string]any{
					"pong":  true,
					"ts":    time.Now().Format(time.RFC3339),
					"alive": true,
				}
			},
			Public: true,
		},
		{
			Name: "system.info", // Info action to return basic service and module info
			Func: func(a typ.IAction) any {
				// Log the info action
				s.logClient.Info("System info requested", map[string]interface{}{"action": "system.info"})
				modules := []string{}
				for _, m := range s.Module.Actions() {
					modules = append(modules, m.Name)
				}
				return map[string]any{
					"name":    s.Module.Name(),
					"modules": modules,
					"ts":      time.Now().Format(time.RFC3339),
				}
			},
			Public: true,
		},
		{
			Name: "system.stats", // Stats action to return statistics on actions and modules
			Func: func(a typ.IAction) any {
				// Log the stats action
				s.logClient.Info("System stats requested", map[string]interface{}{"action": "system.stats"})
				return map[string]any{
					"actions": len(act.Handlers),
					"modules": len(s.Module.Actions()),
					"ts":      time.Now().Format(time.RFC3339),
				}
			},
			Public: true,
		},
	}
}

//---------------------
// Module Creation
//---------------------

// NewSystemModule creates a new instance of SystemModule using the NewModule constructor.
func NewSystemModule(service typ.IService, logClient typ.ILogClient) *SystemModule {
	// Create a new module using NewModule, passing name, service, actions, and nil for OnInit and OnStop
	module := mod.NewModule("sys", service, nil, nil, nil)

	// Log module creation
	logClient.Info("Creating System module", map[string]interface{}{"module": "sys"})

	// Return the SystemModule with the created module
	systemModule := &SystemModule{
		Module:    module,
		logClient: logClient,
	}

	// Register actions for the system module
	for _, action := range systemModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	// Log successful creation of module
	logClient.Info("System module created successfully", map[string]interface{}{"module": "sys"})

	return systemModule
}
