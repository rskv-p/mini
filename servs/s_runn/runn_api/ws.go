package runn_api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/rskv-p/mini/servs/s_runn/runn_serv"
)

type wsContextKey string

// WebSocket upgrader to handle HTTP -> WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (can be improved for production)
		return true
	},
}

// Hub stores all active WebSocket connections and their associated contexts
type Hub struct {
	clients map[*websocket.Conn]context.Context
	mu      sync.Mutex
}

// wsHub is the global instance of Hub
var wsHub = &Hub{
	clients: make(map[*websocket.Conn]context.Context),
}

// HandleWS handles WebSocket connections with authentication
func HandleWS(w http.ResponseWriter, r *http.Request) {
	// Extract token from the request
	tokenStr := extractToken(r)
	if tokenStr == "" {
		http.Error(w, "unauthorized", 401)
		return
	}

	// Parse and verify the token
	claims := &jwtClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || claims.ExpiresAt.Time.Before(time.Now()) {
		http.Error(w, "invalid or expired token", 401)
		return
	}

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "could not open websocket connection", 500)
		return
	}
	defer conn.Close()

	// Send a welcome message to the connected client
	err = conn.WriteMessage(websocket.TextMessage, []byte("Welcome to the WebSocket server!"))
	if err != nil {
		http.Error(w, "error sending message", 500)
		return
	}

	// Optionally, start listening and handling messages from the client here
}

// SendStatusUpdate sends status updates to all connected WebSocket clients
func SendStatusUpdate(proc *runn_serv.Process) {
	event := OutgoingEvent{
		Type:    "status_update",
		Process: proc,
	}
	msg, _ := json.Marshal(event)

	// Lock and send the message to each connected client
	wsHub.mu.Lock()
	defer wsHub.mu.Unlock()
	for conn, ctx := range wsHub.clients {
		username, _, ok := UserFromContext(ctx)
		if ok {
			log.Printf("➡️ pushing update to %s (pid: %d)", username, proc.ID)
		}
		_ = conn.WriteMessage(websocket.TextMessage, msg)
	}
}
