package citation

import (
	"context"
	"io"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
	gslides "google.golang.org/api/slides/v1"
)

// mockCitationDrive implements CitationDriveService for tests.
type mockCitationDrive struct {
	files map[string]*drive.File
}

func (m *mockCitationDrive) GetFile(_ context.Context, fileID string, _ string) (*drive.File, error) {
	if f, ok := m.files[fileID]; ok {
		return f, nil
	}
	return nil, io.ErrUnexpectedEOF
}

func (m *mockCitationDrive) DownloadFile(_ context.Context, fileID string) (io.ReadCloser, error) {
	if f, ok := m.files[fileID]; ok {
		return io.NopCloser(strings.NewReader("content of " + f.Name)), nil
	}
	return nil, io.ErrUnexpectedEOF
}

func (m *mockCitationDrive) ExportFile(_ context.Context, fileID string, _ string) (io.ReadCloser, error) {
	if f, ok := m.files[fileID]; ok {
		return io.NopCloser(strings.NewReader("exported " + f.Name)), nil
	}
	return nil, io.ErrUnexpectedEOF
}

func (m *mockCitationDrive) MoveFile(_ context.Context, fileID string, _ string) (*drive.File, error) {
	if f, ok := m.files[fileID]; ok {
		return f, nil
	}
	return nil, io.ErrUnexpectedEOF
}

// mockCitationSheets implements CitationSheetsService for tests.
type mockCitationSheets struct {
	created *sheets.Spreadsheet
}

func (m *mockCitationSheets) CreateSpreadsheet(_ context.Context, ss *sheets.Spreadsheet) (*sheets.Spreadsheet, error) {
	ss.SpreadsheetId = "mock-sheet-id"
	m.created = ss
	return ss, nil
}

func (m *mockCitationSheets) BatchUpdateValues(_ context.Context, _ string, _ *sheets.BatchUpdateValuesRequest) error {
	return nil
}

func (m *mockCitationSheets) AppendValues(_ context.Context, _, _ string, _ *sheets.ValueRange) error {
	return nil
}

func (m *mockCitationSheets) GetValues(_ context.Context, _, _ string) (*sheets.ValueRange, error) {
	return &sheets.ValueRange{}, nil
}

func (m *mockCitationSheets) UpdateValues(_ context.Context, _, _ string, _ *sheets.ValueRange) error {
	return nil
}

func (m *mockCitationSheets) ClearValues(_ context.Context, _, _ string) error {
	return nil
}

// mockCitationSlides implements CitationSlidesService for tests.
type mockCitationSlides struct{}

func (m *mockCitationSlides) GetPresentation(_ context.Context, _ string) (*gslides.Presentation, error) {
	return &gslides.Presentation{
		Slides: []*gslides.Page{
			{
				PageElements: []*gslides.PageElement{
					{Shape: &gslides.Shape{Text: &gslides.TextContent{
						TextElements: []*gslides.TextElement{
							{TextRun: &gslides.TextRun{Content: "Slide 1 content"}},
						},
					}}},
				},
			},
		},
	}, nil
}

func TestNewRealCitationServiceWithDeps_AddDocuments(t *testing.T) {
	ctx := context.Background()

	mockDrive := &mockCitationDrive{
		files: map[string]*drive.File{
			"f1": {Id: "f1", Name: "report.txt", MimeType: "text/plain", ModifiedTime: "2026-01-01T00:00:00Z"},
		},
	}
	mockSheets := &mockCitationSheets{}

	svc := NewRealCitationServiceWithDeps(mockDrive, mockSheets, &mockCitationSlides{}, &CitationConfig{
		Indexes: map[string]IndexEntry{"idx1": {SheetID: "sheet1"}},
	})

	// Pre-populate a store so we don't hit real Sheets for DualStore creation
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer sqlite.Close()
	svc.stores["idx1"] = &DualStore{indexID: "idx1", sheets: NewSheetsStore(mockSheets, "sheet1"), sqlite: sqlite}

	count, err := svc.AddDocuments(ctx, "idx1", []string{"f1"})
	if err != nil {
		t.Fatalf("AddDocuments: %v", err)
	}
	if count == 0 {
		t.Error("expected at least 1 chunk from text file")
	}
}

func TestNewRealCitationServiceWithDeps_CreateIndex(t *testing.T) {
	ctx := context.Background()

	mockDrive := &mockCitationDrive{files: make(map[string]*drive.File)}
	mockSheets := &mockCitationSheets{}

	svc := NewRealCitationServiceWithDeps(mockDrive, mockSheets, &mockCitationSlides{}, nil)

	info, err := svc.CreateIndex(ctx, "test-index", "")
	if err != nil {
		t.Fatalf("CreateIndex: %v", err)
	}
	if info.IndexID != "test-index" {
		t.Errorf("expected index_id=test-index, got %q", info.IndexID)
	}
	if info.SheetID != "mock-sheet-id" {
		t.Errorf("expected sheet_id=mock-sheet-id, got %q", info.SheetID)
	}
}

func TestNewRealCitationServiceWithDeps_ChunkSlides(t *testing.T) {
	ctx := context.Background()

	mockDrive := &mockCitationDrive{
		files: map[string]*drive.File{
			"pres1": {Id: "pres1", Name: "deck.gslides", MimeType: "application/vnd.google-apps.presentation", ModifiedTime: "2026-01-01T00:00:00Z"},
		},
	}
	mockSheets := &mockCitationSheets{}
	mockSlides := &mockCitationSlides{}

	svc := NewRealCitationServiceWithDeps(mockDrive, mockSheets, mockSlides, &CitationConfig{
		Indexes: map[string]IndexEntry{"idx1": {SheetID: "sheet1"}},
	})

	// Pre-populate store
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer sqlite.Close()
	svc.stores["idx1"] = &DualStore{indexID: "idx1", sheets: NewSheetsStore(mockSheets, "sheet1"), sqlite: sqlite}

	count, err := svc.AddDocuments(ctx, "idx1", []string{"pres1"})
	if err != nil {
		t.Fatalf("AddDocuments (slides): %v", err)
	}
	if count == 0 {
		t.Error("expected at least 1 chunk from slides presentation")
	}
}
