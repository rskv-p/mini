package db_client

import (
	"fmt"
	"log"
	"sync"

	"github.com/rskv-p/mini/mod/m_bus/bus_client"
	"github.com/rskv-p/mini/mod/m_bus/bus_core"
	"github.com/rskv-p/mini/mod/m_bus/bus_req"
)

// Client represents a client for DB operations in the bus system.
type Client struct {
	bus_client.Client
	id            uint64
	bus           *bus_core.Bus
	mu            sync.Mutex
	subs          map[string]func(string, []byte)
	dbStorage     map[string]interface{} // Simulating a simple in-memory DB
	HandleMessage func(*bus_req.Request) // Field to handle incoming messages
}

// NewClient creates a new DB client associated with the bus.
func NewClient(id uint64, bus *bus_core.Bus) *Client {
	return &Client{
		id:        id,
		bus:       bus,
		subs:      make(map[string]func(string, []byte)),
		dbStorage: make(map[string]interface{}),
	}
}

// Publish sends a message to the bus indicating a database operation.
func (c *Client) Publish(subject string, data []byte) error {
	return c.bus.Publish(subject, data)
}

// Create simulates a DB create operation and publishes a message.
func (c *Client) Create(model interface{}, key string) error {
	c.dbStorage[key] = model
	// Publish the "db.create" message to the bus
	return c.Publish("db.create", []byte(fmt.Sprintf("Created model %v with key %s", model, key)))
}

// Find retrieves a model from the simulated DB.
func (c *Client) Find(key string) (interface{}, error) {
	model, exists := c.dbStorage[key]
	if !exists {
		return nil, fmt.Errorf("model not found")
	}
	return model, nil
}

// Subscribe subscribes the client to a subject for DB-related messages.
func (c *Client) Subscribe(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add subscription handler for the given subject
	c.subs[subject] = func(subject string, msg []byte) {
		log.Printf("Client %d received DB message on subject %s: %s", c.id, subject, string(msg))
		// If HandleMessage is defined, process the message
		if c.HandleMessage != nil {
			c.HandleMessage(&bus_req.Request{
				Subject: subject,
				Data:    msg,
			})
		}
	}

	log.Printf("Client %d subscribed to DB subject %s", c.id, subject)
	return nil
}

// HandleSubscribe processes incoming messages for subscriptions.
func (c *Client) HandleSubscribe() {
	// Loop through subscriptions and invoke corresponding handlers
	for subject, handler := range c.subs {
		log.Printf("Handling subscription for subject %s", subject)
		// Here you would typically invoke your message handler
		// For example, simulate receiving a message
		handler(subject, []byte(fmt.Sprintf("Simulated message for %s", subject)))
	}
}
