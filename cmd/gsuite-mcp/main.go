package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/calendar"
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/aliwatters/gsuite-mcp/internal/contacts"
	"github.com/aliwatters/gsuite-mcp/internal/docs"
	"github.com/aliwatters/gsuite-mcp/internal/drive"
	"github.com/aliwatters/gsuite-mcp/internal/gmail"
	"github.com/aliwatters/gsuite-mcp/internal/sheets"
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

	s := server.NewMCPServer(serverName, serverVersion)

	// Register all service tools
	gmail.RegisterTools(s)
	calendar.RegisterTools(s)
	docs.RegisterTools(s)
	tasks.RegisterTools(s)
	drive.RegisterTools(s)
	sheets.RegisterTools(s)
	contacts.RegisterTools(s)

	// Start server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// initializeApp sets up the auth manager and shared dependencies.
// No configuration required - uses dynamic credential discovery.
func initializeApp() error {
	authManager, err := auth.NewManager()
	if err != nil {
		return err
	}

	// Set up shared dependencies for all packages
	common.SetDeps(&common.Deps{
		AuthManager: authManager,
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

No configuration required - just authenticate any Google account on demand.
When tools request an account without credentials, auth flow is triggered automatically.

Configuration:
  Config dir:     %s
  Credentials:    %s
  Client secret:  %s

For more information, see README.md
`, serverName, serverName, serverName, serverName, serverName,
		config.DefaultConfigDir(), config.CredentialsDir(), config.ClientSecretPath())
}

// runInit ensures config directory exists and shows setup instructions.
func runInit() {
	if err := config.EnsureConfigDir(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config directory ready: %s\n", config.DefaultConfigDir())
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
