package cmd_runn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// listCmd retrieves and displays a list of all processes
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return fmt.Errorf("login required via `hades login`")
		}

		// Create the GET request to fetch process list
		req, _ := http.NewRequest("GET", apiURL()+"/api/processes", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check if the response is successful
		if resp.StatusCode != 200 {
			return fmt.Errorf("error: %s", resp.Status)
		}

		// Decode the response body into a list of processes
		var procs []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&procs); err != nil {
			return err
		}

		// Display the list of processes
		for _, p := range procs {
			fmt.Printf("[%v] %s (%s) – %s\n", p["id"], p["cmd"], p["dir"], p["status"])
		}

		return nil
	},
}

func init() {
	// Add the 'list' command to the root command
	RootCmd.AddCommand(listCmd)
}
