package cmd_runn

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// resumeCmd resumes a paused process (SIGCONT)
var resumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a paused process (SIGCONT)",
	Args:  cobra.ExactArgs(1), // Requires exactly one argument (process ID)
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Get the process ID from the arguments
		id := args[0]

		// Create the POST request to resume the process
		req, _ := http.NewRequest("POST", apiURL()+"/api/resume/"+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for successful response
		if resp.StatusCode != 204 {
			return fmt.Errorf("error resuming process: %s", resp.Status)
		}

		// Confirm the process has been resumed
		fmt.Println("▶️  Process resumed:", id)
		return nil
	},
}

func init() {
	// Add the 'resume' command to the root command
	RootCmd.AddCommand(resumeCmd)
}
