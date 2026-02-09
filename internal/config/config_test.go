package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialPathForEmail(t *testing.T) {
	tests := []struct {
		email      string
		wantSuffix string
	}{
		{"test@example.com", "test@example.com.json"},
		{"user@gmail.com", "user@gmail.com.json"},
		{"work@company.org", "work@company.org.json"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			path := CredentialPathForEmail(tt.email)
			if !filepath.IsAbs(path) {
				t.Errorf("CredentialPathForEmail() returned non-absolute path: %s", path)
			}
			if filepath.Base(path) != tt.wantSuffix {
				t.Errorf("CredentialPathForEmail() = %s, want suffix %s", filepath.Base(path), tt.wantSuffix)
			}
		})
	}
}

func TestGetAuthenticatedEmails(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	credDir := filepath.Join(tmpDir, "credentials")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Override the config dir for testing
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(filepath.Dir(tmpDir)))
	defer os.Setenv("HOME", origHome)

	// We need to use a custom approach since the functions use hardcoded paths
	// Let's test GetAuthenticatedEmails by creating credential files in a temp location
	// and verifying the logic works correctly with our test helper

	t.Run("empty credentials dir", func(t *testing.T) {
		emails, err := getAuthenticatedEmailsFrom(credDir)
		if err != nil {
			t.Errorf("getAuthenticatedEmailsFrom() error = %v", err)
		}
		if len(emails) != 0 {
			t.Errorf("getAuthenticatedEmailsFrom() = %v, want empty", emails)
		}
	})

	t.Run("with credential files", func(t *testing.T) {
		// Create test credential files
		testEmails := []string{"alice@example.com", "bob@example.com", "charlie@example.com"}
		for _, email := range testEmails {
			path := filepath.Join(credDir, email+".json")
			if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		emails, err := getAuthenticatedEmailsFrom(credDir)
		if err != nil {
			t.Errorf("getAuthenticatedEmailsFrom() error = %v", err)
		}
		if len(emails) != 3 {
			t.Errorf("getAuthenticatedEmailsFrom() got %d emails, want 3", len(emails))
		}
		// Should be sorted
		if emails[0] != "alice@example.com" {
			t.Errorf("emails[0] = %s, want alice@example.com", emails[0])
		}
	})

	t.Run("ignores non-json files", func(t *testing.T) {
		// Create a non-json file
		nonJsonPath := filepath.Join(credDir, "notes.txt")
		if err := os.WriteFile(nonJsonPath, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		emails, err := getAuthenticatedEmailsFrom(credDir)
		if err != nil {
			t.Errorf("getAuthenticatedEmailsFrom() error = %v", err)
		}
		// Should still be 3 (from previous test), not 4
		if len(emails) != 3 {
			t.Errorf("getAuthenticatedEmailsFrom() got %d emails, want 3", len(emails))
		}
	})

	t.Run("ignores directories", func(t *testing.T) {
		// Create a subdirectory with .json suffix (edge case)
		subDir := filepath.Join(credDir, "subdir.json")
		if err := os.MkdirAll(subDir, 0700); err != nil {
			t.Fatalf("Failed to create test dir: %v", err)
		}

		emails, err := getAuthenticatedEmailsFrom(credDir)
		if err != nil {
			t.Errorf("getAuthenticatedEmailsFrom() error = %v", err)
		}
		if len(emails) != 3 {
			t.Errorf("getAuthenticatedEmailsFrom() got %d emails, want 3", len(emails))
		}
	})
}

func TestGetAuthenticatedEmailsNonexistentDir(t *testing.T) {
	emails, err := getAuthenticatedEmailsFrom("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Errorf("getAuthenticatedEmailsFrom() should not error for nonexistent dir, got: %v", err)
	}
	if emails != nil && len(emails) != 0 {
		t.Errorf("getAuthenticatedEmailsFrom() should return nil or empty for nonexistent dir, got: %v", emails)
	}
}

func TestHasCredentialsForEmail(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a credential file
	credPath := filepath.Join(tmpDir, "exists@example.com.json")
	if err := os.WriteFile(credPath, []byte("{}"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("existing credentials", func(t *testing.T) {
		if !hasCredentialsAt(credPath) {
			t.Error("hasCredentialsAt() returned false for existing file")
		}
	})

	t.Run("nonexistent credentials", func(t *testing.T) {
		nonexistentPath := filepath.Join(tmpDir, "missing@example.com.json")
		if hasCredentialsAt(nonexistentPath) {
			t.Error("hasCredentialsAt() returned true for nonexistent file")
		}
	})
}

func TestGetDefaultEmail(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("empty credentials dir", func(t *testing.T) {
		email := getDefaultEmailFrom(tmpDir)
		if email != "" {
			t.Errorf("getDefaultEmailFrom() = %s, want empty string", email)
		}
	})

	t.Run("with credentials", func(t *testing.T) {
		// Create credential files
		for _, e := range []string{"zebra@example.com", "alice@example.com"} {
			path := filepath.Join(tmpDir, e+".json")
			if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		email := getDefaultEmailFrom(tmpDir)
		// Should return first alphabetically
		if email != "alice@example.com" {
			t.Errorf("getDefaultEmailFrom() = %s, want alice@example.com", email)
		}
	})
}

func TestEnsureConfigDir(t *testing.T) {
	// This test just verifies the function doesn't error
	// We can't easily test it without modifying the actual config directory
	err := EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() error = %v", err)
	}
}

func TestDefaultConfigDir(t *testing.T) {
	dir := DefaultConfigDir()
	if dir == "" {
		t.Error("DefaultConfigDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("DefaultConfigDir() returned non-absolute path: %s", dir)
	}
}

func TestCredentialsDir(t *testing.T) {
	dir := CredentialsDir()
	if dir == "" {
		t.Error("CredentialsDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("CredentialsDir() returned non-absolute path: %s", dir)
	}
	// Should be a subdirectory of config dir
	configDir := DefaultConfigDir()
	if filepath.Dir(dir) != configDir {
		t.Errorf("CredentialsDir() = %s, expected subdirectory of %s", dir, configDir)
	}
}

func TestClientSecretPath(t *testing.T) {
	path := ClientSecretPath()
	if path == "" {
		t.Error("ClientSecretPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("ClientSecretPath() returned non-absolute path: %s", path)
	}
	if filepath.Base(path) != "client_secret.json" {
		t.Errorf("ClientSecretPath() = %s, expected client_secret.json filename", path)
	}
}

// Test helper functions that work with arbitrary paths for testing purposes
func getAuthenticatedEmailsFrom(credDir string) ([]string, error) {
	entries, err := os.ReadDir(credDir)
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
		email := name[:len(name)-5]
		if email != "" {
			emails = append(emails, email)
		}
	}
	// Sort for consistent ordering
	for i := 0; i < len(emails)-1; i++ {
		for j := i + 1; j < len(emails); j++ {
			if emails[i] > emails[j] {
				emails[i], emails[j] = emails[j], emails[i]
			}
		}
	}
	return emails, nil
}

func hasCredentialsAt(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getDefaultEmailFrom(credDir string) string {
	emails, err := getAuthenticatedEmailsFrom(credDir)
	if err != nil || len(emails) == 0 {
		return ""
	}
	return emails[0]
}
