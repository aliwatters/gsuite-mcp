package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

func TestHandleDocsFormatByFind(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("fbf-doc", "Format By Find Test", "Hello World! Hello World!")

	// Add a document with table content for table search testing
	fixtures.MockService.Documents["table-doc"] = &docs.Document{
		DocumentId: "table-doc",
		Title:      "Table Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   10,
					Paragraph: &docs.Paragraph{
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
					EndIndex:   40,
					Table: &docs.Table{
						TableRows: []*docs.TableRow{
							{
								TableCells: []*docs.TableCell{
									{
										Content: []*docs.StructuralElement{
											{
												StartIndex: 12,
												EndIndex:   24,
												Paragraph: &docs.Paragraph{
													Elements: []*docs.ParagraphElement{
														{
															StartIndex: 12,
															EndIndex:   24,
															TextRun:    &docs.TextRun{Content: "Hello Cell\n"},
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

	// Add a document with emoji/UTF-16 surrogate pairs to test index calculation
	// "Hi 👋 World" — the emoji is a surrogate pair (2 UTF-16 code units)
	// Byte layout: H(1) i(1) (1) 👋(4) (1) W(1) o(1) r(1) l(1) d(1) \n(1) = 14 bytes
	// UTF-16 layout: H(1) i(1) (1) 👋(2) (1) W(1) o(1) r(1) l(1) d(1) \n(1) = 12 code units
	fixtures.MockService.Documents["emoji-doc"] = &docs.Document{
		DocumentId: "emoji-doc",
		Title:      "Emoji Doc",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   12,
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   12,
								TextRun:    &docs.TextRun{Content: "Hi 👋 World\n"},
							},
						},
					},
				},
			},
		},
	}

	// Add a document containing curly braces for brace-search testing
	fixtures.AddTestDocument("braces-doc", "Braces Doc", "Hello {value} world")

	// Add a document with multiple opening braces
	fixtures.AddTestDocument("brace-only-doc", "Brace Only Doc", "a{b and c{d")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
		checkCalls  func(t *testing.T, mock *MockDocsService)
	}{
		{
			name: "format all occurrences with bold",
			args: map[string]any{
				"document_id": "fbf-doc",
				"find_text":   "Hello",
				"bold":        true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["success"] != true {
					t.Error("expected success to be true")
				}
				if result["document_id"] != "fbf-doc" {
					t.Errorf("expected document_id 'fbf-doc', got %v", result["document_id"])
				}
				if result["find_text"] != "Hello" {
					t.Errorf("expected find_text 'Hello', got %v", result["find_text"])
				}
				// "Hello World! Hello World!\n" has 2 occurrences of "Hello"
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 2 {
					t.Errorf("expected 2 matches, got %v", matchesFound)
				}
				if result["match_case"] != true {
					t.Errorf("expected match_case true (default), got %v", result["match_case"])
				}
				if result["match_all"] != true {
					t.Errorf("expected match_all true (default), got %v", result["match_all"])
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				if len(mock.Calls.BatchUpdate) == 0 {
					t.Fatal("expected BatchUpdate to be called")
				}
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				if len(lastCall.Requests) != 2 {
					t.Fatalf("expected 2 requests in BatchUpdate, got %d", len(lastCall.Requests))
				}
				// First match: "Hello" at index 0-5
				req0 := lastCall.Requests[0].UpdateTextStyle
				if req0.Range.StartIndex != 0 || req0.Range.EndIndex != 5 {
					t.Errorf("expected first match range 0-5, got %d-%d", req0.Range.StartIndex, req0.Range.EndIndex)
				}
				if !req0.TextStyle.Bold {
					t.Error("expected bold to be true in text style")
				}
				// Second match: "Hello" at index 13-18
				req1 := lastCall.Requests[1].UpdateTextStyle
				if req1.Range.StartIndex != 13 || req1.Range.EndIndex != 18 {
					t.Errorf("expected second match range 13-18, got %d-%d", req1.Range.StartIndex, req1.Range.EndIndex)
				}
			},
		},
		{
			name: "match_all false formats only first",
			args: map[string]any{
				"document_id": "fbf-doc",
				"find_text":   "Hello",
				"match_all":   false,
				"italic":      true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 1 {
					t.Errorf("expected 1 match with match_all=false, got %v", matchesFound)
				}
				if result["match_all"] != false {
					t.Errorf("expected match_all false, got %v", result["match_all"])
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				if len(lastCall.Requests) != 1 {
					t.Fatalf("expected 1 request with match_all=false, got %d", len(lastCall.Requests))
				}
			},
		},
		{
			name: "case insensitive match",
			args: map[string]any{
				"document_id": "fbf-doc",
				"find_text":   "hello",
				"match_case":  false,
				"underline":   true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 2 {
					t.Errorf("expected 2 case-insensitive matches, got %v", matchesFound)
				}
				if result["match_case"] != false {
					t.Errorf("expected match_case false, got %v", result["match_case"])
				}
			},
		},
		{
			name: "find text in table cells",
			args: map[string]any{
				"document_id": "table-doc",
				"find_text":   "Hello",
				"bold":        true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 1 {
					t.Errorf("expected 1 match in table, got %v", matchesFound)
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				req := lastCall.Requests[0].UpdateTextStyle
				// "Hello" starts at offset 0 within the TextRun, element starts at index 12
				if req.Range.StartIndex != 12 || req.Range.EndIndex != 17 {
					t.Errorf("expected table match range 12-17, got %d-%d", req.Range.StartIndex, req.Range.EndIndex)
				}
			},
		},
		{
			name: "UTF-16 surrogate pairs calculate correct indexes",
			args: map[string]any{
				"document_id": "emoji-doc",
				"find_text":   "World",
				"bold":        true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 1 {
					t.Errorf("expected 1 match, got %v", matchesFound)
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				req := lastCall.Requests[0].UpdateTextStyle
				// "Hi 👋 World\n" — "World" starts after "Hi 👋 "
				// UTF-16: H(1) i(1) space(1) 👋(2) space(1) = 6 code units before "World"
				// "World" = 5 UTF-16 code units → range 6-11
				if req.Range.StartIndex != 6 || req.Range.EndIndex != 11 {
					t.Errorf("expected UTF-16 range 6-11, got %d-%d", req.Range.StartIndex, req.Range.EndIndex)
				}
			},
		},
		{
			name: "find text containing curly braces",
			args: map[string]any{
				"document_id": "braces-doc",
				"find_text":   "{value}",
				"bold":        true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 1 {
					t.Errorf("expected 1 match for '{value}', got %v", matchesFound)
				}
			},
			checkCalls: func(t *testing.T, mock *MockDocsService) {
				lastCall := mock.Calls.BatchUpdate[len(mock.Calls.BatchUpdate)-1]
				if len(lastCall.Requests) != 1 {
					t.Fatalf("expected 1 request, got %d", len(lastCall.Requests))
				}
				req := lastCall.Requests[0].UpdateTextStyle
				// "Hello {value} world\n" — "{value}" starts at index 6, length 7
				if req.Range.StartIndex != 6 || req.Range.EndIndex != 13 {
					t.Errorf("expected range 6-13 for '{value}', got %d-%d", req.Range.StartIndex, req.Range.EndIndex)
				}
			},
		},
		{
			name: "find text with opening brace only",
			args: map[string]any{
				"document_id": "brace-only-doc",
				"find_text":   "{",
				"italic":      true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				matchesFound, _ := result["matches_found"].(float64)
				if matchesFound != 2 {
					t.Errorf("expected 2 matches for '{', got %v", matchesFound)
				}
			},
		},
		{
			name:        "missing document_id",
			args:        map[string]any{"find_text": "Hello", "bold": true},
			wantErr:     true,
			errContains: "document_id parameter is required",
		},
		{
			name: "missing find_text",
			args: map[string]any{
				"document_id": "fbf-doc",
				"bold":        true,
			},
			wantErr:     true,
			errContains: "find_text",
		},
		{
			name: "no formatting specified",
			args: map[string]any{
				"document_id": "fbf-doc",
				"find_text":   "Hello",
			},
			wantErr:     true,
			errContains: "at least one formatting option must be specified",
		},
		{
			name: "text not found",
			args: map[string]any{
				"document_id": "fbf-doc",
				"find_text":   "NonexistentText",
				"bold":        true,
			},
			wantErr:     true,
			errContains: "not found in document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableDocsFormatByFind(context.Background(), request, fixtures.Deps)

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
