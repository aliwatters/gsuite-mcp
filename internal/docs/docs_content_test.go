package docs

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleDocsCreate(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "create document successfully",
			args: map[string]any{
				"title": "My New Document",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "My New Document" {
					t.Errorf("expected title 'My New Document', got %v", result["title"])
				}
				if result["document_id"] == nil || result["document_id"] == "" {
					t.Error("expected document_id to be set")
				}
				url, ok := result["url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/document/d/") {
					t.Errorf("expected valid docs URL, got %v", result["url"])
				}
			},
		},
		{
			name:        "missing title parameter",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "title parameter is required",
		},
		{
			name: "empty title parameter",
			args: map[string]any{
				"title": "",
			},
			wantErr:     true,
			errContains: "title parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsCreate(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsGet(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("doc-123", "Test Doc", "Hello World!")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get document by ID",
			args: map[string]any{
				"document_id": "doc-123",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Test Doc" {
					t.Errorf("expected title 'Test Doc', got %v", result["title"])
				}
				if result["document_id"] != "doc-123" {
					t.Errorf("expected document_id 'doc-123', got %v", result["document_id"])
				}
				content, ok := result["content"].(string)
				if !ok || !strings.Contains(content, "Hello World!") {
					t.Errorf("expected content containing 'Hello World!', got %v", result["content"])
				}
			},
		},
		{
			name: "get document by URL",
			args: map[string]any{
				"document_id": "https://docs.google.com/document/d/doc-123/edit",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["document_id"] != "doc-123" {
					t.Errorf("expected document_id 'doc-123', got %v", result["document_id"])
				}
			},
		},
		{
			name:        "missing document_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
		{
			name: "document not found",
			args: map[string]any{
				"document_id": "nonexistent-doc",
			},
			wantErr:     true,
			errContains: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsGet(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsGetMetadata(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("meta-doc", "Metadata Test", "One two three four five")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get metadata successfully",
			args: map[string]any{
				"document_id": "meta-doc",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Metadata Test" {
					t.Errorf("expected title 'Metadata Test', got %v", result["title"])
				}
				if result["revision_id"] == nil {
					t.Error("expected revision_id to be set")
				}
				// Word count should be 5
				wordCount, ok := result["word_count"].(float64)
				if !ok || wordCount != 5 {
					t.Errorf("expected word_count 5, got %v", result["word_count"])
				}
			},
		},
		{
			name:        "missing document_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsGetMetadata(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsAppendText(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("append-doc", "Append Test", "Initial content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "append text successfully",
			args: map[string]any{
				"document_id": "append-doc",
				"text":        " More content",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["document_id"] != "append-doc" {
					t.Errorf("expected document_id 'append-doc', got %v", result["document_id"])
				}
			},
		},
		{
			name: "missing document_id",
			args: map[string]any{
				"text": "Some text",
			},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
		{
			name: "missing text",
			args: map[string]any{
				"document_id": "append-doc",
			},
			wantErr:     true,
			errContains: "text parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsAppendText(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsInsertText(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("insert-doc", "Insert Test", "Initial content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "insert text successfully",
			args: map[string]any{
				"document_id": "insert-doc",
				"text":        "Inserted: ",
				"index":       float64(1),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				msg, ok := result["message"].(string)
				if !ok || !strings.Contains(msg, "index 1") {
					t.Errorf("expected message about index 1, got %v", result["message"])
				}
			},
		},
		{
			name: "missing index",
			args: map[string]any{
				"document_id": "insert-doc",
				"text":        "Some text",
			},
			wantErr:     true,
			errContains: "index parameter is required",
		},
		{
			name: "index less than 1",
			args: map[string]any{
				"document_id": "insert-doc",
				"text":        "Some text",
				"index":       float64(0),
			},
			wantErr:     true,
			errContains: "index must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsInsertText(context.Background(), request, fixtures.Deps)

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

func TestExtractDocumentID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Plain document IDs
		{"abc123", "abc123"},
		{"1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"},

		// URLs with /edit
		{"https://docs.google.com/document/d/abc123/edit", "abc123"},
		{"https://docs.google.com/document/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit", "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"},

		// URLs without /edit
		{"https://docs.google.com/document/d/abc123", "abc123"},

		// URLs with query parameters
		{"https://docs.google.com/document/d/abc123/edit?usp=sharing", "abc123"},

		// Document IDs with special characters
		{"doc-with-dashes_and_underscores", "doc-with-dashes_and_underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := common.ExtractGoogleResourceID(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractGoogleResourceID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDocsServiceErrors(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	t.Run("API error on create", func(t *testing.T) {
		fixtures.MockService.Errors.Create = errors.New("API quota exceeded")
		defer func() { fixtures.MockService.Errors.Create = nil }()

		request := common.CreateMCPRequest(map[string]any{
			"title": "Test Doc",
		})
		result, err := testableDocsCreate(context.Background(), request, fixtures.Deps)

		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}

		if !result.IsError {
			t.Error("expected error result")
		}

		text := getDocsTextContent(result)
		if !strings.Contains(text, "API quota exceeded") {
			t.Errorf("expected error about quota, got %q", text)
		}
	})

	t.Run("API error on get", func(t *testing.T) {
		fixtures.MockService.Errors.GetDocument = errors.New("permission denied")
		defer func() { fixtures.MockService.Errors.GetDocument = nil }()

		request := common.CreateMCPRequest(map[string]any{
			"document_id": "test-doc-1",
		})
		result, err := testableDocsGet(context.Background(), request, fixtures.Deps)

		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}

		if !result.IsError {
			t.Error("expected error result")
		}

		text := getDocsTextContent(result)
		if !strings.Contains(text, "permission denied") {
			t.Errorf("expected error about permission, got %q", text)
		}
	})
}

