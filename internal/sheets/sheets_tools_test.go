package sheets

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/sheets/v4"
)

// makeRequest creates a CallToolRequest with the given arguments.
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Arguments: args,
		},
	}
}

// parseResult unmarshals the JSON text from a tool result into a map.
func parseResult(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	return data
}

// isErrorResult checks if a tool result is an error result.
func isErrorResult(result *mcp.CallToolResult) bool {
	return result.IsError
}

// === SheetsGet Tests ===

func TestTestableSheetsGet_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsGet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["spreadsheet_id"] != "test-sheet-1" {
		t.Errorf("expected spreadsheet_id=test-sheet-1, got %v", data["spreadsheet_id"])
	}
	if data["title"] != "Test Spreadsheet" {
		t.Errorf("expected title=Test Spreadsheet, got %v", data["title"])
	}
}

func TestTestableSheetsGet_MissingID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{})

	result, err := TestableSheetsGet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing spreadsheet_id")
	}
}

func TestTestableSheetsGet_NotFound(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "nonexistent",
	})

	result, err := TestableSheetsGet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for nonexistent spreadsheet")
	}
}

func TestTestableSheetsGet_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.GetSpreadsheet = errors.New("API error")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsGet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === SheetsRead Tests ===

func TestTestableSheetsRead_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A1:B2",
	})

	result, err := TestableSheetsRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["range"] != "Sheet1!A1:B2" {
		t.Errorf("expected range=Sheet1!A1:B2, got %v", data["range"])
	}
	rowCount, ok := data["row_count"].(float64)
	if !ok || rowCount != 2 {
		t.Errorf("expected row_count=2, got %v", data["row_count"])
	}
}

func TestTestableSheetsRead_MissingRange(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range")
	}
}

func TestTestableSheetsRead_MissingSpreadsheetID(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"range": "Sheet1!A1:B2",
	})

	result, err := TestableSheetsRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing spreadsheet_id")
	}
}

// === SheetsWrite Tests ===

func TestTestableSheetsWrite_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A1:B2",
		"values": []any{
			[]any{"Name", "Age"},
			[]any{"Bob", float64(30)},
		},
	})

	result, err := TestableSheetsWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
	if data["updated_range"] != "Sheet1!A1:B2" {
		t.Errorf("expected updated_range=Sheet1!A1:B2, got %v", data["updated_range"])
	}
}

func TestTestableSheetsWrite_RawInput(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id":     "test-sheet-1",
		"range":              "Sheet1!A1:A1",
		"values":             []any{[]any{"=SUM(B1:B10)"}},
		"value_input_option": "raw",
	})

	result, err := TestableSheetsWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}
}

func TestTestableSheetsWrite_MissingValues(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A1:B2",
	})

	result, err := TestableSheetsWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing values")
	}
}

// === SheetsAppend Tests ===

func TestTestableSheetsAppend_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A:B",
		"values": []any{
			[]any{"Alice", "alice@example.com"},
		},
	})

	result, err := TestableSheetsAppend(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
}

func TestTestableSheetsAppend_MissingRange(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"values":         []any{[]any{"data"}},
	})

	result, err := TestableSheetsAppend(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range")
	}
}

// === SheetsCreate Tests ===

func TestTestableSheetsCreate_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"title": "My New Spreadsheet",
	})

	result, err := TestableSheetsCreate(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
	if data["title"] != "My New Spreadsheet" {
		t.Errorf("expected title=My New Spreadsheet, got %v", data["title"])
	}
}

func TestTestableSheetsCreate_MissingTitle(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{})

	result, err := TestableSheetsCreate(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing title")
	}
}

func TestTestableSheetsCreate_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.CreateSpreadsheet = errors.New("quota exceeded")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"title": "Test",
	})

	result, err := TestableSheetsCreate(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === SheetsBatchRead Tests ===

func TestTestableSheetsBatchRead_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"ranges":         []any{"Sheet1!A1:B2", "Sheet1!C1:D2"},
	})

	result, err := TestableSheetsBatchRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["spreadsheet_id"] != "test-sheet-1" {
		t.Errorf("expected spreadsheet_id=test-sheet-1, got %v", data["spreadsheet_id"])
	}
	rangesCount, ok := data["ranges_count"].(float64)
	if !ok || rangesCount != 2 {
		t.Errorf("expected ranges_count=2, got %v", data["ranges_count"])
	}
}

