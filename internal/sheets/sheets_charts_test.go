package sheets

import (
	"context"
	"errors"
	"testing"
)

// === SheetsCreateChart Tests ===

func TestTestableSheetsCreateChart_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_type":     "BAR",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(3),
		"title":          "Sales Chart",
	})

	result, err := TestableSheetsCreateChart(ctx, request, fixtures.Deps)
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
	if data["chart_type"] != "BAR" {
		t.Errorf("expected chart_type=BAR, got %v", data["chart_type"])
	}
}

func TestTestableSheetsCreateChart_MissingChartType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(3),
	})

	result, err := TestableSheetsCreateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing chart_type")
	}
}

func TestTestableSheetsCreateChart_InvalidChartType(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_type":     "INVALID",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(10),
		"end_col":        float64(3),
	})

	result, err := TestableSheetsCreateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for invalid chart_type")
	}
}

func TestTestableSheetsCreateChart_MissingRange(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_type":     "LINE",
	})

	result, err := TestableSheetsCreateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range params")
	}
}

func TestTestableSheetsCreateChart_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.BatchUpdate = errors.New("quota exceeded")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_type":     "PIE",
		"sheet_id":       float64(0),
		"start_row":      float64(0),
		"start_col":      float64(0),
		"end_row":        float64(5),
		"end_col":        float64(2),
	})

	result, err := TestableSheetsCreateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === SheetsUpdateChart Tests ===

func TestTestableSheetsUpdateChart_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_id":       float64(123),
		"title":          "Updated Title",
		"chart_type":     "LINE",
	})

	result, err := TestableSheetsUpdateChart(ctx, request, fixtures.Deps)
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

func TestTestableSheetsUpdateChart_MissingChartID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"title":          "New Title",
	})

	result, err := TestableSheetsUpdateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing chart_id")
	}
}

func TestTestableSheetsUpdateChart_NoUpdateFields(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_id":       float64(123),
	})

	result, err := TestableSheetsUpdateChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for no update fields")
	}
}

// === SheetsDeleteChart Tests ===

func TestTestableSheetsDeleteChart_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_id":       float64(123),
	})

	result, err := TestableSheetsDeleteChart(ctx, request, fixtures.Deps)
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

func TestTestableSheetsDeleteChart_MissingChartID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsDeleteChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing chart_id")
	}
}

func TestTestableSheetsDeleteChart_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.BatchUpdate = errors.New("not found")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"chart_id":       float64(999),
	})

	result, err := TestableSheetsDeleteChart(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === SheetsCreatePivotTable Tests ===

func TestTestableSheetsCreatePivotTable_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":     "test-sheet-1",
		"source_sheet_id":    float64(0),
		"start_row":          float64(0),
		"start_col":          float64(0),
		"end_row":            float64(100),
		"end_col":            float64(5),
		"target_sheet_id":    float64(0),
		"target_row":         float64(0),
		"target_col":         float64(6),
		"row_source_columns": []any{float64(0)},
		"value_columns": []any{
			map[string]any{"column": float64(2), "summarize_function": "SUM"},
		},
	})

	result, err := TestableSheetsCreatePivotTable(ctx, request, fixtures.Deps)
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

func TestTestableSheetsCreatePivotTable_MissingSourceSheetID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":  "test-sheet-1",
		"start_row":       float64(0),
		"start_col":       float64(0),
		"end_row":         float64(10),
		"end_col":         float64(3),
		"target_sheet_id": float64(0),
	})

	result, err := TestableSheetsCreatePivotTable(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing source_sheet_id")
	}
}

func TestTestableSheetsCreatePivotTable_NoGroupingColumns(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":  "test-sheet-1",
		"source_sheet_id": float64(0),
		"start_row":       float64(0),
		"start_col":       float64(0),
		"end_row":         float64(10),
		"end_col":         float64(3),
		"target_sheet_id": float64(0),
	})

	result, err := TestableSheetsCreatePivotTable(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for no grouping columns")
	}
}

func TestTestableSheetsCreatePivotTable_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.BatchUpdate = errors.New("permission denied")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":     "test-sheet-1",
		"source_sheet_id":    float64(0),
		"start_row":          float64(0),
		"start_col":          float64(0),
		"end_row":            float64(10),
		"end_col":            float64(3),
		"target_sheet_id":    float64(0),
		"row_source_columns": []any{float64(0)},
	})

	result, err := TestableSheetsCreatePivotTable(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}
