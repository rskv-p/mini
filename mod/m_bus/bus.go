package m_bus

import (
	"fmt"

	"github.com/rskv-p/mini/act"
	"github.com/rskv-p/mini/mod"
	"github.com/rskv-p/mini/pkg/x_bus"
	"github.com/rskv-p/mini/pkg/x_log"
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

// BusModule creates and returns a bus module with actions for publish, subscribe, and stats.
func BusModule(bus *x_bus.Bus) typ.IModule {
	return &mod.Module{
		ModName: "bus",
		Acts: []typ.ActionDef{
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
				Func:   HandleStats, // Bus statistics
				Public: true,
			},
		},
		OnInit: func() error {
			// Initialize the bus module (no specific initialization required here)
			return nil
		},
		OnStop: func() error {
			// Stop the bus module (no specific stopping required here)
			return nil
		},
	}
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

// HandlePublish processes the action to publish a message through the bus.
func HandlePublish(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		x_log.RootLogger().Errorf("Failed to cast action to *act.Action: %v", err)
		return err
	}

	subject := action.InputString(0)
	msg, ok := action.Inputs[1].([]byte)
	if !ok {
		return fmt.Errorf("invalid message type for publish: expected []byte")
	}

	// Publish the message through the bus
	err = action.Bus.Publish(subject, msg)
	if err != nil {
		x_log.RootLogger().Errorf("Failed to publish message to subject %s: %v", subject, err)
		return err
	}

	x_log.RootLogger().Infof("Successfully published message to subject %s", subject)
	return true
}

// HandleSubscribe processes the action to subscribe to a topic.
func HandleSubscribe(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		x_log.RootLogger().Errorf("Failed to cast action to *act.Action: %v", err)
		return err
	}

	subject := action.InputString(0)

	// Subscribe to the topic through the bus
	err = action.Bus.Subscribe(subject)
	if err != nil {
		x_log.RootLogger().Errorf("Failed to subscribe to subject %s: %v", subject, err)
		return err
	}

	x_log.RootLogger().Infof("Successfully subscribed to subject %s", subject)
	return true
}

// HandleStats returns the bus statistics.
func HandleStats(a typ.IAction) any {
	action, err := handleActionCast(a)
	if err != nil {
		x_log.RootLogger().Errorf("Failed to cast action to *act.Action: %v", err)
		return err
	}

	// Retrieve and return bus statistics
	stats := map[string]interface{}{
		"total_clients":       action.Bus.GetClientCount(),
		"total_subscriptions": action.Bus.GetMsgHandlerCount(),
	}

	x_log.RootLogger().Infof("Bus stats: %v", stats)
	return stats
}
