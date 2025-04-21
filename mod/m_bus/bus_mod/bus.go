package bus_mod

import (
	"fmt"

	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/mod/m_act/act_core"
	"github.com/rskv-p/mini/mod/m_act/act_type"
	"github.com/rskv-p/mini/mod/m_bus/bus_core"
	"github.com/rskv-p/mini/mod/m_log/log_type"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Error Handling
//---------------------

// ErrInvalidActionType is the error returned when an action cannot be cast to *act_core.Action.
var ErrInvalidActionType = fmt.Errorf("invalid action type")

//---------------------
// Bus Module
//---------------------

// BusModule manages the bus system for publishing and subscribing to messages.
type BusModule struct {
	Module    typ.IModule         // The module implementing IModule interface
	Bus       *bus_core.Bus       // The bus object for handling messages
	logClient log_type.ILogClient // Client for logging
}

// Ensure BusModule implements the IModule interface.
var _ typ.IModule = (*BusModule)(nil)

// Name returns the name of the module.
func (m *BusModule) GetName() string {
	return m.Module.GetName()
}

// Stop stops the module (specific shutdown logic can be added).
func (s *BusModule) Start() error {
	// Here you can implement stop logic if needed
	return nil
}

// Stop stops the module (specific shutdown logic can be added).
func (s *BusModule) Init() error {
	// Here you can implement stop logic if needed
	return nil
}

// Stop stops the module and logs the shutdown process.
func (m *BusModule) Stop() error {
	m.logClient.Info("Stopping Bus module", map[string]interface{}{"module": m.GetName()})
	return nil
}

// Actions returns a list of actions that the bus module can perform.
func (m *BusModule) GetActions() []act_type.ActionDef {
	return []act_type.ActionDef{
		{
			Name:   "bus.publish",
			Func:   HandlePublish,
			Public: true,
		},
		{
			Name:   "bus.subscribe",
			Func:   HandleSubscribe,
			Public: true,
		},
		{
			Name:   "bus.stats",
			Func:   HandleStats, // Bus statistics action
			Public: true,
		},
	}
}

//---------------------
// Module Creation
//---------------------

// NewBusModule creates a new instance of BusModule and initializes it.
func NewBusModule(service typ.IService, bus *bus_core.Bus, logClient log_type.ILogClient) *BusModule {
	// Create a new module using NewModule
	module := mod.NewModule("bus", service, nil, nil, nil)

	// Return the BusModule with the created module
	busModule := &BusModule{
		Module:    module,
		Bus:       bus,
		logClient: logClient,
	}

	// Register actions for the bus module
	for _, action := range busModule.GetActions() {
		act_core.Register(action.Name, action.Func)
	}

	return busModule
}

//---------------------
// Helper Functions
//---------------------

// handleActionCast safely casts IAction to *act_core.Action and handles errors.
func handleActionCast(a act_type.IAction) (*act_core.Action, error) {
	action, ok := a.(*act_core.Action)
	if !ok {
		return nil, ErrInvalidActionType
	}
	return action, nil
}

//---------------------
// Handlers for Actions
//---------------------

// HandlePublish handles the action for publishing a message to the bus.
func HandlePublish(a act_type.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act_core.Action", map[string]interface{}{"error": err})
		return err
	}

	subject := action.InputString(0)
	msg, ok := action.Inputs[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid message type for publish: expected []byte")
	}

	// Publish the message to the bus
	err = action.Bus.Publish(subject, msg)
	if err != nil {
		// Log the error
		//action.logClient.Error(fmt.Sprintf("Failed to publish message to subject %s", subject), map[string]interface{}{"subject": subject, "error": err})
		return err
	}

	// Log success
	//action.logClient.Info(fmt.Sprintf("Successfully published message to subject %s", subject), map[string]interface{}{"subject": subject})
	return true
}

// HandleSubscribe handles the action for subscribing to a subject on the bus.
func HandleSubscribe(a act_type.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act_core.Action", map[string]interface{}{"error": err})
		return err
	}

	subject := action.InputString(0)

	// Subscribe to the subject on the bus
	err = action.Bus.Subscribe(subject)
	if err != nil {
		// Log the error
		//action.logClient.Error(fmt.Sprintf("Failed to subscribe to subject %s", subject), map[string]interface{}{"subject": subject, "error": err})
		return err
	}

	// Log success
	//action.logClient.Info(fmt.Sprintf("Successfully subscribed to subject %s", subject), map[string]interface{}{"subject": subject})
	return true
}

// HandleStats returns the bus statistics, including the number of clients and subscriptions.
func HandleStats(a act_type.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act_core.Action", map[string]interface{}{"error": err})
		return err
	}

	// Get and return the bus statistics
	stats := map[string]interface{}{
		"total_clients":       action.Bus.GetClientCount(),
		"total_subscriptions": action.Bus.GetMsgHandlerCount(),
	}

	// Log the statistics
	//action.logClient.Info("Bus stats", map[string]interface{}{"stats": stats})
	return stats
}
