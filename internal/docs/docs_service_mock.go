package docs

import (
	"context"
	"fmt"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// MockDocsService implements DocsService for testing.
type MockDocsService struct {
	// Documents stores mock document data keyed by document ID
	Documents map[string]*docs.Document

	// ExportedPDFs stores mock PDF data keyed by document ID
	ExportedPDFs map[string][]byte

	// ImportedDocs stores mock imported document data
	ImportedDocs []*drive.File

	// Errors allows tests to configure specific errors for methods
	Errors struct {
		GetDocument    error
		Create         error
		BatchUpdate    error
		ExportPDF      error
		ImportDocument error
	}

	// Calls tracks method invocations for verification
	Calls struct {
		GetDocument []string
		Create      []string
		BatchUpdate []struct {
			DocumentID string
			Requests   []*docs.Request
		}
		ExportPDF      []string
		ImportDocument []struct {
			Title       string
			ContentType string
			ParentID    string
		}
	}
}

// NewMockDocsService creates a new mock Docs service with default test data.
func NewMockDocsService() *MockDocsService {
	m := &MockDocsService{
		Documents:    make(map[string]*docs.Document),
		ExportedPDFs: make(map[string][]byte),
	}

	// Add a default test document
	m.Documents["test-doc-1"] = &docs.Document{
		DocumentId: "test-doc-1",
		Title:      "Test Document",
		RevisionId: "rev-1",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   13,
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   13,
								TextRun: &docs.TextRun{
									Content: "Hello World!\n",
								},
							},
						},
					},
				},
			},
		},
	}

	return m
}

// GetDocument retrieves a mock document by ID.
func (m *MockDocsService) GetDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	m.Calls.GetDocument = append(m.Calls.GetDocument, documentID)

	if m.Errors.GetDocument != nil {
		return nil, m.Errors.GetDocument
	}

	doc, ok := m.Documents[documentID]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	return doc, nil
}

// CreateDocument creates a mock document.
func (m *MockDocsService) CreateDocument(ctx context.Context, title string) (*docs.Document, error) {
	m.Calls.Create = append(m.Calls.Create, title)

	if m.Errors.Create != nil {
		return nil, m.Errors.Create
	}

	docID := fmt.Sprintf("new-doc-%d", len(m.Documents)+1)
	doc := &docs.Document{
		DocumentId: docID,
		Title:      title,
		RevisionId: "rev-1",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   1,
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   1,
								TextRun: &docs.TextRun{
									Content: "\n",
								},
							},
						},
					},
				},
			},
		},
	}

	m.Documents[docID] = doc
	return doc, nil
}

// BatchUpdate performs a mock batch update on a document.
func (m *MockDocsService) BatchUpdate(ctx context.Context, documentID string, requests []*docs.Request) (*docs.BatchUpdateDocumentResponse, error) {
	m.Calls.BatchUpdate = append(m.Calls.BatchUpdate, struct {
		DocumentID string
		Requests   []*docs.Request
	}{documentID, requests})

	if m.Errors.BatchUpdate != nil {
		return nil, m.Errors.BatchUpdate
	}

	doc, ok := m.Documents[documentID]
	if !ok {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	// Simulate applying updates (just update revision for now)
	doc.RevisionId = fmt.Sprintf("rev-%d", len(m.Calls.BatchUpdate)+1)

	return &docs.BatchUpdateDocumentResponse{
		DocumentId: documentID,
	}, nil
}

// ExportPDF exports a mock document as PDF.
func (m *MockDocsService) ExportPDF(ctx context.Context, fileID string) ([]byte, *drive.File, error) {
	m.Calls.ExportPDF = append(m.Calls.ExportPDF, fileID)

	if m.Errors.ExportPDF != nil {
		return nil, nil, m.Errors.ExportPDF
	}

	doc, ok := m.Documents[fileID]
	if !ok {
		return nil, nil, fmt.Errorf("document not found: %s", fileID)
	}

	// Return mock PDF data if configured, otherwise generate fake PDF bytes
	pdfData := m.ExportedPDFs[fileID]
	if pdfData == nil {
		pdfData = []byte("%PDF-1.4 mock pdf content")
	}

	file := &drive.File{
		Id:       fileID,
		Name:     doc.Title,
		MimeType: "application/vnd.google-apps.document",
	}

	return pdfData, file, nil
}

// ImportDocument creates a mock imported document.
func (m *MockDocsService) ImportDocument(ctx context.Context, title string, content []byte, contentType string, parentID string) (*drive.File, error) {
	m.Calls.ImportDocument = append(m.Calls.ImportDocument, struct {
		Title       string
		ContentType string
		ParentID    string
	}{title, contentType, parentID})

	if m.Errors.ImportDocument != nil {
		return nil, m.Errors.ImportDocument
	}

	docID := fmt.Sprintf("imported-doc-%d", len(m.Calls.ImportDocument))
	file := &drive.File{
		Id:          docID,
		Name:        title,
		MimeType:    "application/vnd.google-apps.document",
		WebViewLink: fmt.Sprintf("https://docs.google.com/document/d/%s/edit", docID),
	}

	m.ImportedDocs = append(m.ImportedDocs, file)
	return file, nil
}