// === Phase 2 Extended Tools Tests ===

func TestHandleDocsReplaceText(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("replace-doc", "Replace Test", "Hello World! Hello World!")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "replace text successfully",
			args: map[string]any{
				"document_id":  "replace-doc",
				"find_text":    "Hello",
				"replace_text": "Hi",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["document_id"] != "replace-doc" {
					t.Errorf("expected document_id 'replace-doc', got %v", result["document_id"])
				}
			},
		},
		{
			name: "delete by replacing with empty",
			args: map[string]any{
				"document_id":  "replace-doc",
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
			name: "missing document_id",
			args: map[string]any{
				"find_text":    "Hello",
				"replace_text": "Hi",
			},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
		{
			name: "missing find_text",
			args: map[string]any{
				"document_id":  "replace-doc",
				"replace_text": "Hi",
			},
			wantErr:     true,
			errContains: "find_text parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsReplaceText(context.Background(), request, fixtures.Deps)

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

func TestHandleDocsDeleteText(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("delete-doc", "Delete Test", "Hello World!")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "delete text successfully",
			args: map[string]any{
				"document_id": "delete-doc",
				"start_index": float64(1),
				"end_index":   float64(6),
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				msg, ok := result["message"].(string)
				if !ok || !strings.Contains(msg, "1 to 6") {
					t.Errorf("expected message about deleting 1 to 6, got %v", result["message"])
				}
			},
		},
		{
			name: "missing start_index",
			args: map[string]any{
				"document_id": "delete-doc",
				"end_index":   float64(6),
			},
			wantErr:     true,
			errContains: "start_index parameter is required",
		},
		{
			name: "missing end_index",
			args: map[string]any{
				"document_id": "delete-doc",
				"start_index": float64(1),
			},
			wantErr:     true,
			errContains: "end_index parameter is required",
		},
		{
			name: "end_index less than start_index",
			args: map[string]any{
				"document_id": "delete-doc",
				"start_index": float64(10),
				"end_index":   float64(5),
			},
			wantErr:     true,
			errContains: "end_index must be greater than start_index",
		},
		{
			name: "start_index less than 1",
			args: map[string]any{
				"document_id": "delete-doc",
				"start_index": float64(0),
				"end_index":   float64(5),
			},
			wantErr:     true,
			errContains: "start_index must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsDeleteText(context.Background(), request, fixtures.Deps)

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
