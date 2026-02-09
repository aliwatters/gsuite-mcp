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
	DriveFileListFields = "nextPageToken,files(id,name,mimeType,size,createdTime,modifiedTime,parents,webViewLink)"
	// DriveFileGetFields contains fields for single file retrieval (full metadata)
	DriveFileGetFields = "id,name,mimeType,size,createdTime,modifiedTime,parents,webViewLink,webContentLink,description,starred,trashed,owners,permissions"
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
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleDriveSearch         = common.WrapHandler[DriveService](TestableDriveSearch)
	HandleDriveGet            = common.WrapHandler[DriveService](TestableDriveGet)
	HandleDriveDownload       = common.WrapHandler[DriveService](TestableDriveDownload)
	HandleDriveUpload         = common.WrapHandler[DriveService](TestableDriveUpload)
	HandleDriveList           = common.WrapHandler[DriveService](TestableDriveList)
	HandleDriveCreateFolder   = common.WrapHandler[DriveService](TestableDriveCreateFolder)
	HandleDriveMove           = common.WrapHandler[DriveService](TestableDriveMove)
	HandleDriveCopy           = common.WrapHandler[DriveService](TestableDriveCopy)
	HandleDriveTrash          = common.WrapHandler[DriveService](TestableDriveTrash)
	HandleDriveDelete         = common.WrapHandler[DriveService](TestableDriveDelete)
	HandleDriveShare          = common.WrapHandler[DriveService](TestableDriveShare)
	HandleDriveGetPermissions = common.WrapHandler[DriveService](TestableDriveGetPermissions)
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
