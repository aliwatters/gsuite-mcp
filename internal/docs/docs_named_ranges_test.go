package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

func TestHandleDocsListNamedRanges(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	// Add a document with named ranges
	fixtures.MockService.Documents["nr-doc"] = &docs.Document{
		DocumentId: "nr-doc",
		Title:      "Named Ranges Doc",
		NamedRanges: map[string]docs.NamedRanges{
			"bookmark1": {
				Name: "bookmark1",
				NamedRanges: []*docs.NamedRange{
					{
						Name:         "bookmark1",
						NamedRangeId: "nr-id-1",
						Ranges: []*docs.Range{
							{StartIndex: 5, EndIndex: 10},
						},
					},
				},
			},
			"section": {
				Name: "section",
				NamedRanges: []*docs.NamedRange{
					{
						Name:         "section",
						NamedRangeId: "nr-id-2",
						Ranges: []*docs.Range{
							{StartIndex: 20, EndIndex: 50},
						},
					},
				},
			},
		},
		Body: &docs.Body{Content: []*docs.StructuralElement{}},
	}

	// Empty document with no named ranges
	fixtures.MockService.Documents["empty-nr-doc"] = &docs.Document{
		DocumentId: "empty-nr-doc",
		Title:      "Empty NR Doc",
		Body:       &docs.Body{Content: []*docs.StructuralElement{}},
	}

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list named ranges",
			args: map[string]any{"document_id": "nr-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["document_id"] != "nr-doc" {
					t.Errorf("expected document_id 'nr-doc', got %v", result["document_id"])
				}
				count, _ := result["count"].(float64)
				if count != 2 {
					t.Errorf("expected 2 named ranges, got %v", count)
				}
				ranges, ok := result["named_ranges"].([]any)
				if !ok {
					t.Fatalf("expected named_ranges to be array, got %T", result["named_ranges"])
				}
				if len(ranges) != 2 {
					t.Fatalf("expected 2 named ranges, got %d", len(ranges))
				}
			},
		},
		{
			name: "empty document has no named ranges",
			args: map[string]any{"document_id": "empty-nr-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				count, _ := result["count"].(float64)
				if count != 0 {
					t.Errorf("expected 0 named ranges, got %v", count)
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
			name:        "document not found",
			args:        map[string]any{"document_id": "nonexistent"},
			wantErr:     true,
			errContains: "Docs API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsListNamedRanges(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := getDocsTextContent(result)

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

func TestHandleDocsCreateNamedRange(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("create-nr-doc", "Create NR Doc", "Hello World, this is a test document")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
		checkCalls  func(t *testing.T, mock *MockDocsService)
	}{
		{
			name: "create named range",
			args: map[string]any{
				"document_id": "create-nr-doc",
				"name":        "important-section",
				"start_index": float64(1),
				"end_index":   float64(12),
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success true")
				}
				if result["name"] != "important-section" {
					t.Errorf("expected name 'important-section', got %v", result["name"])
				}
				startIdx, _ := result["start_index"].(float64)
				endIdx, _ := result["end_index"].(float64)
				if startIdx != 1 || endIdx != 12 {
					t.Errorf("expected range 1-12, got %v-%v", startIdx, endIdx)
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				if len(mock.Calls.BatchUpdate) == 0 {
					t.Fatal("expected BatchUpdate to be called")
				}
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				if len(lastCall.Requests) != 1 {
					t.Fatalf("expected 1 request, got %d", len(lastCall.Requests))
				}
				req := lastCall.Requests[0].CreateNamedRange
				if req == nil {
					t.Fatal("expected CreateNamedRange request")
				}
				if req.Name != "important-section" {
					t.Errorf("expected name 'important-section', got %q", req.Name)
				}
				if req.Range.StartIndex != 1 || req.Range.EndIndex != 12 {
					t.Errorf("expected range 1-12, got %d-%d", req.Range.StartIndex, req.Range.EndIndex)
				}
			},
		},
		{
			name: "missing name",
			args: map[string]any{
				"document_id": "create-nr-doc",
				"start_index": float64(1),
				"end_index":   float64(5),
			},
			wantErr:     true,
			errContains: "name parameter is required",
		},
		{
			name: "missing start_index",
			args: map[string]any{
				"document_id": "create-nr-doc",
				"name":        "test",
				"end_index":   float64(5),
			},
			wantErr:     true,
			errContains: "start_index",
		},
		{
			name:        "missing document_id",
			args:        map[string]any{"name": "test", "start_index": float64(1), "end_index": float64(5)},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsCreateNamedRange(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := getDocsTextContent(result)

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
				if tt.checkCalls != nil {
					tt.checkCalls(t, fixtures.MockService)
				}
			}
		})
	}
}

func TestHandleDocsDeleteNamedRange(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("delete-nr-doc", "Delete NR Doc", "Hello World")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
		checkCalls  func(t *testing.T, mock *MockDocsService)
	}{
		{
			name: "delete named range",
			args: map[string]any{
				"document_id":    "delete-nr-doc",
				"named_range_id": "nr-id-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success true")
				}
				if result["named_range_id"] != "nr-id-123" {
					t.Errorf("expected named_range_id 'nr-id-123', got %v", result["named_range_id"])
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				if len(mock.Calls.BatchUpdate) == 0 {
					t.Fatal("expected BatchUpdate to be called")
				}
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				req := lastCall.Requests[0].DeleteNamedRange
				if req == nil {
					t.Fatal("expected DeleteNamedRange request")
				}
				if req.NamedRangeId != "nr-id-123" {
					t.Errorf("expected named_range_id 'nr-id-123', got %q", req.NamedRangeId)
				}
			},
		},
		{
			name: "missing named_range_id",
			args: map[string]any{
				"document_id": "delete-nr-doc",
			},
			wantErr:     true,
			errContains: "named_range_id parameter is required",
		},
		{
			name:        "missing document_id",
			args:        map[string]any{"named_range_id": "nr-id-123"},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsDeleteNamedRange(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := getDocsTextContent(result)

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
				if tt.checkCalls != nil {
					tt.checkCalls(t, fixtures.MockService)
				}
			}
		})
	}
}
