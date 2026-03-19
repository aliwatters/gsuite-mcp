package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/calendar"
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/aliwatters/gsuite-mcp/internal/config"
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
	serverName    = "gsuite-mcp"
	serverVersion = "0.1.0"
)

func main() {
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
		case "help", "--help", "-h":
			printUsage()
			return
		case "version", "--version", "-v":
			fmt.Printf("%s %s\n", serverName, serverVersion)
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

	s := server.NewMCPServer(serverName, serverVersion)

	// Register all service tools
	gmail.RegisterTools(s)
	calendar.RegisterTools(s)
	docs.RegisterTools(s)
	tasks.RegisterTools(s)
	drive.RegisterTools(s)
	sheets.RegisterTools(s)
	slides.RegisterTools(s)
	forms.RegisterTools(s)
	contacts.RegisterTools(s)

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
		return err
	}

	// Load config for drive access filtering (optional)
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
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
	common.SetDeps(&common.Deps{
		AuthManager:       authManager,
		DriveAccessFilter: driveFilter,
	})

	return nil
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

No configuration required - just authenticate any Google account on demand.
When tools request an account without credentials, auth flow is triggered automatically.

Configuration:
  Config dir:     %s
  Config file:    %s
  Credentials:    %s
  Client secret:  %s

Environment variables:
  GSUITE_MCP_OAUTH_PORT   Override OAuth callback port (default: %d)

For more information, see README.md
`, serverName, serverName, serverName, serverName, serverName, serverName,
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

// runAccounts lists authenticated accounts (discovered from credentials directory).
func runAccounts() {
	emails, err := config.GetAuthenticatedEmails()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading credentials: %v\n", err)
		os.Exit(1)
	}

	if len(emails) == 0 {
		fmt.Printf("No authenticated accounts found.\n")
		fmt.Printf("\nRun '%s auth' to authenticate a Google account.\n", serverName)
		return
	}

	fmt.Printf("Authenticated accounts:\n\n")
	for _, email := range emails {
		fmt.Printf("  %s\n", email)
	}
	fmt.Printf("\nRun '%s auth' to add another account.\n", serverName)
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
