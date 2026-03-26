package driveactivity

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	da "google.golang.org/api/driveactivity/v2"
)

func TestHandleDriveActivityQuery(t *testing.T) {
	fixtures := NewDriveActivityTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "query by item_id successfully",
			args: map[string]any{
				"item_id": "abc123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["activity_count"].(float64)
				if !ok || count < 1 {
					t.Errorf("expected activity_count >= 1, got %v", result["activity_count"])
				}
				activities, ok := result["activities"].([]any)
				if !ok || len(activities) < 1 {
					t.Errorf("expected at least 1 activity, got %v", result["activities"])
				}
				// Check first activity structure
				if len(activities) > 0 {
					activity, ok := activities[0].(map[string]any)
					if !ok {
						t.Fatal("expected activity to be a map")
					}
					if activity["action"] == nil {
						t.Error("expected action field in activity")
					}
					if activity["actors"] == nil {
						t.Error("expected actors field in activity")
					}
					if activity["targets"] == nil {
						t.Error("expected targets field in activity")
					}
				}
			},
		},
		{
			name: "query by folder_id successfully",
			args: map[string]any{
				"folder_id": "folder123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["activity_count"].(float64)
				if !ok || count < 1 {
					t.Errorf("expected activity_count >= 1, got %v", result["activity_count"])
				}
			},
		},
		{
			name: "query with URL extracts ID",
			args: map[string]any{
				"item_id": "https://drive.google.com/file/d/abc123/view",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				// Should extract ID from URL and query successfully
				if result["activity_count"] == nil {
					t.Error("expected activity_count to be set")
				}
			},
		},
		{
			name: "query with filter",
			args: map[string]any{
				"item_id": "abc123",
				"filter":  "detail.action_detail_case:EDIT",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				// Verify the filter was passed to the service
				calls := fixtures.MockService.Calls.QueryActivity
				lastCall := calls[len(calls)-1]
				if lastCall.Filter != "detail.action_detail_case:EDIT" {
					t.Errorf("expected filter to be passed, got %v", lastCall.Filter)
				}
			},
		},
		{
			name:        "missing both item_id and folder_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "either item_id or folder_id parameter is required",
		},
		{
			name: "item not found returns empty activities",
			args: map[string]any{
				"item_id": "nonexistent",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["activity_count"].(float64)
				if !ok || count != 0 {
					t.Errorf("expected activity_count 0, got %v", result["activity_count"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDriveActivityQuery(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := getTextContent(result)

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestDriveActivityServiceErrors(t *testing.T) {
	fixtures := NewDriveActivityTestFixtures()

	t.Run("API error on query", func(t *testing.T) {
		fixtures.MockService.Errors.QueryActivity = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.QueryActivity = nil }()

		request := common.CreateMCPRequest(map[string]any{"item_id": "abc123"})
		result, err := testableDriveActivityQuery(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
		text := getTextContent(result)
		if !strings.Contains(text, "Drive Activity API error") {
			t.Errorf("expected Drive Activity API error, got: %s", text)
		}
	})
}

func TestFormatActionDetail(t *testing.T) {
	tests := []struct {
		name   string
		detail *da.ActionDetail
		want   string
	}{
		{"edit", &da.ActionDetail{Edit: &da.Edit{}}, "EDIT"},
		{"create", &da.ActionDetail{Create: &da.Create{}}, "CREATE"},
		{"delete", &da.ActionDetail{Delete: &da.Delete{}}, "DELETE"},
		{"move", &da.ActionDetail{Move: &da.Move{}}, "MOVE"},
		{"restore", &da.ActionDetail{Restore: &da.Restore{}}, "RESTORE"},
		{"rename with titles", &da.ActionDetail{Rename: &da.Rename{OldTitle: "A", NewTitle: "B"}}, "RENAME: A -> B"},
		{"rename without titles", &da.ActionDetail{Rename: &da.Rename{}}, "RENAME"},
		{"permission_change", &da.ActionDetail{PermissionChange: &da.PermissionChange{}}, "PERMISSION_CHANGE"},
		{"comment", &da.ActionDetail{Comment: &da.Comment{}}, "COMMENT"},
		{"unknown", &da.ActionDetail{}, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatActionDetail(tt.detail)
			if got != tt.want {
				t.Errorf("formatActionDetail() = %q, want %q", got, tt.want)
			}
		})
	}
}
