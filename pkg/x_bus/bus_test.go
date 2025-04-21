// file: mini/pkg/x_bus/bus_test.go
package x_bus_test

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rskv-p/mini/pkg/x_bus"
	"github.com/rskv-p/mini/pkg/x_req"

	"github.com/stretchr/testify/require"
)

//---------------------
// Helpers
//---------------------

// newTestBus creates a new bus instance for testing.
func newTestBus() *x_bus.Bus {
	return x_bus.NewBus(false, "")
}

//---------------------
// Basic: Subscribe + Publish
//---------------------

// TestBus_SubscribeAndPublish tests the basic subscribe and publish functionality.
func TestBus_SubscribeAndPublish(t *testing.T) {
	b := newTestBus()
	got := make(chan []byte, 1)

	// Subscribe to "foo.bar" topic and handle incoming messages
	err := b.SubscribeWithHandler("foo.bar", func(_ string, data []byte) {
		got <- data
	})
	require.NoError(t, err)

	// Publish a message to "foo.bar"
	msg := []byte("hello")
	err = b.Publish("foo.bar", msg)
	require.NoError(t, err)

	// Wait for the message and validate it
	select {
	case out := <-got:
		require.Equal(t, msg, out)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

//---------------------
// Request / Respond
//---------------------

// TestBus_RequestAndRespond tests the request and respond functionality.
func TestBus_RequestAndRespond(t *testing.T) {
	b := newTestBus()

	// Subscribe to "echo" and respond with the same data
	err := b.SubscribeWithQueue("echo", "", func(subj string, data []byte, reply string) {
		_ = b.Respond(reply, data)
	})
	require.NoError(t, err)

	// Send a request and check the response
	resp, err := b.Request("echo", []byte("test"), time.Second)
	require.NoError(t, err)
	require.Equal(t, []byte("test"), resp)
}

//---------------------
// Unsubscribe
//---------------------

// TestBus_Unsubscribe tests unsubscribing from a topic.
func TestBus_Unsubscribe(t *testing.T) {
	b := newTestBus()
	called := false

	// Subscribe to "foo.baz"
	_ = b.SubscribeWithHandler("foo.baz", func(_ string, _ []byte) {
		called = true
	})

	// Unsubscribe and verify no message is received
	err := b.Unsubscribe("foo.baz")
	require.NoError(t, err)

	err = b.Publish("foo.baz", []byte("ping"))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	require.False(t, called, "should not have received message after unsubscribe")
}

//---------------------
// No Subscribers
//---------------------

// TestBus_PublishWithoutSubscriber tests publishing a message without subscribers.
func TestBus_PublishWithoutSubscriber(t *testing.T) {
	b := newTestBus()
	err := b.Publish("no.match", []byte("test"))
	require.NoError(t, err)
}

//---------------------
// Respond to multiple subscribers
//---------------------

// TestBus_RespondToMultipleSubscribers tests responding to multiple subscribers.
func TestBus_RespondToMultipleSubscribers(t *testing.T) {
	b := newTestBus()
	reply := "_REPLY.multicast"
	data := []byte("hi all")

	var got1, got2 []byte
	wg := sync.WaitGroup{}
	wg.Add(2)

	// First subscriber
	err := b.SubscribeWithNewClient(reply, "", func(_ string, msg []byte, _ string) {
		got1 = msg
		wg.Done()
	})
	require.NoError(t, err)

	// Second subscriber
	err = b.SubscribeWithNewClient(reply, "", func(_ string, msg []byte, _ string) {
		got2 = msg
		wg.Done()
	})
	require.NoError(t, err)

	// Respond to both subscribers
	err = b.Respond(reply, data)
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Verify both subscribers received the message
	select {
	case <-done:
		require.Equal(t, data, got1)
		require.Equal(t, data, got2)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for responses")
	}
}

//---------------------
// Respond to All Leafs
//---------------------

// TestBus_RespondToAllLeafs tests responding to all leaf nodes.
func TestBus_RespondToAllLeafs(t *testing.T) {
	bus := x_bus.NewBus(false, "")

	// Setup TCP Listener
	ln, err := net.Listen("tcp", ":43211")
	require.NoError(t, err)
	defer ln.Close()

	// Accept two incoming leaf connections
	acceptLeaf := func() {
		conn, _ := ln.Accept()
		_, _ = x_bus.AcceptLeaf(conn, bus)
	}
	go acceptLeaf()
	go acceptLeaf()

	// Connect two leaf nodes
	leaf1, err := x_bus.NewLeafNode("localhost:43211", bus)
	require.NoError(t, err)
	leaf2, err := x_bus.NewLeafNode("localhost:43211", bus)
	require.NoError(t, err)

	// Setup message capture
	got1 := make(chan []byte, 1)
	got2 := make(chan []byte, 1)

	leaf1.C.HandleMessage = func(req *x_req.Request) {
		got1 <- req.Data
	}
	leaf2.C.HandleMessage = func(req *x_req.Request) {
		got2 <- req.Data
	}

	// Subscribe both leaves to the reply subject
	subject := "REPLY.test.42"
	leaf1.C.OnSubscribe(subject)
	time.Sleep(200 * time.Millisecond)
	leaf2.C.OnSubscribe(subject)
	time.Sleep(200 * time.Millisecond)

	// Wait for subscriptions to propagate
	time.Sleep(200 * time.Millisecond)

	// Respond via Bus
	msg := []byte("hello leafs")
	err = bus.Respond(subject, msg)
	require.NoError(t, err)

	// Verify both received the message
	select {
	case got := <-got1:
		require.Equal(t, msg, got)
	case <-time.After(time.Second):
		t.Fatal("leaf1 did not receive message")
	}

	select {
	case got := <-got2:
		require.Equal(t, msg, got)
	case <-time.After(time.Second):
		t.Fatal("leaf2 did not receive message")
	}
}

//---------------------
// Remove Client Unsubscribes
//---------------------

// TestBus_RemoveClient_Unsubscribes tests that removing a client unsubscribes it from topics.
func TestBus_RemoveClient_Unsubscribes(t *testing.T) {
	b := newTestBus()
	got := make(chan string, 1)

	client := x_bus.NewClient(999, b)

	// Subscribe manually for the client
	err := b.SubscribeForClient(client, "foo.bar", "", func(_ string, data []byte, _ string) {
		got <- string(data)
	})
	require.NoError(t, err)

	// Check if subscription works
	err = b.Publish("foo.bar", []byte("before"))
	require.NoError(t, err)
	select {
	case msg := <-got:
		require.Equal(t, "before", msg)
	case <-time.After(time.Second):
		t.Fatal("timeout before RemoveClient")
	}

	// Remove the client
	b.RemoveClient(client)

	// Check that subscription is removed
	err = b.Publish("foo.bar", []byte("after"))
	require.NoError(t, err)
	select {
	case msg := <-got:
		t.Fatalf("should not receive message after RemoveClient, but got: %s", msg)
	case <-time.After(100 * time.Millisecond):
		// ok
	}
}
