package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleDocsImportToGoogleDoc(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "import plain text successfully",
			args: map[string]any{
				"title":   "Imported Doc",
				"content": "This is some plain text content.",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Imported Doc" {
					t.Errorf("expected title 'Imported Doc', got %v", result["title"])
				}
				if result["document_id"] == nil || result["document_id"] == "" {
					t.Error("expected document_id to be set")
				}
				if result["source_content_type"] != "text/plain" {
					t.Errorf("expected source_content_type 'text/plain', got %v", result["source_content_type"])
				}
				url, ok := result["url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/document/d/") {
					t.Errorf("expected valid docs URL, got %v", result["url"])
				}
			},
		},
		{
			name: "import HTML content",
			args: map[string]any{
				"title":        "HTML Doc",
				"content":      "<h1>Title</h1><p>Body text</p>",
				"content_type": "text/html",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "HTML Doc" {
					t.Errorf("expected title 'HTML Doc', got %v", result["title"])
				}
				if result["source_content_type"] != "text/html" {
					t.Errorf("expected source_content_type 'text/html', got %v", result["source_content_type"])
				}
			},
		},
		{
			name: "import markdown content",
			args: map[string]any{
				"title":        "Markdown Doc",
				"content":      "# Title\n\nSome **bold** text.",
				"content_type": "text/markdown",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["source_content_type"] != "text/markdown" {
					t.Errorf("expected source_content_type 'text/markdown', got %v", result["source_content_type"])
				}
			},
		},
		{
			name: "invalid content_type",
			args: map[string]any{
				"title":        "Bad Type",
				"content":      "content",
				"content_type": "application/json",
			},
			wantErr:     true,
			errContains: "Invalid content_type",
		},
		{
			name: "missing title",
			args: map[string]any{
				"content": "some content",
			},
			wantErr:     true,
			errContains: "title parameter is required",
		},
		{
			name: "missing content",
			args: map[string]any{
				"title": "No Content",
			},
			wantErr:     true,
			errContains: "content parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsImportToGoogleDoc(context.Background(), request, fixtures.Deps)

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

func TestMarkdownToBasicHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "heading",
			input:    "# Title",
			contains: []string{"<h1>Title</h1>"},
		},
		{
			name:     "bold",
			input:    "**bold**",
			contains: []string{"<b>bold</b>"},
		},
		{
			name:     "italic",
			input:    "*italic*",
			contains: []string{"<i>italic</i>"},
		},
		{
			name:     "list item",
			input:    "- item",
			contains: []string{"<ul><li>item</li></ul>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markdownToBasicHTML(tt.input)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected HTML to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}
