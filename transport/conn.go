// file: arc/service/transport/conn.go
package transport

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rskv-p/mini/service/codec"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
)

// Ensure Conn implements IConn interface.
var _ IConn = (*Conn)(nil)

type IConn interface {
	Publish(subject string, data []byte) error
	Request(subject string, data []byte, timeout time.Duration) (codec.IMessage, error)
	Subscribe(subject string, handler MsgHandler) (*Subscription, error)
	SubscribeOnce(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error)
	Close()
	IsConnected() bool
	Ping() error
}

type Conn struct {
	producer   *nsq.Producer
	opts       *ConnOptions
	mu         sync.RWMutex
	consumers  map[string]*nsq.Consumer
	replyMu    sync.Mutex
	replyChans map[string]chan codec.IMessage
}

type Subscription struct {
	topic    string
	channel  string
	consumer *nsq.Consumer
}

type ConnOptions struct {
	Servers []string
	Timeout time.Duration
	Debug   bool
	Metrics IMetrics
}

func DefaultConnOptions() *ConnOptions {
	return &ConnOptions{
		Servers: []string{"127.0.0.1:4150"},
		Timeout: 5 * time.Second,
	}
}

// Connect initializes the NSQ producer and returns a Conn.
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

// Publish sends a message to the specified subject.
func (c *Conn) Publish(subject string, data []byte) error {
	if c.opts.Debug {
		fmt.Printf("[nsq] → publish %s (%d bytes)\n", subject, len(data))
	}
	return c.producer.Publish(subject, data)
}

// Request sends a message and waits for a reply using replyTo + contextID.
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

	_, err := c.SubscribeOnce(msg.GetReplyTo(), func(data []byte) error {
		resp := codec.NewMessage("")
		if err := codec.Unmarshal(data, resp); err != nil {
			return err
		}
		if c.opts.Debug {
			fmt.Printf("[nsq] ← response %s (contextID=%s)\n", msg.GetReplyTo(), resp.GetContextID())
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

	c.replyMu.Lock()
	c.replyChans[msg.GetContextID()] = replyCh
	c.replyMu.Unlock()
	c.cleanupReplyChan(msg.GetContextID(), replyCh, timeout+5*time.Second)

	raw, err := codec.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if c.opts.Debug {
		fmt.Printf("[nsq] → request %s → %s (contextID=%s)\n", subject, msg.GetReplyTo(), msg.GetContextID())
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

// cleanupReplyChan removes reply channel after TTL or response.
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

// Subscribe registers a persistent consumer on a subject.
func (c *Conn) Subscribe(subject string, handler MsgHandler) (*Subscription, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.opts.Debug {
		fmt.Printf("[nsq] subscribe to: %s\n", subject)
	}

	if _, exists := c.consumers[subject]; exists {
		return nil, fmt.Errorf("consumer for subject %s already exists", subject)
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

// SubscribeWithTTL registers a consumer with automatic cancellation after ttl.
func (c *Conn) SubscribeWithTTL(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error) {
	sub, err := c.Subscribe(subject, handler)
	if err != nil {
		return nil, err
	}
	time.AfterFunc(ttl, func() {
		if c.opts.Debug {
			fmt.Printf("[nsq] ⏱ auto-unsubscribe (TTL): %s\n", subject)
		}
		_ = sub.cancel()
		c.mu.Lock()
		delete(c.consumers, subject)
		c.mu.Unlock()
	})
	return sub, nil
}

// SubscribeOnce registers a one-time subscription with TTL.
func (c *Conn) SubscribeOnce(subject string, handler MsgHandler, ttl time.Duration) (*Subscription, error) {
	return c.SubscribeWithTTL(subject, handler, ttl)
}

// cancel stops the subscription.
func (s *Subscription) cancel() error {
	if s.consumer != nil {
		s.consumer.Stop()
		<-s.consumer.StopChan
	}
	return nil
}

// IsConnected returns true if producer is active.
func (c *Conn) IsConnected() bool {
	return c.producer != nil
}

// Ping sends a no-op to check connection.
func (c *Conn) Ping() error {
	if c.producer == nil {
		return errors.New("producer is nil")
	}
	return c.producer.Ping()
}

// Close shuts down all consumers and the producer.
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
