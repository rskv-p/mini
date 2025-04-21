package typ

import "github.com/rskv-p/mini/mod/m_act/act_type"

// IModule defines the lifecycle of a module, including initialization and stopping.
type IModule interface {
	// Name returns the name of the module
	GetName() string
	Stop() error
	Start() error
	Init() error

	// Actions returns a list of actions associated with the module
	GetActions() []act_type.ActionDef
}

// IService defines the necessary methods for any service.
type IService interface {
	GetName() string             // Get the name of the service.
	Start() error                // Start the service and its modules.
	Stop() error                 // Stop the service and its modules.
	AddModule(mod IModule) error // Add a module to the service.
	//GetModules() []IModule
}
