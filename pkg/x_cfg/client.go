package x_cfg

import (
	"fmt"
	"log"
	"sync"

	"github.com/rskv-p/mini/pkg/x_bus"
	"github.com/rskv-p/mini/pkg/x_req"
)

// Client represents a configuration client in the bus system.
type Client struct {
	id            uint64
	bus           *x_bus.Bus
	mu            sync.Mutex
	subs          map[string]func(string, []byte, string) // Map of subject to handler functions
	config        map[string]interface{}
	HandleMessage func(*x_req.Request) // Handle incoming messages
}

// NewClient creates a new config client associated with the bus.
func NewClient(id uint64, bus *x_bus.Bus) *Client {
	return &Client{
		id:     id,
		bus:    bus,
		subs:   make(map[string]func(string, []byte, string)),
		config: make(map[string]interface{}),
	}
}

// Publish sends a configuration update message to the bus.
func (c *Client) Publish(subject string, data []byte) error {
	return c.bus.Publish(subject, data)
}

// SetConfig sets a configuration value and publishes the update.
func (c *Client) SetConfig(key string, value interface{}) error {
	c.config[key] = value
	// Publish the "config.update" message to notify about the update
	return c.Publish("config.update", []byte(fmt.Sprintf("Updated config %s: %v", key, value)))
}

// GetConfig retrieves a configuration value by key.
func (c *Client) GetConfig(key string) (interface{}, error) {
	value, exists := c.config[key]
	if !exists {
		return nil, fmt.Errorf("config not found")
	}
	return value, nil
}

// Subscribe subscribes the client to a subject for config-related messages.
// Subscribe subscribes the client to a subject for log-related or config-related messages.
func (c *Client) Subscribe(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already subscribed, do not add again
	if _, exists := c.subs[subject]; exists {
		log.Printf("Client %d is already subscribed to subject %s", c.id, subject)
		return nil
	}

	// Add the subscription handler with the correct signature (subject, msg, reply)
	c.subs[subject] = func(subject string, msg []byte, reply string) {
		log.Printf("Client %d received message on subject %s: %s", c.id, subject, string(msg))

		// If HandleMessage is defined, call it with the message
		if c.HandleMessage != nil {
			c.HandleMessage(&x_req.Request{
				Subject: subject,
				Data:    msg,
			})
		}
	}

	// Log subscription
	log.Printf("Client %d subscribed to subject %s", c.id, subject)

	// Subscribe to the bus for this subject with an empty queue if queue is not necessary
	// Pass an empty string for queue if not needed
	return c.bus.SubscribeWithQueue(subject, "", c.subs[subject])
}

// Unsubscribe removes the client's subscription from a subject.
func (c *Client) Unsubscribe(subject string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the subscription handler
	if _, exists := c.subs[subject]; !exists {
		log.Printf("Client %d is not subscribed to subject %s", c.id, subject)
		return nil
	}

	// Remove the subscription handler
	delete(c.subs, subject)

	// Log unsubscription
	log.Printf("Client %d unsubscribed from config subject %s", c.id, subject)

	// Unsubscribe from the bus
	return c.bus.Unsubscribe(subject)
}

// HandleSubscribe processes incoming messages for all subscriptions.
// HandleSubscribe processes incoming messages for all subscriptions.
func (c *Client) HandleSubscribe() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Loop through subscriptions and invoke corresponding handlers
	for subject, handler := range c.subs {
		log.Printf("Client %d handling subscription for subject %s", c.id, subject)
		// Simulate receiving a message (you can replace this with actual message publishing)
		// Ensure the simulated message is relevant to the subscribed subject
		if subject == "log.message" && c.id == 2 { // logClient only handles log.message
			handler(subject, []byte(fmt.Sprintf("Simulated message for %s", subject)), "")
		} else if subject == "config.update" && c.id == 3 { // cfgClient only handles config.update
			handler(subject, []byte(fmt.Sprintf("Simulated message for %s", subject)), "")
		} else if subject == "db.create" && c.id == 1 { // dbClient only handles db.create
			handler(subject, []byte(fmt.Sprintf("Simulated message for %s", subject)), "")
		}
	}
}
