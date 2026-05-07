package drive

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// TestableDriveSearch searches files with query syntax.
func TestableDriveSearch(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	query, errResult := common.RequireStringArg(request.GetArguments(), "query")
	if errResult != nil {
		return errResult, nil
	}

	// Apply friendly file_type filter if provided
	fileType := common.ParseStringArg(request.GetArguments(), "file_type", "")
	if fileType != "" {
		mimeType, ok := friendlyFileTypes[strings.ToLower(fileType)]
		if !ok {
			// Not a friendly name — treat as raw mimeType after validation
			if !validMimeType.MatchString(fileType) && !strings.HasSuffix(fileType, "/") {
				return mcp.NewToolResultError(fmt.Sprintf("invalid file_type %q: use a friendly name (doc, sheet, slides, pdf, folder, image, video, audio) or a valid MIME type", fileType)), nil
			}
			mimeType = fileType
		}
		// Prefix-based types (image/, video/, audio/) use "contains" match
		if strings.HasSuffix(mimeType, "/") {
			query = fmt.Sprintf("(%s) and mimeType contains '%s'", query, mimeType)
		} else {
			query = fmt.Sprintf("(%s) and mimeType = '%s'", query, mimeType)
		}
	}

	maxResults := common.ParseMaxResults(request.GetArguments(), common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.GetArguments(), "page_token", "")
	corpora := common.ParseStringArg(request.GetArguments(), "corpora", "allDrives")

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

	name, errResult := common.RequireStringArg(request.GetArguments(), "name")
	if errResult != nil {
		return errResult, nil
	}

	contentStr, errResult := common.RequireStringArg(request.GetArguments(), "content")
	if errResult != nil {
		return errResult, nil
	}

	// Reject obviously oversized content before decoding
	if int64(len(contentStr)) > common.DriveMaxFileSize*2 {
		return mcp.NewToolResultError(fmt.Sprintf("Content too large. Maximum supported size is %d bytes", common.DriveMaxFileSize)), nil
	}

	// Decode content if base64 encoded
	var data []byte
	encoding := common.ParseStringArg(request.GetArguments(), "encoding", "")
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

	if parentID := common.ParseStringArg(request.GetArguments(), "parent_id", ""); parentID != "" {
		file.Parents = []string{common.ExtractGoogleResourceID(parentID)}
	}

	if mimeType := common.ParseStringArg(request.GetArguments(), "mime_type", ""); mimeType != "" {
		file.MimeType = mimeType
	}

	if description := common.ParseStringArg(request.GetArguments(), "description", ""); description != "" {
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

	folderID := common.ParseStringArg(request.GetArguments(), "folder_id", "root")
	if folderID != "root" {
		folderID = common.ExtractGoogleResourceID(folderID)
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	maxResults := common.ParseMaxResults(request.GetArguments(), common.DriveListDefaultMaxResults, common.DriveListMaxResultsLimit)
	pageToken := common.ParseStringArg(request.GetArguments(), "page_token", "")
	orderBy := common.ParseStringArg(request.GetArguments(), "order_by", "name")

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

	name, errResult := common.RequireStringArg(request.GetArguments(), "name")
	if errResult != nil {
		return errResult, nil
	}

	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentID := common.ParseStringArg(request.GetArguments(), "parent_id", ""); parentID != "" {
		folder.Parents = []string{common.ExtractGoogleResourceID(parentID)}
	}

	if description := common.ParseStringArg(request.GetArguments(), "description", ""); description != "" {
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

	newParentID, errResult := common.RequireStringArg(request.GetArguments(), "new_parent_id")
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

	if name := common.ParseStringArg(request.GetArguments(), "name", ""); name != "" {
		copyFile.Name = name
	}

	if parentID := common.ParseStringArg(request.GetArguments(), "parent_id", ""); parentID != "" {
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
