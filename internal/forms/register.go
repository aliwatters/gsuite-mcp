package forms

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Forms tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Forms Read ===

	// forms_get - Get form metadata and structure
	s.AddTool(mcp.NewTool("forms_get",
		mcp.WithDescription("Get a Google Form's metadata, questions, and structure including item types and options."),
		mcp.WithString("form_id", mcp.Required(), mcp.Description("Form ID or full Google Forms URL")),
		common.WithAccountParam(),
	), HandleFormsGet)

	// forms_list_responses - List form responses
	s.AddTool(mcp.NewTool("forms_list_responses",
		mcp.WithDescription("List all responses submitted to a Google Form."),
		mcp.WithString("form_id", mcp.Required(), mcp.Description("Form ID or full Google Forms URL")),
		common.WithAccountParam(),
	), HandleFormsListResponses)

	// forms_get_response - Get a single form response
	s.AddTool(mcp.NewTool("forms_get_response",
		mcp.WithDescription("Get a single response submitted to a Google Form by response ID."),
		mcp.WithString("form_id", mcp.Required(), mcp.Description("Form ID or full Google Forms URL")),
		mcp.WithString("response_id", mcp.Required(), mcp.Description("Response ID (from forms_list_responses)")),
		common.WithAccountParam(),
	), HandleFormsGetResponse)

	// === Forms Write ===

	// forms_create - Create a new form
	s.AddTool(mcp.NewTool("forms_create",
		mcp.WithDescription("Create a new Google Form with the given title."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Form title")),
		common.WithAccountParam(),
	), HandleFormsCreate)

	// forms_batch_update - Batch update form
	s.AddTool(mcp.NewTool("forms_batch_update",
		mcp.WithDescription("Execute batch update requests on a Google Form. Use to add/update/delete questions, update form info, and modify settings. See Google Forms API batchUpdate documentation for request format."),
		mcp.WithString("form_id", mcp.Required(), mcp.Description("Form ID or full Google Forms URL")),
		mcp.WithString("requests", mcp.Required(), mcp.Description("JSON array of batch update requests (see Google Forms API docs)")),
		common.WithAccountParam(),
	), HandleFormsBatchUpdate)
}
