package runn_client

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// WSClient manages the WebSocket connection and message handling
type WSClient struct {
	URL     string                    // WebSocket server URL
	Handler func(event OutgoingEvent) // Event handler for incoming messages
}

// NewWSClient creates a new WebSocket client with the provided URL and handler
func NewWSClient(url string, handler func(event OutgoingEvent)) *WSClient {
	return &WSClient{
		URL:     url,
		Handler: handler,
	}
}

// Connect opens a WebSocket connection and starts listening for messages
func (ws *WSClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(ws.URL, nil)
	if err != nil {
		return err
	}
	go ws.listen(conn)
	return nil
}

// listen listens for incoming WebSocket messages and invokes the handler
func (ws *WSClient) listen(conn *websocket.Conn) {
	defer conn.Close()
	for {
		// Read incoming message
		_, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ws error: %v", err)
			return
		}

		// Unmarshal the data into an OutgoingEvent
		var event OutgoingEvent
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		// Call the handler with the event if available
		if ws.Handler != nil {
			ws.Handler(event)
		}
	}
}
