package slides

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/option"
	slides "google.golang.org/api/slides/v1"
)

// Type alias using generic types from common package.
type SlidesHandlerDeps = common.HandlerDeps[SlidesService]

// NewSlidesService creates a SlidesService from an authenticated HTTP client.
func NewSlidesService(ctx context.Context, client *http.Client) (SlidesService, error) {
	srv, err := slides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealSlidesService(srv), nil
}

// DefaultSlidesHandlerDeps holds the default dependencies for production use.
var DefaultSlidesHandlerDeps = common.NewDefaultHandlerDeps(NewSlidesService)

// ResolveSlidesServiceOrError resolves a Slides service, returning an MCP error result on failure.
func ResolveSlidesServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (SlidesService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultSlidesHandlerDeps)
}
