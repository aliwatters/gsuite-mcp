package common

import "github.com/mark3labs/mcp-go/mcp"

// WithAccountParam returns the standard account ToolOption used by all tools.
func WithAccountParam() mcp.ToolOption {
	return mcp.WithString("account", mcp.Description("Account label or email (uses default if omitted)"))
}

// WithPageToken returns the standard page_token ToolOption for paginated tools.
func WithPageToken() mcp.ToolOption {
	return mcp.WithString("page_token", mcp.Description("Token for pagination (from previous response)"))
}
