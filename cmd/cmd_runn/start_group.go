package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// startGroupCmd starts a group of processes by the group ID
var startGroupCmd = &cobra.Command{
	Use:   "start-group <group_id>",                 // Command syntax
	Short: "Start a group of processes by group ID", // Short description
	Args:  cobra.ExactArgs(1),                       // Requires exactly one argument (group ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the group ID from arguments
		groupID := args[0]

		// Create the POST request to start the group
		req, _ := http.NewRequest("POST", apiURL()+"/api/start-group/"+groupID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for successful response (204 No Content)
		if resp.StatusCode != 204 {
			return fmt.Errorf("failed to start group: %s", resp.Status)
		}

		// Confirm the group has been started
		fmt.Println("✅ Process group started:", groupID)
		return nil
	},
}

func init() {
	// Add the 'start-group' command to the root command
	RootCmd.AddCommand(startGroupCmd)
}
