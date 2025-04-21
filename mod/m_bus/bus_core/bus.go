package bus_core

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/rskv-p/mini/mod/m_bus/bus_sub"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
	"github.com/rskv-p/mini/mod/m_bus/bus_util"
)

// Error represents a custom error type for the Bus system.
type BusError struct {
	Code    int
	Message string
}

func (e *BusError) Error() string {
	return fmt.Sprintf("Bus Error %d: %s", e.Code, e.Message)
}

// Bus represents the message bus, handling clients, subscriptions, and message routing.
type Bus struct {
	mu          sync.Mutex
	selfClient  bus_type.IBusClient
	clients     map[uint64]bus_type.IBusClient
	sublist     *bus_sub.Sublist
	transform   *bus_sub.SubjectTransform
	authEnabled bool
	secretKey   string
	msgHandlers map[string]func(subject string, msg []byte)
	leaves      []bus_type.ILeaf
	middlewares []bus_type.IMiddleware
}

// NewBus creates and initializes a new Bus instance.
func NewBus(authEnabled bool, secretKey string, clientFactory bus_type.ClientFactory) *Bus {
	bus := &Bus{
		clients:     make(map[uint64]bus_type.IBusClient),
		sublist:     bus_sub.NewSublist(100),
		msgHandlers: make(map[string]func(string, []byte)),
		authEnabled: authEnabled,
		secretKey:   secretKey,
		leaves:      make([]bus_type.ILeaf, 0),
	}

	// Создаем selfClient через фабрику
	bus.selfClient = clientFactory(rand.Uint64(), bus)
	bus.AddClient(bus.selfClient)
	return bus
}

// GetSecretKey returns the secret key used for authentication.
func (b *Bus) GetSecretKey() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.secretKey
}

func (c *Bus) GetMsgHandlerCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.msgHandlers)
}

// Use adds middleware to the bus for processing requests.
func (b *Bus) Use(mw bus_type.IMiddleware) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middlewares = append(b.middlewares, mw)
}

// AddClient adds a client to the bus and starts listening for messages.
func (b *Bus) AddClient(c bus_type.IBusClient) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c.GetID()] = c
}

// RemoveClient removes a client from the bus, cleaning up subscriptions and handlers.
func (b *Bus) RemoveClient(c bus_type.IBusClient) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.clients, c.GetID())

	for _, subject := range c.GetSubscriptions() {
		b.sublist.Remove(&bus_sub.Subscription{
			Subject: []byte(subject),
			Client:  c,
		})

		if c.GetOnUnsubscribe() != nil {
			c.GetOnUnsubscribe()(subject)
		}
	}
}

// AddLeaf adds a leaf node to the bus.
func (b *Bus) AddLeaf(leaf bus_type.ILeaf) {
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
	claims, err := bus_util.VerifyJWT(token, b.secretKey)
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
		leaf.SendResp(reply, data)
		anyForwarded = true
	}

	if anyForwarded {
		return nil
	}

	return &BusError{Code: 404, Message: fmt.Sprintf("no route to: %s", reply)}
}

//---------------------
// bus_req.Request / Reply
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
	client := b.selfClient // Используем фабрику для создания клиента
	b.AddClient(client)
	return b.SubscribeForClient(client, subject, queue, handler)
}

// SubscribeWithQueue subscribes to a subject with a queue and handler.
func (b *Bus) SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte, reply string)) error {
	return b.SubscribeForClient(b.selfClient, subject, queue, handler)
}

// SubscribeForClient subscribes a client to a subject.
func (b *Bus) SubscribeForClient(c bus_type.IBusClient, subject, queue string, handler func(subject string, msg []byte, reply string)) error {
	if handler != nil {
		// Wrap the handler with middleware
		c.SetHandleMessage(func(req *bus_req.Request) {
			for _, mw := range b.middlewares {
				if err := mw.Process(req); err != nil {
					return
				}
			}
			handler(req.Subject, req.Data, req.Reply)
		})
	}

	if err := c.SubscribeWithQueue(subject, queue, func(subject string, msg []byte) {
		if c.GetHandleMessage() != nil {
			c.GetHandleMessage()(&bus_req.Request{
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
	b.sublist.Insert(&bus_sub.Subscription{
		Subject: []byte(subject),
		Queue:   []byte(queue),
		Client:  c,
	})

	return nil
}

// Unsubscribe removes the subscription for the subject.
func (b *Bus) Unsubscribe(subject string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.sublist.Remove(&bus_sub.Subscription{Subject: []byte(subject)})

	return nil
}

func (b *Bus) GetClientCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.clients)
}

//---------------------
// Lifecycle
//---------------------

// Start begins background tasks or metrics if needed.
func (b *Bus) Start() error {
	// TODO: background tasks (metrics, leaf sync, etc.)
	return nil
}
