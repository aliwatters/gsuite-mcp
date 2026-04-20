package drive

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type DriveHandlerDeps = common.HandlerDeps[DriveService]

// NewDriveServiceConstructor returns a ServiceConstructor that creates a DriveService,
// wrapping with access control if a DriveAccessFilter is provided.
// Using a constructor factory allows the DriveAccessFilter to be passed explicitly
// rather than read from the global deps singleton.
func NewDriveServiceConstructor(filter *common.DriveAccessFilter) common.ServiceConstructor[DriveService] {
	return func(ctx context.Context, client *http.Client) (DriveService, error) {
		srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			return nil, fmt.Errorf("creating drive service: %w", err)
		}
		real := NewRealDriveService(srv)
		if filter != nil && filter.IsActive() {
			return NewFilteredDriveService(real, filter), nil
		}
		return real, nil
	}
}

// NewDriveService creates a DriveService from an authenticated HTTP client.
// If a DriveAccessFilter is configured in the global deps, the service is wrapped
// with access control. Prefer NewDriveServiceConstructor for explicit dependency passing.
func NewDriveService(ctx context.Context, client *http.Client) (DriveService, error) {
	d := common.GetDeps()
	var filter *common.DriveAccessFilter
	if d != nil {
		filter = d.DriveAccessFilter
	}
	return NewDriveServiceConstructor(filter)(ctx, client)
}

// InitDefaultDriveHandlerDeps initializes the default Drive handler deps with explicit deps,
// avoiding reliance on the global singleton at call time.
func InitDefaultDriveHandlerDeps(appDeps *common.Deps) {
	var filter *common.DriveAccessFilter
	if appDeps != nil {
		filter = appDeps.DriveAccessFilter
	}
	DefaultDriveHandlerDeps = common.NewDefaultHandlerDeps(NewDriveServiceConstructor(filter), appDeps)
}

// DefaultDriveHandlerDeps holds the default dependencies for production use.
// Initialize with InitDefaultDriveHandlerDeps after SetDeps to pass deps explicitly.
var DefaultDriveHandlerDeps = common.NewDefaultHandlerDeps(NewDriveService)

// ResolveDriveServiceOrError resolves a Drive service, returning an MCP error result on failure.
func ResolveDriveServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (DriveService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultDriveHandlerDeps)
}
