package sheets

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	return NewRealSheetsService(srv), nil
}

// InitDefaultSheetsHandlerDeps initializes the default Sheets handler deps with explicit deps,
// avoiding reliance on the global singleton at call time.
func InitDefaultSheetsHandlerDeps(appDeps *common.Deps) {
	DefaultSheetsHandlerDeps = common.NewDefaultHandlerDeps(NewSheetsService, appDeps)
}

// DefaultSheetsHandlerDeps holds the default dependencies for production use.
// Initialize with InitDefaultSheetsHandlerDeps after SetDeps to pass deps explicitly.
var DefaultSheetsHandlerDeps = common.NewDefaultHandlerDeps(NewSheetsService)

// ResolveSheetsServiceOrError resolves a Sheets service, returning an MCP error result on failure.
func ResolveSheetsServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *SheetsHandlerDeps) (SheetsService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultSheetsHandlerDeps)
}
