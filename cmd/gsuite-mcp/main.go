package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/calendar"
	"github.com/aliwatters/gsuite-mcp/internal/chat"
	"github.com/aliwatters/gsuite-mcp/internal/citation"
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/aliwatters/gsuite-mcp/internal/contacts"
	"github.com/aliwatters/gsuite-mcp/internal/docs"
	"github.com/aliwatters/gsuite-mcp/internal/drive"
	"github.com/aliwatters/gsuite-mcp/internal/driveactivity"
	"github.com/aliwatters/gsuite-mcp/internal/forms"
	"github.com/aliwatters/gsuite-mcp/internal/gmail"
	"github.com/aliwatters/gsuite-mcp/internal/meet"
	"github.com/aliwatters/gsuite-mcp/internal/sheets"
	"github.com/aliwatters/gsuite-mcp/internal/slides"
	"github.com/aliwatters/gsuite-mcp/internal/tasks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "gsuite-mcp"
	serverVersion = "0.6.1"
)

// GitCommit is injected at build time via (run from the repo root):
//
//	go build -ldflags "-X main.GitCommit=$(git rev-parse --short HEAD)" ./cmd/gsuite-mcp/
var GitCommit string

func main() {
	// Parse --config-dir flag or GSUITE_MCP_CONFIG_DIR env var before subcommand dispatch.
	// This must happen first so all config.* path functions resolve correctly.
	parseConfigDir()

	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			runInit()
			return
		case "auth":
			runAuth()
			return
		case "accounts":
			runAccounts()
			return
		case "check":
			runCheck()
			return
		case "call":
			os.Exit(runCall(os.Args[2:], os.Stdout, os.Stderr))
		case "help", "--help", "-h":
			printUsage()
			return
		case "version", "--version", "-v":
			if GitCommit != "" {
				fmt.Printf("%s %s (%s)\n", serverName, serverVersion, GitCommit)
			} else {
				fmt.Printf("%s %s\n", serverName, serverVersion)
			}
			return
		}
	}

	// Initialize config and auth for MCP server mode
	if err := initializeApp(); err != nil {
		fmt.Fprintf(os.Stderr, "Initialization error: %v\n", err)
		os.Exit(1)
	}

	// Start persistent HTTP auth server (non-fatal if port is unavailable)
	authSrv := startAuthServer()
	if authSrv != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			authSrv.Shutdown(ctx)
		}()
	}

	// Write the pidfile so `gsuite-mcp auth` can reliably identify us as the
	// holder of oauth_port and take over instead of failing with EADDRINUSE
	// (#158). Best-effort: a missing pidfile triggers a process-name fallback,
	// so a write failure shouldn't crash the daemon.
	if err := auth.WritePidFile(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write pidfile: %v\n", err)
	} else {
		defer func() {
			if err := auth.RemovePidFile(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not remove pidfile on shutdown: %v\n", err)
			}
		}()
	}

	s := newRegisteredServer()

	// Start server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// initializeApp sets up the auth manager, drive access filter, and shared dependencies.
// No configuration required - uses dynamic credential discovery.
func initializeApp() error {
	authManager, err := auth.NewManager()
	if err != nil {
		return fmt.Errorf("initializing auth manager: %w", err)
	}

	// Load config for drive access filtering (optional)
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	driveFilter := common.NewDriveAccessFilter(cfg.DriveAccess)
	if driveFilter != nil && driveFilter.IsActive() {
		fmt.Fprintf(os.Stderr, "drive access filter active")
		if cfg.DriveAccess != nil {
			if len(cfg.DriveAccess.Allowed) > 0 {
				fmt.Fprintf(os.Stderr, " (allowed: %v)", cfg.DriveAccess.Allowed)
			}
			if len(cfg.DriveAccess.Blocked) > 0 {
				fmt.Fprintf(os.Stderr, " (blocked: %v)", cfg.DriveAccess.Blocked)
			}
		}
		fmt.Fprintln(os.Stderr)
	}

	// Set up shared dependencies for all packages
	appDeps := &common.Deps{
		AuthManager:       authManager,
		DriveAccessFilter: driveFilter,
	}
	// SetDeps retained for backward compatibility with WithDriveAccessCheck/WithLargeContentHint
	// middleware that may receive nil deps. All handler factories now use explicit passing below.
	common.SetDeps(appDeps)

	// Initialize handler deps with explicit dependency passing so handler logic does not call GetDeps().
	drive.InitDefaultDriveHandlerDeps(appDeps)
	gmail.InitDefaultGmailHandlerDeps(appDeps)
	docs.InitDefaultDocsHandlerDeps(appDeps)
	sheets.InitDefaultSheetsHandlerDeps(appDeps)
	calendar.InitDefaultCalendarHandlerDeps(appDeps)
	chat.InitDefaultChatHandlerDeps(appDeps)
	contacts.InitDefaultContactsHandlerDeps(appDeps)
	driveactivity.InitDefaultDriveActivityHandlerDeps(appDeps)
	forms.InitDefaultFormsHandlerDeps(appDeps)
	meet.InitDefaultMeetHandlerDeps(appDeps)
	slides.InitDefaultSlidesHandlerDeps(appDeps)
	tasks.InitDefaultTasksHandlerDeps(appDeps)

	return nil
}

