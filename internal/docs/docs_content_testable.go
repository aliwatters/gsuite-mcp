package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// docsEditURLFormat is the URL template for Google Docs edit links.
const docsEditURLFormat = "https://docs.google.com/document/d/%s/edit"

// extractIndexRange extracts and validates start_index and end_index from a request.
// Both must be present, start_index >= 1, and end_index > start_index.
func extractIndexRange(request mcp.CallToolRequest) (startIndex, endIndex int64, errResult *mcp.CallToolResult) {
	startIndexFloat, ok := request.Params.Arguments["start_index"].(float64)
	if !ok {
		return 0, 0, mcp.NewToolResultError("start_index parameter is required (1-based position in document)")
	}
	startIndex = int64(startIndexFloat)
	if startIndex < 1 {
		return 0, 0, mcp.NewToolResultError("start_index must be at least 1")
	}

	endIndexFloat, ok := request.Params.Arguments["end_index"].(float64)
	if !ok {
		return 0, 0, mcp.NewToolResultError("end_index parameter is required (1-based position in document)")
	}
	endIndex = int64(endIndexFloat)
	if endIndex <= startIndex {
		return 0, 0, mcp.NewToolResultError("end_index must be greater than start_index")
	}

	return startIndex, endIndex, nil
}

// TestableDocsCreate is the testable version of HandleDocsCreate.
func TestableDocsCreate(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title, _ := request.Params.Arguments["title"].(string)
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	created, err := srv.CreateDocument(ctx, title)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"document_id": created.DocumentId,
		"title":       created.Title,
		"url":         fmt.Sprintf(docsEditURLFormat, created.DocumentId),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsGet is the testable version of HandleDocsGet.
func TestableDocsGet(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	content := extractDocumentText(doc)

	result := map[string]any{
		"document_id": doc.DocumentId,
		"title":       doc.Title,
		"content":     content,
		"url":         fmt.Sprintf(docsEditURLFormat, doc.DocumentId),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsGetMetadata is the testable version of HandleDocsGetMetadata.
func TestableDocsGetMetadata(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"document_id": doc.DocumentId,
		"title":       doc.Title,
		"url":         fmt.Sprintf(docsEditURLFormat, doc.DocumentId),
	}

	if doc.RevisionId != "" {
		result["revision_id"] = doc.RevisionId
	}

	content := extractDocumentText(doc)
	result["character_count"] = len(content)
	result["word_count"] = len(strings.Fields(content))

	return common.MarshalToolResult(result)
}

// TestableDocsAppendText is the testable version of HandleDocsAppendText.
func TestableDocsAppendText(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	text, _ := request.Params.Arguments["text"].(string)
	if text == "" {
		return mcp.NewToolResultError("text parameter is required"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	// First get the document to find the end index
	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error getting document: %v", err)), nil
	}

	// Find the end of the document body
	endIndex := int64(1)
	if doc.Body != nil && doc.Body.Content != nil {
		for _, elem := range doc.Body.Content {
			if elem.EndIndex > endIndex {
				endIndex = elem.EndIndex
			}
		}
	}

	// The actual insert position is endIndex - 1 (before the final newline)
	insertIndex := endIndex - 1
	if insertIndex < 1 {
		insertIndex = 1
	}

	requests := []*docs.Request{{
		InsertText: &docs.InsertTextRequest{
			Text:     text,
			Location: &docs.Location{Index: insertIndex},
		},
	}}
	_, err = srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     "Text appended successfully",
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsInsertText is the testable version of HandleDocsInsertText.
func TestableDocsInsertText(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	text, _ := request.Params.Arguments["text"].(string)
	if text == "" {
		return mcp.NewToolResultError("text parameter is required"), nil
	}

	indexFloat, ok := request.Params.Arguments["index"].(float64)
	if !ok {
		return mcp.NewToolResultError("index parameter is required (1-based position in document)"), nil
	}
	index := int64(indexFloat)
	if index < 1 {
		return mcp.NewToolResultError("index must be at least 1"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	requests := []*docs.Request{{
		InsertText: &docs.InsertTextRequest{
			Text:     text,
			Location: &docs.Location{Index: index},
		},
	}}
	_, err := srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Text inserted at index %d", index),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsReplaceText is the testable version of HandleDocsReplaceText.
func TestableDocsReplaceText(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	findText, _ := request.Params.Arguments["find_text"].(string)
	if findText == "" {
		return mcp.NewToolResultError("find_text parameter is required"), nil
	}

	replaceText, _ := request.Params.Arguments["replace_text"].(string)
	// replace_text can be empty (to delete matched text)

	matchCase, _ := request.Params.Arguments["match_case"].(bool)
	// Default to false if not specified

	docID = common.ExtractGoogleResourceID(docID)

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
		"replacements_count": replacementsCount,
		"match_case":         matchCase,
		"message":            fmt.Sprintf("Replaced occurrences of '%s' with '%s'", findText, replaceText),
		"url":                fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsDeleteText is the testable version of HandleDocsDeleteText.
func TestableDocsDeleteText(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	startIndex, endIndex, errResult := extractIndexRange(request)
	if errResult != nil {
		return errResult, nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	requests := []*docs.Request{{
		DeleteContentRange: &docs.DeleteContentRangeRequest{
			Range: &docs.Range{
				StartIndex: startIndex,
				EndIndex:   endIndex,
			},
		},
	}}
	_, err := srv.BatchUpdate(ctx, docID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Deleted text from index %d to %d", startIndex, endIndex),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsBatchUpdate is the testable version of HandleDocsBatchUpdate.
func TestableDocsBatchUpdate(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	requestsJSON, _ := request.Params.Arguments["requests"].(string)
	if requestsJSON == "" {
		return mcp.NewToolResultError("requests parameter is required (JSON array of batch update requests)"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	// Parse the JSON into Docs API request objects
	var docsRequests []*docs.Request
	if err := json.Unmarshal([]byte(requestsJSON), &docsRequests); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse requests JSON: %v", err)), nil
	}

	if len(docsRequests) == 0 {
		return mcp.NewToolResultError("requests array cannot be empty"), nil
	}

	resp, err := srv.BatchUpdate(ctx, docID, docsRequests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	repliesCount := 0
	if resp != nil && resp.Replies != nil {
		repliesCount = len(resp.Replies)
	}

	result := map[string]any{
		"success":        true,
		"document_id":    docID,
		"requests_count": len(docsRequests),
		"replies_count":  repliesCount,
		"message":        fmt.Sprintf("Successfully executed %d batch update request(s)", len(docsRequests)),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
