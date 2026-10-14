package drive

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
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

	query := common.ParseStringArg(request.GetArguments(), "query", "")
	exactName := common.ParseStringArg(request.GetArguments(), "name", "")
	searchParent := common.ParseStringArg(request.GetArguments(), "parent_id", "")
	if query == "" && exactName == "" && searchParent == "" {
		return mcp.NewToolResultError("provide at least one of query, name, or parent_id"), nil
	}
	if exactName != "" {
		clause := fmt.Sprintf("name = '%s'", escapeDriveQueryValue(exactName))
		if query == "" {
			query = clause
		} else {
			query = fmt.Sprintf("(%s) and %s", query, clause)
		}
	}
	if searchParent != "" {
		clause := fmt.Sprintf("'%s' in parents", escapeDriveQueryValue(common.ExtractGoogleResourceID(searchParent)))
		if query == "" {
			query = clause
		} else {
			query = fmt.Sprintf("(%s) and %s", query, clause)
		}
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
		"truncated":       resp.NextPageToken != "",
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

// resolveUploadBody returns the reader for an upload or content update from
// exactly one of "content" or "file_path". The returned close func must be
// called when done. Content-based bodies are capped at DriveMaxFileSize; local
// files are streamed and not size-capped.
func resolveUploadBody(args map[string]any, required bool) (io.Reader, func(), *mcp.CallToolResult) {
	noop := func() {}
	contentStr := common.ParseStringArg(args, "content", "")
	_, hasContent := args["content"]
	filePath := common.ParseStringArg(args, "file_path", "")

	if hasContent && filePath != "" {
		return nil, noop, mcp.NewToolResultError("provide exactly one of content or file_path, not both")
	}
	if !hasContent && filePath == "" {
		if required {
			return nil, noop, mcp.NewToolResultError("exactly one of content or file_path is required")
		}
		return nil, noop, nil
	}

	if filePath != "" {
		f, _, err := openRegularFile(filePath)
		if err != nil {
			log.Printf("drive: upload source rejected: %v", err)
			return nil, noop, mcp.NewToolResultError(err.Error())
		}
		return f, func() { _ = f.Close() }, nil
	}

	// Reject obviously oversized content before decoding
	if int64(len(contentStr)) > common.DriveMaxFileSize*2 {
		return nil, noop, mcp.NewToolResultError(fmt.Sprintf("Content too large. Maximum supported size is %d bytes", common.DriveMaxFileSize))
	}

	var data []byte
	if common.ParseStringArg(args, "encoding", "") == "base64" {
		var err error
		data, err = base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			return nil, noop, mcp.NewToolResultError(fmt.Sprintf("Invalid base64 content: %v", err))
		}
	} else {
		data = []byte(contentStr)
	}
	if int64(len(data)) > common.DriveMaxFileSize {
		return nil, noop, mcp.NewToolResultError(fmt.Sprintf("Content too large (%d bytes). Maximum supported size is %d bytes", len(data), common.DriveMaxFileSize))
	}
	return bytes.NewReader(data), noop, nil
}

// formatUploadResult formats an upload/update result, including size and md5Checksum.
func formatUploadResult(f *drive.File) map[string]any {
	return map[string]any{
		"id":           f.Id,
		"file_id":      f.Id,
		"name":         f.Name,
		"mime_type":    f.MimeType,
		"size":         f.Size,
		"md5Checksum":  f.Md5Checksum,
		"created_time": f.CreatedTime,
		"url":          f.WebViewLink,
	}
}

// TestableDriveUpload uploads a new file from inline content or a local file path.
func TestableDriveUpload(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name, errResult := common.RequireStringArg(request.GetArguments(), "name")
	if errResult != nil {
		return errResult, nil
	}

	body, closeBody, errResult := resolveUploadBody(request.GetArguments(), true)
	if errResult != nil {
		return errResult, nil
	}
	defer closeBody()

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

	created, err := srv.CreateFile(ctx, file, body)
	if err != nil {
		return driveAPIErrorResult("upload", name, err), nil
	}

	return common.MarshalToolResult(formatUploadResult(created))
}

// TestableDriveUpdateFile replaces a file's body and/or renames it.
func TestableDriveUpdateFile(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	args := request.GetArguments()
	newName := common.ParseStringArg(args, "name", "")

	body, closeBody, errResult := resolveUploadBody(args, false)
	if errResult != nil {
		return errResult, nil
	}
	defer closeBody()

	if body == nil && newName == "" {
		return mcp.NewToolResultError("nothing to update: provide name, content, or file_path"), nil
	}

	meta := &drive.File{Name: newName}
	if mimeType := common.ParseStringArg(args, "mime_type", ""); mimeType != "" {
		meta.MimeType = mimeType
	}

	label := newName
	if label == "" {
		label = fileID
	}

	var updated *drive.File
	var err error
	if body != nil {
		updated, err = srv.UpdateFileContent(ctx, fileID, meta, body)
	} else {
		updated, err = srv.UpdateFile(ctx, fileID, meta)
	}
	if err != nil {
		return driveAPIErrorResult("update", label+" ("+fileID+")", err), nil
	}

	result := formatUploadResult(updated)
	result["success"] = true
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

	query := fmt.Sprintf("'%s' in parents and trashed = false", escapeDriveQueryValue(folderID))
	if exactName := common.ParseStringArg(request.GetArguments(), "name", ""); exactName != "" {
		query += fmt.Sprintf(" and name = '%s'", escapeDriveQueryValue(exactName))
	}
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
		"truncated":       resp.NextPageToken != "",
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
