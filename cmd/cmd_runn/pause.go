package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// pauseCmd pauses a process by sending a SIGSTOP signal
var pauseCmd = &cobra.Command{
	Use:   "pause <id>",
	Short: "Pause a process (SIGSTOP)",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID from arguments
		id := args[0]

		// Create the POST request to pause the process
		req, _ := http.NewRequest("POST", apiURL()+"/api/pause/"+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for successful response
		if resp.StatusCode != 204 {
			return fmt.Errorf("failed to pause process: %s", resp.Status)
		}

		// Confirm the process was paused
		fmt.Println("⏸️  Process paused:", id)
		return nil
	},
}

func init() {
	// Add the 'pause' command to the root command
	RootCmd.AddCommand(pauseCmd)
}
