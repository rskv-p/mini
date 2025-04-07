package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// StopCmd stops a process by its ID
var StopCmd = &cobra.Command{
	Use:   "stop <id>",            // Command syntax (requires process ID)
	Short: "Stop a process by ID", // Short description of the command
	Args:  cobra.ExactArgs(1),     // Ensure exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID from arguments
		id := args[0]

		// Create the POST request to stop the process
		req, _ := http.NewRequest("POST", apiURL()+"/api/stop/"+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for successful response (204 No Content)
		if resp.StatusCode != 204 {
			return fmt.Errorf("failed to stop process: %s", resp.Status)
		}

		// Confirm the process has been stopped
		fmt.Println("⛔ Process stopped:", id)
		return nil
	},
}

func init() {
	// Add the 'stop' command to the root command
	RootCmd.AddCommand(StopCmd)
}
