package docs

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestableDocsInsertTable is the testable version of HandleDocsInsertTable.
func TestableDocsInsertTable(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	rowsFloat, ok := request.Params.Arguments["rows"].(float64)
	if !ok || rowsFloat < 1 {
		return mcp.NewToolResultError("rows parameter is required and must be at least 1"), nil
	}
	rows := int64(rowsFloat)

	columnsFloat, ok := request.Params.Arguments["columns"].(float64)
	if !ok || columnsFloat < 1 {
		return mcp.NewToolResultError("columns parameter is required and must be at least 1"), nil
	}
	columns := int64(columnsFloat)

	// Default index to 1 (beginning of document)
	index := int64(1)
	if indexFloat, ok := request.Params.Arguments["index"].(float64); ok {
		index = int64(indexFloat)
		if index < 1 {
			return mcp.NewToolResultError("index must be at least 1"), nil
		}
	}

	docID = common.ExtractGoogleResourceID(docID)

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"rows":        rows,
		"columns":     columns,
		"message":     fmt.Sprintf("Inserted %dx%d table at index %d", rows, columns, index),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsInsertLink is the testable version of HandleDocsInsertLink.
func TestableDocsInsertLink(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
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

	linkURL, _ := request.Params.Arguments["url"].(string)
	if linkURL == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
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

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"text":        text,
		"link_url":    linkURL,
		"message":     fmt.Sprintf("Inserted hyperlink '%s' at index %d", text, index),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsInsertPageBreak is the testable version of HandleDocsInsertPageBreak.
func TestableDocsInsertPageBreak(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
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

	_, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Inserted page break at index %d", index),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// TestableDocsInsertImage is the testable version of HandleDocsInsertImage.
func TestableDocsInsertImage(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	imageURI, _ := request.Params.Arguments["uri"].(string)
	if imageURI == "" {
		return mcp.NewToolResultError("uri parameter is required (URL of the image)"), nil
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

	resp, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	// Extract the inline object ID from the response (may be empty in test mocks)
	var inlineObjectId string
	if resp != nil && resp.Replies != nil && len(resp.Replies) > 0 && resp.Replies[0].InsertInlineImage != nil {
		inlineObjectId = resp.Replies[0].InsertInlineImage.ObjectId
	}

	result := map[string]any{
		"success":          true,
		"document_id":      docID,
		"inline_object_id": inlineObjectId,
		"message":          fmt.Sprintf("Inserted image at index %d", index),
		"url":              fmt.Sprintf(docsEditURLFormat, docID),
	}

	return common.MarshalToolResult(result)
}

// handleDocsCreateHeaderOrFooter is shared logic for creating a document header or footer.
// The sectionType must be "header" or "footer".
func handleDocsCreateHeaderOrFooter(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps, sectionType string) (*mcp.CallToolResult, error) {
	if sectionType != "header" && sectionType != "footer" {
		return mcp.NewToolResultError(fmt.Sprintf("invalid section type %q; must be 'header' or 'footer'", sectionType)), nil
	}

	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, _ := request.Params.Arguments["document_id"].(string)
	if docID == "" {
		return mcp.NewToolResultError("document_id parameter is required"), nil
	}

	docID = common.ExtractGoogleResourceID(docID)

	resp, err := srv.BatchUpdate(ctx, docID, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	var sectionID string
	if resp != nil && resp.Replies != nil && len(resp.Replies) > 0 {
		switch sectionType {
		case "header":
			if resp.Replies[0].CreateHeader != nil {
				sectionID = resp.Replies[0].CreateHeader.HeaderId
			}
		case "footer":
			if resp.Replies[0].CreateFooter != nil {
				sectionID = resp.Replies[0].CreateFooter.FooterId
			}
		}
	}

	// Optionally insert content
	content, hasContent := request.Params.Arguments["content"].(string)
	if hasContent && content != "" && sectionID != "" {
		_, err = srv.BatchUpdate(ctx, docID, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Docs API error inserting %s content: %v", sectionType, err)), nil
		}
	}

	result := map[string]any{
		"success":     true,
		"document_id": docID,
		"message":     fmt.Sprintf("Created document %s", sectionType),
		"url":         fmt.Sprintf(docsEditURLFormat, docID),
	}
	switch sectionType {
	case "header":
		result["header_id"] = sectionID
	case "footer":
		result["footer_id"] = sectionID
	}

	return common.MarshalToolResult(result)
}

// TestableDocsCreateHeader is the testable version of HandleDocsCreateHeader.
func TestableDocsCreateHeader(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	return handleDocsCreateHeaderOrFooter(ctx, request, deps, "header")
}

// TestableDocsCreateFooter is the testable version of HandleDocsCreateFooter.
func TestableDocsCreateFooter(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	return handleDocsCreateHeaderOrFooter(ctx, request, deps, "footer")
}
