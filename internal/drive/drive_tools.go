package drive

import (
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/drive/v3"
)

// Drive API field constants for optimized responses.
// These reduce response payload size by only requesting needed fields.
const (
	// DriveFileListFields contains fields for file listings (search, list)
	DriveFileListFields = "nextPageToken,files(id,name,mimeType,size,createdTime,modifiedTime,parents,driveId,webViewLink)"
	// DriveFileGetFields contains fields for single file retrieval (full metadata)
	DriveFileGetFields = "id,name,mimeType,size,createdTime,modifiedTime,parents,driveId,webViewLink,webContentLink,description,starred,trashed,owners,permissions"
	// DriveFileDownloadFields contains minimal fields for download operations
	DriveFileDownloadFields = "id,name,mimeType,size"
	// DriveFileCreateFields contains fields for file creation responses
	DriveFileCreateFields = "id,name,mimeType,createdTime,webViewLink"
	// DriveFileMoveFields contains fields for move operation responses
	DriveFileMoveFields = "id,name,parents,webViewLink"
	// DriveFileCopyFields contains fields for copy operation responses
	DriveFileCopyFields = "id,name,mimeType,size,createdTime,webViewLink"
	// DriveFileTrashFields contains fields for trash operation responses
	DriveFileTrashFields = "id,name,trashed"
	// DrivePermissionFields contains fields for permission responses
	DrivePermissionFields = "id,type,role,emailAddress"
	// DrivePermissionListFields contains fields for permission list responses
	DrivePermissionListFields = "permissions(id,type,role,emailAddress,displayName,domain)"
	// DriveFileUploadFields contains fields for upload operation responses
	DriveFileUploadFields = "id,name,mimeType,size,createdTime,webViewLink"
	// DriveShareableLinkFields contains fields for shareable link responses
	DriveShareableLinkFields = "id,name,mimeType,webViewLink,webContentLink,sharingUser,permissions(id,type,role,emailAddress,displayName,domain)"
	// DriveCommentFields contains fields for comment responses
	DriveCommentFields = "id,content,author(displayName,emailAddress),createdTime,modifiedTime,resolved,replies(id,content,author(displayName,emailAddress),createdTime,modifiedTime,action)"
	// DriveCommentListFields contains fields for comment list responses
	DriveCommentListFields = "nextPageToken,comments(id,content,author(displayName,emailAddress),createdTime,modifiedTime,resolved,quotedFileContent,replies(id,content,author(displayName,emailAddress),createdTime,modifiedTime,action))"
	// DriveReplyListFields contains fields for reply list responses
	DriveReplyListFields = "nextPageToken,replies(id,content,author(displayName,emailAddress),createdTime,modifiedTime,action)"
	// DriveRevisionListFields contains fields for revision list responses
	DriveRevisionListFields = "nextPageToken,revisions(id,mimeType,modifiedTime,lastModifyingUser(displayName,emailAddress),size,keepForever,originalFilename)"
	// DriveRevisionGetFields contains fields for single revision responses
	DriveRevisionGetFields = "id,mimeType,modifiedTime,lastModifyingUser(displayName,emailAddress),size,keepForever,originalFilename,exportLinks"
)

// friendlyFileTypes maps friendly names to Google Drive MIME types for search filtering.
var friendlyFileTypes = map[string]string{
	"doc":          "application/vnd.google-apps.document",
	"document":     "application/vnd.google-apps.document",
	"sheet":        "application/vnd.google-apps.spreadsheet",
	"spreadsheet":  "application/vnd.google-apps.spreadsheet",
	"slides":       "application/vnd.google-apps.presentation",
	"presentation": "application/vnd.google-apps.presentation",
	"form":         "application/vnd.google-apps.form",
	"drawing":      "application/vnd.google-apps.drawing",
	"folder":       "application/vnd.google-apps.folder",
	"pdf":          "application/pdf",
	"image":        "image/",
	"video":        "video/",
	"audio":        "audio/",
}

// === Handle functions - generated via WrapHandler ===

