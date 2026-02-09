package docs

import (
	"context"
	"fmt"

	"google.golang.org/api/docs/v1"
)

// MockDocsService implements DocsService for testing.
type MockDocsService struct {
	// Documents stores mock document data keyed by document ID
	Documents map[string]*docs.Document

	// Errors allows tests to configure specific errors for methods
	Errors struct {
		GetDocument error
		Create      error
		BatchUpdate error
	}

	// Calls tracks method invocations for verification
	Calls struct {
		GetDocument []string
		Create      []string
		BatchUpdate []struct {
			DocumentID string
			Requests   []*docs.Request
		}
	}
}

// NewMockDocsService creates a new mock Docs service with default test data.
func NewMockDocsService() *MockDocsService {
	m := &MockDocsService{
		Documents: make(map[string]*docs.Document),
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
