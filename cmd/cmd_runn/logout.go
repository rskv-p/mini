package cmd_runn

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// logoutCmd handles logging out by removing the token
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of the system (remove the token)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tokenFile := tokenFilePath()

		// Check if the token file exists
		if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
			return fmt.Errorf("token not found. You are already logged out")
		}

		// Attempt to remove the token file
		if err := os.Remove(tokenFile); err != nil {
			return fmt.Errorf("failed to remove token: %w", err)
		}

		// Notify the user of successful logout
		fmt.Println("✅ Logout successful. Token removed.")
		return nil
	},
}

func init() {
	// Add the 'logout' command to the root command
	RootCmd.AddCommand(logoutCmd)
}
