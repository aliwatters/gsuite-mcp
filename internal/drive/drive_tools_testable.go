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

	query, errResult := common.RequireStringArg(request.Params.Arguments, "query")
	if errResult != nil {
		return errResult, nil
	}

	// Apply friendly file_type filter if provided
	fileType := common.ParseStringArg(request.Params.Arguments, "file_type", "")
	if fileType != "" {
		mimeType, ok := friendlyFileTypes[strings.ToLower(fileType)]
		if !ok {
			// Not a friendly name — treat as raw mimeType
			mimeType = fileType
		}
		// Prefix-based types (image/, video/, audio/) use "contains" match
		if strings.HasSuffix(mimeType, "/") {
			query = fmt.Sprintf("(%s) and mimeType contains '%s'", query, mimeType)
		} else {
			query = fmt.Sprintf("(%s) and mimeType = '%s'", query, mimeType)
		}
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	corpora := common.ParseStringArg(request.Params.Arguments, "corpora", "allDrives")

	resp, err := srv.ListFiles(ctx, &ListFilesOptions{
		Query:     query,
		PageSize:  maxResults,
		PageToken: pageToken,
		Fields:    DriveFileListFields,
		Corpora:   corpora,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	resolver := NewPathResolver(srv)
	files := make([]map[string]any, 0, len(resp.Files))
	for _, f := range resp.Files {
		fm := formatFile(f)
		if path := resolver.ResolvePath(ctx, f.Parents); path != "" {
			fm["path"] = path
		}
		files = append(files, fm)
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

	result := formatFileFull(file)
	resolver := NewPathResolver(srv)
	if path := resolver.ResolvePath(ctx, file.Parents); path != "" {
		result["path"] = path
	}

	return common.MarshalToolResult(result)
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

	name, errResult := common.RequireStringArg(request.Params.Arguments, "name")
	if errResult != nil {
		return errResult, nil
	}

	contentStr, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
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
		Corpora:   "allDrives",
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	resolver := NewPathResolver(srv)
	files := make([]map[string]any, 0, len(resp.Files))
	for _, f := range resp.Files {
		fm := formatFile(f)
		if path := resolver.ResolvePath(ctx, f.Parents); path != "" {
			fm["path"] = path
		}
		files = append(files, fm)
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

	name, errResult := common.RequireStringArg(request.Params.Arguments, "name")
	if errResult != nil {
		return errResult, nil
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

	newParentID, errResult := common.RequireStringArg(request.Params.Arguments, "new_parent_id")
	if errResult != nil {
		return errResult, nil
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

	email, errResult := common.RequireStringArg(request.Params.Arguments, "email")
	if errResult != nil {
		return errResult, nil
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

// TestableDriveGetShareableLink gets a shareable link and sharing status for a file.
func TestableDriveGetShareableLink(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	file, err := srv.GetFile(ctx, fileID, DriveShareableLinkFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"file_id":   file.Id,
		"name":      file.Name,
		"mime_type": file.MimeType,
		"url":       file.WebViewLink,
	}

	if file.WebContentLink != "" {
		result["download_url"] = file.WebContentLink
	}

	// Include permissions/sharing status
	permissions := make([]map[string]any, 0, len(file.Permissions))
	for _, p := range file.Permissions {
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
	result["permissions"] = permissions
	result["permission_count"] = len(permissions)

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

// === Comments & Replies ===

// formatComment formats a comment for output.
func formatComment(c *drive.Comment) map[string]any {
	result := map[string]any{
		"id":      c.Id,
		"content": c.Content,
	}
	if c.Author != nil {
		result["author"] = map[string]any{
			"display_name": c.Author.DisplayName,
			"email":        c.Author.EmailAddress,
		}
	}
	if c.CreatedTime != "" {
		result["created_time"] = c.CreatedTime
	}
	if c.ModifiedTime != "" {
		result["modified_time"] = c.ModifiedTime
	}
	if c.Resolved {
		result["resolved"] = true
	}
	if c.QuotedFileContent != nil && c.QuotedFileContent.Value != "" {
		result["quoted_content"] = c.QuotedFileContent.Value
	}
	if len(c.Replies) > 0 {
		replies := make([]map[string]any, 0, len(c.Replies))
		for _, r := range c.Replies {
			replies = append(replies, formatReply(r))
		}
		result["replies"] = replies
	}
	return result
}

// formatReply formats a reply for output.
func formatReply(r *drive.Reply) map[string]any {
	result := map[string]any{
		"id":      r.Id,
		"content": r.Content,
	}
	if r.Author != nil {
		result["author"] = map[string]any{
			"display_name": r.Author.DisplayName,
			"email":        r.Author.EmailAddress,
		}
	}
	if r.CreatedTime != "" {
		result["created_time"] = r.CreatedTime
	}
	if r.ModifiedTime != "" {
		result["modified_time"] = r.ModifiedTime
	}
	if r.Action != "" {
		result["action"] = r.Action
	}
	return result
}

// TestableDriveListComments lists comments on a file.
func TestableDriveListComments(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	resp, err := srv.ListComments(ctx, fileID, DriveCommentListFields, maxResults, pageToken, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	comments := make([]map[string]any, 0, len(resp.Comments))
	for _, c := range resp.Comments {
		comments = append(comments, formatComment(c))
	}

	result := map[string]any{
		"file_id":         fileID,
		"comments":        comments,
		"count":           len(comments),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGetComment gets a single comment.
func TestableDriveGetComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	comment, err := srv.GetComment(ctx, fileID, commentID, DriveCommentFields, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(comment)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveCreateComment creates a comment on a file.
func TestableDriveCreateComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	comment := &drive.Comment{
		Content: content,
	}

	created, err := srv.CreateComment(ctx, fileID, comment, DriveCommentFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(created)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveUpdateComment updates a comment.
func TestableDriveUpdateComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	comment := &drive.Comment{
		Content: content,
	}

	updated, err := srv.UpdateComment(ctx, fileID, commentID, comment, DriveCommentFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(updated)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveDeleteComment deletes a comment.
func TestableDriveDeleteComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	err := srv.DeleteComment(ctx, fileID, commentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success":    true,
		"file_id":    fileID,
		"comment_id": commentID,
		"message":    "Comment deleted",
	}

	return common.MarshalToolResult(result)
}

// TestableDriveListReplies lists replies on a comment.
func TestableDriveListReplies(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	resp, err := srv.ListReplies(ctx, fileID, commentID, DriveReplyListFields, maxResults, pageToken, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	replies := make([]map[string]any, 0, len(resp.Replies))
	for _, r := range resp.Replies {
		replies = append(replies, formatReply(r))
	}

	result := map[string]any{
		"file_id":         fileID,
		"comment_id":      commentID,
		"replies":         replies,
		"count":           len(replies),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveCreateReply creates a reply on a comment.
func TestableDriveCreateReply(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	reply := &drive.Reply{
		Content: content,
	}

	// Support resolve action
	action := common.ParseStringArg(request.Params.Arguments, "action", "")
	if action != "" {
		reply.Action = action
	}

	replyFields := "id,content,author(displayName,emailAddress),createdTime,modifiedTime,action"
	created, err := srv.CreateReply(ctx, fileID, commentID, reply, replyFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatReply(created)
	result["file_id"] = fileID
	result["comment_id"] = commentID

	return common.MarshalToolResult(result)
}

// === Revisions ===

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

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

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

	revisionID, errResult := common.RequireStringArg(request.Params.Arguments, "revision_id")
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

	revisionID, errResult := common.RequireStringArg(request.Params.Arguments, "revision_id")
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
