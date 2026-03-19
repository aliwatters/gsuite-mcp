package docs

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestableDocsExportToPDF exports a Google Doc, Sheet, or Slides presentation to PDF.
// Uses the Drive API export endpoint with application/pdf MIME type.
func TestableDocsExportToPDF(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	data, file, err := srv.ExportPDF(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Export error: %v", err)), nil
	}

	// Verify it's a supported Google Workspace type
	supportedTypes := map[string]bool{
		"application/vnd.google-apps.document":     true,
		"application/vnd.google-apps.spreadsheet":  true,
		"application/vnd.google-apps.presentation": true,
	}
	if file != nil && !supportedTypes[file.MimeType] {
		return mcp.NewToolResultError(fmt.Sprintf(
			"File type '%s' cannot be exported to PDF. Supported types: Google Docs, Sheets, Slides", file.MimeType)), nil
	}

	if int64(len(data)) > maxExportSize {
		return mcp.NewToolResultError(fmt.Sprintf("PDF export exceeds maximum size of %d bytes", maxExportSize)), nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	fileName := docID
	if file != nil && file.Name != "" {
		fileName = file.Name
	}

	result := map[string]any{
		"document_id":  docID,
		"name":         fileName,
		"pdf_size":     len(data),
		"pdf_base64":   encoded,
		"encoding":     "base64",
		"content_type": "application/pdf",
		"message":      fmt.Sprintf("Exported '%s' to PDF (%d bytes)", fileName, len(data)),
	}

	return common.MarshalToolResult(result)
}
