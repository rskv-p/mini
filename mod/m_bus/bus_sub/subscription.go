package bus_sub

import "github.com/rskv-p/mini/mod/m_bus/bus_type"

// Subscription represents a subscription to a subject with an associated client and queue.
type Subscription struct {
	Subject []byte              // The subject to which the client is subscribed
	Queue   []byte              // The queue associated with the subscription
	Client  bus_type.IBusClient // The client associated with the subscription
}

// NewSubscription creates a new subscription with the provided subject, queue, and client.
func NewSubscription(subject, queue []byte, client bus_type.IBusClient) *Subscription {
	return &Subscription{
		Subject: subject,
		Queue:   queue,
		Client:  client,
	}
}
