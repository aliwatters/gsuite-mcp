package sheets

import (
	"context"

	"google.golang.org/api/sheets/v4"
)

// SheetsService defines the interface for Google Sheets API operations.
// This interface enables dependency injection and testing with mocks.
type SheetsService interface {
	// GetSpreadsheet retrieves spreadsheet metadata.
	GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)

	// GetValues reads values from a range.
	GetValues(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error)

	// UpdateValues writes values to a range.
	UpdateValues(ctx context.Context, spreadsheetID, writeRange string, values [][]any, valueInputOption string) (*sheets.UpdateValuesResponse, error)

	// AppendValues appends values after a table.
	AppendValues(ctx context.Context, spreadsheetID, appendRange string, values [][]any, valueInputOption string) (*sheets.AppendValuesResponse, error)

	// BatchGetValues reads multiple ranges at once.
	BatchGetValues(ctx context.Context, spreadsheetID string, ranges []string) (*sheets.BatchGetValuesResponse, error)

	// BatchUpdateValues writes to multiple ranges at once.
	BatchUpdateValues(ctx context.Context, spreadsheetID string, data []*sheets.ValueRange, valueInputOption string) (*sheets.BatchUpdateValuesResponse, error)

	// ClearValues clears values from a range.
	ClearValues(ctx context.Context, spreadsheetID, clearRange string) (*sheets.ClearValuesResponse, error)

	// CreateSpreadsheet creates a new spreadsheet.
	CreateSpreadsheet(ctx context.Context, title string) (*sheets.Spreadsheet, error)
}

// RealSheetsService wraps the Sheets API client and implements SheetsService.
type RealSheetsService struct {
	service *sheets.Service
}

// NewRealSheetsService creates a new RealSheetsService wrapping the given Sheets API service.
func NewRealSheetsService(service *sheets.Service) *RealSheetsService {
	return &RealSheetsService{service: service}
}

// GetSpreadsheet retrieves spreadsheet metadata.
func (s *RealSheetsService) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	return s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
}

// GetValues reads values from a range.
func (s *RealSheetsService) GetValues(ctx context.Context, spreadsheetID, readRange string) (*sheets.ValueRange, error) {
	return s.service.Spreadsheets.Values.Get(spreadsheetID, readRange).Context(ctx).Do()
}

// UpdateValues writes values to a range.
func (s *RealSheetsService) UpdateValues(ctx context.Context, spreadsheetID, writeRange string, values [][]any, valueInputOption string) (*sheets.UpdateValuesResponse, error) {
	valueRange := &sheets.ValueRange{
		Values: values,
	}
	return s.service.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).
		ValueInputOption(valueInputOption).
		Context(ctx).
		Do()
}

// AppendValues appends values after a table.
func (s *RealSheetsService) AppendValues(ctx context.Context, spreadsheetID, appendRange string, values [][]any, valueInputOption string) (*sheets.AppendValuesResponse, error) {
	valueRange := &sheets.ValueRange{
		Values: values,
	}
	return s.service.Spreadsheets.Values.Append(spreadsheetID, appendRange, valueRange).
		ValueInputOption(valueInputOption).
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
}

// BatchGetValues reads multiple ranges at once.
func (s *RealSheetsService) BatchGetValues(ctx context.Context, spreadsheetID string, ranges []string) (*sheets.BatchGetValuesResponse, error) {
	return s.service.Spreadsheets.Values.BatchGet(spreadsheetID).
		Ranges(ranges...).
		Context(ctx).
		Do()
}

// BatchUpdateValues writes to multiple ranges at once.
func (s *RealSheetsService) BatchUpdateValues(ctx context.Context, spreadsheetID string, data []*sheets.ValueRange, valueInputOption string) (*sheets.BatchUpdateValuesResponse, error) {
	req := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: valueInputOption,
		Data:             data,
	}
	return s.service.Spreadsheets.Values.BatchUpdate(spreadsheetID, req).
		Context(ctx).
		Do()
}

// ClearValues clears values from a range.
func (s *RealSheetsService) ClearValues(ctx context.Context, spreadsheetID, clearRange string) (*sheets.ClearValuesResponse, error) {
	return s.service.Spreadsheets.Values.Clear(spreadsheetID, clearRange, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()
}

// CreateSpreadsheet creates a new spreadsheet.
func (s *RealSheetsService) CreateSpreadsheet(ctx context.Context, title string) (*sheets.Spreadsheet, error) {
	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}
	return s.service.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
}
