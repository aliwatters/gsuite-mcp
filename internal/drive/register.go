package drive

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Drive tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Drive Core (Phase 1) ===

	// drive_search - Search files with query syntax
	s.AddTool(mcp.NewTool("drive_search",
		mcp.WithDescription("Search Google Drive files with query syntax. Use queries like \"name contains 'report'\" or \"mimeType = 'application/pdf'\"."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Drive search query (e.g., \"name contains 'budget'\", \"mimeType = 'application/pdf'\")")),
		mcp.WithNumber("max_results", mcp.Description("Maximum results to return (1-100, default 20)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleDriveSearch)

	// drive_get - Get file metadata
	s.AddTool(mcp.NewTool("drive_get",
		mcp.WithDescription("Get detailed metadata for a Google Drive file."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		common.WithAccountParam(),
	), HandleDriveGet)

	// drive_download - Download file content
	s.AddTool(mcp.NewTool("drive_download",
		mcp.WithDescription("Download file content. Returns text for text files, base64 for binary. Max 10MB. Google Docs/Sheets are exported as text/CSV."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		common.WithAccountParam(),
	), HandleDriveDownload)

	// drive_upload - Upload new file
	s.AddTool(mcp.NewTool("drive_upload",
		mcp.WithDescription("Upload a new file to Google Drive."),
		mcp.WithString("name", mcp.Required(), mcp.Description("File name")),
		mcp.WithString("content", mcp.Required(), mcp.Description("File content (text or base64-encoded)")),
		mcp.WithString("encoding", mcp.Description("Content encoding: utf-8 (default) or base64")),
		mcp.WithString("mime_type", mcp.Description("MIME type (auto-detected if omitted)")),
		mcp.WithString("parent_id", mcp.Description("Parent folder ID (uploads to root if omitted)")),
		mcp.WithString("description", mcp.Description("File description")),
		common.WithAccountParam(),
	), HandleDriveUpload)

	// drive_list - List files in folder
	s.AddTool(mcp.NewTool("drive_list",
		mcp.WithDescription("List files in a Google Drive folder."),
		mcp.WithString("folder_id", mcp.Description("Folder ID (default: root folder)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum results (1-1000, default 100)")),
		common.WithPageToken(),
		mcp.WithString("order_by", mcp.Description("Sort order: name, modifiedTime, createdTime (default: name)")),
		common.WithAccountParam(),
	), HandleDriveList)

	// drive_create_folder - Create new folder
	s.AddTool(mcp.NewTool("drive_create_folder",
		mcp.WithDescription("Create a new folder in Google Drive."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Folder name")),
		mcp.WithString("parent_id", mcp.Description("Parent folder ID (creates in root if omitted)")),
		mcp.WithString("description", mcp.Description("Folder description")),
		common.WithAccountParam(),
	), HandleDriveCreateFolder)

	// drive_move - Move file to different folder
	s.AddTool(mcp.NewTool("drive_move",
		mcp.WithDescription("Move a file to a different folder in Google Drive."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		mcp.WithString("new_parent_id", mcp.Required(), mcp.Description("Destination folder ID")),
		common.WithAccountParam(),
	), HandleDriveMove)

	// drive_copy - Copy a file
	s.AddTool(mcp.NewTool("drive_copy",
		mcp.WithDescription("Create a copy of a Google Drive file."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		mcp.WithString("name", mcp.Description("New file name (default: Copy of original)")),
		mcp.WithString("parent_id", mcp.Description("Destination folder ID (default: same folder)")),
		common.WithAccountParam(),
	), HandleDriveCopy)

	// drive_trash - Move file to trash
	s.AddTool(mcp.NewTool("drive_trash",
		mcp.WithDescription("Move a Google Drive file to trash."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		common.WithAccountParam(),
	), HandleDriveTrash)

	// drive_delete - Permanently delete file
	s.AddTool(mcp.NewTool("drive_delete",
		mcp.WithDescription("Permanently delete a Google Drive file (cannot be undone)."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		common.WithAccountParam(),
	), HandleDriveDelete)

	// === Drive Sharing ===

	// drive_share - Share file with users
	s.AddTool(mcp.NewTool("drive_share",
		mcp.WithDescription("Share a Google Drive file with a user or group."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Email address to share with")),
		mcp.WithString("role", mcp.Description("Permission role: reader, writer, commenter (default: reader)")),
		mcp.WithString("type", mcp.Description("Permission type: user, group, domain, anyone (default: user)")),
		mcp.WithBoolean("send_notification", mcp.Description("Send notification email (default: true)")),
		common.WithAccountParam(),
	), HandleDriveShare)

	// drive_get_permissions - Get file permissions
	s.AddTool(mcp.NewTool("drive_get_permissions",
		mcp.WithDescription("Get permissions/sharing settings for a Google Drive file."),
		mcp.WithString("file_id", mcp.Required(), mcp.Description("File ID or Google Drive URL")),
		common.WithAccountParam(),
	), HandleDriveGetPermissions)
}
