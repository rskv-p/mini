package cmd_runn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// loginCmd handles the login process and saves the token
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the service and save the token",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get username and password from flags
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		// Ensure both username and password are provided
		if username == "" || password == "" {
			return fmt.Errorf("both --username and --password are required")
		}

		// Create the request payload
		payload := map[string]string{
			"username": username,
			"password": password,
		}
		body, _ := json.Marshal(payload)

		// Send login request
		resp, err := http.Post(apiURL()+"/auth/login", "application/json", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("request error: %w", err)
		}
		defer resp.Body.Close()

		// Check for successful login response
		if resp.StatusCode != 200 {
			data, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("authorization error: %s", string(data))
		}

		// Parse the response to get the token
		var result map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		token := result["token"]
		if token == "" {
			return fmt.Errorf("token not received")
		}

		// Ensure directory exists and save the token to file
		if err := os.MkdirAll(filepath.Dir(tokenFilePath()), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(tokenFilePath(), []byte(token), 0600); err != nil {
			return err
		}

		// Notify user of successful login and token save
		fmt.Println("✅ Login successful. Token saved.")
		return nil
	},
}

func init() {
	// Define flags for login command
	loginCmd.Flags().String("username", "", "Username")
	loginCmd.Flags().String("password", "", "Password")
	RootCmd.AddCommand(loginCmd)
}

// tokenFilePath returns the file path where the token is stored
func tokenFilePath() string {
	return "./_data/data/.proc_token"
}

// loadToken reads the saved token from file
func loadToken() (string, error) {
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// apiURL returns the base URL for the API (to be loaded from config in the future)
func apiURL() string {
	// Future: can be configured from a config file
	return "http://localhost:8080"
}
