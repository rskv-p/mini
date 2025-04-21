package bus_client

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/rskv-p/mini/mod/m_bus/bus_req"
	"github.com/rskv-p/mini/mod/m_bus/bus_type"
)

var _ bus_type.IBusClient = (*Client)(nil)

// Client implements the IBusClient interface.
type Client struct {
	id             uint64
	bus            bus_type.IBus
	mu             sync.Mutex
	subs           map[string]func(string, []byte)
	HandleMessage  func(*bus_req.Request)
	OnSubscribe    func(string)
	OnUnsubscribe  func(string)
	HandleError    func(error)
	remoteInterest map[string]struct{}
	conn           net.Conn
}

func (b *Client) GetClientCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.bus.GetClientCount()
}

// GetMsgHandlerCount returns the number of message handlers registered for the client.
func (c *Client) GetMsgHandlerCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.bus.GetMsgHandlerCount()
}

// NewClient creates a new client associated with the bus.
func NewClient(id uint64, bus bus_type.IBus, conn net.Conn) *Client {
	return &Client{
		id:             id,
		bus:            bus,
		subs:           make(map[string]func(string, []byte)),
		remoteInterest: make(map[string]struct{}),
		conn:           conn, // Инициализация поля conn

	}
}

// GetSecretKey returns the secret key used for authentication.
func (c *Client) GetSecretKey() string {
	return c.GetSecretKey()
}

// CloseConn closes the client's connection.
func (c *Client) CloseConn() error {
	if c.GetConn() != nil {
		return c.CloseConn()
	}
	return nil
}

// GetConn returns the client's network connection.
func (c *Client) GetConn() net.Conn {
	return c.GetConn()
}

// StartPingLoop starts a loop to send PING messages periodically.
func (c *Client) StartPingLoop() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if c.GetConn() != nil {
				_, err := c.conn.Write([]byte("PING\n"))
				if err != nil {
					log.Printf("Client %d failed to send PING: %v", c.id, err)
					return
				}
			}
		}
	}()
}

// Subscribe subscribes the client to a subject without a queue.
func (c *Client) Subscribe(subject string) error {
	return c.SubscribeWithQueue(subject, "", func(subject string, msg []byte) {
		log.Printf("Client %d received message on topic %s: %s", c.id, subject, string(msg))
	})
}

// SubscribeWithQueue subscribes the client to a subject with a queue and message handler.
func (c *Client) SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subs[subject] = handler

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

	delete(c.subs, subject)

	if c.OnUnsubscribe != nil {
		c.OnUnsubscribe(subject)
	}

	log.Printf("Client %d unsubscribed from %s", c.id, subject)
}

// Publish sends a message to a subject via the bus.
func (c *Client) Publish(subject string, data []byte) error {
	return c.bus.Publish(subject, data)
}

// PublishWithReply sends a message to the client with a reply subject.
func (c *Client) PublishWithReply(subject string, data []byte, reply string) {
	if c.HandleMessage != nil {
		c.HandleMessage(&bus_req.Request{
			Subject: subject,
			Data:    data,
			Reply:   reply,
		})
	} else {
		log.Printf("Warning: No handler defined for reply on client %d for subject %s", c.id, subject)
	}
}

// Deliver sends a message to the client, invoking the HandleMessage callback.
func (c *Client) Deliver(subject string, data []byte) {
	if c.HandleMessage != nil {
		c.HandleMessage(&bus_req.Request{
			Subject: subject,
			Data:    data,
		})
	} else {
		log.Printf("Warning: No handler defined for client %d on subject %s", c.id, subject)
	}
}

// GetSubscriptions returns a list of all subjects the client is subscribed to.
func (c *Client) GetSubscriptions() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	subscriptions := make([]string, 0, len(c.subs))
	for subject := range c.subs {
		subscriptions = append(subscriptions, subject)
	}
	return subscriptions
}

// IsSubscribed checks if the client is subscribed to a specific subject.
func (c *Client) IsSubscribed(subject string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.subs[subject]
	return exists
}

// ClearSubscriptions removes all subscriptions for the client.
func (c *Client) ClearSubscriptions() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for subject := range c.subs {
		if c.OnUnsubscribe != nil {
			c.OnUnsubscribe(subject)
		}
		delete(c.subs, subject)
	}

	log.Printf("Client %d cleared all subscriptions", c.id)
}

// SetErrorHandler sets a custom error handler for the client.
func (c *Client) SetErrorHandler(handler func(error)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HandleError = handler
}

// GetID returns the ID of the client.
func (c *Client) GetID() uint64 {
	return c.id
}

// HasRemoteInterest checks if the client has a remote interest in a subject.
func (c *Client) HasRemoteInterest(subject string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.remoteInterest[subject]
	return exists
}

// MarkRemoteInterest marks a subject as having remote interest.
func (c *Client) MarkRemoteInterest(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.remoteInterest[subject] = struct{}{}
}

// UnmarkRemoteInterest removes the remote interest for a subject.
func (c *Client) UnmarkRemoteInterest(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.remoteInterest, subject)
}

// HasMatchingInterest checks if the client has a matching interest in a subject.
func (c *Client) HasMatchingInterest(subject string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.subs[subject]
	return exists
}

// GetOnUnsubscribe returns the OnUnsubscribe handler.
func (c *Client) GetOnUnsubscribe() func(string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.OnUnsubscribe
}

// SetOnSubscribe sets a new subscription handler for the client.
func (c *Client) SetOnSubscribe(handler func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.OnSubscribe = handler
}

// SetOnUnsubscribe sets a new unsubscription handler for the client.
func (c *Client) SetOnUnsubscribe(handler func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.OnUnsubscribe = handler
}

// SetHandleMessage sets a new message handler for the client.
func (c *Client) SetHandleMessage(handler func(*bus_req.Request)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.HandleMessage = handler
}

// GetHandleMessage returns the current message handler for the client.
func (c *Client) GetHandleMessage() func(*bus_req.Request) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.HandleMessage
}

// GetBus returns the bus associated with the client.
func (c *Client) GetBus() bus_type.IBus {
	return c.bus
}
