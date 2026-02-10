package sheets

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Type alias using generic types from common package.
type SheetsHandlerDeps = common.HandlerDeps[SheetsService]

// NewSheetsService creates a SheetsService from an authenticated HTTP client.
func NewSheetsService(ctx context.Context, client *http.Client) (SheetsService, error) {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealSheetsService(srv), nil
}

// DefaultSheetsHandlerDeps holds the default dependencies for production use.
var DefaultSheetsHandlerDeps = common.NewDefaultHandlerDeps(NewSheetsService)

// ResolveSheetsServiceOrError resolves a Sheets service, returning an MCP error result on failure.
func ResolveSheetsServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (SheetsService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultSheetsHandlerDeps)
}