// newRegisteredServer creates the same in-process tool registry used by stdio and call mode.
func newRegisteredServer() *server.MCPServer {
	s := server.NewMCPServer(serverName, serverVersion)
	gmail.RegisterTools(s)
	calendar.RegisterTools(s)
	docs.RegisterTools(s)
	tasks.RegisterTools(s)
	drive.RegisterTools(s)
	sheets.RegisterTools(s)
	slides.RegisterTools(s)
	forms.RegisterTools(s)
	contacts.RegisterTools(s)
	meet.RegisterTools(s)
	driveactivity.RegisterTools(s)
	chat.RegisterTools(s)
	registerCitationIfEnabled(s)
	return s
}

const (
	callExitError        = 1
	callExitAuthRequired = 2
)

func runCall(args []string, stdout, stderr io.Writer) int {
	toolName, flagArgs := extractCallToolName(args)
	flags := flag.NewFlagSet("call", flag.ContinueOnError)
	flags.SetOutput(stderr)
	account := flags.String("account", "", "Google account email")
	argsJSON := flags.String("args", "", "Tool arguments as a JSON object")
	argsFile := flags.String("args-file", "", "Read tool arguments from a JSON file")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Usage: gsuite-mcp call <tool> --account <email> (--args '<json>' | --args-file <path>)")
		fmt.Fprintln(stderr, "Exit codes: 0 success, 1 tool/argument error, 2 authentication required")
		flags.PrintDefaults()
	}
	if err := flags.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return callExitError
	}
	if toolName == "" || *account == "" || (*argsJSON == "" && *argsFile == "") || (*argsJSON != "" && *argsFile != "") {
		flags.Usage()
		return callExitError
	}
	var raw []byte
	var err error
	if *argsFile != "" {
		raw, err = os.ReadFile(*argsFile)
	} else {
		raw = []byte(*argsJSON)
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s account=%s: reading arguments: %v\n", toolName, *account, err)
		return callExitError
	}
	toolArgs := make(map[string]any)
	if err := json.Unmarshal(raw, &toolArgs); err != nil {
		fmt.Fprintf(stderr, "%s account=%s: invalid JSON arguments: %v\n", toolName, *account, err)
		return callExitError
	}
	if toolArgs == nil {
		fmt.Fprintf(stderr, "%s account=%s: arguments must be a JSON object\n", toolName, *account)
		return callExitError
	}
	toolArgs["account"] = *account

	registry := newRegisteredServer()
	if registry.GetTool(toolName) == nil {
		fmt.Fprintf(stderr, "%s account=%s: unknown tool\n", toolName, *account)
		return callExitError
	}
	if err := initializeApp(); err != nil {
		fmt.Fprintf(stderr, "%s account=%s: initialization failed: %v\n", toolName, *account, err)
		return callExitError
	}
	log.SetOutput(stderr)
	return executeCall(registry, toolName, *account, toolArgs, stdout, stderr)
}

// extractCallToolName removes the positional tool name so Go's flag parser can
// accept flags both before and after it.
func extractCallToolName(args []string) (string, []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--account" || arg == "--args" || arg == "--args-file" {
			i++
			continue
		}
		if strings.HasPrefix(arg, "--account=") || strings.HasPrefix(arg, "--args=") || strings.HasPrefix(arg, "--args-file=") {
			continue
		}
		if !strings.HasPrefix(arg, "-") {
			remaining := append([]string(nil), args[:i]...)
			remaining = append(remaining, args[i+1:]...)
			return arg, remaining
		}
	}
	return "", args
}

