// Package mcptest provides helpers for MCP protocol-level testing.
// It creates a fully registered MCP server (without auth/credentials)
// for verifying tool registration, schemas, and protocol behavior.
package mcptest

import (
	"github.com/aliwatters/gsuite-mcp/internal/calendar"
	"github.com/aliwatters/gsuite-mcp/internal/contacts"
	"github.com/aliwatters/gsuite-mcp/internal/docs"
	"github.com/aliwatters/gsuite-mcp/internal/drive"
	"github.com/aliwatters/gsuite-mcp/internal/forms"
	"github.com/aliwatters/gsuite-mcp/internal/gmail"
	"github.com/aliwatters/gsuite-mcp/internal/sheets"
	"github.com/aliwatters/gsuite-mcp/internal/slides"
	"github.com/aliwatters/gsuite-mcp/internal/tasks"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "gsuite-mcp"
	ServerVersion = "0.1.0"
)

// ServiceToolCounts maps each service to its expected tool count.
// Update these when adding/removing tools.
var ServiceToolCounts = map[string]int{
	"gmail":    50,
	"calendar": 12,
	"drive":    23,
	"docs":     24,
	"sheets":   16,
	"slides":   5,
	"forms":    5,
	"tasks":    12,
	"contacts": 12,
}

// TotalToolCount returns the sum of all service tool counts.
func TotalToolCount() int {
	total := 0
	for _, count := range ServiceToolCounts {
		total += count
	}
	return total
}

// NewTestServer creates a fully registered MCP server for testing.
// No auth manager or credentials are needed — tools are registered
// but will return auth errors when called without credentials.
func NewTestServer() *server.MCPServer {
	s := server.NewMCPServer(ServerName, ServerVersion)

	gmail.RegisterTools(s)
	calendar.RegisterTools(s)
	docs.RegisterTools(s)
	tasks.RegisterTools(s)
	drive.RegisterTools(s)
	sheets.RegisterTools(s)
	slides.RegisterTools(s)
	forms.RegisterTools(s)
	contacts.RegisterTools(s)

	return s
}
