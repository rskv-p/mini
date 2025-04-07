package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// removeCmd deletes a process by ID, provided it is not active
var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a process by ID (if not active)",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID from the arguments
		id := args[0]

		// Create the DELETE request to remove the process
		req, _ := http.NewRequest("DELETE", apiURL()+"/api/remove/"+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check if the process removal was successful
		if resp.StatusCode != 204 {
			return fmt.Errorf("failed to remove process: %s", resp.Status)
		}

		// Confirm the process has been removed
		fmt.Println("🗑️  Process removed:", id)
		return nil
	},
}

func init() {
	// Add the 'remove' command to the root command
	RootCmd.AddCommand(removeCmd)
}
