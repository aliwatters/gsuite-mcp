package sheets

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// MockSheetsService implements SheetsService for testing.
type MockSheetsService struct {
	// Spreadsheets stores mock spreadsheet data keyed by spreadsheet ID
	Spreadsheets map[string]*sheets.Spreadsheet

	// Values stores mock cell values keyed by "spreadsheetID:range"
	Values map[string][][]any

	// Errors allows tests to configure specific errors for methods
	Errors struct {
		GetSpreadsheet    error
		GetValues         error
		UpdateValues      error
		AppendValues      error
		BatchGetValues    error
		BatchUpdateValues error
		ClearValues       error
		CreateSpreadsheet error
	}

	// Calls tracks method invocations for verification
	Calls struct {
		GetSpreadsheet []string
		GetValues      []struct{ SpreadsheetID, Range string }
		UpdateValues   []struct {
			SpreadsheetID, Range string
			Values               [][]any
		}
		AppendValues []struct {
			SpreadsheetID, Range string
			Values               [][]any
		}
		BatchGetValues []struct {
			SpreadsheetID string
			Ranges        []string
		}
		BatchUpdateValues []struct {
			SpreadsheetID string
			Data          []*sheets.ValueRange
		}
		ClearValues       []struct{ SpreadsheetID, Range string }
		CreateSpreadsheet []string
	}
}

// NewMockSheetsService creates a new mock Sheets service with default test data.
func NewMockSheetsService() *MockSheetsService {
	m := &MockSheetsService{
		Spreadsheets: make(map[string]*sheets.Spreadsheet),
		Values:       make(map[string][][]any),
	}

	// Add a default test spreadsheet
	m.Spreadsheets["test-sheet-1"] = &sheets.Spreadsheet{
		SpreadsheetId: "test-sheet-1",
		Properties: &sheets.SpreadsheetProperties{
			Title: "Test Spreadsheet",
		},
		Sheets: []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					SheetId: 0,
					Title:   "Sheet1",
					Index:   0,
				},
			},
		},
	}

	// Add test values
	m.Values["test-sheet-1:Sheet1!A1:B2"] = [][]any{
		{"Name", "Email"},
		{"Alice", "alice@example.com"},
	}

	return m
}

// GetSpreadsheet retrieves a mock spreadsheet by ID.
func (m *MockSheetsService) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	m.Calls.GetSpreadsheet = append(m.Calls.GetSpreadsheet, spreadsheetID)

	if m.Errors.GetSpreadsheet != nil {
		return nil, m.Errors.GetSpreadsheet
	}

	ss, ok := m.Spreadsheets[spreadsheetID]
	if !ok {
		return nil, fmt.Errorf("spreadsheet not found: %s", spreadsheetID)
	}

	return ss, nil
}

// GetValues reads values from a range.
func (m *MockSheetsService) GetValues(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	m.Calls.GetValues = append(m.Calls.GetValues, struct{ SpreadsheetID, Range string }{spreadsheetID, readRange})

	if m.Errors.GetValues != nil {
		return nil, m.Errors.GetValues
	}

	key := spreadsheetID + ":" + readRange
	values, ok := m.Values[key]
	if !ok {
		// Return empty values if not found
		return &sheets.ValueRange{Range: readRange}, nil
	}

	return &sheets.ValueRange{
		Range:  readRange,
		Values: values,
	}, nil
}

// UpdateValues writes values to a range.
func (m *MockSheetsService) UpdateValues(ctx context.Context, spreadsheetID, writeRange string, values [][]any, valueInputOption string) (*sheets.UpdateValuesResponse, error) {
	m.Calls.UpdateValues = append(m.Calls.UpdateValues, struct {
		SpreadsheetID, Range string
		Values               [][]any
	}{spreadsheetID, writeRange, values})

	if m.Errors.UpdateValues != nil {
		return nil, m.Errors.UpdateValues
	}

	// Store the values
	key := spreadsheetID + ":" + writeRange
	m.Values[key] = values

	return &sheets.UpdateValuesResponse{
		SpreadsheetId:  spreadsheetID,
		UpdatedRange:   writeRange,
		UpdatedRows:    int64(len(values)),
		UpdatedColumns: int64(len(values[0])),
		UpdatedCells:   int64(len(values) * len(values[0])),
	}, nil
}

