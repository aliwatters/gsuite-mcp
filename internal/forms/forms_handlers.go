package forms

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type FormsHandlerDeps = common.HandlerDeps[FormsService]

// NewFormsService creates a FormsService from an authenticated HTTP client.
func NewFormsService(ctx context.Context, client *http.Client) (FormsService, error) {
	srv, err := forms.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealFormsService(srv), nil
}

// DefaultFormsHandlerDeps holds the default dependencies for production use.
var DefaultFormsHandlerDeps = common.NewDefaultHandlerDeps(NewFormsService)

// ResolveFormsServiceOrError resolves a Forms service, returning an MCP error result on failure.
func ResolveFormsServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (FormsService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultFormsHandlerDeps)
}
