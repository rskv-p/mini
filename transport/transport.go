// file: mini/transport/transport.go
package transport

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/rskv-p/mini/codec"
)

var (
	ErrDisconnected   = errors.New("transport: not connected")
	ErrMissingHandler = errors.New("transport: handler not set")
	ErrNotSupported   = errors.New("transport: operation not supported")
)

type MsgHandler func([]byte) error

type ITransport interface {
	Init() error
	Close() error

	Options() Options
	SetHandler(MsgHandler)
	Use(MiddlewareFunc)

	Subscribe() error
	Unsubscribe() error
	IsConnected() bool
	Ping() error
	Health() error

	Request(subject string, req []byte, handler ResponseHandler) error
	RequestWithContext(ctx context.Context, subject string, req []byte, handler ResponseHandler) error
	Publish(subject string, data []byte) error
	Respond(replyTo string, msg codec.IMessage) error
	SendFile(msg codec.IMessage, subject string, file []byte) error
	Broadcast(subjects []string, data []byte) error
	Broadcastf(format string, keys ...string) error

	SubscribeTopic(topic string, handler MsgHandler) error
	SubscribePrefix(prefix string, handler MsgHandler) error
}

type Transport struct {
	conn        IConn
	sub         *Subscription
	opts        Options
	handler     MsgHandler
	mu          sync.Mutex
	middlewares []MiddlewareFunc
	active      sync.WaitGroup
}

var _ ITransport = (*Transport)(nil)

const (
	DefaultRequestTimeout = 15 * time.Second
	defaultRetryAttempts  = 3
	defaultRetryDelay     = 100 * time.Millisecond
)

func New(opts ...Option) *Transport {
	options := WithDefaults()
	for _, o := range opts {
		o(&options)
	}
	t := &Transport{opts: options}
	t.Use(TraceMiddleware())
	return t
}

var DefaultTransport = New()

func (t *Transport) Init() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var addrs []string
	for _, addr := range t.opts.Addrs {
		if addr != "" {
			addrs = append(addrs, strings.TrimPrefix(addr, "bus://"))
		}
	}
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:4150"}
	}

	connOpts := DefaultConnOptions()
	connOpts.Servers = addrs
	connOpts.Timeout = t.opts.Timeout
	connOpts.Debug = t.opts.Debug
	connOpts.Metrics = t.opts.Metrics

	conn, err := connOpts.Connect()
	if err != nil {
		return err
	}
	t.conn = conn
	if t.opts.Logger != nil {
		t.opts.Logger.Info("transport initialized: %v", addrs)
	}
	return nil
}

func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.active.Wait()

	if t.sub != nil {
		_ = t.sub.cancel()
		t.sub = nil
	}
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	if t.opts.Logger != nil {
		t.opts.Logger.Info("transport closed")
	}
	return nil
}

func (t *Transport) Options() Options        { return t.opts }
func (t *Transport) SetHandler(h MsgHandler) { t.handler = h }
func (t *Transport) Use(mw MiddlewareFunc)   { t.middlewares = append(t.middlewares, mw) }

func (t *Transport) wrapChain(final func(string, []byte) error) func(string, []byte) error {
	return func(subject string, data []byte) error {
		t.active.Add(1)
		defer t.active.Done()
		wrapped := final
		for i := len(t.middlewares) - 1; i >= 0; i-- {
			next := wrapped
			mw := t.middlewares[i]
			wrapped = func(subj string, d []byte) error {
				return mw(subj, d, next)
			}
		}
		return wrapped(subject, data)
	}
}

func (t *Transport) Subscribe() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return ErrDisconnected
	}
	if t.handler == nil {
		return ErrMissingHandler
	}
	if t.opts.Logger != nil {
		t.opts.Logger.Debug("subscribe to: %s", t.opts.Subject)
	}

	sub, err := t.conn.Subscribe(t.opts.Subject, func(data []byte) error {
		return t.wrapChain(func(_ string, d []byte) error {
			return t.handler(d)
		})(t.opts.Subject, data)
	})
	if err != nil {
		return err
	}
	t.sub = sub
	return nil
}

func (t *Transport) Unsubscribe() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sub != nil {
		err := t.sub.cancel()
		t.sub = nil
		return err
	}
	return nil
}

func (t *Transport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn != nil && t.conn.IsConnected()
}

func (t *Transport) Ping() error {
	if t.conn == nil {
		return ErrDisconnected
	}
	return t.conn.Ping()
}

func (t *Transport) Health() error {
	if err := t.Ping(); err != nil {
		return err
	}
	if !t.IsConnected() {
		return ErrDisconnected
	}
	return nil
}

func (t *Transport) reconnect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn != nil {
		t.conn.Close()
	}

	clientOpts := DefaultConnOptions()
	clientOpts.Servers = t.opts.Addrs
	clientOpts.Timeout = t.opts.Timeout
	clientOpts.Debug = t.opts.Debug
	clientOpts.Metrics = t.opts.Metrics

	conn, err := clientOpts.Connect()
	if err != nil {
		return err
	}
	t.conn = conn
	if t.opts.Logger != nil {
		t.opts.Logger.Info("transport reconnected")
	}
	return nil
}

func (t *Transport) SubscribeTopic(topic string, handler MsgHandler) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return ErrDisconnected
	}
	if t.opts.Logger != nil {
		t.opts.Logger.Debug("subscribe to topic: %s", topic)
	}
	_, err := t.conn.Subscribe(topic, func(data []byte) error {
		return t.wrapChain(func(_ string, d []byte) error {
			return handler(d)
		})(topic, data)
	})
	return err
}

func (t *Transport) SubscribePrefix(prefix string, handler MsgHandler) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	lister, ok := t.conn.(interface{ ListTopics() ([]string, error) })
	if !ok {
		return ErrNotSupported
	}

	topics, err := lister.ListTopics()
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if strings.HasPrefix(topic, prefix) {
			_, err := t.conn.Subscribe(topic, func(data []byte) error {
				return t.wrapChain(func(_ string, d []byte) error {
					return handler(d)
				})(topic, data)
			})
			if err != nil && t.opts.Logger != nil {
				t.opts.Logger.Warn("prefix subscribe failed for %s: %v", topic, err)
			}
		}
	}
	return nil
}
