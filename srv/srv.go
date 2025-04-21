package srv

import (
	"fmt"
	"log"

	"github.com/rskv-p/mini/mod/m_api/api_mod"
	"github.com/rskv-p/mini/mod/m_auth/auth_mod"
	"github.com/rskv-p/mini/mod/m_bus/bus_mod"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_mod"
	"github.com/rskv-p/mini/mod/m_cfg/cfg_type"
	"github.com/rskv-p/mini/mod/m_db/db_mod"
	"github.com/rskv-p/mini/mod/m_db/db_type"
	"github.com/rskv-p/mini/mod/m_log/log_mod"
	"github.com/rskv-p/mini/mod/m_log/log_type"
	"github.com/rskv-p/mini/mod/m_rtm/rtm_mod"
	"github.com/rskv-p/mini/mod/m_sys/sys_mod"
	"github.com/rskv-p/mini/typ"
)

// Service struct represents the main service.
type Service struct {
	Name    string
	Modules []typ.IModule // List of modules registered in the service
}

// New creates a new service and automatically registers built-in services.
func New(name string, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) *Service {
	s := &Service{
		Name:    name,
		Modules: []typ.IModule{},
	}

	// Register built-in modules here (this can be extended later)
	if err := RegisterBuildInModules(s, dbClient, logClient, busClient, cfgClient); err != nil {
		// Handle error registering built-in modules
		log.Fatalf("Error registering built-in modules: %v", err)
	}

	return s
}

// Start initializes all modules and registers their actions.
func (s *Service) GetName() string {
	return s.Name
}

// Start initializes all modules and registers their actions.
func (s *Service) Start() error {
	// Initialize and start all modules
	for _, mod := range s.Modules {
		// Ensure each module's start method is called (if implemented)
		if err := mod.Start(); err != nil {
			return fmt.Errorf("error starting module '%s': %v", mod.GetName(), err)
		}
	}
	return nil
}

// Stop gracefully stops all modules.
func (s *Service) Stop() error {
	for _, mod := range s.Modules {
		if err := mod.Stop(); err != nil {
			return fmt.Errorf("error stopping module '%s': %v", mod.GetName(), err)
		}
	}
	return nil
}

// AddModule adds a new module to the service.
func (s *Service) AddModule(mod typ.IModule) error {
	s.Modules = append(s.Modules, mod)
	return nil
}

//---------------------
// Module Management
//---------------------

// RegisterModule registers a new module by adding it to the service.
func (s *Service) RegisterModule(mod typ.IModule) error {
	return s.AddModule(mod)
}

// GetModules returns the list of modules currently registered in the service.
func (s *Service) GetModules() []typ.IModule {
	return s.Modules
}

//---------------------
// BuildInModules
//---------------------

var BuildInModules = []func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule{
	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return api_mod.NewApiModule(service)
	}, // m_api

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return auth_mod.NewMAuthModule(service, dbClient, logClient) // m_auth
	},

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return cfg_mod.NewConfigModule(service, nil, logClient) // m_cfg
	},

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return bus_mod.NewBusModule(service, nil, logClient) // m_bus
	},

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return db_mod.NewDBModule(service, nil, logClient) // m_db
	},

	// This is where the types need to match: use busClient and cfgClient for log_mod.NewLogModule
	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return log_mod.NewLogModule(service, busClient, cfgClient) // m_log
	},

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return rtm_mod.NewRuntimeModule(service, logClient) // m_rtm
	},

	func(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) typ.IModule {
		return sys_mod.NewSystemModule(service, logClient) // m_sys
	},

	// Add other modules here similarly
}

// RegisterBuildInModules registers all built-in modules.
// RegisterBuildInModules registers all built-in modules.
func RegisterBuildInModules(service typ.IService, dbClient db_type.IDBClient, logClient log_type.ILogClient, busClient bus_type.IBusClient, cfgClient cfg_type.IConfigClient) error {
	for _, newModule := range BuildInModules {
		// Create the module by passing the necessary clients and the IService
		module := newModule(service, dbClient, logClient, busClient, cfgClient)
		if err := service.AddModule(module); err != nil {
			return err
		}
	}
	return nil
}
