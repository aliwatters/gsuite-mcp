package docs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// DocsService defines the interface for Google Docs API operations.
// This interface enables dependency injection and testing with mocks.
type DocsService interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, documentID string) (*docs.Document, error)

	// GetDocumentWithSuggestions retrieves a document with suggestions inline.
	// This mode shows suggested insertions and deletions alongside the document content.
	GetDocumentWithSuggestions(ctx context.Context, documentID string) (*docs.Document, error)

	// CreateDocument creates a new document with the given title.
	CreateDocument(ctx context.Context, title string) (*docs.Document, error)

	// BatchUpdate performs a batch update on a document.
	BatchUpdate(ctx context.Context, documentID string, requests []*docs.Request) (*docs.BatchUpdateDocumentResponse, error)

	// BatchUpdateRaw performs a batch update using raw JSON, bypassing Go client
	// struct limitations. This supports request types (like updateNamedStyle) that
	// the Go client library has not yet added to its typed Request struct.
	BatchUpdateRaw(ctx context.Context, documentID string, requestsJSON json.RawMessage) (*docs.BatchUpdateDocumentResponse, error)

	// ExportPDF exports a document to PDF format, returning the raw PDF bytes.
	ExportPDF(ctx context.Context, fileID string) ([]byte, *drive.File, error)

	// ImportDocument creates a new Google Doc by uploading content with conversion.
	ImportDocument(ctx context.Context, title string, content []byte, contentType string, parentID string) (*drive.File, error)
}

// RealDocsService wraps the Docs API client and implements DocsService.
type RealDocsService struct {
	service      *docs.Service
	driveService *drive.Service
	httpClient   *http.Client
}

// NewRealDocsService creates a new RealDocsService wrapping the given API services.
func NewRealDocsService(service *docs.Service, driveService *drive.Service) *RealDocsService {
	return &RealDocsService{service: service, driveService: driveService}
}

// NewRealDocsServiceWithHTTP creates a new RealDocsService with an explicit HTTP client
// for raw API calls that bypass the typed Go client structs.
func NewRealDocsServiceWithHTTP(service *docs.Service, driveService *drive.Service, httpClient *http.Client) *RealDocsService {
	return &RealDocsService{service: service, driveService: driveService, httpClient: httpClient}
}

// GetDocument retrieves a document by ID.
func (s *RealDocsService) GetDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	return s.service.Documents.Get(documentID).Context(ctx).Do()
}

// GetDocumentWithSuggestions retrieves a document with suggestions shown inline.
func (s *RealDocsService) GetDocumentWithSuggestions(ctx context.Context, documentID string) (*docs.Document, error) {
	return s.service.Documents.Get(documentID).SuggestionsViewMode("SUGGESTIONS_INLINE").Context(ctx).Do()
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

// BatchUpdateRaw performs a batch update using raw JSON bytes. This bypasses the
// typed docs.Request struct, allowing request types not yet supported by the Go
// client library (e.g., updateNamedStyle) to be sent to the API.
func (s *RealDocsService) BatchUpdateRaw(ctx context.Context, documentID string, requestsJSON json.RawMessage) (*docs.BatchUpdateDocumentResponse, error) {
	if s.httpClient == nil {
		return nil, fmt.Errorf("raw batch update requires an HTTP client; use NewRealDocsServiceWithHTTP")
	}

	body := struct {
		Requests json.RawMessage `json:"requests"`
	}{Requests: requestsJSON}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal batch update body: %w", err)
	}

	url := fmt.Sprintf("%sdocuments/%s:batchUpdate", s.service.BasePath, documentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result docs.BatchUpdateDocumentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
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
