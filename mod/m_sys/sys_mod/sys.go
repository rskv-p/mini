package sys_mod

import (
	"time"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_log/log_type"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// SystemModule
//---------------------

// SystemModule represents the system that provides system actions.
type SystemModule struct {
	Module    typ.IModule         // The module implementing IModule interface
	logClient log_type.ILogClient // Client for logging
}

// Ensure SystemModule implements the IModule interface.
var _ typ.IModule = (*SystemModule)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (s *SystemModule) GetName() string {
	return s.Module.GetName()
}

// Stop stops the module (specific shutdown logic can be added).
func (s *SystemModule) Stop() error {
	// Log the stop action
	s.logClient.Info("Stopping System module", map[string]interface{}{"module": s.GetName()})
	return nil
}

// Start starts the module (specific logic can be added).
func (s *SystemModule) Start() error {
	// Log the start action
	s.logClient.Info("Starting System module", map[string]interface{}{"module": s.GetName()})
	return nil
}

// Init initializes the module (specific logic can be added).
func (s *SystemModule) Init() error {
	// Log the initialization
	s.logClient.Info("Initializing System module", map[string]interface{}{"module": s.GetName()})
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions for the system module.
func (s *SystemModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		{
			Name: "system.ping", // Ping action to check service status
			Func: func(a act_type.IAction) any {
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
			Func: func(a act_type.IAction) any {
				// Log the info action
				s.logClient.Info("System info requested", map[string]interface{}{"action": "system.info"})
				modules := []string{}
				for _, m := range s.Module.GetActions() {
					modules = append(modules, m.Name)
				}
				return map[string]any{
					"name":    s.Module.GetName(),
					"modules": modules,
					"ts":      time.Now().Format(time.RFC3339),
				}
			},
			Public: true,
		},
		{
			Name: "system.stats", // Stats action to return statistics on actions and modules
			Func: func(a act_type.IAction) any {
				// Log the stats action
				s.logClient.Info("System stats requested", map[string]interface{}{"action": "system.stats"})
				return map[string]any{
					"actions": len(act_core.Handlers),
					"modules": len(s.Module.GetActions()),
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
func NewSystemModule(service typ.IService, logClient log_type.ILogClient) *SystemModule {
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
	for _, action := range systemModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	// Log successful creation of module
	logClient.Info("System module created successfully", map[string]interface{}{"module": "sys"})

	return systemModule
}
