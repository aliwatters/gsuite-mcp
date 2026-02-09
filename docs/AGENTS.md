# AI Agent Guidelines for gsuite-mcp

This document provides AI agents with project-specific context and guidelines for contributing to gsuite-mcp.

## Quick Reference

| Item | Value |
|------|-------|
| **Language** | Go 1.23+ |
| **Framework** | [mcp-go](https://github.com/mark3labs/mcp-go) |
| **Config** | `~/.config/gsuite-mcp/` |
| **Build** | `go build -o gsuite-mcp` |
| **Test** | `go test ./...` |

## Project Overview

**gsuite-mcp** is a Go-based MCP (Model Context Protocol) server providing complete Google Workspace operations with true multi-account support.

### Key Differentiators
- Per-operation account selection via `account` parameter
- JSON Schema draft 2020-12 (not draft-07 which causes Claude errors)
- Single Go binary (no Python, Node, or runtime dependencies)
- Full inbox management: archive, trash, labels, batch operations

### Services Supported
- Gmail (36 tools)
- Google Calendar (10 tools)
- Google Docs (16 tools)
- Google Tasks (10 tools)
- Google Sheets (8 tools)
- Google Contacts (12 tools)

## Architecture

### Directory Structure
```
gsuite-mcp/
├── main.go                 # Entry point, CLI, tool registration
├── CLAUDE.md               # Claude Code entry point (pointers only)
├── internal/
│   ├── auth/               # OAuth2 authentication
│   ├── common/             # Shared helpers, constants, types
│   ├── config/             # Configuration and account management
│   ├── calendar/           # Google Calendar tools
│   ├── contacts/           # Google Contacts tools
│   ├── docs/               # Google Docs tools
│   ├── drive/              # Google Drive tools
│   ├── gmail/              # Gmail tools
│   ├── sheets/             # Google Sheets tools
│   └── tasks/              # Google Tasks tools
├── docs/
│   ├── AGENTS.md           # This file
│   └── ROADMAP.md          # Development phases
└── README.md               # User documentation
```

### Configuration Files
```
~/.config/gsuite-mcp/
├── config.json             # Account configuration
├── client_secret.json      # Google OAuth app credentials
└── credentials/
    └── {label}.json        # Per-account OAuth tokens
```

## Key Rules

- Every tool must include the `account` parameter
- Use JSON Schema draft 2020-12 (not draft-07)
- Return errors via `mcp.NewToolResultError()`, not Go errors
- Update README.md when adding new tools
- Read relevant source files before making changes
- Check [ROADMAP.md](ROADMAP.md) for current priorities

## Key Patterns

### 1. Account Resolution

All tools accept an optional `account` parameter. Resolution order:

1. If `account` provided: match by label first, then by email
2. If omitted: use `default_account` from config
3. If no default and single account: use that account
4. Otherwise: return error

```go
func handleTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    accountParam, _ := request.Params.Arguments["account"].(string)
    account, err := appConfig.ResolveAccount(accountParam)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }
    // Use account...
}
```

**Reference**: `internal/config/config.go:108-135`

### 2. Tool Registration

Tools use the mcp-go fluent API:

```go
s.AddTool(mcp.NewTool("gmail_tool_name",
    mcp.WithDescription("Tool description"),
    mcp.WithString("required_param", mcp.Required(), mcp.Description("...")),
    mcp.WithString("optional_param", mcp.Description("...")),
    mcp.WithArray("array_param", mcp.Description("...")),
), handlerFunction)
```

**Every tool MUST include the account parameter:**
```go
mcp.WithString("account", mcp.Description("Account label or email (uses default if omitted)"))
```

**Reference**: `main.go:179-232`

### 3. Authentication

Get an authenticated HTTP client:

```go
import (
    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
)

client, err := authManager.GetClient(ctx, account)
if err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("Authentication error: %v", err)), nil
}

srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
if err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("Failed to create Gmail service: %v", err)), nil
}
```

**Reference**: `internal/auth/auth.go:78-110`

### 4. Error Handling

Return user-friendly errors via MCP:

```go
// For expected errors (bad input, not found, etc.)
return mcp.NewToolResultError("Clear error message"), nil

// For successful results
return mcp.NewToolResultText("Success message or JSON"), nil

// Only return Go errors for unexpected failures
return nil, fmt.Errorf("unexpected error: %w", err)
```

**Good error message:**
```go
return mcp.NewToolResultError("Account 'work' not found - check config.json or use 'gsuite-mcp accounts' to list available accounts"), nil
```

**Bad error message:**
```go
return mcp.NewToolResultError("not found"), nil  // Don't do this
```

## Testing

### Building and Running

```bash
# Build
go build -o gsuite-mcp

# Run MCP server (for integration)
./gsuite-mcp

# CLI commands for setup
./gsuite-mcp init           # Create default config
./gsuite-mcp auth <label>   # Authenticate an account
./gsuite-mcp accounts       # List configured accounts
```

### Unit Tests

```bash
go test ./...
go test -v ./internal/config/
```

### Integration Testing

Since this is an MCP server, test tools by:
1. Adding to MCP config and using with Claude/Cursor/etc.
2. Using an MCP client library for programmatic testing

### MCP Configuration for Testing

```json
{
  "mcpServers": {
    "gsuite-mcp": {
      "command": "/path/to/gsuite-mcp"
    }
  }
}
```

## Common Tasks

### Adding a New Tool

1. Add tool registration in the appropriate `register*Tools()` function in `main.go`
2. Create handler function following the pattern:
   ```go
   func handleNewTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
       account, err := resolveAccountFromRequest(request)
       if err != nil {
           return mcp.NewToolResultError(err.Error()), nil
       }

       client, err := authManager.GetClient(ctx, account)
       if err != nil {
           return mcp.NewToolResultError(fmt.Sprintf("Authentication error: %v", err)), nil
       }

       // Create API service and make call
       srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
       // ...
   }
   ```
3. Update README.md tools table
4. Update ROADMAP.md if applicable

### Adding Configuration Options

1. Add field to `Config` struct in `internal/config/config.go`
2. Update `Validate()` if the field has constraints
3. Update `CreateDefault()` with sensible default
4. Document in README.md

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use descriptive variable names
- Keep functions focused and small
- Document exported functions and types
- Prefer returning errors over panicking
- Use `fmt.Errorf` with `%w` for error wrapping

## Commit Messages

Use conventional commits with scope:

```
feat(gmail): add gmail_vacation_set tool
fix(auth): handle token refresh edge case
docs: update installation guide
refactor(config): simplify account resolution
test(config): add tests for multi-account
chore: update dependencies
```

## Common Mistakes to Avoid

1. **Don't commit credentials** - Never include real tokens, emails, or client secrets
2. **Don't break multi-account** - All tools must support the account parameter
3. **Don't use draft-07 schema** - This project uses JSON Schema draft 2020-12
4. **Don't add unnecessary dependencies** - Keep the binary lean
5. **Don't forget to update docs** - README.md tools table must stay current

## Quick Commands

```bash
# Development cycle
go build && ./gsuite-mcp accounts

# Run tests
go test ./...

# Check for issues
go vet ./...

# Format code
gofmt -w .

# Full build check
go build && go test ./... && go vet ./...
```

## References

- [Gmail API](https://developers.google.com/gmail/api/reference/rest)
- [Calendar API](https://developers.google.com/calendar/api/v3/reference)
- [Docs API](https://developers.google.com/docs/api/reference/rest)
- [Tasks API](https://developers.google.com/tasks/reference/rest)
- [Sheets API](https://developers.google.com/sheets/api/reference/rest)
- [People API](https://developers.google.com/people/api/rest)
- [mcp-go Framework](https://github.com/mark3labs/mcp-go)
- [MCP Specification](https://modelcontextprotocol.io/)
- [ROADMAP.md](ROADMAP.md) — Development phases
