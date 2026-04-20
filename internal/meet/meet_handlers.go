package meet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	meet "google.golang.org/api/meet/v2"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type MeetHandlerDeps = common.HandlerDeps[MeetService]

// NewMeetService creates a MeetService from an authenticated HTTP client.
func NewMeetService(ctx context.Context, client *http.Client) (MeetService, error) {
	srv, err := meet.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating meet service: %w", err)
	}
	return NewRealMeetService(srv), nil
}

// DefaultMeetHandlerDeps holds the default dependencies for production use.
var DefaultMeetHandlerDeps = common.NewDefaultHandlerDeps(NewMeetService)

// ResolveMeetServiceOrError resolves a Meet service, returning an MCP error result on failure.
func ResolveMeetServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (MeetService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultMeetHandlerDeps)
}
