package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/cli/config"
)

// LoginCmd represents the login command
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Studio API",
	Long:  `Login to the Studio platform and store credentials locally.`,
	RunE:  runLogin,
}

var (
	loginEmail    string
	loginPassword string
	loginAPIURL   string
)

func init() {
	LoginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Email address")
	LoginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password")
	LoginCmd.Flags().StringVar(&loginAPIURL, "api-url", "http://localhost:8080", "API URL")
	LoginCmd.MarkFlagRequired("email")
	LoginCmd.MarkFlagRequired("password")
}

func runLogin(cmd *cobra.Command, args []string) error {
	fmt.Println("🔐 Logging in to Studio platform...")

	// Prepare login request
	reqBody := map[string]string{
		"email":    loginEmail,
		"password": loginPassword,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send login request
	resp, err := http.Post(
		loginAPIURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Save credentials
	store, err := config.NewCredentialsStore()
	if err != nil {
		return fmt.Errorf("failed to create credentials store: %w", err)
	}

	creds := &config.Credentials{
		APIUrl:      loginAPIURL,
		AccessToken: result.Data.AccessToken,
		Email:       loginEmail,
	}

	if err := store.Save(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("✅ Login successful!")
	fmt.Printf("📧 Logged in as: %s\n", loginEmail)
	fmt.Printf("🔗 API URL: %s\n", loginAPIURL)

	return nil
}