var (
	HandleDriveSearch           = common.WrapHandler[DriveService](TestableDriveSearch)
	HandleDriveGet              = common.WrapHandler[DriveService](TestableDriveGet)
	HandleDriveDownload         = common.WrapHandler[DriveService](TestableDriveDownload)
	HandleDriveUpload           = common.WrapHandler[DriveService](TestableDriveUpload)
	HandleDriveList             = common.WrapHandler[DriveService](TestableDriveList)
	HandleDriveCreateFolder     = common.WrapHandler[DriveService](TestableDriveCreateFolder)
	HandleDriveMove             = common.WrapHandler[DriveService](TestableDriveMove)
	HandleDriveCopy             = common.WrapHandler[DriveService](TestableDriveCopy)
	HandleDriveTrash            = common.WrapHandler[DriveService](TestableDriveTrash)
	HandleDriveDelete           = common.WrapHandler[DriveService](TestableDriveDelete)
	HandleDriveShare            = common.WrapHandler[DriveService](TestableDriveShare)
	HandleDriveGetPermissions   = common.WrapHandler[DriveService](TestableDriveGetPermissions)
	HandleDriveGetShareableLink = common.WrapHandler[DriveService](TestableDriveGetShareableLink)

	// Comments & Replies
	HandleDriveListComments  = common.WrapHandler[DriveService](TestableDriveListComments)
	HandleDriveGetComment    = common.WrapHandler[DriveService](TestableDriveGetComment)
	HandleDriveCreateComment = common.WrapHandler[DriveService](TestableDriveCreateComment)
	HandleDriveUpdateComment = common.WrapHandler[DriveService](TestableDriveUpdateComment)
	HandleDriveDeleteComment = common.WrapHandler[DriveService](TestableDriveDeleteComment)
	HandleDriveListReplies   = common.WrapHandler[DriveService](TestableDriveListReplies)
	HandleDriveCreateReply   = common.WrapHandler[DriveService](TestableDriveCreateReply)

	// Revisions
	HandleDriveListRevisions    = common.WrapHandler[DriveService](TestableDriveListRevisions)
	HandleDriveGetRevision      = common.WrapHandler[DriveService](TestableDriveGetRevision)
	HandleDriveDownloadRevision = common.WrapHandler[DriveService](TestableDriveDownloadRevision)
)

// formatFile formats a file for compact output
func formatFile(file *drive.File) map[string]any {
	result := map[string]any{
		"id":        file.Id,
		"name":      file.Name,
		"mime_type": file.MimeType,
	}

	if file.Size > 0 {
		result["size"] = file.Size
	}

	if file.CreatedTime != "" {
		result["created_time"] = file.CreatedTime
	}

	if file.ModifiedTime != "" {
		result["modified_time"] = file.ModifiedTime
	}

	if file.WebViewLink != "" {
		result["url"] = file.WebViewLink
	}

	return result
}

// formatFileFull formats a file with all details
func formatFileFull(file *drive.File) map[string]any {
	result := formatFile(file)

	if file.Description != "" {
		result["description"] = file.Description
	}

	if file.Starred {
		result["starred"] = true
	}

	if file.Trashed {
		result["trashed"] = true
	}

	if len(file.Parents) > 0 {
		result["parents"] = file.Parents
	}

	if file.WebContentLink != "" {
		result["download_url"] = file.WebContentLink
	}

	if len(file.Owners) > 0 {
		owners := make([]map[string]any, 0, len(file.Owners))
		for _, o := range file.Owners {
			owners = append(owners, map[string]any{
				"email":        o.EmailAddress,
				"display_name": o.DisplayName,
			})
		}
		result["owners"] = owners
	}

	return result
}

// isTextMimeType checks if the mime type represents text content
func isTextMimeType(mimeType string) bool {
	textTypes := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-yaml",
		"application/x-sh",
	}

	for _, t := range textTypes {
		if strings.HasPrefix(mimeType, t) {
			return true
		}
	}

	return false
}
