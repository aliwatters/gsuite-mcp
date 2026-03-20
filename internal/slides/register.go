package slides

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Slides tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Slides Read (Phase 1) ===

	// slides_get_presentation - Get presentation metadata and slide list
	s.AddTool(mcp.NewTool("slides_get_presentation",
		mcp.WithDescription("Get a Google Slides presentation's metadata, slide list with text previews, and structure."),
		mcp.WithString("presentation_id", mcp.Required(), mcp.Description("Presentation ID or full Google Slides URL")),
		common.WithAccountParam(),
	), common.WithLargeContentHint(common.WithDriveAccessCheck(HandleSlidesGetPresentation, "presentation_id")))

	// slides_get_page - Get a single slide/page with full element details
	s.AddTool(mcp.NewTool("slides_get_page",
		mcp.WithDescription("Get full details of a single slide including all elements (shapes, images, tables, text)."),
		mcp.WithString("presentation_id", mcp.Required(), mcp.Description("Presentation ID or full Google Slides URL")),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Page/slide object ID (from slides_get_presentation response)")),
		common.WithAccountParam(),
	), common.WithDriveAccessCheck(HandleSlidesGetPage, "presentation_id"))

	// slides_get_thumbnail - Get slide thumbnail image URL
	s.AddTool(mcp.NewTool("slides_get_thumbnail",
		mcp.WithDescription("Get a thumbnail image URL for a slide. Returns a temporary URL to a PNG image."),
		mcp.WithString("presentation_id", mcp.Required(), mcp.Description("Presentation ID or full Google Slides URL")),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Page/slide object ID")),
		common.WithAccountParam(),
	), common.WithDriveAccessCheck(HandleSlidesGetThumbnail, "presentation_id"))

	// === Slides Write (Phase 2) ===

	// slides_create - Create a new presentation
	s.AddTool(mcp.NewTool("slides_create",
		mcp.WithDescription("Create a new Google Slides presentation with the given title."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Presentation title")),
		common.WithAccountParam(),
	), HandleSlidesCreate)

	// slides_batch_update - Batch update presentation
	s.AddTool(mcp.NewTool("slides_batch_update",
		mcp.WithDescription("Execute batch update requests on a Google Slides presentation. Power user escape hatch for adding/modifying/deleting slides, text, shapes, and images. See Google Slides API batchUpdate documentation for request format."),
		mcp.WithString("presentation_id", mcp.Required(), mcp.Description("Presentation ID or full Google Slides URL")),
		mcp.WithString("requests", mcp.Required(), mcp.Description("JSON array of batch update requests (see Google Slides API docs)")),
		common.WithAccountParam(),
	), common.WithDriveAccessCheck(HandleSlidesBatchUpdate, "presentation_id"))
}
