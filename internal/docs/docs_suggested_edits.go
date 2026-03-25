package docs

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// suggestedEdit represents a single suggested edit found in the document.
type suggestedEdit struct {
	SuggestionID string `json:"suggestion_id"`
	Type         string `json:"type"` // "insertion", "deletion", "formatting"
	Content      string `json:"content,omitempty"`
	StartIndex   int64  `json:"start_index"`
	EndIndex     int64  `json:"end_index"`
}

// collectSuggestedEdits walks the document and collects all suggested edits.
func collectSuggestedEdits(doc *docs.Document) []map[string]any {
	var edits []map[string]any
	if doc == nil || doc.Body == nil {
		return edits
	}

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			edits = collectParagraphSuggestions(elem.Paragraph, edits)
		}
		if elem.Table != nil {
			for _, row := range elem.Table.TableRows {
				for _, cell := range row.TableCells {
					for _, cellContent := range cell.Content {
						if cellContent.Paragraph != nil {
							edits = collectParagraphSuggestions(cellContent.Paragraph, edits)
						}
					}
				}
			}
		}
	}

	return edits
}

// collectParagraphSuggestions extracts suggested edits from paragraph elements.
func collectParagraphSuggestions(para *docs.Paragraph, edits []map[string]any) []map[string]any {
	if para.Elements == nil {
		return edits
	}

	for _, elem := range para.Elements {
		if elem.TextRun == nil {
			continue
		}

		// Suggested insertions
		for _, id := range elem.TextRun.SuggestedInsertionIds {
			edits = append(edits, map[string]any{
				"suggestion_id": id,
				"type":          "insertion",
				"content":       elem.TextRun.Content,
				"start_index":   elem.StartIndex,
				"end_index":     elem.EndIndex,
			})
		}

		// Suggested deletions
		for _, id := range elem.TextRun.SuggestedDeletionIds {
			edits = append(edits, map[string]any{
				"suggestion_id": id,
				"type":          "deletion",
				"content":       elem.TextRun.Content,
				"start_index":   elem.StartIndex,
				"end_index":     elem.EndIndex,
			})
		}

		// Suggested text style changes (formatting)
		for id := range elem.TextRun.SuggestedTextStyleChanges {
			edits = append(edits, map[string]any{
				"suggestion_id": id,
				"type":          "formatting",
				"content":       elem.TextRun.Content,
				"start_index":   elem.StartIndex,
				"end_index":     elem.EndIndex,
			})
		}
	}

	return edits
}

// TestableDocsGetSuggestedEdits retrieves all suggested edits in a document.
func TestableDocsGetSuggestedEdits(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	doc, err := srv.GetDocumentWithSuggestions(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	edits := collectSuggestedEdits(doc)

	result := map[string]any{
		"document_id":     docID,
		"suggested_edits": edits,
		"count":           len(edits),
		"url":             fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
