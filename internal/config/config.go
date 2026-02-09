// Package config provides configuration management for gsuite-mcp.
// Uses dynamic account discovery - no pre-configuration required.
package config

import (
	"os"
	"path/filepath"
	"sort"
)

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
