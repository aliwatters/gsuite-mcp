package docs

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// TestableDocsFindAndReplace searches and replaces text across a Google Doc.
// Uses the Docs API batchUpdate with ReplaceAllText requests.
func TestableDocsFindAndReplace(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
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

	replaceText := common.ParseStringArg(request.Params.Arguments, "replace_text", "")
	matchCase := common.ParseBoolArg(request.Params.Arguments, "match_case", true)

	replaceReq := &docs.ReplaceAllTextRequest{
		ContainsText: &docs.SubstringMatchCriteria{
			Text:      findText,
			MatchCase: matchCase,
		},
		ReplaceText:     replaceText,
		ForceSendFields: []string{"ReplaceText"},
	}
	requests := []*docs.Request{{ReplaceAllText: replaceReq}}

	resp, err := srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	replacementsCount := int64(0)
	if resp != nil && resp.Replies != nil && len(resp.Replies) > 0 && resp.Replies[0].ReplaceAllText != nil {
		replacementsCount = resp.Replies[0].ReplaceAllText.OccurrencesChanged
	}

	result := map[string]any{
		"success":            true,
		"document_id":        docID,
		"find_text":          findText,
		"replace_text":       replaceText,
		"match_case":         matchCase,
		"replacements_count": replacementsCount,
		"message":            fmt.Sprintf("Replaced %d occurrence(s) of '%s' with '%s'", replacementsCount, findText, replaceText),
		"url":                fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
