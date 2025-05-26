// file: mini/transport/conn.go
package transport

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/rskv-p/mini/codec"
)

// Ensure Conn implements IConn interface.
var _ IConn = (*Conn)(nil)

// IConn defines the internal transport connection interface.
type IConn interface {
	Publish(subject string, data []byte) error
	Request(subject string, data []byte, timeout time.Duration) (codec.IMessage, error)
	Subscribe(subject string, handler MsgHandler) (*Subscription, error)
	SubscribeOnce(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error)
	Close()
	IsConnected() bool
	Ping() error
}

// Conn wraps nsq.Producer and dynamic consumer map.
type Conn struct {
	producer   *nsq.Producer
	opts       *ConnOptions
	mu         sync.RWMutex
	consumers  map[string]*nsq.Consumer
	replyMu    sync.Mutex
	replyChans map[string]chan codec.IMessage
}

// Subscription wraps a topic/channel-bound consumer.
type Subscription struct {
	topic    string
	channel  string
	consumer *nsq.Consumer
}

// ConnOptions defines how to connect to NSQ.
type ConnOptions struct {
	Servers []string
	Timeout time.Duration
	Debug   bool
	Metrics IMetrics
}

// DefaultConnOptions returns base connection settings.
func DefaultConnOptions() *ConnOptions {
	return &ConnOptions{
		Servers: []string{"127.0.0.1:4150"},
		Timeout: 5 * time.Second,
	}
}

// Connect creates a new Conn with producer initialized.
func (o *ConnOptions) Connect() (*Conn, error) {
	cfg := nsq.NewConfig()
	prod, err := nsq.NewProducer(o.Servers[0], cfg)
	if err != nil {
		return nil, fmt.Errorf("create producer: %w", err)
	}
	return &Conn{
		producer:   prod,
		opts:       o,
		consumers:  make(map[string]*nsq.Consumer),
		replyChans: make(map[string]chan codec.IMessage),
	}, nil
}

func (c *Conn) Publish(subject string, data []byte) error {
	if c.opts.Debug {
		fmt.Printf("[nsq] → publish: %s (%d bytes)\n", subject, len(data))
	}
	return c.producer.Publish(subject, data)
}

func (c *Conn) Request(subject string, data []byte, timeout time.Duration) (codec.IMessage, error) {
	msg := codec.NewMessage("")
	if err := codec.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("unmarshal request: %w", err)
	}

	if msg.GetContextID() == "" {
		msg.SetContextID(uuid.NewString())
	}
	if msg.GetReplyTo() == "" {
		msg.SetReplyTo("reply." + msg.GetContextID())
	}

	replyCh := make(chan codec.IMessage, 1)

	// Subscribe to reply
	_, err := c.SubscribeOnce(msg.GetReplyTo(), func(data []byte) error {
		resp := codec.NewMessage("")
		if err := codec.Unmarshal(data, resp); err != nil {
			return err
		}
		if c.opts.Debug {
			fmt.Printf("[nsq] ← response from %s (ctx=%s)\n", msg.GetReplyTo(), resp.GetContextID())
		}
		c.replyMu.Lock()
		if ch, ok := c.replyChans[resp.GetContextID()]; ok {
			select {
			case ch <- resp:
			default:
			}
		}
		c.replyMu.Unlock()
		return nil
	}, timeout+5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("subscribe to reply: %w", err)
	}

	// Store channel
	c.replyMu.Lock()
	c.replyChans[msg.GetContextID()] = replyCh
	c.replyMu.Unlock()
	c.cleanupReplyChan(msg.GetContextID(), replyCh, timeout+5*time.Second)

	// Marshal and send
	raw, err := codec.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if c.opts.Debug {
		fmt.Printf("[nsq] → request %s → %s (ctx=%s)\n", subject, msg.GetReplyTo(), msg.GetContextID())
	}
	if err := c.Publish(subject, raw); err != nil {
		return nil, err
	}

	select {
	case resp := <-replyCh:
		return resp, nil
	case <-time.After(timeout):
		return nil, errors.New("request timeout")
	}
}

func (c *Conn) cleanupReplyChan(contextID string, ch chan codec.IMessage, ttl time.Duration) {
	timer := time.NewTimer(ttl)
	go func() {
		defer timer.Stop()
		select {
		case <-ch:
		case <-timer.C:
			if c.opts.Debug {
				fmt.Printf("[nsq] ⏱ replyChan timeout: %s\n", contextID)
			}
		}
		c.replyMu.Lock()
		delete(c.replyChans, contextID)
		c.replyMu.Unlock()
	}()
}

func (c *Conn) Subscribe(subject string, handler MsgHandler) (*Subscription, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.consumers[subject]; exists {
		return nil, fmt.Errorf("consumer for subject %s already exists", subject)
	}

	if c.opts.Debug {
		fmt.Printf("[nsq] subscribe to: %s\n", subject)
	}

	cfg := nsq.NewConfig()
	channel := "channel-" + uuid.NewString()
	consumer, err := nsq.NewConsumer(subject, channel, cfg)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(nsqMsg *nsq.Message) error {
		if c.opts.Debug {
			fmt.Printf("[nsq] ← %s (%d bytes)\n", subject, len(nsqMsg.Body))
		}
		return handler(nsqMsg.Body)
	}))

	if err := consumer.ConnectToNSQD(c.opts.Servers[0]); err != nil {
		return nil, fmt.Errorf("connect to NSQD: %w", err)
	}

	c.consumers[subject] = consumer

	if c.opts.Metrics != nil {
		c.opts.Metrics.IncCounter("conn_subscribed_total")
	}

	return &Subscription{topic: subject, channel: channel, consumer: consumer}, nil
}

func (c *Conn) SubscribeOnce(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error) {
	return c.SubscribeWithTTL(subject, handler, ttl)
}

func (c *Conn) SubscribeWithTTL(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error) {
	sub, err := c.Subscribe(subject, handler)
	if err != nil {
		return nil, err
	}
	c.scheduleTTLUnsubscribe(subject, sub, ttl)
	return sub, nil
}

func (c *Conn) scheduleTTLUnsubscribe(subject string, sub *Subscription, ttl time.Duration) {
	time.AfterFunc(ttl, func() {
		if c.opts.Debug {
			fmt.Printf("[nsq] ⏱ TTL unsubscribe: %s\n", subject)
		}
		_ = sub.cancel()
		c.mu.Lock()
		delete(c.consumers, subject)
		c.mu.Unlock()
	})
}

func (s *Subscription) cancel() error {
	if s.consumer != nil {
		s.consumer.Stop()
		<-s.consumer.StopChan
	}
	return nil
}

func (c *Conn) IsConnected() bool {
	return c.producer != nil
}

func (c *Conn) Ping() error {
	if c.producer == nil {
		return errors.New("producer is nil")
	}
	return c.producer.Ping()
}

func (c *Conn) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.opts.Debug {
		fmt.Printf("[nsq] closing transport\n")
	}

	for _, consumer := range c.consumers {
		consumer.Stop()
	}
	c.producer.Stop()
}
