package sheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/sheets/v4"
)

// === Testable Sheets Tool Handlers ===
// These functions accept SheetsHandlerDeps for dependency injection.

// TestableSheetsGet retrieves spreadsheet metadata.
func TestableSheetsGet(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	spreadsheet, err := srv.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	sheetsInfo := make([]map[string]any, 0, len(spreadsheet.Sheets))
	for _, sheet := range spreadsheet.Sheets {
		props := sheet.Properties
		sheetInfo := map[string]any{
			"sheet_id": props.SheetId,
			"title":    props.Title,
			"index":    props.Index,
		}
		if props.GridProperties != nil {
			sheetInfo["row_count"] = props.GridProperties.RowCount
			sheetInfo["column_count"] = props.GridProperties.ColumnCount
		}
		sheetsInfo = append(sheetsInfo, sheetInfo)
	}

	result := map[string]any{
		"spreadsheet_id": spreadsheet.SpreadsheetId,
		"title":          spreadsheet.Properties.Title,
		"locale":         spreadsheet.Properties.Locale,
		"time_zone":      spreadsheet.Properties.TimeZone,
		"sheets":         sheetsInfo,
		"sheets_count":   len(spreadsheet.Sheets),
		"url":            spreadsheet.SpreadsheetUrl,
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsRead reads cell values from a range.
func TestableSheetsRead(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	readRange, _ := request.Params.Arguments["range"].(string)
	if readRange == "" {
		return mcp.NewToolResultError("range parameter is required (A1 notation, e.g., 'Sheet1!A1:C10')"), nil
	}

	resp, err := srv.GetValues(ctx, spreadsheetID, readRange)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id": spreadsheetID,
		"range":          resp.Range,
		"values":         resp.Values,
		"row_count":      len(resp.Values),
	}

	if len(resp.Values) > 0 {
		maxCols := 0
		for _, row := range resp.Values {
			if len(row) > maxCols {
				maxCols = len(row)
			}
		}
		result["column_count"] = maxCols
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsWrite writes values to a cell range.
func TestableSheetsWrite(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	writeRange, _ := request.Params.Arguments["range"].(string)
	if writeRange == "" {
		return mcp.NewToolResultError("range parameter is required (A1 notation, e.g., 'Sheet1!A1:C3')"), nil
	}

	valuesRaw, ok := request.Params.Arguments["values"]
	if !ok {
		return mcp.NewToolResultError("values parameter is required (2D array of values)"), nil
	}

	values, err := parseValues(valuesRaw)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid values format: %v", err)), nil
	}

	valueInputOption := common.ValueInputUserEntered
	if opt, ok := request.Params.Arguments["value_input_option"].(string); ok && opt != "" {
		opt = strings.ToUpper(opt)
		if opt == common.ValueInputRaw || opt == common.ValueInputUserEntered {
			valueInputOption = opt
		}
	}

	resp, err := srv.UpdateValues(ctx, spreadsheetID, writeRange, values, valueInputOption)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id":  resp.SpreadsheetId,
		"updated_range":   resp.UpdatedRange,
		"updated_rows":    resp.UpdatedRows,
		"updated_columns": resp.UpdatedColumns,
		"updated_cells":   resp.UpdatedCells,
		"success":         true,
		"message":         fmt.Sprintf("Updated %d cells", resp.UpdatedCells),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsAppend appends rows to a sheet.
func TestableSheetsAppend(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	appendRange, _ := request.Params.Arguments["range"].(string)
	if appendRange == "" {
		return mcp.NewToolResultError("range parameter is required (A1 notation for table to append to, e.g., 'Sheet1!A:C')"), nil
	}

	valuesRaw, ok := request.Params.Arguments["values"]
	if !ok {
		return mcp.NewToolResultError("values parameter is required (2D array of values)"), nil
	}

	values, err := parseValues(valuesRaw)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid values format: %v", err)), nil
	}

	valueInputOption := common.ValueInputUserEntered
	if opt, ok := request.Params.Arguments["value_input_option"].(string); ok && opt != "" {
		opt = strings.ToUpper(opt)
		if opt == common.ValueInputRaw || opt == common.ValueInputUserEntered {
			valueInputOption = opt
		}
	}

	resp, err := srv.AppendValues(ctx, spreadsheetID, appendRange, values, valueInputOption)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id": resp.SpreadsheetId,
		"table_range":    resp.TableRange,
		"success":        true,
		"message":        fmt.Sprintf("Appended %d rows", len(values)),
	}

	if resp.Updates != nil {
		result["updated_range"] = resp.Updates.UpdatedRange
		result["updated_rows"] = resp.Updates.UpdatedRows
		result["updated_columns"] = resp.Updates.UpdatedColumns
		result["updated_cells"] = resp.Updates.UpdatedCells
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsCreate creates a new spreadsheet.
func TestableSheetsCreate(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title, _ := request.Params.Arguments["title"].(string)
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	created, err := srv.CreateSpreadsheet(ctx, title)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id": created.SpreadsheetId,
		"title":          created.Properties.Title,
		"url":            created.SpreadsheetUrl,
		"success":        true,
		"message":        "Spreadsheet created successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsBatchRead reads multiple ranges at once.
func TestableSheetsBatchRead(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	rangesRaw, ok := request.Params.Arguments["ranges"].([]any)
	if !ok || len(rangesRaw) == 0 {
		return mcp.NewToolResultError("ranges parameter is required (array of A1 notation ranges)"), nil
	}

	ranges := make([]string, 0, len(rangesRaw))
	for _, r := range rangesRaw {
		if rangeStr, ok := r.(string); ok && rangeStr != "" {
			ranges = append(ranges, rangeStr)
		}
	}

	if len(ranges) == 0 {
		return mcp.NewToolResultError("at least one valid range is required"), nil
	}

	resp, err := srv.BatchGetValues(ctx, spreadsheetID, ranges)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	valueRanges := make([]map[string]any, 0, len(resp.ValueRanges))
	for _, vr := range resp.ValueRanges {
		valueRanges = append(valueRanges, map[string]any{
			"range":     vr.Range,
			"values":    vr.Values,
			"row_count": len(vr.Values),
		})
	}

	result := map[string]any{
		"spreadsheet_id": resp.SpreadsheetId,
		"value_ranges":   valueRanges,
		"ranges_count":   len(valueRanges),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsBatchWrite writes to multiple ranges at once.
func TestableSheetsBatchWrite(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	dataRaw, ok := request.Params.Arguments["data"].([]any)
	if !ok || len(dataRaw) == 0 {
		return mcp.NewToolResultError("data parameter is required (array of {range, values} objects)"), nil
	}

	valueInputOption := common.ValueInputUserEntered
	if opt, ok := request.Params.Arguments["value_input_option"].(string); ok && opt != "" {
		opt = strings.ToUpper(opt)
		if opt == common.ValueInputRaw || opt == common.ValueInputUserEntered {
			valueInputOption = opt
		}
	}

	var data []*sheets.ValueRange
	for i, entry := range dataRaw {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("data[%d] must be an object with 'range' and 'values' fields", i)), nil
		}

		rangeStr, ok := entryMap["range"].(string)
		if !ok || rangeStr == "" {
			return mcp.NewToolResultError(fmt.Sprintf("data[%d].range is required", i)), nil
		}

		values, err := parseValues(entryMap["values"])
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("data[%d].values: %v", i, err)), nil
		}

		data = append(data, &sheets.ValueRange{
			Range:  rangeStr,
			Values: values,
		})
	}

	resp, err := srv.BatchUpdateValues(ctx, spreadsheetID, data, valueInputOption)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id":        resp.SpreadsheetId,
		"total_updated_rows":    resp.TotalUpdatedRows,
		"total_updated_columns": resp.TotalUpdatedColumns,
		"total_updated_cells":   resp.TotalUpdatedCells,
		"total_updated_sheets":  resp.TotalUpdatedSheets,
		"success":               true,
		"message":               fmt.Sprintf("Updated %d cells across %d ranges", resp.TotalUpdatedCells, len(data)),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsClear clears values from a cell range.
func TestableSheetsClear(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, _ := request.Params.Arguments["spreadsheet_id"].(string)
	if spreadsheetID == "" {
		return mcp.NewToolResultError("spreadsheet_id parameter is required"), nil
	}
	spreadsheetID = common.ExtractSpreadsheetID(spreadsheetID)

	clearRange, _ := request.Params.Arguments["range"].(string)
	if clearRange == "" {
		return mcp.NewToolResultError("range parameter is required (A1 notation, e.g., 'Sheet1!A1:C10')"), nil
	}

	resp, err := srv.ClearValues(ctx, spreadsheetID, clearRange)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"spreadsheet_id": resp.SpreadsheetId,
		"cleared_range":  resp.ClearedRange,
		"success":        true,
		"message":        fmt.Sprintf("Cleared range %s", resp.ClearedRange),
	}

	return common.MarshalToolResult(result)
}
