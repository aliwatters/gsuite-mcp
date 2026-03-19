package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/sheets/v4"
)

// sheetsEditURLFormat is the URL template for Google Sheets edit links.
const sheetsEditURLFormat = "https://docs.google.com/spreadsheets/d/%s/edit"

// parseCellRange parses the sheet_id, start_row, start_col, end_row, end_col parameters
// into a sheets.GridRange. Returns the range and nil on success, or nil and an error result.
func parseCellRange(args map[string]any) (*sheets.GridRange, *mcp.CallToolResult) {
	sheetIDFloat, ok := args["sheet_id"].(float64)
	if !ok {
		return nil, mcp.NewToolResultError("sheet_id parameter is required (numeric sheet ID, found in sheets_get response)")
	}
	sheetID := int64(sheetIDFloat)

	startRowFloat, ok := args["start_row"].(float64)
	if !ok {
		return nil, mcp.NewToolResultError("start_row parameter is required (0-based row index)")
	}
	startRow := int64(startRowFloat)

	startColFloat, ok := args["start_col"].(float64)
	if !ok {
		return nil, mcp.NewToolResultError("start_col parameter is required (0-based column index)")
	}
	startCol := int64(startColFloat)

	endRowFloat, ok := args["end_row"].(float64)
	if !ok {
		return nil, mcp.NewToolResultError("end_row parameter is required (0-based row index, exclusive)")
	}
	endRow := int64(endRowFloat)

	endColFloat, ok := args["end_col"].(float64)
	if !ok {
		return nil, mcp.NewToolResultError("end_col parameter is required (0-based column index, exclusive)")
	}
	endCol := int64(endColFloat)

	if endRow <= startRow {
		return nil, mcp.NewToolResultError("end_row must be greater than start_row")
	}
	if endCol <= startCol {
		return nil, mcp.NewToolResultError("end_col must be greater than start_col")
	}

	return &sheets.GridRange{
		SheetId:          sheetID,
		StartRowIndex:    startRow,
		StartColumnIndex: startCol,
		EndRowIndex:      endRow,
		EndColumnIndex:   endCol,
	}, nil
}

// parseHexColor parses a hex color string into a sheets.Color.
func parseHexColor(colorStr string) (*sheets.Color, error) {
	colorStr = strings.TrimPrefix(colorStr, "#")
	if len(colorStr) != 6 {
		return nil, fmt.Errorf("invalid color format: expected 6 hex characters, got %d", len(colorStr))
	}

	r, err := parseHexByte(colorStr[0:2])
	if err != nil {
		return nil, fmt.Errorf("invalid red component: %w", err)
	}
	g, err := parseHexByte(colorStr[2:4])
	if err != nil {
		return nil, fmt.Errorf("invalid green component: %w", err)
	}
	b, err := parseHexByte(colorStr[4:6])
	if err != nil {
		return nil, fmt.Errorf("invalid blue component: %w", err)
	}

	return &sheets.Color{
		Red:   float64(r) / 255.0,
		Green: float64(g) / 255.0,
		Blue:  float64(b) / 255.0,
	}, nil
}

// parseHexByte parses a 2-character hex string into a byte value.
func parseHexByte(s string) (byte, error) {
	var val byte
	for i := 0; i < 2; i++ {
		val <<= 4
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			val |= c - '0'
		case c >= 'A' && c <= 'F':
			val |= c - 'A' + 10
		case c >= 'a' && c <= 'f':
			val |= c - 'a' + 10
		default:
			return 0, fmt.Errorf("invalid hex character: %c", c)
		}
	}
	return val, nil
}

