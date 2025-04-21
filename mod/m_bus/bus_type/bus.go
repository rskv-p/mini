package bus_type

import (
	"net"
	"time"

	"github.com/rskv-p/mini/mod/m_bus/bus_req"
)

// ClientFactory defines a function type for creating new clients.
type ClientFactory func(id uint64, bus IBus) IBusClient

// IBus defines the interface for the message bus.
type IBus interface {
	// Middleware management
	Use(mw IMiddleware)

	// Client management
	AddClient(c IBusClient)
	RemoveClient(c IBusClient)

	// Leaf management
	AddLeaf(leaf ILeaf)

	// Authentication
	Authenticate(token string) error

	// Message publishing
	Publish(subject string, msg []byte) error
	PublishWithReply(subject string, msg []byte, reply string) error
	Respond(reply string, data []byte) error

	// Request/Reply
	Request(subject string, msg []byte, timeout time.Duration) ([]byte, error)

	// Subscriptions
	Subscribe(subject string) error
	SubscribeWithHandler(subject string, handler func(subject string, msg []byte)) error
	SubscribeWithNewClient(subject, queue string, handler func(subject string, msg []byte, reply string)) error
	SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte, reply string)) error
	SubscribeForClient(c IBusClient, subject, queue string, handler func(subject string, msg []byte, reply string)) error
	Unsubscribe(subject string) error

	// Lifecycle
	Start() error

	GetClientCount() int
	GetMsgHandlerCount() int
}

// IMiddleware defines the interface for middleware.
type IMiddleware interface {
	Process(req *bus_req.Request) error
}

// IBusClient defines the interface for a client in the bus system.
type IBusClient interface {
	// Subscription management
	Subscribe(subject string) error
	SubscribeWithQueue(subject, queue string, handler func(subject string, msg []byte)) error
	Unsubscribe(subject string)
	ClearSubscriptions()
	GetSubscriptions() []string
	IsSubscribed(subject string) bool

	// Message publishing
	Publish(subject string, data []byte) error
	PublishWithReply(subject string, data []byte, reply string)

	// Message delivery
	Deliver(subject string, data []byte)

	// Handlers
	SetHandleMessage(handler func(*bus_req.Request))
	GetHandleMessage() func(*bus_req.Request)
	SetErrorHandler(handler func(error))
	GetOnUnsubscribe() func(string)
	SetOnSubscribe(handler func(string))
	SetOnUnsubscribe(handler func(string))

	// Client metadata
	GetID() uint64
	HasRemoteInterest(subject string) bool
	MarkRemoteInterest(subject string)
	UnmarkRemoteInterest(subject string)
	HasMatchingInterest(subject string) bool

	// Bus reference
	GetBus() IBus

	// GetSecretKey returns the secret key used for authentication.
	GetSecretKey() string
	CloseConn() error
	GetConn() net.Conn

	//initHandlers()
	StartPingLoop()

	GetClientCount() int
	GetMsgHandlerCount() int
}

// ILeaf defines the interface for a leaf node in the bus system.
type ILeaf interface {
	// Connection management
	GetConn() net.Conn
	CloseConn() error

	// Message handling
	Send(subject string, msg []byte)
	SendWithReply(subject string, msg []byte, reply string)
	SendSub(subject string)
	SendUnsub(subject string)
	SendResp(subject string, msg []byte)

	// Handlers
	InitHandlers()
	ReadLoop()

	// Ping management
	StartPingLoop()
}