func executeCall(registry *server.MCPServer, toolName, account string, toolArgs map[string]any, stdout, stderr io.Writer) int {
	registered := registry.GetTool(toolName)
	if registered == nil {
		fmt.Fprintf(stderr, "%s account=%s: unknown tool\n", toolName, account)
		return callExitError
	}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: toolName, Arguments: toolArgs}}
	result, err := registered.Handler(context.Background(), request)
	if err != nil {
		fmt.Fprintf(stderr, "%s account=%s: %v\n", toolName, account, err)
		return classifyCallError(err.Error())
	}
	if result == nil {
		fmt.Fprintf(stderr, "%s account=%s: tool returned no result\n", toolName, account)
		return callExitError
	}
	if result.IsError {
		reason := toolResultText(result)
		if classifyCallError(reason) == callExitAuthRequired && !strings.Contains(reason, "gsuite-mcp auth "+account) {
			reason += fmt.Sprintf("; re-authenticate with 'gsuite-mcp auth %s'", account)
		}
		fmt.Fprintf(stderr, "%s account=%s: %s\n", toolName, account, reason)
		return classifyCallError(reason)
	}
	if result.StructuredContent != nil {
		if err := json.NewEncoder(stdout).Encode(result.StructuredContent); err != nil {
			fmt.Fprintf(stderr, "%s account=%s: encoding result: %v\n", toolName, account, err)
			return callExitError
		}
		return 0
	}
	text := toolResultText(result)
	if text == "" {
		text = "null"
	}
	if _, err := fmt.Fprintln(stdout, text); err != nil {
		fmt.Fprintf(stderr, "%s account=%s: writing result: %v\n", toolName, account, err)
		return callExitError
	}
	return 0
}

func toolResultText(result *mcp.CallToolResult) string {
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}

func classifyCallError(reason string) int {
	lower := strings.ToLower(reason)
	if strings.Contains(lower, "auth_expired") || strings.Contains(lower, "invalid_grant") || strings.Contains(lower, "no credentials") || strings.Contains(lower, "re-authenticate") || strings.Contains(lower, "reauth") {
		return callExitAuthRequired
	}
	return callExitError
}

// printUsage prints CLI help.
func printUsage() {
	fmt.Printf(`%s - Google Workspace MCP server with dynamic multi-account support

Usage:
  %s              Start MCP server (for Claude/AI integration)
  %s init         Create initial configuration
  %s auth         Authenticate a Google account (opens browser)
  %s accounts     List authenticated accounts
  %s check        Verify setup (config, tokens, API access)
  %s call <tool> --account <email> (--args '<json>' | --args-file <path>)
                    Call a registered tool without starting an auth flow
                    Exit codes: 0 success, 1 tool/argument error, 2 authentication required

No configuration required - just authenticate any Google account on demand.
When tools request an account without credentials, auth flow is triggered automatically.

Flags:
  --config-dir DIR  Use DIR for config, credentials, and client_secret.json
                    (default: %s)

Configuration:
  Config dir:     %s
  Config file:    %s
  Credentials:    %s
  Client secret:  %s

Environment variables:
  GSUITE_MCP_CONFIG_DIR    Override config directory (same as --config-dir)
  GSUITE_MCP_OAUTH_PORT    Override OAuth callback port (default: %d)
  GSUITE_OAUTH_TOKEN_CMD   Command to retrieve an OAuth refresh token from an external
                            secret source. The command receives the account email as its
                            first argument and must print the token to stdout; non-zero
                            exit is treated as an error. When unset, tokens are read from
                            local JSON files (default behaviour).

For more information, see README.md
`, serverName, serverName, serverName, serverName, serverName, serverName, serverName,
		config.DefaultConfigDir(),
		config.DefaultConfigDir(), config.ConfigPath(), config.CredentialsDir(), config.ClientSecretPath(),
		config.DefaultOAuthPort)
}

// runInit ensures config directory exists, creates default config, and shows setup instructions.
func runInit() {
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config directory ready: %s\n", config.DefaultConfigDir())

	created, err := config.WriteDefaultConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config.json: %v\n", err)
		os.Exit(1)
	}
	if created {
		fmt.Printf("Created config.json:    %s\n", config.ConfigPath())
	} else {
		fmt.Printf("Config.json exists:     %s\n", config.ConfigPath())
	}

	fmt.Printf("\nSetup steps:\n")
	fmt.Printf("1. Create OAuth credentials in Google Cloud Console\n")
	fmt.Printf("2. Download and save as: %s\n", config.ClientSecretPath())
	fmt.Printf("3. Run '%s auth' to authenticate your Google account\n", serverName)
}

