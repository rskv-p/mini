package log_client

import (
	"fmt"
	"log"
	"sync"

	"github.com/rskv-p/mini/mod/m_bus/bus_core"
	"github.com/rskv-p/mini/mod/m_bus/bus_req"
)

// Client represents a log client in the bus system.
type Client struct {
	id            uint64
	bus           *bus_core.Bus
	mu            sync.Mutex
	subs          map[string]func(string, []byte, string) // Map of subject to handler functions
	HandleMessage func(*bus_req.Request)                  // Handle incoming messages
}

// NewClient creates a new log client associated with the bus.
func NewClient(id uint64, bus *bus_core.Bus) *Client {
	return &Client{
		id:   id,
		bus:  bus,
		subs: make(map[string]func(string, []byte, string)),
	}
}

// Publish sends a log message to the bus.
func (c *Client) Publish(subject string, data []byte) error {
	return c.bus.Publish(subject, data)
}

// Log sends a log message with a given level to the bus.
func (c *Client) Log(level string, message string) {
	logMessage := fmt.Sprintf("[%s] %s", level, message)
	// Publish the log message on the "log.message" subject
	if err := c.Publish("log.message", []byte(logMessage)); err != nil {
		log.Printf("Error publishing log message: %v", err)
	}
}

// Subscribe subscribes the client to a subject for log-related messages.
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
			c.HandleMessage(&bus_req.Request{
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
	log.Printf("Client %d unsubscribed from log subject %s", c.id, subject)

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
