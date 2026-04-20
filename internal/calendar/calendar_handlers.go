package calendar

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Type alias using generic types from common package.
type CalendarHandlerDeps = common.HandlerDeps[CalendarService]

// NewCalendarService creates a CalendarService from an authenticated HTTP client.
func NewCalendarService(ctx context.Context, client *http.Client) (CalendarService, error) {
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealCalendarService(srv), nil
}

// InitDefaultCalendarHandlerDeps initializes the default Calendar handler deps with explicit deps,
// avoiding reliance on the global singleton at call time.
func InitDefaultCalendarHandlerDeps(appDeps *common.Deps) {
	DefaultCalendarHandlerDeps = common.NewDefaultHandlerDeps(NewCalendarService, appDeps)
}

// DefaultCalendarHandlerDeps holds the default dependencies for production use.
// Initialize with InitDefaultCalendarHandlerDeps after SetDeps to pass deps explicitly.
var DefaultCalendarHandlerDeps = common.NewDefaultHandlerDeps(NewCalendarService)

// ResolveCalendarServiceOrError resolves a Calendar service, returning an MCP error result on failure.
func ResolveCalendarServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (CalendarService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultCalendarHandlerDeps)
}
