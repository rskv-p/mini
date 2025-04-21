package m_bus

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_bus"
	"github.com/rskv-p/mini/typ"
)

//---------------------
// Error Handling
//---------------------

// ErrInvalidActionType is the error returned when an action cannot be cast to *act.Action.
var ErrInvalidActionType = fmt.Errorf("invalid action type")

//---------------------
// Bus Module
//---------------------

// BusModule manages the bus system for publishing and subscribing to messages.
type BusModule struct {
	Module    typ.IModule    // The module implementing IModule interface
	Bus       *x_bus.Bus     // The bus object for handling messages
	logClient typ.ILogClient // Client for logging
}

// Ensure BusModule implements the IModule interface.
var _ typ.IModule = (*BusModule)(nil)

// Name returns the name of the module.
func (m *BusModule) Name() string {
	return m.Module.Name()
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
	m.logClient.Info("Stopping Bus module", map[string]interface{}{"module": m.Name()})
	return nil
}

// Actions returns a list of actions that the bus module can perform.
func (m *BusModule) Actions() []typ.ActionDef {
	return []typ.ActionDef{
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
func NewBusModule(service typ.IService, bus *x_bus.Bus, logClient typ.ILogClient) *BusModule {
	// Create a new module using NewModule
	module := mod.NewModule("bus", service, nil, nil, nil)

	// Return the BusModule with the created module
	busModule := &BusModule{
		Module:    module,
		Bus:       bus,
		logClient: logClient,
	}

	// Register actions for the bus module
	for _, action := range busModule.Actions() {
		act.Register(action.Name, action.Func)
	}

	return busModule
}

//---------------------
// Helper Functions
//---------------------

// handleActionCast safely casts IAction to *act.Action and handles errors.
func handleActionCast(a typ.IAction) (*act.Action, error) {
	action, ok := a.(*act.Action)
	if !ok {
		return nil, ErrInvalidActionType
	}
	return action, nil
}

//---------------------
// Handlers for Actions
//---------------------

// HandlePublish handles the action for publishing a message to the bus.
func HandlePublish(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act.Action", map[string]interface{}{"error": err})
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
func HandleSubscribe(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act.Action", map[string]interface{}{"error": err})
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
func HandleStats(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		// Log the error
		//action.logClient.Error("Failed to cast action to *act.Action", map[string]interface{}{"error": err})
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
