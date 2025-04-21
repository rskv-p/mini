package mod

import (
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Module
//---------------------

// Module represents a reusable component that registers actions and interacts with IService.
type Module struct {
	Name    string               // Module name
	Actions []act_type.ActionDef // List of actions associated with the module
	Service typ.IService         // Service the module interacts with
	OnInit  func() error         // Optional initialization function
	OnStart func() error         // Optional stop function
	OnStop  func() error         // Optional stop function
}

// Ensure Module implements the IModule interface.
var _ typ.IModule = (*Module)(nil)

//---------------------
// Module Lifecycle
//---------------------

// Name returns the name of the module.
func (m *Module) GetName() string {
	return m.Name
}

// Stop stops the module and performs cleanup if the OnStop function is provided.
func (m *Module) Init() error {
	// Call the stop function if provided
	if m.OnStop != nil {
		return m.OnInit()
	}
	return nil
}

// Stop stops the module and performs cleanup if the OnStop function is provided.
func (m *Module) Stop() error {
	// Call the stop function if provided
	if m.OnStop != nil {
		return m.OnStop()
	}
	return nil
}

// Start starts the module and performs any initialization or setup if the OnStart function is provided.
func (m *Module) Start() error {
	// Call the start function if provided
	if m.OnStart != nil {
		return m.OnStart()
	}
	return nil
}

//---------------------
// Actions
//---------------------

// Actions returns a list of actions associated with the module.
func (m *Module) GetActions() []act_type.ActionDef {
	return m.Actions
}

// AddAction adds a new action to the module.
func (m *Module) AddAction(name string, action act_type.Handler) {
	m.Actions = append(m.Actions, act_type.ActionDef{Name: name, Func: action})
}

//---------------------
// Module Creation
//---------------------

// NewModule creates a new instance of Module with the given name, service, actions, and optional OnInit and OnStop functions.
func NewModule(modName string, service typ.IService, actions []act_type.ActionDef, onInit func() error, onStop func() error) *Module {
	// Create a new module with the provided parameters
	module := &Module{
		Name:    modName,
		Actions: actions,
		Service: service,
		OnInit:  onInit,
		OnStop:  onStop,
	}

	// Call the initialization function if provided
	if module.OnInit != nil {
		if err := module.OnInit(); err != nil {
			return nil
		}
	}

	// Register actions for the module
	for _, a := range actions {
		act_core.Register(a.Name, a.Func)
	}

	return module
}
