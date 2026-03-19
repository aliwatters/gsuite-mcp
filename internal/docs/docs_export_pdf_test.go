package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleDocsExportToPDF(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("pdf-doc", "PDF Export Test", "Some content")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "export to PDF successfully",
			args: map[string]any{
				"document_id": "pdf-doc",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["document_id"] != "pdf-doc" {
					t.Errorf("expected document_id 'pdf-doc', got %v", result["document_id"])
				}
				if result["name"] != "PDF Export Test" {
					t.Errorf("expected name 'PDF Export Test', got %v", result["name"])
				}
				if result["content_type"] != "application/pdf" {
					t.Errorf("expected content_type 'application/pdf', got %v", result["content_type"])
				}
				if result["encoding"] != "base64" {
					t.Errorf("expected encoding 'base64', got %v", result["encoding"])
				}
				if result["pdf_base64"] == nil || result["pdf_base64"] == "" {
					t.Error("expected pdf_base64 to be set")
				}
				pdfSize, ok := result["pdf_size"].(float64)
				if !ok || pdfSize <= 0 {
					t.Errorf("expected positive pdf_size, got %v", result["pdf_size"])
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
				"document_id": "nonexistent",
			},
			wantErr:     true,
			errContains: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsExportToPDF(context.Background(), request, fixtures.Deps)

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