func TestTestableSheetsBatchRead_MissingRanges(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsBatchRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing ranges")
	}
}

func TestTestableSheetsBatchRead_EmptyRanges(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"ranges":         []any{},
	})

	result, err := TestableSheetsBatchRead(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for empty ranges")
	}
}

// === SheetsBatchWrite Tests ===

func TestTestableSheetsBatchWrite_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"data": []any{
			map[string]any{
				"range":  "Sheet1!A1:A1",
				"values": []any{[]any{"Hello"}},
			},
			map[string]any{
				"range":  "Sheet1!B1:B1",
				"values": []any{[]any{"World"}},
			},
		},
	})

	result, err := TestableSheetsBatchWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
}

func TestTestableSheetsBatchWrite_MissingData(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsBatchWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing data")
	}
}

func TestTestableSheetsBatchWrite_InvalidDataEntry(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"data":           []any{"not a map"},
	})

	result, err := TestableSheetsBatchWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for invalid data entry")
	}
}

func TestTestableSheetsBatchWrite_MissingRangeInEntry(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"data": []any{
			map[string]any{
				"values": []any{[]any{"data"}},
			},
		},
	})

	result, err := TestableSheetsBatchWrite(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range in data entry")
	}
}

// === SheetsClear Tests ===

func TestTestableSheetsClear_Success(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A1:B2",
	})

	result, err := TestableSheetsClear(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success, got error result")
	}

	data := parseResult(t, result)
	if data["success"] != true {
		t.Error("expected success=true")
	}
	if data["cleared_range"] != "Sheet1!A1:B2" {
		t.Errorf("expected cleared_range=Sheet1!A1:B2, got %v", data["cleared_range"])
	}
}

func TestTestableSheetsClear_MissingRange(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
	})

	result, err := TestableSheetsClear(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for missing range")
	}
}

func TestTestableSheetsClear_APIError(t *testing.T) {
	fixtures := NewSheetsTestFixtures()
	fixtures.MockService.Errors.ClearValues = errors.New("permission denied")
	ctx := context.Background()

	request := makeRequest(map[string]any{
		"spreadsheet_id": "test-sheet-1",
		"range":          "Sheet1!A1:B2",
	})

	result, err := TestableSheetsClear(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isErrorResult(result) {
		t.Fatal("expected error result for API error")
	}
}

// === parseValues Tests ===

func TestParseValues_NilInput(t *testing.T) {
	_, err := parseValues(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestParseValues_2DArray(t *testing.T) {
	input := []any{
		[]any{"a", "b"},
		[]any{"c", "d"},
	}

	result, err := parseValues(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
	if len(result[0]) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(result[0]))
	}
}

func TestParseValues_InvalidRow(t *testing.T) {
	input := []any{
		"not an array",
	}

	_, err := parseValues(input)
	if err == nil {
		t.Fatal("expected error for invalid row")
	}
}

func TestParseValues_InvalidType(t *testing.T) {
	_, err := parseValues("not an array")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

// === URL Extraction Test ===

func TestTestableSheetsGet_URLExtraction(t *testing.T) {
	fixtures := NewSheetsTestFixtures()

	// Add a spreadsheet with an ID that would be extracted from a URL
	fixtures.MockService.Spreadsheets["abc123def456"] = &sheets.Spreadsheet{
		SpreadsheetId: "abc123def456",
		Properties: &sheets.SpreadsheetProperties{
			Title: "URL Test Sheet",
		},
		SpreadsheetUrl: "https://docs.google.com/spreadsheets/d/abc123def456/edit",
		Sheets:         []*sheets.Sheet{},
	}

	ctx := context.Background()

	// Pass a full URL — the handler should extract the ID
	request := makeRequest(map[string]any{
		"spreadsheet_id": "https://docs.google.com/spreadsheets/d/abc123def456/edit#gid=0",
	})

	result, err := TestableSheetsGet(ctx, request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(result) {
		t.Fatal("expected success when passing URL, got error result")
	}

	data := parseResult(t, result)
	if data["title"] != "URL Test Sheet" {
		t.Errorf("expected title=URL Test Sheet, got %v", data["title"])
	}
}
