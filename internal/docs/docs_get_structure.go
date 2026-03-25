package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// paragraphTextPreview extracts a text preview from a paragraph, truncated to maxLen runes.
func paragraphTextPreview(para *docs.Paragraph, maxLen int) string {
	var sb strings.Builder
	for _, elem := range para.Elements {
		if elem.TextRun != nil {
			sb.WriteString(elem.TextRun.Content)
		}
	}
	text := strings.TrimSpace(sb.String())
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return text
}

// TestableDocsGetStructure returns the document's structural elements with real character indices.
func TestableDocsGetStructure(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	var elements []map[string]any

	if doc.Body != nil && doc.Body.Content != nil {
		for _, elem := range doc.Body.Content {
			entry := map[string]any{
				"startIndex": elem.StartIndex,
				"endIndex":   elem.EndIndex,
			}

			switch {
			case elem.Paragraph != nil:
				entry["type"] = "paragraph"
				style := ""
				if elem.Paragraph.ParagraphStyle != nil {
					style = elem.Paragraph.ParagraphStyle.NamedStyleType
				}
				entry["style"] = style
				entry["text"] = paragraphTextPreview(elem.Paragraph, 80)

			case elem.Table != nil:
				entry["type"] = "table"
				entry["rows"] = int64(len(elem.Table.TableRows))
				entry["cols"] = elem.Table.Columns

			case elem.TableOfContents != nil:
				entry["type"] = "tableOfContents"

			case elem.SectionBreak != nil:
				entry["type"] = "sectionBreak"

			default:
				entry["type"] = "unknown"
			}

			elements = append(elements, entry)
		}
	}

	result := map[string]any{
		"document_id": docID,
		"title":       doc.Title,
		"elements":    elements,
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
