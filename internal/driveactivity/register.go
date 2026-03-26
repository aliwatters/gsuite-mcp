package driveactivity

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Drive Activity tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// driveactivity_query - Query activity for a Drive item or folder
	s.AddTool(mcp.NewTool("driveactivity_query",
		mcp.WithDescription("Query activity history for a Google Drive file or folder. Returns audit trail including who edited, created, moved, renamed, or shared items. Supports filtering by time range and action type."),
		mcp.WithString("item_id", mcp.Description("Drive file ID or URL to query activity for (mutually exclusive with folder_id)")),
		mcp.WithString("folder_id", mcp.Description("Drive folder ID or URL to query activity for all items within (mutually exclusive with item_id)")),
		mcp.WithString("filter", mcp.Description("Filter string (e.g., 'time > \"2024-01-01T00:00:00Z\"', 'detail.action_detail_case:EDIT', 'detail.action_detail_case:(CREATE EDIT)')")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleDriveActivityQuery)
}
