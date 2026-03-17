// Package config provides configuration management for gsuite-mcp.
// Uses dynamic account discovery - no pre-configuration required.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DefaultOAuthPort is the default port used for the OAuth callback server.
const DefaultOAuthPort = 8100

// Config holds the application configuration loaded from config.json.
type Config struct {
	OAuthPort int `json:"oauth_port"`
}

// ConfigPath returns the path to config.json.
func ConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.json")
}

// LoadConfig loads config from the default config.json path.
func LoadConfig() (Config, error) {
	return loadConfigFromPath(ConfigPath())
}

// loadConfigFromPath loads config from the given path.
// Returns default Config if the file doesn't exist (config.json is optional).
func loadConfigFromPath(path string) (Config, error) {
	cfg := Config{OAuthPort: DefaultOAuthPort}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return cfg, fmt.Errorf("parsing config.json: %w", err)
	}

	if fileCfg.OAuthPort == 0 {
		fileCfg.OAuthPort = DefaultOAuthPort
	}

	return fileCfg, nil
}

// WriteDefaultConfig creates config.json with defaults if it doesn't already exist.
// Returns true if the file was created, false if it already existed.
func WriteDefaultConfig() (bool, error) {
	return writeDefaultConfigTo(ConfigPath())
}

// writeDefaultConfigTo creates config.json at the given path with defaults.
// Returns true if the file was created, false if it already existed.
func writeDefaultConfigTo(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("checking config.json: %w", err)
	}

	cfg := Config{OAuthPort: DefaultOAuthPort}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshaling default config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return false, fmt.Errorf("writing config.json: %w", err)
	}

	return true, nil
}

// DefaultConfigDir returns the default configuration directory.
func DefaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gsuite-mcp")
}

// CredentialsDir returns the credentials directory path.
func CredentialsDir() string {
	return filepath.Join(DefaultConfigDir(), "credentials")
}

// ClientSecretPath returns the OAuth client secret file path.
func ClientSecretPath() string {
	return filepath.Join(DefaultConfigDir(), "client_secret.json")
}

// CredentialPathForEmail returns the credential file path for an email address.
func CredentialPathForEmail(email string) string {
	return filepath.Join(CredentialsDir(), email+".json")
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
func EnsureConfigDir() error {
	if err := os.MkdirAll(CredentialsDir(), 0700); err != nil {
		return err
	}
	return nil
}

// GetAuthenticatedEmails returns a list of all emails with saved credentials.
// Returns emails sorted alphabetically.
func GetAuthenticatedEmails() ([]string, error) {
	entries, err := os.ReadDir(CredentialsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var emails []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		// Extract email from filename (format: email.json)
		email := name[:len(name)-5]
		if email != "" {
			emails = append(emails, email)
		}
	}
	sort.Strings(emails)
	return emails, nil
}

// HasCredentialsForEmail checks if credentials exist for an email.
func HasCredentialsForEmail(email string) bool {
	path := CredentialPathForEmail(email)
	_, err := os.Stat(path)
	return err == nil
}

// GetDefaultEmail returns the first authenticated email, or empty if none.
func GetDefaultEmail() string {
	emails, err := GetAuthenticatedEmails()
	if err != nil || len(emails) == 0 {
		return ""
	}
	return emails[0]
}
