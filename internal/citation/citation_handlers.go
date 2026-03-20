package citation

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// CitationHandlerDeps is the handler dependency type for citation tools.
type CitationHandlerDeps = common.HandlerDeps[CitationService]

// DefaultCitationHandlerDeps holds the default production dependencies.
// Initialized lazily because CitationConfig must be loaded first.
var DefaultCitationHandlerDeps *CitationHandlerDeps

// InitDefaultDeps initializes the default handler deps with the given config.
func InitDefaultDeps(cfg *CitationConfig) {
	constructor := func(ctx context.Context, client *http.Client) (CitationService, error) {
		return NewRealCitationService(ctx, client, cfg)
	}
	DefaultCitationHandlerDeps = common.NewDefaultHandlerDeps(constructor)
}

// ResolveCitationServiceOrError resolves a citation service from the request.
func ResolveCitationServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (CitationService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultCitationHandlerDeps)
}
