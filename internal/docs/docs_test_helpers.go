package docs

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// Aliases for testable functions used by test files.
var (
	testableDocsCreate      = TestableDocsCreate
	testableDocsGet         = TestableDocsGet
	testableDocsGetMetadata = TestableDocsGetMetadata
	testableDocsAppendText  = TestableDocsAppendText
	testableDocsInsertText  = TestableDocsInsertText
	testableDocsReplaceText = TestableDocsReplaceText
	testableDocsDeleteText  = TestableDocsDeleteText
	testableDocsInsertTable = TestableDocsInsertTable
	testableDocsInsertLink  = TestableDocsInsertLink
	testableDocsBatchUpdate = TestableDocsBatchUpdate
)

// getDocsTextContent extracts text content from an MCP CallToolResult.
func getDocsTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return ""
}

// DocsTestFixtures contains test fixtures for Docs tool testing.
type DocsTestFixtures struct {
	DefaultEmail string
	MockService  *MockDocsService
	Deps         *DocsHandlerDeps
}

// NewDocsTestFixtures creates a new set of test fixtures.
func NewDocsTestFixtures() *DocsTestFixtures {
	mockService := NewMockDocsService()
	f := common.NewTestFixtures[DocsService](mockService)

	return &DocsTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}

// AddTestDocument adds a test document to the mock service.
func (f *DocsTestFixtures) AddTestDocument(id, title, content string) {
	f.MockService.Documents[id] = &docs.Document{
		DocumentId: id,
		Title:      title,
		RevisionId: "rev-1",
		Body: &docs.Body{
			Content: []*docs.StructuralElement{
				{
					StartIndex: 0,
					EndIndex:   int64(len(content) + 1),
					Paragraph: &docs.Paragraph{
						Elements: []*docs.ParagraphElement{
							{
								StartIndex: 0,
								EndIndex:   int64(len(content) + 1),
								TextRun: &docs.TextRun{
									Content: content + "\n",
								},
							},
						},
					},
				},
			},
		},
	}
}
