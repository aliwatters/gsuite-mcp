package sheets

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// === SheetsFormatCells Tests ===

func TestTestableSheetsFormatCells_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":   "test-sheet-1",
		"sheet_id":         float64(0),
		"start_row":        float64(0),
		"start_col":        float64(0),
		"end_row":          float64(5),
		"end_col":          float64(3),
		"bold":             true,
		"font_size":        float64(14),
		"background_color": "#FF0000",
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
	if data["spreadsheet_id"] != "test-sheet-1" {
		t.Errorf("expected spreadsheet_id=test-sheet-1, got %v", data["spreadsheet_id"])
	}
}

func TestTestableSheetsFormatCells_MissingSpreadsheetID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"sheet_id":  float64(0),
		"start_row": float64(0),
		"start_col": float64(0),
		"end_row":   float64(5),
		"end_col":   float64(3),
		"bold":      true,
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing spreadsheet_id")
	}
}

func TestTestableSheetsFormatCells_MissingRange(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"bold":           true,
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range params")
	}
}

func TestTestableSheetsFormatCells_NoFormattingOptions(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(5),
		"end_col":        float64(3),
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result when no formatting options specified")
	}
}

func TestTestableSheetsFormatCells_InvalidColor(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":   "test-sheet-1",
		"sheet_id":         float64(0),
		"start_row":        float64(0),
		"start_col":        float64(0),
		"end_row":          float64(5),
		"end_col":          float64(3),
		"background_color": "ZZZZZZ",
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for invalid color")
	}
}

func TestTestableSheetsFormatCells_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.BatchUpdate = errors.New("permission denied")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(5),
		"end_col":        float64(3),
		"bold":           true,
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

func TestTestableSheetsFormatCells_NumberFormat(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":     "test-sheet-1",
		"sheet_id":           float64(0),
		"start_row":          float64(0),
		"start_col":          float64(0),
		"end_row":            float64(10),
		"end_col":            float64(1),
		"number_format":      "#,##0.00",
		"number_format_type": "NUMBER",
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}
}

func TestTestableSheetsFormatCells_Alignment(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":       "test-sheet-1",
		"sheet_id":             float64(0),
		"start_row":            float64(0),
		"start_col":            float64(0),
		"end_row":              float64(1),
		"end_col":              float64(5),
		"horizontal_alignment": "CENTER",
		"vertical_alignment":   "MIDDLE",
	})

	result, err := TestableSheetsFormatCells(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}
}

// === SheetsAddConditionalFormat Tests ===

func TestTestableSheetsAddConditionalFormat_BooleanSuccess(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":          "test-sheet-1",
		"sheet_id":                float64(0),
		"start_row":               float64(0),
		"start_col":               float64(0),
		"end_row":                 float64(10),
		"end_col":                 float64(3),
		"rule_type":               "BOOLEAN",
		"condition_type":          "NUMBER_GREATER",
		"condition_values":        []any{"100"},
		"format_background_color": "#00FF00",
		"format_bold":             true,
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}

	data := parseResult(t, result)
	if data["rule_type"] != "BOOLEAN" {
		t.Errorf("expected rule_type=BOOLEAN, got %v", data["rule_type"])
	}
}

func TestTestableSheetsAddConditionalFormat_GradientSuccess(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
		"rule_type":      "GRADIENT",
		"min_color":      "#FF0000",
		"mid_color":      "#FFFF00",
		"max_color":      "#00FF00",
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}

	data := parseResult(t, result)
	if data["rule_type"] != "GRADIENT" {
		t.Errorf("expected rule_type=GRADIENT, got %v", data["rule_type"])
	}
}

func TestTestableSheetsAddConditionalFormat_MissingRuleType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing rule_type")
	}
}

func TestTestableSheetsAddConditionalFormat_InvalidRuleType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
		"rule_type":      "INVALID",
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for invalid rule_type")
	}
}

func TestTestableSheetsAddConditionalFormat_BooleanMissingConditionType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
		"rule_type":      "BOOLEAN",
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing condition_type")
	}
}

func TestTestableSheetsAddConditionalFormat_GradientMissingColors(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
		"rule_type":      "GRADIENT",
		"min_color":      "#FF0000",
	})

	result, err := TestableSheetsAddConditionalFormat(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing max_color")
	}
}

// === SheetsAddDataValidation Tests ===

func TestTestableSheetsAddDataValidation_DropdownSuccess(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":  "test-sheet-1",
		"sheet_id":        float64(0),
		"start_row":       float64(0),
		"start_col":       float64(0),
		"end_row":         float64(10),
		"end_col":         float64(1),
		"validation_type": "ONE_OF_LIST",
		"values":          []any{"Option A", "Option B", "Option C"},
		"show_dropdown":   true,
	})

	result, err := TestableSheetsAddDataValidation(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}

	data := parseResult(t, result)
	if data["validation_type"] != "ONE_OF_LIST" {
		t.Errorf("expected validation_type=ONE_OF_LIST, got %v", data["validation_type"])
	}
}

func TestTestableSheetsAddDataValidation_MissingType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(1),
	})

	result, err := TestableSheetsAddDataValidation(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing validation_type")
	}
}

func TestTestableSheetsAddDataValidation_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.BatchUpdate = errors.New("quota exceeded")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":  "test-sheet-1",
		"sheet_id":        float64(0),
		"start_row":       float64(0),
		"start_col":       float64(0),
		"end_row":         float64(10),
		"end_col":         float64(1),
		"validation_type": "ONE_OF_LIST",
		"values":          []any{"A", "B"},
	})

	result, err := TestableSheetsAddDataValidation(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === SheetsBatchUpdateSpreadsheet Tests ===

func TestTestableSheetsBatchUpdateSpreadsheet_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"requests":       `[{"updateSheetProperties":{"properties":{"sheetId":0,"title":"Renamed"},"fields":"title"}}]`,
	})

	result, err := TestableSheetsBatchUpdateSpreadsheet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatalf("expected success, got error: %s", getTextContent(result))
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
}

func TestTestableSheetsBatchUpdateSpreadsheet_MissingRequests(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsBatchUpdateSpreadsheet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing requests")
	}
}

func TestTestableSheetsBatchUpdateSpreadsheet_InvalidJSON(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"requests":       "not valid json",
	})

	result, err := TestableSheetsBatchUpdateSpreadsheet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for invalid JSON")
	}
}

func TestTestableSheetsBatchUpdateSpreadsheet_EmptyArray(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"requests":       "[]",
	})

	result, err := TestableSheetsBatchUpdateSpreadsheet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for empty requests array")
	}
}

// getTextContent extracts text content from an MCP result for error messages.
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}
