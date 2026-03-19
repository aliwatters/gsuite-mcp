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

	// === Sheets Formatting (Phase 3) ===

	// sheets_format_cells - Apply formatting to cell range
	s.AddTool(mcp.NewTool("sheets_format_cells",
		mcp.WithDescription("Apply formatting to a cell range: background color, font, bold, italic, number format, alignment."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("sheet_id", mcp.Required(), mcp.Description("Numeric sheet ID (from sheets_get response)")),
		mcp.WithNumber("start_row", mcp.Required(), mcp.Description("Start row index (0-based)")),
		mcp.WithNumber("start_col", mcp.Required(), mcp.Description("Start column index (0-based)")),
		mcp.WithNumber("end_row", mcp.Required(), mcp.Description("End row index (0-based, exclusive)")),
		mcp.WithNumber("end_col", mcp.Required(), mcp.Description("End column index (0-based, exclusive)")),
		mcp.WithString("background_color", mcp.Description("Background color as hex (e.g., '#FF0000' or 'FF0000')")),
		mcp.WithBoolean("bold", mcp.Description("Bold text")),
		mcp.WithBoolean("italic", mcp.Description("Italic text")),
		mcp.WithBoolean("underline", mcp.Description("Underline text")),
		mcp.WithBoolean("strikethrough", mcp.Description("Strikethrough text")),
		mcp.WithString("font_family", mcp.Description("Font family (e.g., 'Arial', 'Times New Roman')")),
		mcp.WithNumber("font_size", mcp.Description("Font size in points")),
		mcp.WithString("foreground_color", mcp.Description("Text color as hex (e.g., '#0000FF')")),
		mcp.WithString("number_format", mcp.Description("Number format pattern (e.g., '#,##0.00', 'yyyy-mm-dd')")),
		mcp.WithString("number_format_type", mcp.Description("Format type: TEXT, NUMBER, PERCENT, CURRENCY, DATE, TIME, DATE_TIME, SCIENTIFIC")),
		mcp.WithString("horizontal_alignment", mcp.Description("Horizontal alignment: LEFT, CENTER, RIGHT")),
		mcp.WithString("vertical_alignment", mcp.Description("Vertical alignment: TOP, MIDDLE, BOTTOM")),
		mcp.WithString("wrap_strategy", mcp.Description("Text wrap: OVERFLOW_CELL, LEGACY_WRAP, CLIP, WRAP")),
		common.WithAccountParam(),
	), HandleSheetsFormatCells)

	// sheets_add_conditional_format - Add conditional formatting rule
	s.AddTool(mcp.NewTool("sheets_add_conditional_format",
		mcp.WithDescription("Add a conditional formatting rule (boolean or gradient) to a cell range."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("sheet_id", mcp.Required(), mcp.Description("Numeric sheet ID")),
		mcp.WithNumber("start_row", mcp.Required(), mcp.Description("Start row index (0-based)")),
		mcp.WithNumber("start_col", mcp.Required(), mcp.Description("Start column index (0-based)")),
		mcp.WithNumber("end_row", mcp.Required(), mcp.Description("End row index (0-based, exclusive)")),
		mcp.WithNumber("end_col", mcp.Required(), mcp.Description("End column index (0-based, exclusive)")),
		mcp.WithString("rule_type", mcp.Required(), mcp.Description("Rule type: BOOLEAN or GRADIENT")),
		// Boolean rule params
		mcp.WithString("condition_type", mcp.Description("Condition type for BOOLEAN rules (e.g., NUMBER_GREATER, TEXT_CONTAINS, CUSTOM_FORMULA)")),
		mcp.WithArray("condition_values", mcp.Description("Condition values (array of strings)")),
		mcp.WithString("format_background_color", mcp.Description("Background color for matching cells (hex)")),
		mcp.WithString("format_text_color", mcp.Description("Text color for matching cells (hex)")),
		mcp.WithBoolean("format_bold", mcp.Description("Bold for matching cells")),
		mcp.WithBoolean("format_italic", mcp.Description("Italic for matching cells")),
		// Gradient rule params
		mcp.WithString("min_color", mcp.Description("Gradient min color (hex)")),
		mcp.WithString("min_type", mcp.Description("Gradient min type: MIN, NUMBER, PERCENT, PERCENTILE")),
		mcp.WithString("min_value", mcp.Description("Gradient min value")),
		mcp.WithString("mid_color", mcp.Description("Gradient midpoint color (hex, optional)")),
		mcp.WithString("mid_type", mcp.Description("Gradient mid type: NUMBER, PERCENT, PERCENTILE")),
		mcp.WithString("mid_value", mcp.Description("Gradient mid value")),
		mcp.WithString("max_color", mcp.Description("Gradient max color (hex)")),
		mcp.WithString("max_type", mcp.Description("Gradient max type: MAX, NUMBER, PERCENT, PERCENTILE")),
		mcp.WithString("max_value", mcp.Description("Gradient max value")),
		common.WithAccountParam(),
	), HandleSheetsAddConditionalFormat)

	// sheets_add_data_validation - Add data validation rules
	s.AddTool(mcp.NewTool("sheets_add_data_validation",
		mcp.WithDescription("Add data validation to a cell range (dropdowns, number constraints, custom formulas)."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("sheet_id", mcp.Required(), mcp.Description("Numeric sheet ID")),
		mcp.WithNumber("start_row", mcp.Required(), mcp.Description("Start row index (0-based)")),
		mcp.WithNumber("start_col", mcp.Required(), mcp.Description("Start column index (0-based)")),
		mcp.WithNumber("end_row", mcp.Required(), mcp.Description("End row index (0-based, exclusive)")),
		mcp.WithNumber("end_col", mcp.Required(), mcp.Description("End column index (0-based, exclusive)")),
		mcp.WithString("validation_type", mcp.Required(), mcp.Description("Validation type: ONE_OF_LIST, ONE_OF_RANGE, NUMBER_BETWEEN, NUMBER_GREATER, TEXT_CONTAINS, CUSTOM_FORMULA, etc.")),
		mcp.WithArray("values", mcp.Description("Validation values (e.g., dropdown options for ONE_OF_LIST, or [min, max] for NUMBER_BETWEEN)")),
		mcp.WithBoolean("strict", mcp.Description("Reject invalid input (default: true)")),
		mcp.WithBoolean("show_dropdown", mcp.Description("Show dropdown UI for list validations (default: true)")),
		mcp.WithString("input_message", mcp.Description("Help text shown when cell is selected")),
		common.WithAccountParam(),
	), HandleSheetsAddDataValidation)

	// === Sheets Charts & Pivot Tables (Phase 3) ===

	// sheets_create_chart - Create embedded chart
	s.AddTool(mcp.NewTool("sheets_create_chart",
		mcp.WithDescription("Create an embedded chart from spreadsheet data."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("chart_type", mcp.Required(), mcp.Description("Chart type: BAR, LINE, AREA, COLUMN, SCATTER, COMBO, STEPPED_AREA, PIE")),
		mcp.WithNumber("sheet_id", mcp.Required(), mcp.Description("Sheet ID containing source data")),
		mcp.WithNumber("start_row", mcp.Required(), mcp.Description("Source data start row (0-based)")),
		mcp.WithNumber("start_col", mcp.Required(), mcp.Description("Source data start column (0-based)")),
		mcp.WithNumber("end_row", mcp.Required(), mcp.Description("Source data end row (0-based, exclusive)")),
		mcp.WithNumber("end_col", mcp.Required(), mcp.Description("Source data end column (0-based, exclusive)")),
		mcp.WithString("title", mcp.Description("Chart title")),
		mcp.WithString("legend_position", mcp.Description("Legend position: BOTTOM_LEGEND, LEFT_LEGEND, RIGHT_LEGEND, TOP_LEGEND, NO_LEGEND")),
		mcp.WithNumber("anchor_row", mcp.Description("Row index for chart placement (default: 0)")),
		mcp.WithNumber("anchor_col", mcp.Description("Column index for chart placement (default: 0)")),
		common.WithAccountParam(),
	), HandleSheetsCreateChart)

	// sheets_update_chart - Update existing chart
	s.AddTool(mcp.NewTool("sheets_update_chart",
		mcp.WithDescription("Update an existing embedded chart's title or type."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("chart_id", mcp.Required(), mcp.Description("Chart ID (from sheets_create_chart response)")),
		mcp.WithString("title", mcp.Description("New chart title")),
		mcp.WithString("chart_type", mcp.Description("New chart type: BAR, LINE, AREA, COLUMN, SCATTER, COMBO, STEPPED_AREA, PIE")),
		common.WithAccountParam(),
	), HandleSheetsUpdateChart)

	// sheets_delete_chart - Delete embedded chart
	s.AddTool(mcp.NewTool("sheets_delete_chart",
		mcp.WithDescription("Delete an embedded chart from a spreadsheet."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("chart_id", mcp.Required(), mcp.Description("Chart ID to delete")),
		common.WithAccountParam(),
	), HandleSheetsDeleteChart)

	// sheets_create_pivot_table - Create pivot table
	s.AddTool(mcp.NewTool("sheets_create_pivot_table",
		mcp.WithDescription("Create a pivot table for data summarization."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithNumber("source_sheet_id", mcp.Required(), mcp.Description("Sheet ID containing source data")),
		mcp.WithNumber("start_row", mcp.Required(), mcp.Description("Source data start row (0-based)")),
		mcp.WithNumber("start_col", mcp.Required(), mcp.Description("Source data start column (0-based)")),
		mcp.WithNumber("end_row", mcp.Required(), mcp.Description("Source data end row (0-based, exclusive)")),
		mcp.WithNumber("end_col", mcp.Required(), mcp.Description("Source data end column (0-based, exclusive)")),
		mcp.WithNumber("target_sheet_id", mcp.Required(), mcp.Description("Sheet ID where pivot table is placed")),
		mcp.WithNumber("target_row", mcp.Description("Target row for pivot table (default: 0)")),
		mcp.WithNumber("target_col", mcp.Description("Target column for pivot table (default: 0)")),
		mcp.WithArray("row_source_columns", mcp.Description("Array of column offsets (0-based) for row grouping")),
		mcp.WithArray("col_source_columns", mcp.Description("Array of column offsets (0-based) for column grouping")),
		mcp.WithArray("value_columns", mcp.Description("Array of {column, summarize_function} objects. summarize_function: SUM, COUNTA, COUNT, AVERAGE, MAX, MIN, CUSTOM")),
		common.WithAccountParam(),
	), HandleSheetsCreatePivotTable)

	// sheets_batch_update - Raw batchUpdate for power users
	s.AddTool(mcp.NewTool("sheets_batch_update",
		mcp.WithDescription("Execute raw batchUpdate requests on a spreadsheet. Power user escape hatch for advanced operations."),
		mcp.WithString("spreadsheet_id", mcp.Required(), mcp.Description("Spreadsheet ID or full Google Sheets URL")),
		mcp.WithString("requests", mcp.Required(), mcp.Description("JSON array of batch update requests (see Google Sheets API docs)")),
		common.WithAccountParam(),
	), HandleSheetsBatchUpdateSpreadsheet)
}
