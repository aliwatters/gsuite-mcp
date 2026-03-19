package forms

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// Aliases for testable functions used by test files.
var (
	testableFormsGet           = TestableFormsGet
	testableFormsCreate        = TestableFormsCreate
	testableFormsBatchUpdate   = TestableFormsBatchUpdate
	testableFormsListResponses = TestableFormsListResponses
	testableFormsGetResponse   = TestableFormsGetResponse
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

// FormsTestFixtures contains test fixtures for Forms tool testing.
type FormsTestFixtures struct {
	DefaultEmail string
	MockService  *MockFormsService
	Deps         *FormsHandlerDeps
}

// NewFormsTestFixtures creates a new set of test fixtures.
func NewFormsTestFixtures() *FormsTestFixtures {
	mockService := NewMockFormsService()
	f := common.NewTestFixtures[FormsService](mockService)

	return &FormsTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
