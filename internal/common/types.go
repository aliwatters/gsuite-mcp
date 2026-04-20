package common

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// MethodCall records a method call for test verification.
// Args is a slice to support both positional and named arguments:
// positional: Args = []any{arg1, arg2}
// named: Args = []any{map[string]any{"key": value}}
type MethodCall struct {
	Method string
	Args   []any
}

// GetLastCall returns the last call from a slice of MethodCalls, or nil if empty.
func GetLastCall(calls []MethodCall) *MethodCall {
	if len(calls) == 0 {
		return nil
	}
	return &calls[len(calls)-1]
}

// WasMethodCalled checks if a method was called in the given call slice.
func WasMethodCalled(calls []MethodCall, method string) bool {
	for _, call := range calls {
		if call.Method == method {
			return true
		}
	}
	return false
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
