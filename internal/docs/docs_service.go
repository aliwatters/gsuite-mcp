package docs

import (
	"bytes"
	"context"
	"io"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// DocsService defines the interface for Google Docs API operations.
// This interface enables dependency injection and testing with mocks.
type DocsService interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, documentID string) (*docs.Document, error)

	// CreateDocument creates a new document with the given title.
	CreateDocument(ctx context.Context, title string) (*docs.Document, error)

	// BatchUpdate performs a batch update on a document.
	BatchUpdate(ctx context.Context, documentID string, requests []*docs.Request) (*docs.BatchUpdateDocumentResponse, error)

	// ExportPDF exports a document to PDF format, returning the raw PDF bytes.
	ExportPDF(ctx context.Context, fileID string) ([]byte, *drive.File, error)

	// ImportDocument creates a new Google Doc by uploading content with conversion.
	ImportDocument(ctx context.Context, title string, content []byte, contentType string, parentID string) (*drive.File, error)
}

// RealDocsService wraps the Docs API client and implements DocsService.
type RealDocsService struct {
	service      *docs.Service
	driveService *drive.Service
}

// NewRealDocsService creates a new RealDocsService wrapping the given API services.
func NewRealDocsService(service *docs.Service, driveService *drive.Service) *RealDocsService {
	return &RealDocsService{service: service, driveService: driveService}
}

// GetDocument retrieves a document by ID.
func (s *RealDocsService) GetDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	return s.service.Documents.Get(documentID).Context(ctx).Do()
}

// CreateDocument creates a new document with the given title.
func (s *RealDocsService) CreateDocument(ctx context.Context, title string) (*docs.Document, error) {
	doc := &docs.Document{
		Title: title,
	}
	return s.service.Documents.Create(doc).Context(ctx).Do()
}

// BatchUpdate performs a batch update on a document.
func (s *RealDocsService) BatchUpdate(ctx context.Context, documentID string, requests []*docs.Request) (*docs.BatchUpdateDocumentResponse, error) {
	req := &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}
	return s.service.Documents.BatchUpdate(documentID, req).Context(ctx).Do()
}

// ExportPDF exports a document to PDF format using the Drive API.
func (s *RealDocsService) ExportPDF(ctx context.Context, fileID string) ([]byte, *drive.File, error) {
	// Get file metadata
	file, err := s.driveService.Files.Get(fileID).Fields("id,name,mimeType").Context(ctx).Do()
	if err != nil {
		return nil, nil, err
	}

	// Export as PDF
	resp, err := s.driveService.Files.Export(fileID, "application/pdf").Context(ctx).Download()
	if err != nil {
		return nil, file, err
	}
	defer resp.Body.Close()

	// Read with size limit
	limited := io.LimitReader(resp.Body, maxExportSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, file, err
	}

	return data, file, nil
}

// ImportDocument creates a new Google Doc by uploading content with conversion.
func (s *RealDocsService) ImportDocument(ctx context.Context, title string, content []byte, contentType string, parentID string) (*drive.File, error) {
	file := &drive.File{
		Name:     title,
		MimeType: "application/vnd.google-apps.document",
	}
	if parentID != "" {
		file.Parents = []string{parentID}
	}

	return s.driveService.Files.Create(file).
		Media(bytes.NewReader(content)).
		Context(ctx).
		Do()
}
