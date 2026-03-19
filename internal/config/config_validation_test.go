package config

import (
	"testing"
)

func TestConfig_Validate_BothAllowedAndBlocked(t *testing.T) {
	cfg := Config{
		DriveAccess: &DriveAccess{
			Allowed: []string{"A"},
			Blocked: []string{"B"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when both allowed and blocked are set")
	}
}

func TestConfig_Validate_AllowedOnly(t *testing.T) {
	cfg := Config{
		DriveAccess: &DriveAccess{
			Allowed: []string{"Marketing"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_BlockedOnly(t *testing.T) {
	cfg := Config{
		DriveAccess: &DriveAccess{
			Blocked: []string{"SENSITIVE"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_NoDriveAccess(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_EmptyDriveAccess(t *testing.T) {
	cfg := Config{DriveAccess: &DriveAccess{}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
