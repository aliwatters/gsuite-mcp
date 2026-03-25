package chat

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Chat tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Spaces ===

	// chat_list_spaces - List Chat spaces
	s.AddTool(mcp.NewTool("chat_list_spaces",
		mcp.WithDescription("List Google Chat spaces the authenticated user is a member of."),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleChatListSpaces)

	// chat_get_space - Get space details
	s.AddTool(mcp.NewTool("chat_get_space",
		mcp.WithDescription("Get details about a Google Chat space including display name, type, and threading state."),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Space resource name (e.g., 'spaces/AAAA1234')")),
		common.WithAccountParam(),
	), HandleChatGetSpace)

	// chat_create_space - Create a new space
	s.AddTool(mcp.NewTool("chat_create_space",
		mcp.WithDescription("Create a new Google Chat space."),
		mcp.WithString("display_name", mcp.Required(), mcp.Description("Display name for the space")),
		mcp.WithString("space_type", mcp.Description("Space type: SPACE (default) or GROUP_CHAT")),
		common.WithAccountParam(),
	), HandleChatCreateSpace)

	// === Messages ===

	// chat_list_messages - List messages in a space
	s.AddTool(mcp.NewTool("chat_list_messages",
		mcp.WithDescription("List messages in a Google Chat space. Returns messages in reverse chronological order."),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Space resource name (e.g., 'spaces/AAAA1234')")),
		mcp.WithString("filter", mcp.Description("Filter messages (e.g., 'createTime > \"2024-01-01T00:00:00Z\"')")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleChatListMessages)

	// chat_get_message - Get a specific message
	s.AddTool(mcp.NewTool("chat_get_message",
		mcp.WithDescription("Get a specific Google Chat message by its resource name."),
		mcp.WithString("message_name", mcp.Required(), mcp.Description("Message resource name (e.g., 'spaces/AAAA1234/messages/msg123')")),
		common.WithAccountParam(),
	), HandleChatGetMessage)

	// chat_send_message - Send a message to a space
	s.AddTool(mcp.NewTool("chat_send_message",
		mcp.WithDescription("Send a text message to a Google Chat space. Optionally reply in a thread."),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Space resource name (e.g., 'spaces/AAAA1234')")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message text content")),
		mcp.WithString("thread_name", mcp.Description("Thread resource name to reply in (e.g., 'spaces/AAAA1234/threads/thread123')")),
		common.WithAccountParam(),
	), HandleChatSendMessage)

	// === Reactions ===

	// chat_create_reaction - Add a reaction to a message
	s.AddTool(mcp.NewTool("chat_create_reaction",
		mcp.WithDescription("Add a Unicode emoji reaction to a Google Chat message."),
		mcp.WithString("message_name", mcp.Required(), mcp.Description("Message resource name (e.g., 'spaces/AAAA1234/messages/msg123')")),
		mcp.WithString("emoji", mcp.Required(), mcp.Description("Unicode emoji (e.g., '\U0001F44D', '\u2764\uFE0F')")),
		common.WithAccountParam(),
	), HandleChatCreateReaction)

	// chat_delete_reaction - Remove a reaction
	s.AddTool(mcp.NewTool("chat_delete_reaction",
		mcp.WithDescription("Remove a reaction from a Google Chat message."),
		mcp.WithString("reaction_name", mcp.Required(), mcp.Description("Reaction resource name (e.g., 'spaces/AAAA1234/messages/msg123/reactions/react456')")),
		common.WithAccountParam(),
	), HandleChatDeleteReaction)

	// === Members ===

	// chat_list_members - List members of a space
	s.AddTool(mcp.NewTool("chat_list_members",
		mcp.WithDescription("List members of a Google Chat space including their roles."),
		mcp.WithString("space_name", mcp.Required(), mcp.Description("Space resource name (e.g., 'spaces/AAAA1234')")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleChatListMembers)
}
