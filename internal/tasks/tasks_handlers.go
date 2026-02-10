package tasks

import (
	"context"
	"net/http"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

// Type alias using generic types from common package.
type TasksHandlerDeps = common.HandlerDeps[TasksService]

// NewTasksService creates a TasksService from an authenticated HTTP client.
func NewTasksService(ctx context.Context, client *http.Client) (TasksService, error) {
	srv, err := tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return NewRealTasksService(srv), nil
}

// DefaultTasksHandlerDeps holds the default dependencies for production use.
var DefaultTasksHandlerDeps = common.NewDefaultHandlerDeps(NewTasksService)

// ResolveTasksServiceOrError resolves a Tasks service, returning an MCP error result on failure.
func ResolveTasksServiceOrError(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (TasksService, *mcp.CallToolResult, bool) {
	return common.ResolveServiceOrError(ctx, request, deps, DefaultTasksHandlerDeps)
}
