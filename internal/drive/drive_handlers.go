package drive

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Type aliases using generic types from common package.
type DriveHandlerDeps = common.HandlerDeps[DriveService]
type DriveServiceFactory = common.ServiceFactory[DriveService]
type MockDriveServiceFactory = common.MockServiceFactory[DriveService]

// NewDriveService creates a DriveService from an authenticated HTTP client.
func NewDriveService(ctx context.Context, client *http.Client) (DriveService, error) {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealDriveService(srv), nil
}

// DefaultDriveHandlerDeps holds the default dependencies for production use.
var DefaultDriveHandlerDeps = common.NewDefaultHandlerDeps(NewDriveService)

// ResolveDriveServiceOrError resolves a Drive service, returning an MCP error result on failure.
func ResolveDriveServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (DriveService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultDriveHandlerDeps)
}
