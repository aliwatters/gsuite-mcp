package common

import (
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Deps holds shared dependencies for all service handlers.
type Deps struct {
	AuthManager *auth.Manager
}

// Global instance set during initialization
var deps *Deps

// SetDeps initializes the global dependencies.
func SetDeps(d *Deps) {
	deps = d
}

// GetDeps returns the global dependencies.
func GetDeps() *Deps {
	return deps
}

// ResolveAccountFromRequest extracts and validates the account parameter.
func ResolveAccountFromRequest(request mcp.CallToolRequest) (string, error) {
	accountParam, _ := request.Params.Arguments["account"].(string)

	if accountParam == "" {
		// No account specified - use first authenticated email
		email := config.GetDefaultEmail()
		if email == "" {
			return "", fmt.Errorf("no authenticated accounts found; run 'gsuite-mcp auth' to authenticate")
		}
		return email, nil
	}

	// Account param specified - check if credentials exist
	if config.HasCredentialsForEmail(accountParam) {
		return accountParam, nil
	}

	return "", fmt.Errorf("no credentials for %s; run 'gsuite-mcp auth' and sign in with that account", accountParam)
}
