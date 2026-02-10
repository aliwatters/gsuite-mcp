package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleDocsInsertTable(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("table-doc", "Table Test", "Some content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "insert table successfully",
			args: map[string]any{
				"document_id": "table-doc",
				"rows":        float64(3),
				"columns":     float64(4),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["rows"].(float64) != 3 {
					t.Errorf("expected rows 3, got %v", result["rows"])
				}
				if result["columns"].(float64) != 4 {
					t.Errorf("expected columns 4, got %v", result["columns"])
				}
			},
		},
		{
			name: "insert table at specific index",
			args: map[string]any{
				"document_id": "table-doc",
				"rows":        float64(2),
				"columns":     float64(2),
				"index":       float64(5),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				msg, ok := result["message"].(string)
				if !ok || !strings.Contains(msg, "index 5") {
					t.Errorf("expected message about index 5, got %v", result["message"])
				}
			},
		},
		{
			name: "missing rows",
			args: map[string]any{
				"document_id": "table-doc",
				"columns":     float64(3),
			},
			wantErr:     true,
			errContains: "rows parameter is required",
		},
		{
			name: "missing columns",
			args: map[string]any{
				"document_id": "table-doc",
				"rows":        float64(3),
			},
			wantErr:     true,
			errContains: "columns parameter is required",
		},
		{
			name: "invalid index",
			args: map[string]any{
				"document_id": "table-doc",
				"rows":        float64(2),
				"columns":     float64(2),
				"index":       float64(0),
			},
			wantErr:     true,
			errContains: "index must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsInsertTable(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsInsertLink(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("link-doc", "Link Test", "Some content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "insert link successfully",
			args: map[string]any{
				"document_id": "link-doc",
				"text":        "Click here",
				"url":         "https://example.com",
				"index":       float64(1),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["text"] != "Click here" {
					t.Errorf("expected text 'Click here', got %v", result["text"])
				}
				if result["link_url"] != "https://example.com" {
					t.Errorf("expected link_url 'https://example.com', got %v", result["link_url"])
				}
			},
		},
		{
			name: "missing text",
			args: map[string]any{
				"document_id": "link-doc",
				"url":         "https://example.com",
				"index":       float64(1),
			},
			wantErr:     true,
			errContains: "text parameter is required",
		},
		{
			name: "missing url",
			args: map[string]any{
				"document_id": "link-doc",
				"text":        "Click here",
				"index":       float64(1),
			},
			wantErr:     true,
			errContains: "url parameter is required",
		},
		{
			name: "missing index",
			args: map[string]any{
				"document_id": "link-doc",
				"text":        "Click here",
				"url":         "https://example.com",
			},
			wantErr:     true,
			errContains: "index parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsInsertLink(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsBatchUpdate(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("batch-doc", "Batch Test", "Some content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "batch update successfully",
			args: map[string]any{
				"document_id": "batch-doc",
				"requests":    `[{"insertText":{"text":"Hello","location":{"index":1}}}]`,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["requests_count"].(float64) != 1 {
					t.Errorf("expected requests_count 1, got %v", result["requests_count"])
				}
			},
		},
		{
			name: "multiple requests",
			args: map[string]any{
				"document_id": "batch-doc",
				"requests":    `[{"insertText":{"text":"A","location":{"index":1}}},{"insertText":{"text":"B","location":{"index":2}}}]`,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["requests_count"].(float64) != 2 {
					t.Errorf("expected requests_count 2, got %v", result["requests_count"])
				}
			},
		},
		{
			name: "missing requests",
			args: map[string]any{
				"document_id": "batch-doc",
			},
			wantErr:     true,
			errContains: "requests parameter is required",
		},
		{
			name: "invalid JSON",
			args: map[string]any{
				"document_id": "batch-doc",
				"requests":    "not valid json",
			},
			wantErr:     true,
			errContains: "Failed to parse requests JSON",
		},
		{
			name: "empty requests array",
			args: map[string]any{
				"document_id": "batch-doc",
				"requests":    "[]",
			},
			wantErr:     true,
			errContains: "requests array cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsBatchUpdate(context.Background(), request, fixtures.Deps)

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
