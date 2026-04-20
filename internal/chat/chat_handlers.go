package chat

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	chatapi "google.golang.org/api/chat/v1"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type ChatHandlerDeps = common.HandlerDeps[ChatService]

// NewChatService creates a ChatService from an authenticated HTTP client.
func NewChatService(ctx context.Context, client *http.Client) (ChatService, error) {
	srv, err := chatapi.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating chat service: %w", err)
	}
	return NewRealChatService(srv), nil
}

// InitDefaultChatHandlerDeps initializes the default Chat handler deps with explicit deps.
func InitDefaultChatHandlerDeps(appDeps *common.Deps) {
	DefaultChatHandlerDeps = common.NewDefaultHandlerDeps(NewChatService, appDeps)
}

// DefaultChatHandlerDeps holds the default dependencies for production use.
var DefaultChatHandlerDeps = common.NewDefaultHandlerDeps(NewChatService)

// ResolveChatServiceOrError resolves a Chat service, returning an MCP error result on failure.
func ResolveChatServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (ChatService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultChatHandlerDeps)
}
