package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

func TestHandleDocsGetStructure(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	// Add a document with multiple paragraph styles
	fixtures.MockService.Documents["struct-doc"] = &docs.Document{
		DocumentId: "struct-doc",
		Title:      "Structured Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   20,
					Paragraph: &docs.Paragraph{
						ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "TITLE"},
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   20,
								TextRun:    &docs.TextRun{Content: "My Document Title\n"},
							},
						},
					},
				},
				{
					StartIndex: 20,
					EndIndex:   45,
					Paragraph: &docs.Paragraph{
						ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "HEADING_1"},
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 20,
								EndIndex:   45,
								TextRun:    &docs.TextRun{Content: "Introduction\n"},
							},
						},
					},
				},
				{
					StartIndex: 45,
					EndIndex:   80,
					Paragraph: &docs.Paragraph{
						ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 45,
								EndIndex:   80,
								TextRun:    &docs.TextRun{Content: "Some body text here.\n"},
							},
						},
					},
				},
			},
		},
	}

	// Add a document with a table
	fixtures.MockService.Documents["table-struct-doc"] = &docs.Document{
		DocumentId: "table-struct-doc",
		Title:      "Table Struct Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   10,
					Paragraph: &docs.Paragraph{
						ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   10,
								TextRun:    &docs.TextRun{Content: "Preamble.\n"},
							},
						},
					},
				},
				{
					StartIndex: 10,
					EndIndex:   100,
					Table: &docs.Table{
						Columns: 4,
						TableRows: []*docs.TableRow{
							{TableCells: []*docs.TableCell{{}, {}, {}, {}}},
							{TableCells: []*docs.TableCell{{}, {}, {}, {}}},
							{TableCells: []*docs.TableCell{{}, {}, {}, {}}},
						},
					},
				},
			},
		},
	}

	// Add a document with a long paragraph for truncation testing
	longText := strings.Repeat("abcdefghij", 10) + "\n" // 101 chars
	fixtures.MockService.Documents["long-doc"] = &docs.Document{
		DocumentId: "long-doc",
		Title:      "Long Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   int64(len(longText)),
					Paragraph: &docs.Paragraph{
						ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   int64(len(longText)),
								TextRun:    &docs.TextRun{Content: longText},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "basic paragraph structure",
			args: map[string]any{"document_id": "struct-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["document_id"] != "struct-doc" {
					t.Errorf("expected document_id 'struct-doc', got %v", result["document_id"])
				}
				if result["title"] != "Structured Doc" {
					t.Errorf("expected title 'Structured Doc', got %v", result["title"])
				}

				elements, ok := result["elements"].([]any)
				if !ok {
					t.Fatalf("expected elements to be array, got %T", result["elements"])
				}
				if len(elements) != 3 {
					t.Fatalf("expected 3 elements, got %d", len(elements))
				}

				// Check first element (TITLE)
				e0 := elements[0].(map[string]any)
				if e0["type"] != "paragraph" {
					t.Errorf("expected type 'paragraph', got %v", e0["type"])
				}
				if e0["style"] != "TITLE" {
					t.Errorf("expected style 'TITLE', got %v", e0["style"])
				}
				if e0["startIndex"].(float64) != 0 {
					t.Errorf("expected startIndex 0, got %v", e0["startIndex"])
				}
				if e0["endIndex"].(float64) != 20 {
					t.Errorf("expected endIndex 20, got %v", e0["endIndex"])
				}
				if e0["text"] != "My Document Title" {
					t.Errorf("expected text 'My Document Title', got %v", e0["text"])
				}

				// Check second element (HEADING_1)
				e1 := elements[1].(map[string]any)
				if e1["style"] != "HEADING_1" {
					t.Errorf("expected style 'HEADING_1', got %v", e1["style"])
				}
				if e1["startIndex"].(float64) != 20 {
					t.Errorf("expected startIndex 20, got %v", e1["startIndex"])
				}

				// Check third element (NORMAL_TEXT)
				e2 := elements[2].(map[string]any)
				if e2["style"] != "NORMAL_TEXT" {
					t.Errorf("expected style 'NORMAL_TEXT', got %v", e2["style"])
				}
				if e2["text"] != "Some body text here." {
					t.Errorf("expected text 'Some body text here.', got %v", e2["text"])
				}
			},
		},
		{
			name: "document with table",
			args: map[string]any{"document_id": "table-struct-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				elements := result["elements"].([]any)
				if len(elements) != 2 {
					t.Fatalf("expected 2 elements, got %d", len(elements))
				}

				// First element is a paragraph
				e0 := elements[0].(map[string]any)
				if e0["type"] != "paragraph" {
					t.Errorf("expected type 'paragraph', got %v", e0["type"])
				}

				// Second element is a table
				e1 := elements[1].(map[string]any)
				if e1["type"] != "table" {
					t.Errorf("expected type 'table', got %v", e1["type"])
				}
				if e1["rows"].(float64) != 3 {
					t.Errorf("expected rows 3, got %v", e1["rows"])
				}
				if e1["cols"].(float64) != 4 {
					t.Errorf("expected cols 4, got %v", e1["cols"])
				}
				if e1["startIndex"].(float64) != 10 {
					t.Errorf("expected startIndex 10, got %v", e1["startIndex"])
				}
				if e1["endIndex"].(float64) != 100 {
					t.Errorf("expected endIndex 100, got %v", e1["endIndex"])
				}
			},
		},
		{
			name: "text preview truncation",
			args: map[string]any{"document_id": "long-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				elements := result["elements"].([]any)
				e0 := elements[0].(map[string]any)
				text := e0["text"].(string)
				if len(text) != 83 { // 80 chars + "..."
					t.Errorf("expected truncated text of length 83, got %d", len(text))
				}
				if !strings.HasSuffix(text, "...") {
					t.Errorf("expected text to end with '...', got %q", text)
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
			args:        map[string]any{"document_id": "nonexistent-doc"},
			wantErr:     true,
			errContains: "Docs API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsGetStructure(context.Background(), request, fixtures.Deps)

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
