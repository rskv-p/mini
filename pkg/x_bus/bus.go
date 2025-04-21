package x_bus

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/rskv-p/mini/pkg/x_log"
	"github.com/rskv-p/mini/pkg/x_req"
	"github.com/rskv-p/mini/pkg/x_sub"
	"github.com/rskv-p/mini/typ"
)

// Error represents a custom error type for the Bus system.
type BusError struct {
	Code    int
	Message string
}

func (e *BusError) Error() string {
	return fmt.Sprintf("Bus Error %d: %s", e.Code, e.Message)
}

// Middleware represents a function that processes a request before it's handled.
type Middleware func(*x_req.Request) error

// Bus represents the message bus, handling clients, subscriptions, and message routing.
type Bus struct {
	mu            sync.Mutex
	selfClient    *Client
	clients       map[uint64]*Client
	sublist       *x_sub.Sublist
	transform     *x_sub.SubjectTransform
	authEnabled   bool
	secretKey     string
	msgHandlers   map[string]func(subject string, msg []byte)
	eventHandlers map[string][]func(event typ.Event)
	leaves        []*Leaf
	middlewares   []Middleware
}

// NewBus creates and initializes a new Bus instance.
func NewBus(authEnabled bool, secretKey string) *Bus {
	bus := &Bus{
		clients:       make(map[uint64]*Client),
		sublist:       x_sub.NewSublist(100),
		msgHandlers:   make(map[string]func(string, []byte)),
		eventHandlers: make(map[string][]func(typ.Event)),
		authEnabled:   authEnabled,
		secretKey:     secretKey,
		leaves:        make([]*Leaf, 0),
	}
	bus.selfClient = NewClient(rand.Uint64(), bus)
	bus.AddClient(bus.selfClient)
	return bus
}

// Use adds middleware to the bus for processing requests.
func (b *Bus) Use(mw Middleware) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middlewares = append(b.middlewares, mw)
}

// AddClient adds a client to the bus and starts listening for messages.
func (b *Bus) AddClient(c *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c.id] = c
}

// RemoveClient removes a client from the bus, cleaning up subscriptions and handlers.
func (b *Bus) RemoveClient(c *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, c.id)

	for subject := range c.subs {
		b.sublist.Remove(&typ.Subscription{
			Subject: []byte(subject),
			Client:  c,
		})

		if c.OnUnsubscribe != nil {
			c.OnUnsubscribe(subject)
		}

		x_log.RootLogger().Structured().Info("bus unsubscribed due to client removal",
			x_log.FAny("client_id", c.id),
			x_log.FString("subject", subject),
		)
	}
}

// AddLeaf adds a leaf node to the bus.
func (b *Bus) AddLeaf(leaf *Leaf) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.leaves = append(b.leaves, leaf)
}

//---------------------
// Auth
//---------------------

// Authenticate validates the provided JWT token if authentication is enabled.
func (b *Bus) Authenticate(token string) error {
	if !b.authEnabled {
		return nil
	}
	claims, err := verifyJWT(token, b.secretKey)
	if err != nil || (*claims)["exp"] == nil || (*claims)["sub"] == nil {
		return &BusError{Code: 401, Message: "invalid token"}
	}
	return nil
}

//---------------------
// Publish / Respond
//---------------------

// Publish publishes a message to the bus, notifying all matching subscribers.
func (b *Bus) Publish(subject string, msg []byte) error {
	b.mu.Lock()
	subs := b.sublist.Match([]byte(subject)).Psubs
	b.mu.Unlock()

	// If it's an event, handle it as an event
	if strings.HasPrefix(subject, "event.") {
		return b.PublishEvent(typ.Event{
			Name:    subject,
			Payload: msg,
		})
	}

	// Regular message publication
	for _, sub := range subs {
		if sub.Client != nil {
			sub.Client.Deliver(subject, msg)
		}
	}
	return nil
}

// PublishWithReply publishes a message and waits for a reply.
func (b *Bus) PublishWithReply(subject string, msg []byte, reply string) error {
	b.mu.Lock()
	subs := b.sublist.Match([]byte(subject)).Psubs
	b.mu.Unlock()

	// If it's an event, handle it as an event
	if strings.HasPrefix(subject, "event.") {
		return b.PublishEvent(typ.Event{
			Name:    subject,
			Payload: msg,
		})
	}

	// Regular message publication with reply
	for _, sub := range subs {
		if sub.Client != nil {
			sub.Client.PublishWithReply(subject, msg, reply)
		}
	}
	return nil
}

