package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// textMatch represents a match found in the document with real Google Docs indexes.
type textMatch struct {
	StartIndex int64
	EndIndex   int64
}

// findTextOccurrences walks the document structure and returns ranges where findText appears.
// Each range has the real Google Docs startIndex and endIndex.
func findTextOccurrences(doc *docs.Document, findText string, matchCase bool, matchAll bool) []textMatch {
	var matches []textMatch
	if doc == nil || doc.Body == nil {
		return matches
	}

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			matches = findInParagraph(elem.Paragraph, findText, matchCase, matchAll, matches)
			if !matchAll && len(matches) > 0 {
				return matches
			}
		}
		if elem.Table != nil {
			for _, row := range elem.Table.TableRows {
				for _, cell := range row.TableCells {
					for _, cellContent := range cell.Content {
						if cellContent.Paragraph != nil {
							matches = findInParagraph(cellContent.Paragraph, findText, matchCase, matchAll, matches)
							if !matchAll && len(matches) > 0 {
								return matches
							}
						}
					}
				}
			}
		}
	}

	return matches
}

// findInParagraph searches for text within a paragraph's elements and appends matches.
func findInParagraph(para *docs.Paragraph, findText string, matchCase bool, matchAll bool, matches []textMatch) []textMatch {
	if para.Elements == nil {
		return matches
	}

	for _, paraElem := range para.Elements {
		if paraElem.TextRun == nil {
			continue
		}

		content := paraElem.TextRun.Content
		searchContent := content
		searchText := findText
		if !matchCase {
			searchContent = strings.ToLower(content)
			searchText = strings.ToLower(findText)
		}

		offset := 0
		for {
			idx := strings.Index(searchContent[offset:], searchText)
			if idx < 0 {
				break
			}

			realStart := paraElem.StartIndex + int64(offset+idx)
			realEnd := realStart + int64(len(findText))

			matches = append(matches, textMatch{
				StartIndex: realStart,
				EndIndex:   realEnd,
			})

			if !matchAll {
				return matches
			}

			offset += idx + len(searchText)
			if offset >= len(searchContent) {
				break
			}
		}
	}

	return matches
}

// TestableDocsFormatByFind finds text in a document and applies formatting to all matches.
// This avoids the need to manually calculate Google Docs internal character indexes.
func TestableDocsFormatByFind(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	findText, errResult := common.RequireStringArg(request.Params.Arguments, "find_text")
	if errResult != nil {
		return errResult, nil
	}

	matchCase := common.ParseBoolArg(request.Params.Arguments, "match_case", true)
	matchAll := common.ParseBoolArg(request.Params.Arguments, "match_all", true)

	// Validate and collect formatting fields
	fields, validationErr := collectFields(request.Params.Arguments, textFormatFields)
	if validationErr != nil {
		return validationErr, nil
	}
	if len(fields) == 0 {
		return mcp.NewToolResultError("at least one formatting option must be specified"), nil
	}

	// Get the document to find real indexes
	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error getting document: %v", err)), nil
	}

	// Find all occurrences with real indexes
	occurrences := findTextOccurrences(doc, findText, matchCase, matchAll)
	if len(occurrences) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("text %q not found in document", findText)), nil
	}

	// Build formatting requests for each match
	textStyle := buildTextStyle(request.Params.Arguments)
	fieldMask := strings.Join(fields, ",")

	var requests []*docs.Request
	for _, m := range occurrences {
		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: m.StartIndex,
					EndIndex:   m.EndIndex,
				},
				TextStyle: textStyle,
				Fields:    fieldMask,
			},
		})
	}

	_, err = srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"find_text":      findText,
		"matches_found":  len(occurrences),
		"fields_updated": fields,
		"match_case":     matchCase,
		"match_all":      matchAll,
		"message":        fmt.Sprintf("Applied formatting to %d occurrence(s) of %q", len(occurrences), findText),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
