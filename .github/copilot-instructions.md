# Copilot Instructions for gsuite-mcp

## Project Overview

gsuite-mcp is a Go-based MCP (Model Context Protocol) server providing Gmail, Google Calendar, Google Docs, Google Tasks, Google Sheets, and Google Contacts operations with true multi-account support. It's designed as a single binary alternative to Python-based Google Workspace integrations.

**Key differentiators:**
- Per-operation account selection via `account` parameter
- JSON Schema draft 2020-12 (not draft-07)
- Single Go binary, no runtime dependencies
- Full inbox management (archive, trash, labels, batch operations)

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.23+ |
| MCP Framework | [mcp-go](https://github.com/mark3labs/mcp-go) by mark3labs |
| Auth | OAuth2 via `golang.org/x/oauth2` with Google provider |
| APIs | Gmail, Calendar, Docs, Tasks, Sheets, People (Contacts) |

## Project Layout

```
gsuite-mcp/
├── main.go                 # Entry point, CLI commands, tool registration
├── CLAUDE.md               # Claude Code entry point (pointers only)
├── internal/
│   ├── auth/               # OAuth2 authentication and token management
│   ├── common/             # Shared helpers, constants, types
│   ├── config/             # Configuration loading and account management
│   ├── calendar/           # Google Calendar tools
│   ├── contacts/           # Google Contacts tools
│   ├── docs/               # Google Docs tools
│   ├── drive/              # Google Drive tools
│   ├── gmail/              # Gmail tools
│   ├── sheets/             # Google Sheets tools
│   └── tasks/              # Google Tasks tools
├── docs/
│   ├── AGENTS.md           # AI agent development guidelines
│   └── ROADMAP.md          # Development phases
├── README.md               # User documentation
└── INSTALLATION.md         # Getting started guide
```

## Build and Test Commands

```bash
# Build the binary
go build -o gsuite-mcp

# Run tests
go test ./...

# Check for issues
go vet ./...

# Format code
gofmt -w .

# CLI commands (after building)
./gsuite-mcp --help
./gsuite-mcp init           # Create default config
./gsuite-mcp auth <label>   # Authenticate an account
./gsuite-mcp accounts       # List configured accounts
```

## Configuration Locations

```
~/.config/gsuite-mcp/
├── config.json             # Account configuration
├── credentials/
│   └── {label}.json        # OAuth tokens per account
└── client_secret.json      # Google OAuth app credentials
```

## Code Patterns

### Account Resolution

All tools accept an optional `account` parameter. Resolution order:
1. If provided: match by label first, then by email
2. If omitted: use `default_account` from config
3. If no default and single account: use that account
4. Otherwise: return error

### Tool Registration

Use the mcp-go fluent API. Always include the account parameter:

```go
s.AddTool(mcp.NewTool("gmail_tool_name",
    mcp.WithDescription("Tool description"),
    mcp.WithString("required_param", mcp.Required(), mcp.Description("...")),
    mcp.WithString("account", mcp.Description("Account label or email (uses default if omitted)")),
), handlerFunction)
```

### Error Handling

Return user-friendly errors via MCP:

```go
// For expected errors
return mcp.NewToolResultError("Clear error message"), nil

// For successful results
return mcp.NewToolResultText("Success message or JSON"), nil

// Only return Go errors for unexpected failures
return nil, fmt.Errorf("unexpected error: %w", err)
```

### Gmail API Calls

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

## Common Mistakes to Avoid

1. **Don't break multi-account support** - All tools must accept the `account` parameter
2. **Don't use JSON Schema draft-07** - This project uses draft 2020-12
3. **Don't commit credentials** - Never include tokens, emails, or client secrets
4. **Don't add unnecessary dependencies** - Keep the binary lean
5. **Don't return Go errors for user errors** - Use `mcp.NewToolResultError()` instead

## Commit Message Format

Use conventional commits:
```
feat(gmail): add gmail_send tool
fix(auth): handle token refresh edge case
docs: update README with new tool
refactor(config): simplify account resolution
test(config): add tests for edge cases
```

## References

- [docs/AGENTS.md](../docs/AGENTS.md) - Full AI agent development guidelines
- [docs/ROADMAP.md](../docs/ROADMAP.md) - Development phases and priorities
- [Gmail API Documentation](https://developers.google.com/gmail/api/reference/rest)
- [mcp-go Framework](https://github.com/mark3labs/mcp-go)
- [MCP Specification](https://modelcontextprotocol.io/)
