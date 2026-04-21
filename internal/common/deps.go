package common

import (
	"fmt"
	"net/url"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
)

// Deps holds shared dependencies for all service handlers.
type Deps struct {
	AuthManager       *auth.Manager
	DriveAccessFilter *DriveAccessFilter
	CitationEnabled   bool // true when large_doc_indexing feature is on
}

// Global instance set during initialization.
// Deprecated: All service handlers should receive deps explicitly via Init*HandlerDeps.
// This global is retained for middleware (WithDriveAccessCheck, WithLargeContentHint)
// fallbacks and for callers such as ResolveAccountFromRequest that have not yet been
// updated to accept deps explicitly.
var deps *Deps

// SetDeps initializes the global dependencies. Called exactly once during startup.
// Deprecated: Prefer passing *Deps explicitly to Init*HandlerDeps functions.
func SetDeps(d *Deps) {
	deps = d
}

// GetDeps returns the global dependencies.
// Deprecated: Prefer receiving *Deps explicitly. This remains for middleware fallback paths.
func GetDeps() *Deps {
	return deps
}

// ResolveAccountFromRequest extracts and validates the account parameter.
// It reads from the global deps singleton; use ResolveAccountFromRequestWithDeps
// to pass deps explicitly.
func ResolveAccountFromRequest(request mcp.CallToolRequest) (string, error) {
	return ResolveAccountFromRequestWithDeps(request, GetDeps())
}

// ResolveAccountFromRequestWithDeps extracts and validates the account parameter
// using explicitly provided dependencies. Passing nil falls back to the global singleton.
func ResolveAccountFromRequestWithDeps(request mcp.CallToolRequest, d *Deps) (string, error) {
	if d == nil {
		d = GetDeps()
	}
	accountParam := ParseStringArg(request.Params.Arguments, "account", "")

	if accountParam == "" {
		// No account specified - use first authenticated email
		email := config.GetDefaultEmail()
		if email == "" {
			if d != nil && d.AuthManager.AuthServerURL != "" {
				return "", fmt.Errorf("no authenticated accounts found; open %s to authenticate", d.AuthManager.AuthServerURL)
			}
			return "", fmt.Errorf("no authenticated accounts found; run 'gsuite-mcp auth' to authenticate")
		}
		return email, nil
	}

	// Account param specified - check if credentials exist
	if config.HasCredentialsForEmail(accountParam) {
		return accountParam, nil
	}

	if d != nil && d.AuthManager.AuthServerURL != "" {
		return "", fmt.Errorf("no credentials for %s; open %s?account=%s to authenticate", accountParam, d.AuthManager.AuthServerURL, url.QueryEscape(accountParam))
	}
	return "", fmt.Errorf("no credentials for %s; run 'gsuite-mcp auth' and sign in with that account", accountParam)
}
