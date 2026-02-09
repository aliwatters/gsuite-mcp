package docs

import (
	"context"

	"google.golang.org/api/docs/v1"
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
}

// RealDocsService wraps the Docs API client and implements DocsService.
type RealDocsService struct {
	service *docs.Service
}

// NewRealDocsService creates a new RealDocsService wrapping the given Docs API service.
func NewRealDocsService(service *docs.Service) *RealDocsService {
	return &RealDocsService{service: service}
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
