package drive

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type DriveHandlerDeps = common.HandlerDeps[DriveService]

// NewDriveService creates a DriveService from an authenticated HTTP client.
// If a DriveAccessFilter is configured, the service is wrapped with access control.
func NewDriveService(ctx context.Context, client *http.Client) (DriveService, error) {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	real := NewRealDriveService(srv)

	d := common.GetDeps()
	if d != nil && d.DriveAccessFilter != nil && d.DriveAccessFilter.IsActive() {
		return NewFilteredDriveService(real, d.DriveAccessFilter), nil
	}

	return real, nil
}

// DefaultDriveHandlerDeps holds the default dependencies for production use.
var DefaultDriveHandlerDeps = common.NewDefaultHandlerDeps(NewDriveService)

// ResolveDriveServiceOrError resolves a Drive service, returning an MCP error result on failure.
func ResolveDriveServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (DriveService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultDriveHandlerDeps)
}
