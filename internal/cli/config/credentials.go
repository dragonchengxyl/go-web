package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials stores API authentication credentials
type Credentials struct {
	APIUrl      string `json:"api_url"`
	AccessToken string `json:"access_token"`
	Email       string `json:"email"`
}

// CredentialsStore manages credential storage
type CredentialsStore struct {
	configDir string
}

// NewCredentialsStore creates a new credentials store
func NewCredentialsStore() (*CredentialsStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".studio-cli")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &CredentialsStore{
		configDir: configDir,
	}, nil
}

// Save saves credentials to disk
func (s *CredentialsStore) Save(creds *Credentials) error {
	credPath := filepath.Join(s.configDir, "credentials.json")

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// Load loads credentials from disk
func (s *CredentialsStore) Load() (*Credentials, error) {
	credPath := filepath.Join(s.configDir, "credentials.json")

	data, err := os.ReadFile(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in, please run 'studio-cli login' first")
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// Clear removes stored credentials
func (s *CredentialsStore) Clear() error {
	credPath := filepath.Join(s.configDir, "credentials.json")
	if err := os.Remove(credPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}
	return nil
}
