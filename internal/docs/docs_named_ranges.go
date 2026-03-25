package docs

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// TestableDocsListNamedRanges lists all named ranges in a document.
func TestableDocsListNamedRanges(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
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

	var namedRanges []map[string]any
	for name, nrs := range doc.NamedRanges {
		for _, nr := range nrs.NamedRanges {
			var ranges []map[string]any
			for _, r := range nr.Ranges {
				rangeEntry := map[string]any{
					"startIndex": r.StartIndex,
					"endIndex":   r.EndIndex,
				}
				if r.SegmentId != "" {
					rangeEntry["segmentId"] = r.SegmentId
				}
				ranges = append(ranges, rangeEntry)
			}
			namedRanges = append(namedRanges, map[string]any{
				"name":         name,
				"namedRangeId": nr.NamedRangeId,
				"ranges":       ranges,
			})
		}
	}

	result := map[string]any{
		"document_id":  docID,
		"named_ranges": namedRanges,
		"count":        len(namedRanges),
		"url":          fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsCreateNamedRange creates a named range in a document.
func TestableDocsCreateNamedRange(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	name, errResult := common.RequireStringArg(request.Params.Arguments, "name")
	if errResult != nil {
		return errResult, nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	requests := []*docs.Request{
		{
			CreateNamedRange: &docs.CreateNamedRangeRequest{
				Name: name,
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
			},
		},
	}

	resp, err := srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	namedRangeID := ""
	if resp != nil && len(resp.Replies) > 0 && resp.Replies[0].CreateNamedRange != nil {
		namedRangeID = resp.Replies[0].CreateNamedRange.NamedRangeId
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"name":           name,
		"named_range_id": namedRangeID,
		"start_index":    startIndex,
		"end_index":      endIndex,
		"message":        fmt.Sprintf("Created named range %q covering indexes %d-%d", name, startIndex, endIndex),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsDeleteNamedRange deletes a named range from a document.
func TestableDocsDeleteNamedRange(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	namedRangeID, errResult := common.RequireStringArg(request.Params.Arguments, "named_range_id")
	if errResult != nil {
		return errResult, nil
	}

	requests := []*docs.Request{
		{
			DeleteNamedRange: &docs.DeleteNamedRangeRequest{
				NamedRangeId: namedRangeID,
			},
		},
	}

	_, err := srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"named_range_id": namedRangeID,
		"message":        fmt.Sprintf("Deleted named range %s", namedRangeID),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
