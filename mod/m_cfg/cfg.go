package m_cfg

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_cfg"
	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Config Module
//---------------------

// ConfigModule creates and returns a config module with actions for getting, setting, deleting, and reloading configurations.
func ConfigModule(client typ.IConfigClient) typ.IModule {
	return &mod.Module{
		ModName: "config",
		Acts: []typ.ActionDef{
			// Get configuration
			{Name: "config.get", Func: HandleGet, Public: true},
			// Set configuration
			{Name: "config.set", Func: HandleSet},
			// Delete configuration
			{Name: "config.del", Func: HandleDel},
			// Get all configurations
			{Name: "config.all", Func: HandleAll, Public: true},
			// Reload configurations
			{Name: "config.reload", Func: HandleReload},
			// Publish configuration through the bus
			{
				Name: "config.publish",
				Func: func(a typ.IAction) any {
					action, err := castAction(a)
					if err != nil {
						return err
					}

					// Extract the key and value from the action
					key := action.InputString(0)
					val := action.Inputs[1]

					// Publish the configuration through the client (using IConfigClient)
					err = client.PublishConfig(key, val)
					if err != nil {
						x_log.RootLogger().Errorf("Failed to publish config '%s': %v", key, err)
						return err
					}

					// Publish an event on the bus about the configuration update
					event := typ.Event{
						Name:    "config.updated",
						Payload: map[string]any{"key": key, "value": val},
					}
					err = client.PublishConfig("events", event)
					if err != nil {
						x_log.RootLogger().Errorf("Failed to publish config update event: %v", err)
						return err
					}

					return true
				},
				Public: true,
			},
		},
		OnInit: func() error {
			// Reload configuration on module initialization
			return x_cfg.Reload()
		},
	}
}

//---------------------
// Helper Functions
//---------------------

// castAction safely casts the IAction to *act.Action and returns an error if casting fails.
func castAction(a typ.IAction) (*act.Action, error) {
	action, ok := a.(*act.Action)
	if !ok {
		return nil, fmt.Errorf("Invalid action type: expected *act.Action")
	}
	return action, nil
}

//---------------------
// Handlers for Actions
//---------------------

// HandleGet retrieves the configuration value for the given key.
func HandleGet(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	key := action.InputString(0)
	return x_cfg.Get(key)
}

// HandleSet sets a configuration key-value pair.
func HandleSet(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	key := action.InputString(0)
	val := action.Inputs[1]
	x_cfg.Set(key, val)
	return true
}

// HandleDel deletes the configuration for the given key.
func HandleDel(a typ.IAction) any {
	action, err := castAction(a)
	if err != nil {
		return err
	}
	key := action.InputString(0)
	x_cfg.Delete(key)
	return true
}

// HandleAll returns all configurations.
func HandleAll(a typ.IAction) any {
	return x_cfg.All()
}

// HandleReload reloads the configuration.
func HandleReload(a typ.IAction) any {
	return x_cfg.Reload() == nil
}
