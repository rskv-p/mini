package typ

import (
	"context"

	"gorm.io/gorm"
)

//---------------------
// Action and Handler
//---------------------

// Handler defines the function signature for handlers that process actions.
type Handler func(IAction) any

// ActionDef represents a definition of an action, including its name and handler.
type ActionDef struct {
	Name   string  // Action name
	Func   Handler // Function to handle the action
	Public bool    // Future use (e.g., to expose the action via an API)
}

//---------------------
// Interfaces
//---------------------

// IModule defines the lifecycle of a module, including initialization and stopping.
type IModule interface {
	// Name returns the name of the module
	Name() string

	// Init initializes the module
	Init() error

	// Stop stops the module
	Stop() error

	// Actions returns a list of actions associated with the module
	Actions() []ActionDef
}

// IAction defines the core behavior that an action must implement.
type IAction interface {
	// GetName returns the name of the action
	GetName() string

	// WithContext sets and returns the context associated with the action
	WithContext(ctx context.Context) IAction

	// Context returns the context associated with the action
	Context() context.Context

	// Execute runs the action and returns any error encountered
	Execute() (err error)

	// Exec runs the action and returns the result or error
	Exec() (any, error)

	// Run runs the action and returns the result
	Run() any

	// Release releases any resources used by the action
	Release()

	// Dispose clears the inputs and results of the action
	Dispose()

	// Output returns the result of the action after execution
	Output() any

	// String returns a string representation of the action
	String() string

	// ValidateInputsNumber checks if the number of inputs is sufficient
	ValidateInputsNumber(length int) error

	// NumberOfInputs returns the number of inputs to the action
	NumberOfInputs() int

	// NumberOfInputsIs checks if the number of inputs matches the specified number
	NumberOfInputsIs(num int) bool

	// InputNotNull ensures that a specific input is not nil
	InputNotNull(i int) error

	// InputString parses a string input at the specified index
	InputString(i int, defaults ...string) string

	// InputInt parses an integer input at the specified index
	InputInt(i int, defaults ...int) int

	// InputUint32 parses a uint32 input at the specified index
	InputUint32(i int, defaults ...uint32) uint32

	// InputBool parses a boolean input at the specified index
	InputBool(i int, defaults ...bool) bool

	// InputURL parses a URL input at the specified index and retrieves a specific key
	InputURL(i int, key string, defaults ...string) string

	// InputMap parses an input as a map
	InputMap(i int) map[string]any

	// InputArray parses an input as an array
	InputArray(i int) []any

	// InputStrings parses an input as an array of strings
	InputStrings(i int) []string

	// InputsRecords parses an input as an array of records (maps)
	InputsRecords(i int) []map[string]any
}

//---------------------
// Subscription
//---------------------

// Subscription represents a subscription to a subject with an associated client and queue.
type Subscription struct {
	Subject []byte  // The subject to which the client is subscribed
	Queue   []byte  // The queue associated with the subscription
	Client  IClient // The client associated with the subscription
}

// NewSubscription creates a new subscription with the provided subject, queue, and client.
func NewSubscription(subject, queue []byte, client IClient) *Subscription {
	return &Subscription{
		Subject: subject,
		Queue:   queue,
		Client:  client,
	}
}

//---------------------
// IClient
//---------------------

// IClient defines the behavior of a client that can subscribe, publish, and handle messages.
type IClient interface {
	// Subscribe subscribes the client to the specified subject
	Subscribe(subject string) error

	// Unsubscribe removes the client's subscription to the specified subject
	Unsubscribe(subject string)

	// Deliver sends a message to the client
	Deliver(subject string, data []byte)

	// Publish sends a message to the bus
	Publish(subject string, data []byte) error

	// PublishWithReply sends a message to the bus and expects a reply
	PublishWithReply(subject string, data []byte, reply string)

	// SubscribeWithQueue subscribes the client to the specified subject with a queue and handler
	SubscribeWithQueue(subject string, queue string, handler func(subject string, msg []byte)) error

	// GetClientCount returns the total number of clients
	GetClientCount() int

	// GetMsgHandlerCount returns the total number of message handlers
	GetMsgHandlerCount() int
}

//---------------------
// IMessageSender
//---------------------

// IMessageSender defines the behavior of a sender that can publish messages.
type IMessageSender interface {
	// Publish sends a message to the bus
	Publish(subject string, msg []byte) error

	// PublishWithReply sends a message to the bus and expects a reply
	PublishWithReply(subject string, msg []byte, reply string) error
}

//---------------------
// ISubscriber
//---------------------

// ISubscriber defines the behavior of a subscriber that can subscribe to messages.
type ISubscriber interface {
	// Subscribe subscribes to the specified subject
	Subscribe(subject string, handler func(subject string, msg []byte)) error

	// Unsubscribe removes the subscription to the specified subject
	Unsubscribe(subject string) error
}

//---------------------
// IEventPublisher
//---------------------

// IEventPublisher defines the behavior of a publisher that can send events.
type IEventPublisher interface {
	// PublishEvent sends an event to the event bus
	PublishEvent(event Event) error
}

//---------------------
// IEventSubscriber
//---------------------

// IEventSubscriber defines the behavior of a subscriber that can subscribe to events.
type IEventSubscriber interface {
	// SubscribeEvent subscribes to events with the specified name
	SubscribeEvent(eventName string, handler func(event Event)) error
}

//---------------------
// Event
//---------------------

// Event represents an event with a name and a payload.
type Event struct {
	Name    string      // Event name
	Payload interface{} // Event data
}

//---------------------
// Logging and Configuration Clients
//---------------------

// ILogClient defines methods for logging various levels of logs.
type ILogClient interface {
	Trace(message string)
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string)
}

// IConfigClient defines an interface for accessing and manipulating configurations.
type IConfigClient interface {
	GetConfig(key string) any
	SetConfig(key string, value any) error
	DeleteConfig(key string) error
	PublishConfig(key string, value any) error
}

//---------------------
// Runtime Client
//---------------------

// RuntimeClient defines the contract for interacting with external systems from within a runtime.
type RuntimeClient interface {
	ExecuteAction(action IAction) (any, error) // Execute an action in the runtime.
}

//---------------------
// Database Client
//---------------------

// IDBClient defines methods for interacting with a database.
type IDBClient interface {
	// GetDB returns the database connection
	GetDB() (*gorm.DB, error)

	// Migrate applies the database migrations for the provided models
	Migrate(models ...interface{}) error

	// First finds the first record that matches the conditions
	First(dest interface{}, conds ...interface{}) *gorm.DB

	// Create inserts a new record into the database
	Create(model interface{}) error

	// Find retrieves records that match the conditions
	Find(model interface{}, conditions ...interface{}) error
}
