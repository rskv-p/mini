package x_bus

import (
	"log"
	"sync"

	"github.com/rskv-p/mini/pkg/x_req"
	"github.com/rskv-p/mini/typ"
)

// Client represents a client in the bus system.
type Client struct {
	id             uint64
	bus            *Bus
	mu             sync.Mutex
	subs           map[string]func(string, []byte) // Map of subscriptions with handlers
	HandleMessage  func(*x_req.Request)
	OnSubscribe    func(string)
	OnUnsubscribe  func(string)
	remoteInterest map[string]struct{} // Remote interest map for subjects
}

// NewClient creates a new client associated with the bus.
func NewClient(id uint64, bus *Bus) *Client {
	return &Client{
		id:             id,
		bus:            bus,
		subs:           make(map[string]func(string, []byte)),
		remoteInterest: make(map[string]struct{}),
	}
}

// SubscribeWithQueue subscribes the client to a subject with a queue and message handler.
func (c *Client) SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add subscription with handler
	c.subs[subject] = handler

	// Invoke OnSubscribe callback if provided
	if c.OnSubscribe != nil {
		c.OnSubscribe(subject)
	}

	log.Printf("Client %d subscribed to %s with queue %s", c.id, subject, queue)
	return nil
}

// Unsubscribe removes the client's subscription to a subject.
func (c *Client) Unsubscribe(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove subscription
	delete(c.subs, subject)

	// Invoke OnUnsubscribe callback if provided
	if c.OnUnsubscribe != nil {
		c.OnUnsubscribe(subject)
	}

	log.Printf("Client %d unsubscribed from %s", c.id, subject)
}

// Deliver sends a message to the client, invoking the HandleMessage callback.
func (c *Client) Deliver(subject string, data []byte) {
	if c.HandleMessage != nil {
		c.HandleMessage(&x_req.Request{
			Subject: subject,
			Data:    data,
		})
	} else {
		log.Printf("Warning: No handler defined for client %d on subject %s", c.id, subject)
	}
}

// GetClientCount returns the number of clients in the bus.
func (c *Client) GetClientCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.bus.clients)
}

// GetMsgHandlerCount returns the number of registered message handlers.
func (c *Client) GetMsgHandlerCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.bus.msgHandlers)
}

// PublishWithReply sends a message to the client with a reply subject.
func (c *Client) PublishWithReply(subject string, data []byte, reply string) {
	if c.HandleMessage != nil {
		c.HandleMessage(&x_req.Request{
			Subject: subject,
			Data:    data,
			Reply:   reply,
		})
	} else {
		log.Printf("Warning: No handler defined for reply on client %d for subject %s", c.id, subject)
	}
}

// markRemoteInterest adds a subject to the client's remote interest map.
func (c *Client) markRemoteInterest(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remoteInterest[subject] = struct{}{}
}

// unmarkRemoteInterest removes a subject from the client's remote interest map.
func (c *Client) unmarkRemoteInterest(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.remoteInterest, subject)
}

// hasMatchingInterest checks if the client has an interest in a subject.
func (c *Client) hasMatchingInterest(subject string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.remoteInterest[subject]
	return ok
}

// Subscribe subscribes the client to a subject without a queue.
func (c *Client) Subscribe(subject string) error {
	// Subscribe to the subject with an empty queue
	return c.SubscribeWithQueue(subject, "", func(subject string, msg []byte) {
		// For example, print the message to the console
		log.Printf("Client %d received message on topic %s: %s", c.id, subject, string(msg))
	})
}

// SubscribeEvent subscribes to an event by name.
func (c *Client) SubscribeEvent(eventName string, handler func(typ.Event)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Add event handler for the event
	c.bus.eventHandlers[eventName] = append(c.bus.eventHandlers[eventName], handler)

	log.Printf("Client %d subscribed to event %s", c.id, eventName)
}

// Publish sends a message to a subject via the bus.
func (c *Client) Publish(subject string, data []byte) error {
	return c.bus.Publish(subject, data)
}
