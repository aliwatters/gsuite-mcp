package slides

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// Aliases for testable functions used by test files.
var (
	testableSlidesGetPresentation = TestableSlidesGetPresentation
	testableSlidesGetPage         = TestableSlidesGetPage
	testableSlidesGetThumbnail    = TestableSlidesGetThumbnail
	testableSlidesCreate          = TestableSlidesCreate
	testableSlidesBatchUpdate     = TestableSlidesBatchUpdate
)

// getTextContent extracts text content from an MCP CallToolResult.
func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return ""
}

// SlidesTestFixtures contains test fixtures for Slides tool testing.
type SlidesTestFixtures struct {
	DefaultEmail string
	MockService  *MockSlidesService
	Deps         *SlidesHandlerDeps
}

// NewSlidesTestFixtures creates a new set of test fixtures.
func NewSlidesTestFixtures() *SlidesTestFixtures {
	mockService := NewMockSlidesService()
	f := common.NewTestFixtures[SlidesService](mockService)

	return &SlidesTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
