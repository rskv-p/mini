package cmd_runn

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// whoamiCmd retrieves and displays information about the current user
var whoamiCmd = &cobra.Command{
	Use:   "whoami",                                     // Command name
	Short: "Display information about the current user", // Command description
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the authorization token
		token, err := loadToken()
		if err != nil {
			return err
		}

		// Split the token into its parts (header, payload, signature)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			return fmt.Errorf("invalid token")
		}

		// Decode the payload (middle part of the token)
		decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return err
		}

		// Unmarshal the payload into a map
		var payload map[string]interface{}
		if err := json.Unmarshal(decoded, &payload); err != nil {
			return err
		}

		// Display user information (subject, role, expiration)
		fmt.Println("User:", payload["sub"])
		fmt.Println("Role:", payload["role"])
		fmt.Println("Expires:", payload["exp"])
		return nil
	},
}

func init() {
	// Add the 'whoami' command to the root command
	RootCmd.AddCommand(whoamiCmd)
}
