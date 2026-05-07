package sheets

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/sheets/v4"
)

// validChartTypes is the set of supported chart types for create/update operations.
var validChartTypes = map[string]bool{
	"BAR": true, "LINE": true, "AREA": true, "COLUMN": true,
	"SCATTER": true, "COMBO": true, "STEPPED_AREA": true, "PIE": true,
}

// TestableSheetsCreateChart creates an embedded chart in a spreadsheet.
func TestableSheetsCreateChart(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	chartType := common.ParseStringArg(request.GetArguments(), "chart_type", "")
	if chartType == "" {
		return mcp.NewToolResultError("chart_type parameter is required (BAR, LINE, AREA, COLUMN, SCATTER, COMBO, STEPPED_AREA, PIE)"), nil
	}
	if !validChartTypes[chartType] {
		return mcp.NewToolResultError("chart_type must be one of: BAR, LINE, AREA, COLUMN, SCATTER, COMBO, STEPPED_AREA, PIE"), nil
	}

	title := common.ParseStringArg(request.GetArguments(), "title", "")

	// Parse source data range
	sourceRange, rangeErr := parseCellRange(request.GetArguments())
	if rangeErr != nil {
		return rangeErr, nil
	}

	// Build BasicChart spec
	basicChart := &sheets.BasicChartSpec{
		ChartType: chartType,
		Domains: []*sheets.BasicChartDomain{{
			Domain: &sheets.ChartData{
				SourceRange: &sheets.ChartSourceRange{
					Sources: []*sheets.GridRange{sourceRange},
				},
			},
		}},
	}

	// Optionally set legend position
	if legendPos, ok := request.GetArguments()["legend_position"].(string); ok && legendPos != "" {
		basicChart.LegendPosition = legendPos
	}

	chartSpec := &sheets.ChartSpec{
		Title:      title,
		BasicChart: basicChart,
	}

	// Overlay position (default to anchoring at row 0, col 0 of the same sheet)
	anchorRow := int64(0)
	anchorCol := int64(0)
	if ar, ok := request.GetArguments()["anchor_row"].(float64); ok {
		anchorRow = int64(ar)
	}
	if ac, ok := request.GetArguments()["anchor_col"].(float64); ok {
		anchorCol = int64(ac)
	}

	requests := []*sheets.Request{{
		AddChart: &sheets.AddChartRequest{
			Chart: &sheets.EmbeddedChart{
				Spec: chartSpec,
				Position: &sheets.EmbeddedObjectPosition{
					OverlayPosition: &sheets.OverlayPosition{
						AnchorCell: &sheets.GridCoordinate{
							SheetId:     sourceRange.SheetId,
							RowIndex:    anchorRow,
							ColumnIndex: anchorCol,
						},
					},
				},
			},
		},
	}}

	resp, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	var chartID int64
	if resp != nil && resp.Replies != nil && len(resp.Replies) > 0 && resp.Replies[0].AddChart != nil {
		chartID = resp.Replies[0].AddChart.Chart.ChartId
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"chart_id":       chartID,
		"chart_type":     chartType,
		"title":          title,
		"message":        fmt.Sprintf("Created %s chart", chartType),
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsUpdateChart updates an existing embedded chart.
func TestableSheetsUpdateChart(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	chartIDFloat, ok := request.GetArguments()["chart_id"].(float64)
	if !ok {
		return mcp.NewToolResultError("chart_id parameter is required"), nil
	}
	chartID := int64(chartIDFloat)

	chartSpec := &sheets.ChartSpec{}
	var fields []string

	if title, ok := request.GetArguments()["title"].(string); ok {
		chartSpec.Title = title
		fields = append(fields, "title")
	}

	if chartType, ok := request.GetArguments()["chart_type"].(string); ok && chartType != "" {
		if !validChartTypes[chartType] {
			return mcp.NewToolResultError("chart_type must be one of: BAR, LINE, AREA, COLUMN, SCATTER, COMBO, STEPPED_AREA, PIE"), nil
		}
		chartSpec.BasicChart = &sheets.BasicChartSpec{
			ChartType: chartType,
		}
		fields = append(fields, "basicChart.chartType")
	}

	if len(fields) == 0 {
		return mcp.NewToolResultError("at least one update field must be specified (title, chart_type)"), nil
	}

	requests := []*sheets.Request{{
		UpdateChartSpec: &sheets.UpdateChartSpecRequest{
			ChartId: chartID,
			Spec:    chartSpec,
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"chart_id":       chartID,
		"fields_updated": fields,
		"message":        "Chart updated",
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsDeleteChart deletes an embedded chart.
func TestableSheetsDeleteChart(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	chartIDFloat, ok := request.GetArguments()["chart_id"].(float64)
	if !ok {
		return mcp.NewToolResultError("chart_id parameter is required"), nil
	}
	chartID := int64(chartIDFloat)

	requests := []*sheets.Request{{
		DeleteEmbeddedObject: &sheets.DeleteEmbeddedObjectRequest{
			ObjectId: chartID,
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"chart_id":       chartID,
		"message":        "Chart deleted",
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsCreatePivotTable creates a pivot table in a spreadsheet.
func TestableSheetsCreatePivotTable(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	// Source data range (uses "source_sheet_id" instead of "sheet_id")
	sourceRange, rangeErr := parseGridRange(request.GetArguments(), "source_sheet_id")
	if rangeErr != nil {
		return rangeErr, nil
	}

	// Target location
	targetSheetIDFloat, ok := request.GetArguments()["target_sheet_id"].(float64)
	if !ok {
		return mcp.NewToolResultError("target_sheet_id parameter is required (sheet ID where pivot table is placed)"), nil
	}
	targetSheetID := int64(targetSheetIDFloat)

	targetRow := int64(0)
	targetCol := int64(0)
	if tr, ok := request.GetArguments()["target_row"].(float64); ok {
		targetRow = int64(tr)
	}
	if tc, ok := request.GetArguments()["target_col"].(float64); ok {
		targetCol = int64(tc)
	}

	// Build pivot table
	pivotTable := &sheets.PivotTable{
		Source: sourceRange,
	}

	// Row grouping columns (array of column indices)
	if rowsRaw, ok := request.GetArguments()["row_source_columns"].([]any); ok {
		for _, r := range rowsRaw {
			if colIdx, ok := r.(float64); ok {
				pivotTable.Rows = append(pivotTable.Rows, &sheets.PivotGroup{
					SourceColumnOffset: int64(colIdx),
					SortOrder:          "ASCENDING",
				})
			}
		}
	}

	// Column grouping columns
	if colsRaw, ok := request.GetArguments()["col_source_columns"].([]any); ok {
		for _, c := range colsRaw {
			if colIdx, ok := c.(float64); ok {
				pivotTable.Columns = append(pivotTable.Columns, &sheets.PivotGroup{
					SourceColumnOffset: int64(colIdx),
					SortOrder:          "ASCENDING",
				})
			}
		}
	}

	// Values (aggregation)
	if valuesRaw, ok := request.GetArguments()["value_columns"].([]any); ok {
		for _, v := range valuesRaw {
			if valMap, ok := v.(map[string]any); ok {
				colIdx, colOk := valMap["column"].(float64)
				summarizeFunc := "SUM"
				if sf, ok := valMap["summarize_function"].(string); ok && sf != "" {
					summarizeFunc = sf
				}
				if colOk {
					pivotTable.Values = append(pivotTable.Values, &sheets.PivotValue{
						SourceColumnOffset: int64(colIdx),
						SummarizeFunction:  summarizeFunc,
					})
				}
			}
		}
	}

	if len(pivotTable.Rows) == 0 && len(pivotTable.Columns) == 0 {
		return mcp.NewToolResultError("at least one of row_source_columns or col_source_columns must be specified"), nil
	}

	// Create the pivot table by updating the target cell
	requests := []*sheets.Request{{
		UpdateCells: &sheets.UpdateCellsRequest{
			Start: &sheets.GridCoordinate{
				SheetId:     targetSheetID,
				RowIndex:    targetRow,
				ColumnIndex: targetCol,
			},
			Rows: []*sheets.RowData{{
				Values: []*sheets.CellData{{
					PivotTable: pivotTable,
				}},
			}},
			Fields: "pivotTable",
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":         true,
		"spreadsheet_id":  spreadsheetID,
		"target_sheet_id": targetSheetID,
		"target_row":      targetRow,
		"target_col":      targetCol,
		"message":         "Pivot table created",
		"url":             fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}
