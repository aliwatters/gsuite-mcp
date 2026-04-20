package driveactivity

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/driveactivity/v2"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type DriveActivityHandlerDeps = common.HandlerDeps[DriveActivityService]

// NewDriveActivityService creates a DriveActivityService from an authenticated HTTP client.
func NewDriveActivityService(ctx context.Context, client *http.Client) (DriveActivityService, error) {
	srv, err := driveactivity.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating drive activity service: %w", err)
	}
	return NewRealDriveActivityService(srv), nil
}

// InitDefaultDriveActivityHandlerDeps initializes the default DriveActivity handler deps with explicit deps.
func InitDefaultDriveActivityHandlerDeps(appDeps *common.Deps) {
	DefaultDriveActivityHandlerDeps = common.NewDefaultHandlerDeps(NewDriveActivityService, appDeps)
}

// DefaultDriveActivityHandlerDeps holds the default dependencies for production use.
var DefaultDriveActivityHandlerDeps = common.NewDefaultHandlerDeps(NewDriveActivityService)

// ResolveDriveActivityServiceOrError resolves a Drive Activity service, returning an MCP error result on failure.
func ResolveDriveActivityServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *DriveActivityHandlerDeps) (DriveActivityService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultDriveActivityHandlerDeps)
}
