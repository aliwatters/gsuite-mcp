package meet

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// Aliases for testable functions used by test files.
var (
	testableListConferenceRecords = TestableListConferenceRecords
	testableGetConferenceRecord   = TestableGetConferenceRecord
	testableListParticipants      = TestableListParticipants
	testableListTranscripts       = TestableListTranscripts
	testableGetTranscript         = TestableGetTranscript
	testableListTranscriptEntries = TestableListTranscriptEntries
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

// MeetTestFixtures contains test fixtures for Meet tool testing.
type MeetTestFixtures struct {
	DefaultEmail string
	MockService  *MockMeetService
	Deps         *MeetHandlerDeps
}

// NewMeetTestFixtures creates a new set of test fixtures.
func NewMeetTestFixtures() *MeetTestFixtures {
	mockService := NewMockMeetService()
	f := common.NewTestFixtures[MeetService](mockService)

	return &MeetTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
