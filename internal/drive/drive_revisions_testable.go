package drive

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// formatRevision formats a revision for output.
func formatRevision(r *drive.Revision) map[string]any {
	result := map[string]any{
		"id": r.Id,
	}
	if r.MimeType != "" {
		result["mime_type"] = r.MimeType
	}
	if r.ModifiedTime != "" {
		result["modified_time"] = r.ModifiedTime
	}
	if r.LastModifyingUser != nil {
		result["last_modifying_user"] = map[string]any{
			"display_name": r.LastModifyingUser.DisplayName,
			"email":        r.LastModifyingUser.EmailAddress,
		}
	}
	if r.Size > 0 {
		result["size"] = r.Size
	}
	if r.KeepForever {
		result["keep_forever"] = true
	}
	if r.OriginalFilename != "" {
		result["original_filename"] = r.OriginalFilename
	}
	if len(r.ExportLinks) > 0 {
		result["export_links"] = r.ExportLinks
	}
	return result
}

// TestableDriveListRevisions lists revisions of a file.
func TestableDriveListRevisions(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	maxResults := common.ParseMaxResults(request.GetArguments(), common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.GetArguments(), "page_token", "")

	resp, err := srv.ListRevisions(ctx, fileID, DriveRevisionListFields, maxResults, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	revisions := make([]map[string]any, 0, len(resp.Revisions))
	for _, r := range resp.Revisions {
		revisions = append(revisions, formatRevision(r))
	}

	result := map[string]any{
		"file_id":         fileID,
		"revisions":       revisions,
		"count":           len(revisions),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGetRevision gets a specific revision.
func TestableDriveGetRevision(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	revisionID, errResult := common.RequireStringArg(request.GetArguments(), "revision_id")
	if errResult != nil {
		return errResult, nil
	}

	revision, err := srv.GetRevision(ctx, fileID, revisionID, DriveRevisionGetFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatRevision(revision)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveDownloadRevision downloads the content of a specific revision.
func TestableDriveDownloadRevision(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	revisionID, errResult := common.RequireStringArg(request.GetArguments(), "revision_id")
	if errResult != nil {
		return errResult, nil
	}

	// Get file info first for mime type detection
	file, err := srv.GetFile(ctx, fileID, DriveFileDownloadFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error getting file info: %v", err)), nil
	}

	body, err := srv.DownloadRevision(ctx, fileID, revisionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error downloading revision: %v", err)), nil
	}
	defer body.Close()

	content, err := readLimited(body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error reading revision content: %v", err)), nil
	}

	isText := isTextMimeType(file.MimeType)

	result := map[string]any{
		"file_id":     fileID,
		"revision_id": revisionID,
		"name":        file.Name,
		"mime_type":   file.MimeType,
		"size":        len(content),
	}

	if isText {
		result["content"] = string(content)
		result["encoding"] = "utf-8"
	} else {
		result["content"] = base64.StdEncoding.EncodeToString(content)
		result["encoding"] = "base64"
	}

	return common.MarshalToolResult(result)
}
