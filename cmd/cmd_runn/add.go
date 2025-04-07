package cmd_runn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// addCmd adds a new process via the REST API
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new process",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get command and directory flags
		cmdStr, _ := cmd.Flags().GetString("cmd")
		dir, _ := cmd.Flags().GetString("dir")

		// Ensure the command is provided
		if cmdStr == "" {
			return fmt.Errorf("--cmd is required")
		}

		// Create payload for the POST request
		payload := map[string]string{
			"cmd": cmdStr,
			"dir": dir,
		}
		body, _ := json.Marshal(payload)

		// Create the POST request with authorization and JSON content type
		req, _ := http.NewRequest("POST", apiURL()+"/api/add", bytes.NewReader(body))
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
			return fmt.Errorf("error adding process: %s", resp.Status)
		}

		// Parse the response to get the process ID
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		// Print the process ID
		fmt.Printf("✅ Process added with ID: %v\n", result["id"])
		return nil
	},
}

func init() {
	// Define flags for the 'add' command
	addCmd.Flags().String("cmd", "", "Command to run")
	addCmd.Flags().String("dir", "", "Working directory")
	RootCmd.AddCommand(addCmd)
}
