package docs

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type DocsHandlerDeps = common.HandlerDeps[DocsService]

// NewDocsService creates a DocsService from an authenticated HTTP client.
func NewDocsService(ctx context.Context, client *http.Client) (DocsService, error) {
	srv, err := docs.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealDocsService(srv), nil
}

// DefaultDocsHandlerDeps holds the default dependencies for production use.
var DefaultDocsHandlerDeps = common.NewDefaultHandlerDeps(NewDocsService)

// ResolveDocsServiceOrError resolves a Docs service, returning an MCP error result on failure.
func ResolveDocsServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (DocsService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultDocsHandlerDeps)
}