// runAuth authenticates a Google account using dynamic OAuth flow.
// No pre-configuration required - just opens browser and saves credentials by email.
func runAuth() {
	mgr, err := auth.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	email, err := mgr.AuthenticateDynamic(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully authenticated: %s\n", email)
}

// accountInfo holds per-account display information for the accounts command.
type accountInfo struct {
	Email  string `json:"email"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// runAccounts lists authenticated accounts with token health status.
// Flags:
//
//	--json   output machine-readable JSON
func runAccounts() {
	jsonMode := false
	for _, arg := range os.Args[2:] {
		if arg == "--json" {
			jsonMode = true
		}
	}

	emails, err := config.GetAuthenticatedEmails()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		os.Exit(1)
	}

	if len(emails) == 0 {
		if jsonMode {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode([]accountInfo{}) //nolint:errcheck
		} else {
			fmt.Printf("No authenticated accounts found.\n")
			fmt.Printf("\nRun '%s auth' to authenticate a Google account.\n", serverName)
		}
		return
	}

	mgr, err := auth.NewManager()
	if err != nil {
		// If we can't build the manager (missing client_secret.json), list accounts without health.
		if jsonMode {
			var infos []accountInfo
			for _, e := range emails {
				infos = append(infos, accountInfo{Email: e, Status: "unknown", Detail: "cannot read client_secret.json"})
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(infos) //nolint:errcheck
		} else {
			fmt.Printf("Authenticated accounts:\n\n")
			for _, e := range emails {
				fmt.Printf("  %s  [status unknown — cannot read client_secret.json]\n", e)
			}
			fmt.Printf("\nRun '%s auth' to add another account.\n", serverName)
		}
		return
	}

	ctx := context.Background()
	var infos []accountInfo

	for _, email := range emails {
		_, clientErr := mgr.GetClientForEmail(ctx, email)
		info := accountInfo{Email: email}
		switch {
		case clientErr == nil:
			info.Status = "valid"
		case errors.Is(clientErr, auth.ErrAuthExpired):
			info.Status = "expired"
			info.Detail = fmt.Sprintf("re-auth required: run 'gsuite-mcp auth'")
		default:
			info.Status = "error"
			info.Detail = clientErr.Error()
		}
		infos = append(infos, info)
	}

	if jsonMode {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(infos) //nolint:errcheck
		return
	}

	fmt.Printf("Authenticated accounts:\n\n")
	needsReauth := false
	for _, info := range infos {
		switch info.Status {
		case "valid":
			fmt.Printf("  ✓  %s  [valid]\n", info.Email)
		case "expired":
			fmt.Printf("  ✗  %s  [expired — re-auth required]\n", info.Email)
			needsReauth = true
		default:
			fmt.Printf("  ?  %s  [error: %s]\n", info.Email, info.Detail)
		}
	}

	fmt.Printf("\nRun '%s auth' to add another account.\n", serverName)
	if needsReauth {
		fmt.Printf("Run '%s auth' again for each expired account to re-authenticate.\n", serverName)
	}
}

// registerCitationIfEnabled registers citation tools if the large_doc_indexing feature is enabled.
func registerCitationIfEnabled(s *server.MCPServer) {
	cfg, err := config.LoadConfig()
	if err != nil || cfg.Features == nil || !cfg.Features.LargeDocIndexing {
		return
	}

	var citCfg *citation.CitationConfig
	if cfg.Citation != nil {
		citCfg = &citation.CitationConfig{
			Indexes: make(map[string]citation.IndexEntry),
		}
		for k, v := range cfg.Citation.Indexes {
			citCfg.Indexes[k] = citation.IndexEntry{SheetID: v.SheetID}
		}
	}

	citation.InitDefaultDeps(citCfg)
	citation.RegisterTools(s)

	// Enable citation hints on large content responses
	if d := common.GetDeps(); d != nil {
		d.CitationEnabled = true
	}

	fmt.Fprintln(os.Stderr, "citation tools enabled (large_doc_indexing)")
}

// parseConfigDir checks for --config-dir flag or GSUITE_MCP_CONFIG_DIR env var
// and sets the config directory override. The flag is consumed from os.Args so
// subcommand parsing is unaffected.
func parseConfigDir() {
	// Check for --config-dir flag (must appear before subcommand)
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--config-dir" && i+1 < len(os.Args) {
			dir := os.Args[i+1]
			absDir, err := filepath.Abs(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving --config-dir %q: %v\n", dir, err)
				os.Exit(1)
			}
			config.SetConfigDir(absDir)
			// Remove the flag and its value from os.Args
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			return
		}
	}

	// Fall back to environment variable
	if dir := os.Getenv("GSUITE_MCP_CONFIG_DIR"); dir != "" {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving GSUITE_MCP_CONFIG_DIR %q: %v\n", dir, err)
			os.Exit(1)
		}
		config.SetConfigDir(absDir)
	}
}

// startAuthServer starts a persistent HTTP auth server for browser-based re-authentication.
// Returns nil if the port cannot be bound (non-fatal — MCP server continues without it).
func startAuthServer() *auth.AuthServer {
	port, _, err := auth.ResolveOAuthPort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not resolve OAuth port: %v\n", err)
		return nil
	}

	d := common.GetDeps()
	srv := auth.NewAuthServer(d.AuthManager, port)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: auth server not started: %v\n", err)
		return nil
	}

	authURL := fmt.Sprintf("http://localhost:%d/auth", port)
	d.AuthManager.AuthServerURL = authURL
	fmt.Fprintf(os.Stderr, "auth server listening on %s\n", authURL)

	return srv
}
