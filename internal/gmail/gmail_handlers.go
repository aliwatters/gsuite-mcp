package gmail

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type GmailHandlerDeps = common.HandlerDeps[GmailService]

// NewGmailService creates a GmailService from an authenticated HTTP client.
func NewGmailService(ctx context.Context, client *http.Client) (GmailService, error) {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealGmailService(srv), nil
}

// InitDefaultGmailHandlerDeps initializes the default Gmail handler deps with explicit deps,
// avoiding reliance on the global singleton at call time.
func InitDefaultGmailHandlerDeps(appDeps *common.Deps) {
	DefaultGmailHandlerDeps = common.NewDefaultHandlerDeps(NewGmailService, appDeps)
}

// DefaultGmailHandlerDeps holds the default dependencies for production use.
// Initialize with InitDefaultGmailHandlerDeps after SetDeps to pass deps explicitly.
var DefaultGmailHandlerDeps = common.NewDefaultHandlerDeps(NewGmailService)

// ResolveGmailServiceOrError resolves a Gmail service, returning an MCP error result on failure.
func ResolveGmailServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (GmailService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultGmailHandlerDeps)
}
