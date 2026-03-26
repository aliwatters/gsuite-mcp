package driveactivity

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// Aliases for testable functions used by test files.
var (
	testableDriveActivityQuery = TestableDriveActivityQuery
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

// DriveActivityTestFixtures contains test fixtures for Drive Activity tool testing.
type DriveActivityTestFixtures struct {
	DefaultEmail string
	MockService  *MockDriveActivityService
	Deps         *DriveActivityHandlerDeps
}

// NewDriveActivityTestFixtures creates a new set of test fixtures.
func NewDriveActivityTestFixtures() *DriveActivityTestFixtures {
	mockService := NewMockDriveActivityService()
	f := common.NewTestFixtures[DriveActivityService](mockService)

	return &DriveActivityTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