// AppendValues appends values after a table.
func (m *MockSheetsService) AppendValues(ctx context.Context, spreadsheetID, appendRange string, values [][]any, valueInputOption string) (*sheets.AppendValuesResponse, error) {
	m.Calls.AppendValues = append(m.Calls.AppendValues, struct {
		SpreadsheetID, Range string
		Values               [][]any
	}{spreadsheetID, appendRange, values})

	if m.Errors.AppendValues != nil {
		return nil, m.Errors.AppendValues
	}

	return &sheets.AppendValuesResponse{
		SpreadsheetId: spreadsheetID,
		Updates: &sheets.UpdateValuesResponse{
			UpdatedRange:   appendRange,
			UpdatedRows:    int64(len(values)),
			UpdatedColumns: int64(len(values[0])),
			UpdatedCells:   int64(len(values) * len(values[0])),
		},
	}, nil
}

// BatchGetValues reads multiple ranges at once.
func (m *MockSheetsService) BatchGetValues(ctx context.Context, spreadsheetID string, ranges []string) (*sheets.BatchGetValuesResponse, error) {
	m.Calls.BatchGetValues = append(m.Calls.BatchGetValues, struct {
		SpreadsheetID string
		Ranges        []string
	}{spreadsheetID, ranges})

	if m.Errors.BatchGetValues != nil {
		return nil, m.Errors.BatchGetValues
	}

	var valueRanges []*sheets.ValueRange
	for _, r := range ranges {
		key := spreadsheetID + ":" + r
		values := m.Values[key]
		valueRanges = append(valueRanges, &sheets.ValueRange{
			Range:  r,
			Values: values,
		})
	}

	return &sheets.BatchGetValuesResponse{
		SpreadsheetId: spreadsheetID,
		ValueRanges:   valueRanges,
	}, nil
}

// BatchUpdateValues writes to multiple ranges at once.
func (m *MockSheetsService) BatchUpdateValues(ctx context.Context, spreadsheetID string, data []*sheets.ValueRange, valueInputOption string) (*sheets.BatchUpdateValuesResponse, error) {
	m.Calls.BatchUpdateValues = append(m.Calls.BatchUpdateValues, struct {
		SpreadsheetID string
		Data          []*sheets.ValueRange
	}{spreadsheetID, data})

	if m.Errors.BatchUpdateValues != nil {
		return nil, m.Errors.BatchUpdateValues
	}

	var totalCells int64
	for _, vr := range data {
		key := spreadsheetID + ":" + vr.Range
		m.Values[key] = vr.Values
		for _, row := range vr.Values {
			totalCells += int64(len(row))
		}
	}

	return &sheets.BatchUpdateValuesResponse{
		SpreadsheetId:       spreadsheetID,
		TotalUpdatedRows:    int64(len(data)),
		TotalUpdatedColumns: 0, // Simplified
		TotalUpdatedCells:   totalCells,
	}, nil
}

// ClearValues clears values from a range.
func (m *MockSheetsService) ClearValues(ctx context.Context, spreadsheetID, clearRange string) (*sheets.ClearValuesResponse, error) {
	m.Calls.ClearValues = append(m.Calls.ClearValues, struct{ SpreadsheetID, Range string }{spreadsheetID, clearRange})

	if m.Errors.ClearValues != nil {
		return nil, m.Errors.ClearValues
	}

	key := spreadsheetID + ":" + clearRange
	delete(m.Values, key)

	return &sheets.ClearValuesResponse{
		SpreadsheetId: spreadsheetID,
		ClearedRange:  clearRange,
	}, nil
}

// CreateSpreadsheet creates a new spreadsheet.
func (m *MockSheetsService) CreateSpreadsheet(ctx context.Context, title string) (*sheets.Spreadsheet, error) {
	m.Calls.CreateSpreadsheet = append(m.Calls.CreateSpreadsheet, title)

	if m.Errors.CreateSpreadsheet != nil {
		return nil, m.Errors.CreateSpreadsheet
	}

	ssID := fmt.Sprintf("new-sheet-%d", len(m.Spreadsheets)+1)
	ss := &sheets.Spreadsheet{
		SpreadsheetId:  ssID,
		SpreadsheetUrl: fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", ssID),
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
		Sheets: []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					SheetId: 0,
					Title:   "Sheet1",
					Index:   0,
				},
			},
		},
	}

	m.Spreadsheets[ssID] = ss
	return ss, nil
}
