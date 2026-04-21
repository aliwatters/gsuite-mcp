package citation

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
	gslides "google.golang.org/api/slides/v1"
)

// CitationDriveService defines the Drive operations used by the citation package.
// Keeping this interface minimal (only methods actually called) enables testability
// without mocking the entire Drive API.
type CitationDriveService interface {
	// GetFile retrieves file metadata by ID.
	GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error)

	// DownloadFile downloads a file's raw content.
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)

	// ExportFile exports a Google Workspace file to the specified MIME type.
	ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error)

	// AddParentFolder adds a parent folder to a file (Drive AddParents API).
	// This does not remove existing parents; use the Drive package's MoveFile for a full move.
	AddParentFolder(ctx context.Context, fileID string, parentFolderID string) (*drive.File, error)
}

// CitationSheetsService defines the Sheets operations used by the citation package.
type CitationSheetsService interface {
	// CreateSpreadsheet creates a new spreadsheet.
	CreateSpreadsheet(ctx context.Context, ss *sheets.Spreadsheet) (*sheets.Spreadsheet, error)

	// BatchUpdateValues writes multiple ranges of values.
	BatchUpdateValues(ctx context.Context, sheetID string, req *sheets.BatchUpdateValuesRequest) error

	// AppendValues appends rows to a sheet range.
	AppendValues(ctx context.Context, sheetID, rangeStr string, vr *sheets.ValueRange) error

	// GetValues reads values from a sheet range.
	GetValues(ctx context.Context, sheetID, rangeStr string) (*sheets.ValueRange, error)

	// UpdateValues writes values to a specific range (overwrites existing data).
	UpdateValues(ctx context.Context, sheetID, rangeStr string, vr *sheets.ValueRange) error

	// ClearValues clears values from a sheet range.
	ClearValues(ctx context.Context, sheetID, rangeStr string) error
}

// CitationSlidesService defines the Slides operations used by the citation package.
type CitationSlidesService interface {
	// GetPresentation retrieves a presentation by ID.
	GetPresentation(ctx context.Context, presentationID string) (*gslides.Presentation, error)
}

// realCitationDriveService wraps *drive.Service to implement CitationDriveService.
type realCitationDriveService struct {
	svc *drive.Service
}

func (r *realCitationDriveService) GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error) {
	call := r.svc.Files.Get(fileID).SupportsAllDrives(true).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

func (r *realCitationDriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := r.svc.Files.Get(fileID).SupportsAllDrives(true).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (r *realCitationDriveService) ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error) {
	resp, err := r.svc.Files.Export(fileID, mimeType).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (r *realCitationDriveService) AddParentFolder(ctx context.Context, fileID string, parentFolderID string) (*drive.File, error) {
	return r.svc.Files.Update(fileID, nil).
		AddParents(parentFolderID).
		SupportsAllDrives(true).
		Context(ctx).Do()
}

// realCitationSheetsService wraps *sheets.Service to implement CitationSheetsService.
type realCitationSheetsService struct {
	svc *sheets.Service
}

func (r *realCitationSheetsService) CreateSpreadsheet(ctx context.Context, ss *sheets.Spreadsheet) (*sheets.Spreadsheet, error) {
	return r.svc.Spreadsheets.Create(ss).Context(ctx).Do()
}

func (r *realCitationSheetsService) BatchUpdateValues(ctx context.Context, sheetID string, req *sheets.BatchUpdateValuesRequest) error {
	_, err := r.svc.Spreadsheets.Values.BatchUpdate(sheetID, req).Context(ctx).Do()
	return err
}

func (r *realCitationSheetsService) AppendValues(ctx context.Context, sheetID, rangeStr string, vr *sheets.ValueRange) error {
	_, err := r.svc.Spreadsheets.Values.Append(sheetID, rangeStr, vr).
		ValueInputOption("RAW").Context(ctx).Do()
	return err
}

func (r *realCitationSheetsService) GetValues(ctx context.Context, sheetID, rangeStr string) (*sheets.ValueRange, error) {
	return r.svc.Spreadsheets.Values.Get(sheetID, rangeStr).Context(ctx).Do()
}

func (r *realCitationSheetsService) UpdateValues(ctx context.Context, sheetID, rangeStr string, vr *sheets.ValueRange) error {
	_, err := r.svc.Spreadsheets.Values.Update(sheetID, rangeStr, vr).
		ValueInputOption("RAW").Context(ctx).Do()
	return err
}

func (r *realCitationSheetsService) ClearValues(ctx context.Context, sheetID, rangeStr string) error {
	_, err := r.svc.Spreadsheets.Values.Clear(sheetID, rangeStr, &sheets.ClearValuesRequest{}).
		Context(ctx).Do()
	return err
}

// realCitationSlidesService wraps *gslides.Service to implement CitationSlidesService.
type realCitationSlidesService struct {
	svc *gslides.Service
}

func (r *realCitationSlidesService) GetPresentation(ctx context.Context, presentationID string) (*gslides.Presentation, error) {
	return r.svc.Presentations.Get(presentationID).Context(ctx).Do()
}
