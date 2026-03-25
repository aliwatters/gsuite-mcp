package chat

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleChatListSpaces     = common.WrapHandler[ChatService](TestableChatListSpaces)
	HandleChatGetSpace       = common.WrapHandler[ChatService](TestableChatGetSpace)
	HandleChatCreateSpace    = common.WrapHandler[ChatService](TestableChatCreateSpace)
	HandleChatListMessages   = common.WrapHandler[ChatService](TestableChatListMessages)
	HandleChatGetMessage     = common.WrapHandler[ChatService](TestableChatGetMessage)
	HandleChatSendMessage    = common.WrapHandler[ChatService](TestableChatSendMessage)
	HandleChatCreateReaction = common.WrapHandler[ChatService](TestableChatCreateReaction)
	HandleChatDeleteReaction = common.WrapHandler[ChatService](TestableChatDeleteReaction)
	HandleChatListMembers    = common.WrapHandler[ChatService](TestableChatListMembers)
)
