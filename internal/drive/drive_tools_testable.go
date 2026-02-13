package drive

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// googleWorkspaceExportMIME maps Google Workspace MIME types to their export formats.
var googleWorkspaceExportMIME = map[string]string{
	"application/vnd.google-apps.document":     "text/plain",
	"application/vnd.google-apps.spreadsheet":  "text/csv",
	"application/vnd.google-apps.presentation": "text/plain",
}

// extractRequiredFileID extracts, validates, and normalizes the file_id parameter.
// Returns the cleaned ID or an error result if missing.
func extractRequiredFileID(request mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	fileID := common.ParseStringArg(request.Params.Arguments, "file_id", "")
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
		return nil, err
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

// TestableDriveSearch searches files with query syntax.
func TestableDriveSearch(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	query := common.ParseStringArg(request.Params.Arguments, "query", "")
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	resp, err := srv.ListFiles(ctx, &ListFilesOptions{
		Query:     query,
		PageSize:  maxResults,
		PageToken: pageToken,
		Fields:    DriveFileListFields,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	files := make([]map[string]any, 0, len(resp.Files))
	for _, f := range resp.Files {
		files = append(files, formatFile(f))
	}

	result := map[string]any{
		"files":           files,
		"count":           len(files),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGet gets file metadata.
func TestableDriveGet(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	file, err := srv.GetFile(ctx, fileID, DriveFileGetFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatFileFull(file))
}

// TestableDriveDownload downloads file content.
func TestableDriveDownload(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	file, err := srv.GetFile(ctx, fileID, DriveFileDownloadFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error getting file info: %v", err)), nil
	}

	if file.Size > common.DriveMaxFileSize {
		return mcp.NewToolResultError(fmt.Sprintf("File too large (%d bytes). Maximum supported size is %d bytes", file.Size, common.DriveMaxFileSize)), nil
	}

	content, exportMimeType, dlErrResult := downloadFileContent(ctx, srv, fileID, file)
	if dlErrResult != nil {
		return dlErrResult, nil
	}

	isText := isTextMimeType(file.MimeType) || exportMimeType != ""

	result := map[string]any{
		"file_id":   file.Id,
		"name":      file.Name,
		"mime_type": file.MimeType,
		"size":      len(content),
	}

	if exportMimeType != "" {
		result["exported_as"] = exportMimeType
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

// TestableDriveUpload uploads a new file.
func TestableDriveUpload(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name := common.ParseStringArg(request.Params.Arguments, "name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	contentStr := common.ParseStringArg(request.Params.Arguments, "content", "")
	if contentStr == "" {
		return mcp.NewToolResultError("content parameter is required"), nil
	}

	// Reject obviously oversized content before decoding
	if int64(len(contentStr)) > common.DriveMaxFileSize*2 {
		return mcp.NewToolResultError(fmt.Sprintf("Content too large. Maximum supported size is %d bytes", common.DriveMaxFileSize)), nil
	}

	// Decode content if base64 encoded
	var data []byte
	encoding := common.ParseStringArg(request.Params.Arguments, "encoding", "")
	if encoding == "base64" {
		var err error
		data, err = base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid base64 content: %v", err)), nil
		}
	} else {
		data = []byte(contentStr)
	}

	if int64(len(data)) > common.DriveMaxFileSize {
		return mcp.NewToolResultError(fmt.Sprintf("Content too large (%d bytes). Maximum supported size is %d bytes", len(data), common.DriveMaxFileSize)), nil
	}

	file := &drive.File{
		Name: name,
	}

	if parentID := common.ParseStringArg(request.Params.Arguments, "parent_id", ""); parentID != "" {
		file.Parents = []string{common.ExtractGoogleResourceID(parentID)}
	}

	if mimeType := common.ParseStringArg(request.Params.Arguments, "mime_type", ""); mimeType != "" {
		file.MimeType = mimeType
	}

	if description := common.ParseStringArg(request.Params.Arguments, "description", ""); description != "" {
		file.Description = description
	}

	created, err := srv.CreateFile(ctx, file, bytes.NewReader(data))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"file_id":      created.Id,
		"name":         created.Name,
		"mime_type":    created.MimeType,
		"size":         created.Size,
		"created_time": created.CreatedTime,
		"url":          created.WebViewLink,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveList lists files in a folder.
func TestableDriveList(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	folderID := common.ParseStringArg(request.Params.Arguments, "folder_id", "root")
	if folderID != "root" {
		folderID = common.ExtractGoogleResourceID(folderID)
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveListDefaultMaxResults, common.DriveListMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	orderBy := common.ParseStringArg(request.Params.Arguments, "order_by", "name")

	resp, err := srv.ListFiles(ctx, &ListFilesOptions{
		Query:     query,
		PageSize:  maxResults,
		PageToken: pageToken,
		OrderBy:   orderBy,
		Fields:    DriveFileListFields,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	files := make([]map[string]any, 0, len(resp.Files))
	for _, f := range resp.Files {
		files = append(files, formatFile(f))
	}

	result := map[string]any{
		"folder_id":       folderID,
		"files":           files,
		"count":           len(files),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveCreateFolder creates a new folder.
func TestableDriveCreateFolder(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name := common.ParseStringArg(request.Params.Arguments, "name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentID := common.ParseStringArg(request.Params.Arguments, "parent_id", ""); parentID != "" {
		folder.Parents = []string{common.ExtractGoogleResourceID(parentID)}
	}

	if description := common.ParseStringArg(request.Params.Arguments, "description", ""); description != "" {
		folder.Description = description
	}

	created, err := srv.CreateFile(ctx, folder, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"folder_id":    created.Id,
		"name":         created.Name,
		"created_time": created.CreatedTime,
		"url":          created.WebViewLink,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveMove moves a file to a different folder.
func TestableDriveMove(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	newParentID := common.ParseStringArg(request.Params.Arguments, "new_parent_id", "")
	if newParentID == "" {
		return mcp.NewToolResultError("new_parent_id parameter is required"), nil
	}
	newParentID = common.ExtractGoogleResourceID(newParentID)

	// Get current parents to remove
	file, err := srv.GetFile(ctx, fileID, "id,parents")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error getting file: %v", err)), nil
	}

	previousParents := strings.Join(file.Parents, ",")

	updated, err := srv.MoveFile(ctx, fileID, newParentID, previousParents)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success":       true,
		"file_id":       updated.Id,
		"name":          updated.Name,
		"new_parent_id": newParentID,
		"url":           updated.WebViewLink,
		"message":       "File moved successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableDriveCopy copies a file.
func TestableDriveCopy(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	copyFile := &drive.File{}

	if name := common.ParseStringArg(request.Params.Arguments, "name", ""); name != "" {
		copyFile.Name = name
	}

	if parentID := common.ParseStringArg(request.Params.Arguments, "parent_id", ""); parentID != "" {
		copyFile.Parents = []string{common.ExtractGoogleResourceID(parentID)}
	}

	copied, err := srv.CopyFile(ctx, fileID, copyFile)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"file_id":        copied.Id,
		"name":           copied.Name,
		"mime_type":      copied.MimeType,
		"size":           copied.Size,
		"created_time":   copied.CreatedTime,
		"url":            copied.WebViewLink,
		"source_file_id": fileID,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveTrash moves a file to trash.
func TestableDriveTrash(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	updated, err := srv.UpdateFile(ctx, fileID, &drive.File{Trashed: true})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success": true,
		"file_id": updated.Id,
		"name":    updated.Name,
		"trashed": updated.Trashed,
		"message": "File moved to trash",
	}

	return common.MarshalToolResult(result)
}

// TestableDriveDelete permanently deletes a file.
func TestableDriveDelete(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	err := srv.DeleteFile(ctx, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success": true,
		"file_id": fileID,
		"message": "File permanently deleted",
	}

	return common.MarshalToolResult(result)
}

// TestableDriveShare shares a file with users.
func TestableDriveShare(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	email := common.ParseStringArg(request.Params.Arguments, "email", "")
	if email == "" {
		return mcp.NewToolResultError("email parameter is required"), nil
	}

	role := common.ParseStringArg(request.Params.Arguments, "role", "reader")
	permType := common.ParseStringArg(request.Params.Arguments, "type", "user")

	permission := &drive.Permission{
		Type:         permType,
		Role:         role,
		EmailAddress: email,
	}

	sendNotification := common.ParseBoolArg(request.Params.Arguments, "send_notification", true)

	created, err := srv.CreatePermission(ctx, fileID, permission, sendNotification)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success":       true,
		"file_id":       fileID,
		"permission_id": created.Id,
		"type":          created.Type,
		"role":          created.Role,
		"email":         created.EmailAddress,
		"message":       fmt.Sprintf("File shared with %s as %s", email, role),
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGetPermissions gets file permissions.
func TestableDriveGetPermissions(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	resp, err := srv.ListPermissions(ctx, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	permissions := make([]map[string]any, 0, len(resp.Permissions))
	for _, p := range resp.Permissions {
		perm := map[string]any{
			"id":   p.Id,
			"type": p.Type,
			"role": p.Role,
		}
		if p.EmailAddress != "" {
			perm["email"] = p.EmailAddress
		}
		if p.DisplayName != "" {
			perm["display_name"] = p.DisplayName
		}
		if p.Domain != "" {
			perm["domain"] = p.Domain
		}
		permissions = append(permissions, perm)
	}

	result := map[string]any{
		"file_id":     fileID,
		"permissions": permissions,
		"count":       len(permissions),
	}

	return common.MarshalToolResult(result)
}
