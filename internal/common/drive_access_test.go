package common

import (
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/config"
)

func TestNewDriveAccessFilter_NilConfig(t *testing.T) {
	f := NewDriveAccessFilter(nil)
	if f != nil {
		t.Fatal("expected nil filter for nil config")
	}
}

func TestNewDriveAccessFilter_EmptyConfig(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{})
	if f != nil {
		t.Fatal("expected nil filter for empty config")
	}
}

func TestDriveAccessFilter_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		filter *DriveAccessFilter
		want   bool
	}{
		{"nil filter", nil, false},
		{"allowed set", NewDriveAccessFilter(&config.DriveAccess{Allowed: []string{"Marketing"}}), true},
		{"blocked set", NewDriveAccessFilter(&config.DriveAccess{Blocked: []string{"SENSITIVE"}}), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDriveAccessFilter_Check_Allowlist(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{Allowed: []string{"Marketing"}})
	f.ResolveDriveNames([]DriveInfo{
		{ID: "drive-marketing", Name: "Marketing"},
		{ID: "drive-sensitive", Name: "SENSITIVE"},
	})

	tests := []struct {
		name    string
		driveID string
		wantErr bool
	}{
		{"my drive always allowed", "", false},
		{"allowed drive", "drive-marketing", false},
		{"blocked drive", "drive-sensitive", true},
		{"unknown drive", "drive-unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.Check(tt.driveID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check(%q) error = %v, wantErr %v", tt.driveID, err, tt.wantErr)
			}
		})
	}
}

func TestDriveAccessFilter_Check_Blocklist(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{Blocked: []string{"SENSITIVE", "HR"}})
	f.ResolveDriveNames([]DriveInfo{
		{ID: "drive-marketing", Name: "Marketing"},
		{ID: "drive-sensitive", Name: "SENSITIVE"},
		{ID: "drive-hr", Name: "HR"},
	})

	tests := []struct {
		name    string
		driveID string
		wantErr bool
	}{
		{"my drive always allowed", "", false},
		{"unblocked drive", "drive-marketing", false},
		{"blocked drive sensitive", "drive-sensitive", true},
		{"blocked drive hr", "drive-hr", true},
		{"unknown drive allowed", "drive-unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.Check(tt.driveID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Check(%q) error = %v, wantErr %v", tt.driveID, err, tt.wantErr)
			}
		})
	}
}

func TestDriveAccessFilter_Check_CaseInsensitive(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{Allowed: []string{"marketing"}})
	f.ResolveDriveNames([]DriveInfo{
		{ID: "drive-marketing", Name: "Marketing"},
	})

	if err := f.Check("drive-marketing"); err != nil {
		t.Errorf("lowercase config should match titlecase drive name: %v", err)
	}
}

func TestDriveAccessFilter_Check_DirectID(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{Allowed: []string{"drive-direct-id"}})
	f.ResolveDriveNames([]DriveInfo{
		{ID: "drive-direct-id", Name: "Some Drive"},
	})

	if err := f.Check("drive-direct-id"); err != nil {
		t.Errorf("direct ID in config should work: %v", err)
	}
}

func TestDriveAccessFilter_Check_UnresolvedFailsOpen(t *testing.T) {
	f := NewDriveAccessFilter(&config.DriveAccess{Blocked: []string{"SENSITIVE"}})
	// Don't call ResolveDriveNames — should fail open

	if err := f.Check("any-drive"); err != nil {
		t.Errorf("unresolved filter should fail open: %v", err)
	}
}

func TestDriveAccessFilter_NilSafe(t *testing.T) {
	var f *DriveAccessFilter
	f.ResolveDriveNames(nil)
	if err := f.Check("any"); err != nil {
		t.Errorf("nil filter should always allow: %v", err)
	}
}
