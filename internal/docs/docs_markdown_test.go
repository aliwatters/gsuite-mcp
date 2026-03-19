package docs

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

func TestHandleDocsGetAsMarkdown(t *testing.T) {
	fixtures := NewDocsTestFixtures()
	fixtures.AddTestDocument("md-doc", "Markdown Test", "Hello World!")

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get markdown successfully",
			args: map[string]any{
				"document_id": "md-doc",
			},
			wantErr: false,
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Markdown Test" {
					t.Errorf("expected title 'Markdown Test', got %v", result["title"])
				}
				if result["document_id"] != "md-doc" {
					t.Errorf("expected document_id 'md-doc', got %v", result["document_id"])
				}
				md, ok := result["markdown"].(string)
				if !ok || !strings.Contains(md, "Hello World!") {
					t.Errorf("expected markdown containing 'Hello World!', got %v", result["markdown"])
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
			result, err := testableDocsGetAsMarkdown(context.Background(), request, fixtures.Deps)

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

func TestExtractDocumentMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		doc      *docs.Document
		contains []string
	}{
		{
			name: "plain text",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{Content: "Hello World\n"}},
								},
							},
						},
					},
				},
			},
			contains: []string{"Hello World"},
		},
		{
			name: "heading",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "HEADING_1"},
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{Content: "Title\n"}},
								},
							},
						},
					},
				},
			},
			contains: []string{"# Title"},
		},
		{
			name: "bold text",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{
										Content:   "bold text\n",
										TextStyle: &docs.TextStyle{Bold: true},
									}},
								},
							},
						},
					},
				},
			},
			contains: []string{"**bold text**"},
		},
		{
			name: "italic text",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{
										Content:   "italic text\n",
										TextStyle: &docs.TextStyle{Italic: true},
									}},
								},
							},
						},
					},
				},
			},
			contains: []string{"*italic text*"},
		},
		{
			name: "bold and italic text",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{
										Content:   "bold italic\n",
										TextStyle: &docs.TextStyle{Bold: true, Italic: true},
									}},
								},
							},
						},
					},
				},
			},
			contains: []string{"***bold italic***"},
		},
		{
			name: "link text",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Paragraph: &docs.Paragraph{
								Elements: []*docs.ParagraphElement{
									{TextRun: &docs.TextRun{
										Content: "click here\n",
										TextStyle: &docs.TextStyle{
											Link: &docs.Link{Url: "https://example.com"},
										},
									}},
								},
							},
						},
					},
				},
			},
			contains: []string{"[click here](https://example.com)"},
		},
		{
			name: "table",
			doc: &docs.Document{
				Body: &docs.Body{
					Content: []*docs.StructuralElement{
						{
							Table: &docs.Table{
								Rows:    2,
								Columns: 2,
								TableRows: []*docs.TableRow{
									{
										TableCells: []*docs.TableCell{
											{Content: []*docs.StructuralElement{{Paragraph: &docs.Paragraph{Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "Name\n"}}}}}}},
											{Content: []*docs.StructuralElement{{Paragraph: &docs.Paragraph{Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "Age\n"}}}}}}},
										},
									},
									{
										TableCells: []*docs.TableCell{
											{Content: []*docs.StructuralElement{{Paragraph: &docs.Paragraph{Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "Alice\n"}}}}}}},
											{Content: []*docs.StructuralElement{{Paragraph: &docs.Paragraph{Elements: []*docs.ParagraphElement{{TextRun: &docs.TextRun{Content: "30\n"}}}}}}},
										},
									},
								},
							},
						},
					},
				},
			},
			contains: []string{"| Name | Age |", "|---|---|", "| Alice | 30 |"},
		},
		{
			name: "nil document",
			doc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDocumentMarkdown(tt.doc)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected markdown to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