// TestableSheetsFormatCells applies formatting to a cell range.
func TestableSheetsFormatCells(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	gridRange, rangeErr := parseCellRange(request.Params.Arguments)
	if rangeErr != nil {
		return rangeErr, nil
	}

	// Build CellFormat from arguments
	cellFormat := &sheets.CellFormat{}
	var fields []string

	// Background color
	if bgColor, ok := request.Params.Arguments["background_color"].(string); ok && bgColor != "" {
		color, err := parseHexColor(bgColor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid background_color: %v", err)), nil
		}
		cellFormat.BackgroundColor = color
		fields = append(fields, "userEnteredFormat.backgroundColor")
	}

	// Text format
	textFormat := &sheets.TextFormat{}
	hasTextFormat := false

	if bold, ok := request.Params.Arguments["bold"].(bool); ok {
		textFormat.Bold = bold
		if !bold {
			textFormat.ForceSendFields = append(textFormat.ForceSendFields, "Bold")
		}
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.bold")
	}
	if italic, ok := request.Params.Arguments["italic"].(bool); ok {
		textFormat.Italic = italic
		if !italic {
			textFormat.ForceSendFields = append(textFormat.ForceSendFields, "Italic")
		}
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.italic")
	}
	if underline, ok := request.Params.Arguments["underline"].(bool); ok {
		textFormat.Underline = underline
		if !underline {
			textFormat.ForceSendFields = append(textFormat.ForceSendFields, "Underline")
		}
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.underline")
	}
	if strikethrough, ok := request.Params.Arguments["strikethrough"].(bool); ok {
		textFormat.Strikethrough = strikethrough
		if !strikethrough {
			textFormat.ForceSendFields = append(textFormat.ForceSendFields, "Strikethrough")
		}
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.strikethrough")
	}
	if fontFamily, ok := request.Params.Arguments["font_family"].(string); ok && fontFamily != "" {
		textFormat.FontFamily = fontFamily
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.fontFamily")
	}
	if fontSize, ok := request.Params.Arguments["font_size"].(float64); ok && fontSize > 0 {
		textFormat.FontSize = int64(fontSize)
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.fontSize")
	}
	if fgColor, ok := request.Params.Arguments["foreground_color"].(string); ok && fgColor != "" {
		color, err := parseHexColor(fgColor)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid foreground_color: %v", err)), nil
		}
		textFormat.ForegroundColor = color
		hasTextFormat = true
		fields = append(fields, "userEnteredFormat.textFormat.foregroundColor")
	}

	if hasTextFormat {
		cellFormat.TextFormat = textFormat
	}

	// Number format
	if numberFormat, ok := request.Params.Arguments["number_format"].(string); ok && numberFormat != "" {
		numberFormatType := common.ParseStringArg(request.Params.Arguments, "number_format_type", "NUMBER")
		validTypes := map[string]bool{
			"TEXT": true, "NUMBER": true, "PERCENT": true, "CURRENCY": true,
			"DATE": true, "TIME": true, "DATE_TIME": true, "SCIENTIFIC": true,
		}
		if !validTypes[numberFormatType] {
			return mcp.NewToolResultError("number_format_type must be one of: TEXT, NUMBER, PERCENT, CURRENCY, DATE, TIME, DATE_TIME, SCIENTIFIC"), nil
		}
		cellFormat.NumberFormat = &sheets.NumberFormat{
			Type:    numberFormatType,
			Pattern: numberFormat,
		}
		fields = append(fields, "userEnteredFormat.numberFormat")
	}

	// Horizontal alignment
	if hAlign, ok := request.Params.Arguments["horizontal_alignment"].(string); ok && hAlign != "" {
		validAlignments := map[string]bool{"LEFT": true, "CENTER": true, "RIGHT": true}
		if !validAlignments[hAlign] {
			return mcp.NewToolResultError("horizontal_alignment must be one of: LEFT, CENTER, RIGHT"), nil
		}
		cellFormat.HorizontalAlignment = hAlign
		fields = append(fields, "userEnteredFormat.horizontalAlignment")
	}

	// Vertical alignment
	if vAlign, ok := request.Params.Arguments["vertical_alignment"].(string); ok && vAlign != "" {
		validAlignments := map[string]bool{"TOP": true, "MIDDLE": true, "BOTTOM": true}
		if !validAlignments[vAlign] {
			return mcp.NewToolResultError("vertical_alignment must be one of: TOP, MIDDLE, BOTTOM"), nil
		}
		cellFormat.VerticalAlignment = vAlign
		fields = append(fields, "userEnteredFormat.verticalAlignment")
	}

	// Wrap strategy
	if wrapStrategy, ok := request.Params.Arguments["wrap_strategy"].(string); ok && wrapStrategy != "" {
		validStrategies := map[string]bool{"OVERFLOW_CELL": true, "LEGACY_WRAP": true, "CLIP": true, "WRAP": true}
		if !validStrategies[wrapStrategy] {
			return mcp.NewToolResultError("wrap_strategy must be one of: OVERFLOW_CELL, LEGACY_WRAP, CLIP, WRAP"), nil
		}
		cellFormat.WrapStrategy = wrapStrategy
		fields = append(fields, "userEnteredFormat.wrapStrategy")
	}

	if len(fields) == 0 {
		return mcp.NewToolResultError("at least one formatting option must be specified"), nil
	}

	requests := []*sheets.Request{{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: gridRange,
			Cell: &sheets.CellData{
				UserEnteredFormat: cellFormat,
			},
			Fields: strings.Join(fields, ","),
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"fields_updated": fields,
		"message":        fmt.Sprintf("Applied formatting to range (sheet %d, rows %d-%d, cols %d-%d)", gridRange.SheetId, gridRange.StartRowIndex, gridRange.EndRowIndex, gridRange.StartColumnIndex, gridRange.EndColumnIndex),
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsAddConditionalFormat adds a conditional formatting rule.
func TestableSheetsAddConditionalFormat(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	gridRange, rangeErr := parseCellRange(request.Params.Arguments)
	if rangeErr != nil {
		return rangeErr, nil
	}

	ruleType := common.ParseStringArg(request.Params.Arguments, "rule_type", "")
	if ruleType == "" {
		return mcp.NewToolResultError("rule_type parameter is required (BOOLEAN or GRADIENT)"), nil
	}

	var rule *sheets.ConditionalFormatRule

	switch ruleType {
	case "BOOLEAN":
		conditionType := common.ParseStringArg(request.Params.Arguments, "condition_type", "")
		if conditionType == "" {
			return mcp.NewToolResultError("condition_type is required for BOOLEAN rules (e.g., NUMBER_GREATER, TEXT_CONTAINS, CUSTOM_FORMULA)"), nil
		}

		// Parse condition values
		var conditionValues []*sheets.ConditionValue
		if valuesRaw, ok := request.Params.Arguments["condition_values"].([]any); ok {
			for _, v := range valuesRaw {
				if s, ok := v.(string); ok {
					conditionValues = append(conditionValues, &sheets.ConditionValue{UserEnteredValue: s})
				}
			}
		}

		boolRule := &sheets.BooleanRule{
			Condition: &sheets.BooleanCondition{
				Type:   conditionType,
				Values: conditionValues,
			},
		}

		// Parse format for matching cells
		format := &sheets.CellFormat{}
		if bgColor, ok := request.Params.Arguments["format_background_color"].(string); ok && bgColor != "" {
			color, err := parseHexColor(bgColor)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid format_background_color: %v", err)), nil
			}
			format.BackgroundColor = color
		}
		if fgColor, ok := request.Params.Arguments["format_text_color"].(string); ok && fgColor != "" {
			color, err := parseHexColor(fgColor)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid format_text_color: %v", err)), nil
			}
			format.TextFormat = &sheets.TextFormat{ForegroundColor: color}
		}
		if bold, ok := request.Params.Arguments["format_bold"].(bool); ok && bold {
			if format.TextFormat == nil {
				format.TextFormat = &sheets.TextFormat{}
			}
			format.TextFormat.Bold = true
		}
		if italic, ok := request.Params.Arguments["format_italic"].(bool); ok && italic {
			if format.TextFormat == nil {
				format.TextFormat = &sheets.TextFormat{}
			}
			format.TextFormat.Italic = true
		}

		boolRule.Format = format
		rule = &sheets.ConditionalFormatRule{
			Ranges:      []*sheets.GridRange{gridRange},
			BooleanRule: boolRule,
		}

	case "GRADIENT":
		gradientRule := &sheets.GradientRule{}

		if minColor, ok := request.Params.Arguments["min_color"].(string); ok && minColor != "" {
			color, err := parseHexColor(minColor)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid min_color: %v", err)), nil
			}
			minType := common.ParseStringArg(request.Params.Arguments, "min_type", "MIN")
			gradientRule.Minpoint = &sheets.InterpolationPoint{
				Color: color,
				Type:  minType,
			}
			if minValue, ok := request.Params.Arguments["min_value"].(string); ok && minValue != "" {
				gradientRule.Minpoint.Value = minValue
			}
		}

		if midColor, ok := request.Params.Arguments["mid_color"].(string); ok && midColor != "" {
			color, err := parseHexColor(midColor)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid mid_color: %v", err)), nil
			}
			midType := common.ParseStringArg(request.Params.Arguments, "mid_type", "PERCENTILE")
			gradientRule.Midpoint = &sheets.InterpolationPoint{
				Color: color,
				Type:  midType,
			}
			if midValue, ok := request.Params.Arguments["mid_value"].(string); ok && midValue != "" {
				gradientRule.Midpoint.Value = midValue
			}
		}

		if maxColor, ok := request.Params.Arguments["max_color"].(string); ok && maxColor != "" {
			color, err := parseHexColor(maxColor)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid max_color: %v", err)), nil
			}
			maxType := common.ParseStringArg(request.Params.Arguments, "max_type", "MAX")
			gradientRule.Maxpoint = &sheets.InterpolationPoint{
				Color: color,
				Type:  maxType,
			}
			if maxValue, ok := request.Params.Arguments["max_value"].(string); ok && maxValue != "" {
				gradientRule.Maxpoint.Value = maxValue
			}
		}

		if gradientRule.Minpoint == nil || gradientRule.Maxpoint == nil {
			return mcp.NewToolResultError("GRADIENT rules require at least min_color and max_color"), nil
		}

		rule = &sheets.ConditionalFormatRule{
			Ranges:       []*sheets.GridRange{gridRange},
			GradientRule: gradientRule,
		}

	default:
		return mcp.NewToolResultError("rule_type must be BOOLEAN or GRADIENT"), nil
	}

	requests := []*sheets.Request{{
		AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
			Rule:  rule,
			Index: 0,
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"rule_type":      ruleType,
		"message":        "Conditional formatting rule added",
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsAddDataValidation adds data validation rules to a range.
func TestableSheetsAddDataValidation(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	gridRange, rangeErr := parseCellRange(request.Params.Arguments)
	if rangeErr != nil {
		return rangeErr, nil
	}

	validationType := common.ParseStringArg(request.Params.Arguments, "validation_type", "")
	if validationType == "" {
		return mcp.NewToolResultError("validation_type parameter is required (e.g., ONE_OF_LIST, NUMBER_BETWEEN, CUSTOM_FORMULA)"), nil
	}

	// Parse condition values
	var conditionValues []*sheets.ConditionValue
	if valuesRaw, ok := request.Params.Arguments["values"].([]any); ok {
		for _, v := range valuesRaw {
			if s, ok := v.(string); ok {
				conditionValues = append(conditionValues, &sheets.ConditionValue{UserEnteredValue: s})
			}
		}
	}

	strict := common.ParseBoolArg(request.Params.Arguments, "strict", true)
	showDropdown := common.ParseBoolArg(request.Params.Arguments, "show_dropdown", true)
	inputMessage := common.ParseStringArg(request.Params.Arguments, "input_message", "")

	dataValidation := &sheets.DataValidationRule{
		Condition: &sheets.BooleanCondition{
			Type:   validationType,
			Values: conditionValues,
		},
		Strict:          strict,
		ShowCustomUi:    showDropdown,
		InputMessage:    inputMessage,
		ForceSendFields: []string{"Strict", "ShowCustomUi"},
	}

	requests := []*sheets.Request{{
		SetDataValidation: &sheets.SetDataValidationRequest{
			Range: gridRange,
			Rule:  dataValidation,
		},
	}}

	_, err := srv.BatchUpdate(ctx, spreadsheetID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	result := map[string]any{
		"success":         true,
		"spreadsheet_id":  spreadsheetID,
		"validation_type": validationType,
		"message":         "Data validation rule applied",
		"url":             fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}

// TestableSheetsBatchUpdateSpreadsheet executes raw batchUpdate requests for power users.
func TestableSheetsBatchUpdateSpreadsheet(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSheetsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spreadsheetID, idErrResult := extractRequiredSpreadsheetID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	requestsJSON := common.ParseStringArg(request.Params.Arguments, "requests", "")
	if requestsJSON == "" {
		return mcp.NewToolResultError("requests parameter is required (JSON array of batch update requests)"), nil
	}

	var sheetsRequests []*sheets.Request
	if err := json.Unmarshal([]byte(requestsJSON), &sheetsRequests); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse requests JSON: %v", err)), nil
	}

	if len(sheetsRequests) == 0 {
		return mcp.NewToolResultError("requests array cannot be empty"), nil
	}

	resp, err := srv.BatchUpdate(ctx, spreadsheetID, sheetsRequests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Sheets API error: %v", err)), nil
	}

	repliesCount := 0
	if resp != nil && resp.Replies != nil {
		repliesCount = len(resp.Replies)
	}

	result := map[string]any{
		"success":        true,
		"spreadsheet_id": spreadsheetID,
		"requests_count": len(sheetsRequests),
		"replies_count":  repliesCount,
		"message":        fmt.Sprintf("Successfully executed %d batch update request(s)", len(sheetsRequests)),
		"url":            fmt.Sprintf(sheetsEditURLFormat, spreadsheetID),
	}

	return common.MarshalToolResult(result)
}
