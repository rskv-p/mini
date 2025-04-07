package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// startCmd starts a process by its ID
var startCmd = &cobra.Command{
	Use:   "start <id>",            // Command syntax (process ID required)
	Short: "Start a process by ID", // Short description of the command
	Args:  cobra.ExactArgs(1),      // Ensure exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID from arguments
		id := args[0]

		// Create the POST request to start the process
		req, _ := http.NewRequest("POST", apiURL()+"/api/start/"+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for a successful response (204 No Content)
		if resp.StatusCode != 204 {
			return fmt.Errorf("failed to start process: %s", resp.Status)
		}

		// Confirm that the process has been started
		fmt.Println("✅ Process started:", id)
		return nil
	},
}

func init() {
	// Add the 'start' command to the root command
	RootCmd.AddCommand(startCmd)
}