// Respond sends a response to the client.
func (b *Bus) Respond(reply string, data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	matched := false
	if res := b.sublist.Match([]byte(reply)); len(res.Psubs) > 0 {
		for _, sub := range res.Psubs {
			if sub.Client != nil {
				sub.Client.Deliver(reply, data)
				matched = true
			}
		}
	}

	if matched {
		return nil
	}

	anyForwarded := false
	for _, leaf := range b.leaves {
		x_log.RootLogger().Structured().Info("forwarding RESP to leaf",
			x_log.FString("reply", reply),
			x_log.FBinary("data", data),
		)
		leaf.SendResp(reply, data)
		anyForwarded = true
	}

	if anyForwarded {
		return nil
	}

	return &BusError{Code: 404, Message: fmt.Sprintf("no route to: %s", reply)}
}

//---------------------
// x_req.Request / Reply
//---------------------

// Request sends a request and waits for a response with a timeout.
func (b *Bus) Request(subject string, msg []byte, timeout time.Duration) ([]byte, error) {
	reply := fmt.Sprintf("_INBOX.%d", rand.Uint64())
	respCh := make(chan []byte, 1)

	err := b.SubscribeWithHandler(reply, func(_ string, resp []byte) {
		respCh <- resp
	})
	if err != nil {
		return nil, err
	}

	b.PublishWithReply(subject, msg, reply)

	select {
	case data := <-respCh:
		b.Unsubscribe(reply)
		return data, nil
	case <-time.After(timeout):
		b.Unsubscribe(reply)
		return nil, &BusError{Code: 408, Message: "request timeout"}
	}
}

//---------------------
// Subscriptions
//---------------------

// Subscribe subscribes to a subject on the bus.
func (b *Bus) Subscribe(subject string) error {
	return b.SubscribeWithHandler(subject, nil)
}

// SubscribeWithHandler subscribes to a subject with a handler.
func (b *Bus) SubscribeWithHandler(subject string, handler func(subject string, msg []byte)) error {
	wrapped := func(subj string, msg []byte, _ string) {
		if handler != nil {
			handler(subj, msg)
		}
	}
	return b.SubscribeWithQueue(subject, "", wrapped)
}

// SubscribeWithNewClient subscribes a new client to a subject with a queue.
func (b *Bus) SubscribeWithNewClient(subject, queue string, handler func(subject string, msg []byte, reply string)) error {
	client := NewClient(rand.Uint64(), b)
	b.AddClient(client)
	return b.SubscribeForClient(client, subject, queue, handler)
}

// SubscribeWithQueue subscribes to a subject with a queue and handler.
func (b *Bus) SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte, reply string)) error {
	return b.SubscribeForClient(b.selfClient, subject, queue, handler)
}

// SubscribeForClient subscribes a client to a subject.
func (b *Bus) SubscribeForClient(c *Client, subject, queue string, handler func(subject string, msg []byte, reply string)) error {
	if handler != nil {
		// Wrap the handler with middleware
		c.HandleMessage = func(req *x_req.Request) {
			for _, mw := range c.bus.middlewares {
				if err := mw(req); err != nil {
					x_log.RootLogger().Structured().Warn("middleware blocked request",
						x_log.FError(err),
						x_log.FString("subject", req.Subject),
					)
					return
				}
			}
			handler(req.Subject, req.Data, req.Reply)
		}
	}

	if err := c.SubscribeWithQueue(subject, queue, func(subject string, msg []byte) {
		if c.HandleMessage != nil {
			c.HandleMessage(&x_req.Request{
				Subject: subject,
				Data:    msg,
				Reply:   "",
			})
		}
	}); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Add subscription to list
	b.sublist.Insert(&typ.Subscription{
		Subject: []byte(subject),
		Queue:   []byte(queue),
		Client:  c,
	})

	x_log.RootLogger().Structured().Info("bus subscription added",
		x_log.FAny("client_id", c.id),
		x_log.FString("subject", subject),
		x_log.FString("queue", queue),
	)

	return nil
}

// Unsubscribe removes the subscription for the subject.
func (b *Bus) Unsubscribe(subject string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.sublist.Remove(&typ.Subscription{Subject: []byte(subject)})

	x_log.RootLogger().Structured().Info("bus unsubscribed",
		x_log.FString("subject", subject),
	)

	return nil
}

//---------------------
// Event Handling
//---------------------

// SubscribeEvent subscribes to an event by name.
func (b *Bus) SubscribeEvent(eventName string, handler func(typ.Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add event handler for the event
	b.eventHandlers[eventName] = append(b.eventHandlers[eventName], handler)
}

// PublishEvent publishes an event and triggers all associated handlers.
func (b *Bus) PublishEvent(event typ.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Handle all event handlers for the event
	if handlers, exists := b.eventHandlers[event.Name]; exists {
		for _, handler := range handlers {
			go handler(event) // Handle asynchronously
		}
	}
	return nil
}

//---------------------
// Lifecycle
//---------------------

// Start begins background tasks or metrics if needed.
func (b *Bus) Start() error {
	// TODO: background tasks (metrics, leaf sync, etc.)
	return nil
}
