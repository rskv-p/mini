package mod

import (
	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Types
//---------------------

// Module is a reusable component that can register actions.
type Module struct {
	ModName string
	Acts    []typ.ActionDef

	OnInit func() error
	OnStop func() error
}

//---------------------
// Lifecycle
//---------------------

// Name returns the module name.
func (m *Module) Name() string {
	return m.ModName
}

// Init initializes the module and registers actions.
func (m *Module) Init() error {
	// Register actions
	for _, a := range m.Acts {
		act.Register(a.Name, a.Func)
	}

	// Run custom init (if any)
	if m.OnInit != nil {
		return m.OnInit()
	}

	return nil
}

// Stop stops the module and executes custom stop logic (if any).
func (m *Module) Stop() error {
	// Run custom stop (if any)
	if m.OnStop != nil {
		return m.OnStop()
	}

	return nil
}

// Actions returns the list of actions available in the module.
func (m *Module) Actions() []typ.ActionDef {
	return m.Acts
}

//---------------------
// Module Registration
//---------------------

var modules = map[string]typ.IModule{}

// Register adds the module to the global registry.
func Register(m typ.IModule) {
	modules[m.Name()] = m
}

// GetModules returns all registered modules.
func GetModules() map[string]typ.IModule {
	return modules
}

// RegisterBuiltinModules registers all built-in modules.
func RegisterBuiltinModules() {
	// Registering built-in modules dynamically
	RegisterBuiltInModule("m_api", registerAPI)
	RegisterBuiltInModule("m_bus", registerBus)
	RegisterBuiltInModule("m_cfg", registerConfig)
	RegisterBuiltInModule("m_log", registerLog)
	RegisterBuiltInModule("m_rtm", registerRuntime)
}

// RegisterBuiltInModule registers a specific built-in module with initialization logic.
func RegisterBuiltInModule(modName string, initFunc func() error) {
	module := &Module{
		ModName: modName,
		OnInit:  initFunc,
	}

	// Dynamically register the module
	Register(module)
}

// InitAll initializes all modules.
func InitAll() error {
	for _, m := range modules {
		if err := m.Init(); err != nil {
			return err
		}
	}
	return nil
}

// StopAll stops all modules.
func StopAll() {
	for _, m := range modules {
		_ = m.Stop()
	}
}

//---------------------
// Built-in Module Registrations
//---------------------

// registerAPI registers actions for the API module.
func registerAPI() error {
	apiActions := []typ.ActionDef{
		{
			Name: "api.get_info",
			Func: func(a typ.IAction) any {
				// Logic to retrieve API info
				return "API Info"
			},
		},
	}

	// Register actions for m_api
	for _, action := range apiActions {
		act.Register(action.Name, action.Func)
	}

	return nil
}

// registerBus registers actions for the Bus module.
func registerBus() error {
	busActions := []typ.ActionDef{
		{
			Name: "bus.publish",
			Func: func(a typ.IAction) any {
				// Logic to publish a message
				return "Published to Bus"
			},
		},
	}

	// Register actions for m_bus
	for _, action := range busActions {
		act.Register(action.Name, action.Func)
	}

	return nil
}

// registerConfig registers actions for the Config module.
func registerConfig() error {
	configActions := []typ.ActionDef{
		{
			Name: "config.get",
			Func: func(a typ.IAction) any {
				// Logic to get config value
				return "Config Value"
			},
		},
	}

	// Register actions for m_cfg
	for _, action := range configActions {
		act.Register(action.Name, action.Func)
	}

	return nil
}

// registerLog registers actions for the Log module.
func registerLog() error {
	logActions := []typ.ActionDef{
		{
			Name: "log.info",
			Func: func(a typ.IAction) any {
				// Logic to log info
				return "Log Info"
			},
		},
	}

	// Register actions for m_log
	for _, action := range logActions {
		act.Register(action.Name, action.Func)
	}

	return nil
}

// registerRuntime registers actions for the Runtime module.
func registerRuntime() error {
	runtimeActions := []typ.ActionDef{
		{
			Name: "runtime.execute",
			Func: func(a typ.IAction) any {
				// Logic to execute runtime actions
				return "Runtime Executed"
			},
		},
	}

	// Register actions for m_rtm
	for _, action := range runtimeActions {
		act.Register(action.Name, action.Func)
	}

	return nil
}
