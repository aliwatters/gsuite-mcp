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

// extractRequiredDocID extracts, validates, and normalizes the document_id parameter.
// Returns the cleaned ID or an error result if missing.
func extractRequiredDocID(request mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	docID := common.ParseStringArg(request.GetArguments(), "document_id", "")
	if docID == "" {
		return "", mcp.NewToolResultError("document_id parameter is required")
	}
	return common.ExtractGoogleResourceID(docID), nil
}

// extractIndexRange extracts and validates start_index and end_index from a request.
// Both must be present, start_index >= 1, and end_index > start_index.
func extractIndexRange(request mcp.CallToolRequest) (startIndex, endIndex int64, errResult *mcp.CallToolResult) {
	startIndexFloat, ok := request.GetArguments()["start_index"].(float64)
	if !ok {
		return 0, 0, mcp.NewToolResultError("start_index parameter is required (1-based position in document)")
	}
	startIndex = int64(startIndexFloat)
	if startIndex < 1 {
		return 0, 0, mcp.NewToolResultError("start_index must be at least 1")
	}

	endIndexFloat, ok := request.GetArguments()["end_index"].(float64)
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

	title, errResult := common.RequireStringArg(request.GetArguments(), "title")
	if errResult != nil {
		return errResult, nil
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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	text, errResult := common.RequireStringArg(request.GetArguments(), "text")
	if errResult != nil {
		return errResult, nil
	}

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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	text, errResult := common.RequireStringArg(request.GetArguments(), "text")
	if errResult != nil {
		return errResult, nil
	}

	indexFloat, ok := request.GetArguments()["index"].(float64)
	if !ok {
		return mcp.NewToolResultError("index parameter is required (1-based position in document)"), nil
	}
	index := int64(indexFloat)
	if index < 1 {
		return mcp.NewToolResultError("index must be at least 1"), nil
	}

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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	findText, errResult := common.RequireStringArg(request.GetArguments(), "find_text")
	if errResult != nil {
		return errResult, nil
	}

	replaceText := common.ParseStringArg(request.GetArguments(), "replace_text", "")
	// replace_text can be empty (to delete matched text)

	matchCase := common.ParseBoolArg(request.GetArguments(), "match_case", false)

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

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	startIndex, endIndex, idxErrResult := extractIndexRange(request)
	if idxErrResult != nil {
		return idxErrResult, nil
	}

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
// Uses raw JSON passthrough to support all Docs API request types, including
// those not yet in the Go client library's typed Request struct (e.g., updateNamedStyle).
func TestableDocsBatchUpdate(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	requestsJSON := common.ParseStringArg(request.GetArguments(), "requests", "")
	if requestsJSON == "" {
		return mcp.NewToolResultError("requests parameter is required (JSON array of batch update requests)"), nil
	}

	// Validate that the JSON is a non-empty array
	var rawRequests []json.RawMessage
	if err := json.Unmarshal([]byte(requestsJSON), &rawRequests); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse requests JSON: %v", err)), nil
	}

	if len(rawRequests) == 0 {
		return mcp.NewToolResultError("requests array cannot be empty"), nil
	}

	// Use raw JSON passthrough to support all request types, including those
	// the Go client library doesn't have typed structs for yet.
	resp, err := srv.BatchUpdateRaw(ctx, docID, json.RawMessage(requestsJSON))
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
		"requests_count": len(rawRequests),
		"replies_count":  repliesCount,
		"message":        fmt.Sprintf("Successfully executed %d batch update request(s)", len(rawRequests)),
		"url":            fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}
