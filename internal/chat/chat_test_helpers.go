package chat

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// Aliases for testable functions used by test files.
var (
	testableChatListSpaces     = TestableChatListSpaces
	testableChatGetSpace       = TestableChatGetSpace
	testableChatCreateSpace    = TestableChatCreateSpace
	testableChatListMessages   = TestableChatListMessages
	testableChatGetMessage     = TestableChatGetMessage
	testableChatSendMessage    = TestableChatSendMessage
	testableChatCreateReaction = TestableChatCreateReaction
	testableChatDeleteReaction = TestableChatDeleteReaction
	testableChatListMembers    = TestableChatListMembers
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

// ChatTestFixtures contains test fixtures for Chat tool testing.
type ChatTestFixtures struct {
	DefaultEmail string
	MockService  *MockChatService
	Deps         *ChatHandlerDeps
}

// NewChatTestFixtures creates a new set of test fixtures.
func NewChatTestFixtures() *ChatTestFixtures {
	mockService := NewMockChatService()
	f := common.NewTestFixtures[ChatService](mockService)

	return &ChatTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
