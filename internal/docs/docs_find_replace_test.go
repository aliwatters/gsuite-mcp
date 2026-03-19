package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleDocsFindAndReplace(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("far-doc", "Find Replace Test", "Hello World! Hello World!")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "find and replace successfully",
			args: map[string]any{
				"document_id":  "far-doc",
				"find_text":    "Hello",
				"replace_text": "Hi",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["document_id"] != "far-doc" {
					t.Errorf("expected document_id 'far-doc', got %v", result["document_id"])
				}
				if result["find_text"] != "Hello" {
					t.Errorf("expected find_text 'Hello', got %v", result["find_text"])
				}
				if result["replace_text"] != "Hi" {
					t.Errorf("expected replace_text 'Hi', got %v", result["replace_text"])
				}
				if result["match_case"] != true {
					t.Errorf("expected match_case true (default), got %v", result["match_case"])
				}
			},
		},
		{
			name: "case insensitive replace",
			args: map[string]any{
				"document_id":  "far-doc",
				"find_text":    "hello",
				"replace_text": "Hi",
				"match_case":   false,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["match_case"] != false {
					t.Errorf("expected match_case false, got %v", result["match_case"])
				}
			},
		},
		{
			name: "delete by replacing with empty",
			args: map[string]any{
				"document_id":  "far-doc",
				"find_text":    "World",
				"replace_text": "",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
			},
		},
		{
			name:        "missing document_id",
			args:        map[string]any{"find_text": "Hello"},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
		{
			name: "missing find_text",
			args: map[string]any{
				"document_id": "far-doc",
			},
			wantErr:     true,
			errContains: "find_text parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsFindAndReplace(context.Background(), request, fixtures.Deps)

			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := getDocsTextContent(result)

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result")
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
