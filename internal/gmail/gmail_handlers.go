package gmail

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Type aliases using generic types from common package.
type GmailHandlerDeps = common.HandlerDeps[GmailService]
type GmailServiceFactory = common.ServiceFactory[GmailService]
type MockGmailServiceFactory = common.MockServiceFactory[GmailService]

// NewGmailService creates a GmailService from an authenticated HTTP client.
func NewGmailService(ctx context.Context, client *http.Client) (GmailService, error) {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealGmailService(srv), nil
}

// DefaultGmailHandlerDeps holds the default dependencies for production use.
var DefaultGmailHandlerDeps = common.NewDefaultHandlerDeps(NewGmailService)

// ResolveGmailServiceOrError resolves a Gmail service, returning an MCP error result on failure.
func ResolveGmailServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (GmailService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultGmailHandlerDeps)
}
