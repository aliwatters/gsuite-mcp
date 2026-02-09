package common

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
)

// AuthManagerInterface defines the interface for auth.Manager.
// This interface enables dependency injection for testing.
type AuthManagerInterface interface {
	GetClientForEmail(ctx context.Context, email string) (*http.Client, error)
	GetClientOrAuthenticate(ctx context.Context, email string, interactive bool) (*http.Client, error)
}

// MethodCall records a method call for test verification.
type MethodCall struct {
	Method string
	Args   map[string]any
}

// CreateMCPRequest creates an MCP request with the given arguments.
// This helper function is used across test files.
func CreateMCPRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Arguments: args,
		},
	}
}
