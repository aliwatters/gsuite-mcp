package common

import (
	"testing"
)

func TestRequireStringArg(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		key       string
		wantVal   string
		wantError bool
	}{
		{
			name:    "present and non-empty",
			args:    map[string]any{"key": "value"},
			key:     "key",
			wantVal: "value",
		},
		{
			name:      "missing key",
			args:      map[string]any{},
			key:       "key",
			wantError: true,
		},
		{
			name:      "empty string",
			args:      map[string]any{"key": ""},
			key:       "key",
			wantError: true,
		},
		{
			name:      "wrong type",
			args:      map[string]any{"key": 123},
			key:       "key",
			wantError: true,
		},
		{
			name:      "nil args",
			args:      nil,
			key:       "key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, errResult := RequireStringArg(tt.args, tt.key)
			if tt.wantError {
				if errResult == nil {
					t.Error("expected error result, got nil")
				}
				if val != "" {
					t.Errorf("expected empty string on error, got %q", val)
				}
			} else {
				if errResult != nil {
					t.Errorf("unexpected error result: %v", errResult)
				}
				if val != tt.wantVal {
					t.Errorf("got %q, want %q", val, tt.wantVal)
				}
			}
		})
	}
}

func TestParseStringArg(t *testing.T) {
	args := map[string]any{"name": "alice", "empty": ""}
	if got := ParseStringArg(args, "name", "default"); got != "alice" {
		t.Errorf("got %q, want %q", got, "alice")
	}
	if got := ParseStringArg(args, "missing", "default"); got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
	if got := ParseStringArg(args, "empty", "default"); got != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
}

func TestParseBoolArg(t *testing.T) {
	args := map[string]any{"flag": true}
	if got := ParseBoolArg(args, "flag", false); got != true {
		t.Errorf("got %v, want true", got)
	}
	if got := ParseBoolArg(args, "missing", true); got != true {
		t.Errorf("got %v, want true", got)
	}
}

func TestParseMaxResults(t *testing.T) {
	args := map[string]any{"max_results": float64(50)}
	if got := ParseMaxResults(args, 10, 100); got != 50 {
		t.Errorf("got %d, want 50", got)
	}
	if got := ParseMaxResults(args, 10, 30); got != 30 {
		t.Errorf("got %d, want 30 (capped)", got)
	}
	if got := ParseMaxResults(map[string]any{}, 10, 100); got != 10 {
		t.Errorf("got %d, want 10 (default)", got)
	}
}
