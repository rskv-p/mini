package cmd_runn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// watchCmd subscribes to process updates via WebSocket
var watchCmd = &cobra.Command{
	Use:   "watch",                                        // Command name
	Short: "Subscribe to process updates (via WebSocket)", // Command description
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Create WebSocket connection URL
		u := url.URL{
			Scheme:   "ws",
			Host:     "localhost:8080", // WebSocket server address
			Path:     "/ws",
			RawQuery: "token=" + token, // Append token as query parameter
		}

		// Establish WebSocket connection
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to connect to WebSocket: %w", err)
		}
		defer conn.Close()

		// Notify user that the connection is established
		fmt.Println("🔌 Connected. Waiting for events...")

		// Continuously read messages from the WebSocket
		for {
			// Read message from WebSocket
			_, data, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("WebSocket error: %w", err)
			}

			// Parse the received message
			var evt map[string]interface{}
			if err := json.Unmarshal(data, &evt); err != nil {
				log.Println("Failed to parse message:", err)
				continue
			}

			// Extract and display process information
			process := evt["process"].(map[string]interface{})
			fmt.Printf("📢 [%v] %s (%s): %s\n", process["id"], process["cmd"], process["dir"], process["status"])
		}
	},
}

func init() {
	// Add the 'watch' command to the root command
	RootCmd.AddCommand(watchCmd)
}
