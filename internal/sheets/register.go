package sheets

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Sheets tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Sheets Core (Phase 1) ===

	// sheets_get - Get spreadsheet metadata
	s.AddTool(mcp.NewTool("sheets_get",
		mcp.WithDescription("Get spreadsheet metadata including title, sheets list, and properties."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		common.WithAccountParam(),
	), HandleSheetsGet)

	// sheets_read - Read cell values from range
	s.AddTool(mcp.NewTool("sheets_read",
		mcp.WithDescription("Read cell values from a range using A1 notation."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range (e.g., 'Sheet1!A1:C10', 'A1:B5')")),
		common.WithAccountParam(),
	), HandleSheetsRead)

	// sheets_write - Write values to cell range
	s.AddTool(mcp.NewTool("sheets_write",
		mcp.WithDescription("Write values to a cell range. Values are parsed as formulas when using USER_ENTERED."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range (e.g., 'Sheet1!A1:C3')")),
		mcp.WithArray("values", mcp.Required(), mcp.Description("2D array of values (rows of cells)")),
		mcp.WithString("value_input_option", mcp.Description("How to interpret input: RAW (as-is) or USER_ENTERED (parse formulas, default)")),
		common.WithAccountParam(),
	), HandleSheetsWrite)

	// sheets_append - Append rows to sheet
	s.AddTool(mcp.NewTool("sheets_append",
		mcp.WithDescription("Append rows after the last row of a table. New rows are inserted below existing data."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range identifying the table (e.g., 'Sheet1!A:C')")),
		mcp.WithArray("values", mcp.Required(), mcp.Description("2D array of values (rows to append)")),
		mcp.WithString("value_input_option", mcp.Description("How to interpret input: RAW or USER_ENTERED (default)")),
		common.WithAccountParam(),
	), HandleSheetsAppend)

	// === Sheets Extended (Phase 2) ===

	// sheets_create - Create new spreadsheet
	s.AddTool(mcp.NewTool("sheets_create",
		mcp.WithDescription("Create a new Google Sheets spreadsheet."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title for the new spreadsheet")),
		common.WithAccountParam(),
	), HandleSheetsCreate)

	// sheets_batch_read - Read multiple ranges at once
	s.AddTool(mcp.NewTool("sheets_batch_read",
		mcp.WithDescription("Read multiple ranges from a spreadsheet in one request."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithArray("ranges", mcp.Required(), mcp.Description("Array of A1 notation ranges to read")),
		common.WithAccountParam(),
	), HandleSheetsBatchRead)

	// sheets_batch_write - Write to multiple ranges at once
	s.AddTool(mcp.NewTool("sheets_batch_write",
		mcp.WithDescription("Write to multiple ranges in one request."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithArray("data", mcp.Required(), mcp.Description("Array of {range, values} objects")),
		mcp.WithString("value_input_option", mcp.Description("How to interpret input: RAW or USER_ENTERED (default)")),
		common.WithAccountParam(),
	), HandleSheetsBatchWrite)

	// sheets_clear - Clear cell range
	s.AddTool(mcp.NewTool("sheets_clear",
		mcp.WithDescription("Clear values from a cell range (removes values but keeps formatting)."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("range", mcp.Required(), mcp.Description("A1 notation range to clear (e.g., 'Sheet1!A1:C10')")),
		common.WithAccountParam(),
	), HandleSheetsClear)
}
