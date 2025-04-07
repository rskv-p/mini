package cmd_runn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// editCmd allows editing a process by ID
var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a process by ID",
	Args:  cobra.ExactArgs(1), // Require exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID, command, and directory from flags
		id := args[0]
		cmdStr, _ := cmd.Flags().GetString("cmd")
		dir, _ := cmd.Flags().GetString("dir")

		// Ensure at least one of --cmd or --dir is provided
		if cmdStr == "" && dir == "" {
			return fmt.Errorf("at least one of --cmd or --dir must be provided")
		}

		// Create the payload for the update request
		payload := map[string]string{
			"cmd": cmdStr,
			"dir": dir,
		}
		body, _ := json.Marshal(payload)

		// Create the POST request with the token and content type
		req, _ := http.NewRequest("POST", apiURL()+"/api/edit/"+id, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for a successful response
		if resp.StatusCode != 200 {
			return fmt.Errorf("edit failed: %s", resp.Status)
		}

		// Print the success message
		fmt.Println("✅ Process edited:", id)
		return nil
	},
}

func init() {
	// Define flags for the 'edit' command
	editCmd.Flags().String("cmd", "", "New command for the process")
	editCmd.Flags().String("dir", "", "New directory for the process")
	RootCmd.AddCommand(editCmd)
}
