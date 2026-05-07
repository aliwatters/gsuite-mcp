package drive

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// validMimeType matches standard MIME type patterns (e.g. "application/pdf", "image/").
var validMimeType = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9!#$&\-^_.+]*\/[a-zA-Z0-9!#$&\-^_.+]*$`)

// googleWorkspaceExportMIME maps Google Workspace MIME types to their export formats.
var googleWorkspaceExportMIME = map[string]string{
	"application/vnd.google-apps.document":     "text/plain",
	"application/vnd.google-apps.spreadsheet":  "text/csv",
	"application/vnd.google-apps.presentation": "text/plain",
}

// extractRequiredFileID extracts, validates, and normalizes the file_id parameter.
// Returns the cleaned ID or an error result if missing.
func extractRequiredFileID(request mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	fileID := common.ParseStringArg(request.GetArguments(), "file_id", "")
	if fileID == "" {
		return "", mcp.NewToolResultError("file_id parameter is required")
	}
	return common.ExtractGoogleResourceID(fileID), nil
}

// readLimited reads up to DriveMaxFileSize+1 bytes to detect oversized responses
// (Google Workspace files may report Size=0 in metadata).
func readLimited(r io.Reader) ([]byte, error) {
	limited := io.LimitReader(r, common.DriveMaxFileSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("reading file content: %w", err)
	}
	if int64(len(data)) > common.DriveMaxFileSize {
		return nil, fmt.Errorf("response body exceeds maximum size of %d bytes", common.DriveMaxFileSize)
	}
	return data, nil
}

// downloadFileContent handles the download or export of a file, returning raw bytes
// and the export MIME type (empty for non-Workspace files).
func downloadFileContent(ctx context.Context, srv DriveService, fileID string, file *drive.File) ([]byte, string, *mcp.CallToolResult) {
	if strings.HasPrefix(file.MimeType, "application/vnd.google-apps.") {
		exportMimeType, ok := googleWorkspaceExportMIME[file.MimeType]
		if !ok {
			return nil, "", mcp.NewToolResultError(fmt.Sprintf("Cannot download Google Workspace file of type: %s", file.MimeType))
		}

		body, err := srv.ExportFile(ctx, fileID, exportMimeType)
		if err != nil {
			return nil, "", mcp.NewToolResultError(fmt.Sprintf("Drive API error exporting file: %v", err))
		}
		defer body.Close()

		content, err := readLimited(body)
		if err != nil {
			return nil, "", mcp.NewToolResultError(fmt.Sprintf("Error reading file content: %v", err))
		}
		return content, exportMimeType, nil
	}

	body, err := srv.DownloadFile(ctx, fileID)
	if err != nil {
		return nil, "", mcp.NewToolResultError(fmt.Sprintf("Drive API error downloading file: %v", err))
	}
	defer body.Close()

	content, err := readLimited(body)
	if err != nil {
		return nil, "", mcp.NewToolResultError(fmt.Sprintf("Error reading file content: %v", err))
	}
	return content, "", nil
}
