package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

func TestHandleDocsGetSuggestedEdits(t *testing.T) {
	fixtures := NewDocsTestFixtures()

	// Document with suggested insertions and deletions
	fixtures.MockService.Documents["suggested-doc"] = &docs.Document{
		DocumentId: "suggested-doc",
		Title:      "Suggested Edits Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   20,
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   5,
								TextRun: &docs.TextRun{
									Content:               "Hello",
									SuggestedInsertionIds: []string{"suggestion-1"},
								},
							},
							{
								StartIndex: 5,
								EndIndex:   11,
								TextRun: &docs.TextRun{
									Content:              " World",
									SuggestedDeletionIds: []string{"suggestion-2"},
								},
							},
							{
								StartIndex: 11,
								EndIndex:   20,
								TextRun: &docs.TextRun{
									Content: " testing\n",
									SuggestedTextStyleChanges: map[string]docs.SuggestedTextStyle{
										"suggestion-3": {},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Document with no suggestions
	fixtures.MockService.Documents["no-suggestions-doc"] = &docs.Document{
		DocumentId: "no-suggestions-doc",
		Title:      "No Suggestions Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   13,
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   13,
								TextRun:    &docs.TextRun{Content: "Hello World!\n"},
							},
						},
					},
				},
			},
		},
	}

	// Document with suggestions in a table
	fixtures.MockService.Documents["table-suggestions-doc"] = &docs.Document{
		DocumentId: "table-suggestions-doc",
		Title:      "Table Suggestions Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   40,
					Table: &docs.Table{
						TableRows: []*docs.TableRow{
							{
								TableCells: []*docs.TableCell{
									{
										Content: []*docs.StructuralElement{
											{
												StartIndex: 2,
												EndIndex:   10,
												Paragraph: &docs.Paragraph{
													Elements: []*docs.ParagraphElement{
														{
															StartIndex: 2,
															EndIndex:   10,
															TextRun: &docs.TextRun{
																Content:               "cell fix",
																SuggestedInsertionIds: []string{"table-suggestion-1"},
															},
														},
													},
												},
											},
										},
									},
								},
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
			name: "document with suggestions",
			args: map[string]any{"document_id": "suggested-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["document_id"] != "suggested-doc" {
					t.Errorf("expected document_id 'suggested-doc', got %v", result["document_id"])
				}
				count, _ := result["count"].(float64)
				if count != 3 {
					t.Errorf("expected 3 suggestions, got %v", count)
				}

				edits, ok := result["suggested_edits"].([]any)
				if !ok {
					t.Fatalf("expected suggested_edits array, got %T", result["suggested_edits"])
				}

				// Check types
				types := make(map[string]int)
				for _, e := range edits {
					edit := e.(map[string]any)
					types[edit["type"].(string)]++
				}
				if types["insertion"] != 1 {
					t.Errorf("expected 1 insertion, got %d", types["insertion"])
				}
				if types["deletion"] != 1 {
					t.Errorf("expected 1 deletion, got %d", types["deletion"])
				}
				if types["formatting"] != 1 {
					t.Errorf("expected 1 formatting, got %d", types["formatting"])
				}
			},
		},
		{
			name: "document without suggestions",
			args: map[string]any{"document_id": "no-suggestions-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				count, _ := result["count"].(float64)
				if count != 0 {
					t.Errorf("expected 0 suggestions, got %v", count)
				}
			},
		},
		{
			name: "suggestions in table cells",
			args: map[string]any{"document_id": "table-suggestions-doc"},
			checkResult: func(t *testing.T, result map[string]any) {
				count, _ := result["count"].(float64)
				if count != 1 {
					t.Errorf("expected 1 suggestion in table, got %v", count)
				}
				edits := result["suggested_edits"].([]any)
				edit := edits[0].(map[string]any)
				if edit["suggestion_id"] != "table-suggestion-1" {
					t.Errorf("expected suggestion_id 'table-suggestion-1', got %v", edit["suggestion_id"])
				}
				if edit["type"] != "insertion" {
					t.Errorf("expected type 'insertion', got %v", edit["type"])
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
			result, err := testableDocsGetSuggestedEdits(context.Background(), request, fixtures.Deps)
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
