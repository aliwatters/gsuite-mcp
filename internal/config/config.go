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
const DefaultOAuthPort = 38917

// DriveAccess configures which shared drives are accessible via MCP tools.
// Set either Allowed (allowlist) or Blocked (blocklist), not both.
// My Drive is always accessible. If neither is set, all drives are accessible.
type DriveAccess struct {
	Allowed []string `json:"allowed,omitempty"` // Only these shared drives + My Drive
	Blocked []string `json:"blocked,omitempty"` // Everything except these shared drives
}

// CitationIndex maps an index name to its backing Sheet ID.
type CitationIndex struct {
	SheetID string `json:"sheet_id"`
}

// CitationConfig holds citation/large-doc-indexing settings.
type CitationConfig struct {
	Indexes map[string]CitationIndex `json:"indexes,omitempty"`
}

// Features holds feature flags.
type Features struct {
	LargeDocIndexing bool `json:"large_doc_indexing,omitempty"`
}

// Config holds the application configuration loaded from config.json.
type Config struct {
	OAuthPort   int             `json:"oauth_port"`
	DriveAccess *DriveAccess    `json:"drive_access,omitempty"`
	Features    *Features       `json:"features,omitempty"`
	Citation    *CitationConfig `json:"citation,omitempty"`
}

// Validate checks the configuration for errors.
func (c Config) Validate() error {
	if c.DriveAccess != nil && len(c.DriveAccess.Allowed) > 0 && len(c.DriveAccess.Blocked) > 0 {
		return fmt.Errorf("drive_access: cannot set both 'allowed' and 'blocked' — choose one mode")
	}
	return nil
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

	if err := fileCfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
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

// configDirOverride is set by SetConfigDir to use a custom config directory.
var configDirOverride string

// SetConfigDir overrides the default config directory.
// All path functions (ConfigPath, ClientSecretPath, CredentialsDir, etc.)
// derive from DefaultConfigDir, so they automatically respect this override.
func SetConfigDir(dir string) {
	configDirOverride = dir
}

// DefaultConfigDir returns the default configuration directory.
func DefaultConfigDir() string {
	if configDirOverride != "" {
		return configDirOverride
	}
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
