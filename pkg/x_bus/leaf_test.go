// file:mini/pkg/x_bus/leaf_test.go
package x_bus_test

import (
	"net"
	"testing"
	"time"

	"github.com/rskv-p/mini/pkg/x_bus"

	"github.com/stretchr/testify/require"
)

func TestLeaf_PublishAndRespond(t *testing.T) {
	bus := x_bus.NewBus(false, "")
	go bus.Start()

	//---------------------
	// Setup TCP Listener
	//---------------------

	ln, err := net.Listen("tcp", ":43210")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		conn, _ := ln.Accept()
		_, _ = x_bus.AcceptLeaf(conn, bus)
	}()

	//---------------------
	// Leaf connects
	//---------------------

	leaf, err := x_bus.NewLeafNode("localhost:43210", bus)
	require.NoError(t, err)

	//---------------------
	// Subscribe on local bus
	//---------------------

	got := make(chan string, 1)

	_ = bus.SubscribeWithHandler("test.echo", func(_ string, msg []byte) {
		got <- string(msg)
	})

	//---------------------
	// Send message from leaf
	//---------------------

	leaf.Send("test.echo", []byte("ping"))

	select {
	case msg := <-got:
		require.Equal(t, "ping", msg)
	case <-time.After(1 * time.Second):
		t.Fatal("did not receive message")
	}
}
