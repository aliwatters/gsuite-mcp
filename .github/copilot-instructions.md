# Copilot Instructions for gsuite-mcp

## Project Overview

gsuite-mcp is a Go-based MCP (Model Context Protocol) server providing Gmail, Google Calendar,
Google Docs, Google Tasks, Google Sheets, Google Slides, Google Forms, Google Contacts, Google Drive,
and Google Meet operations with true multi-account support. It is designed as a single binary
alternative to Python-based Google Workspace integrations.

**Key differentiators:**
- Per-operation account selection via `account` parameter
- Dynamic multi-account support — no pre-configuration required
- Single Go binary, no runtime dependencies
- Full inbox management (archive, trash, labels, batch operations)

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| MCP Framework | [mcp-go](https://github.com/mark3labs/mcp-go) by mark3labs |
| Auth | OAuth2 via `golang.org/x/oauth2` with Google provider |
| APIs | Gmail, Calendar, Docs, Drive, Tasks, Sheets, Slides, Forms, Contacts, Meet, Chat |
| Concurrency | `golang.org/x/sync/errgroup` for bounded parallel API calls |

## Project Layout

```
gsuite-mcp/
├── cmd/gsuite-mcp/         # Entry point: main.go, check.go (CLI commands)
├── internal/
│   ├── auth/               # OAuth2 auth, token management, persistent auth server
│   ├── common/             # Shared helpers, WrapHandler, MarshalToolResult
│   ├── config/             # Configuration loading and credential discovery
│   ├── calendar/           # Google Calendar tools
│   ├── citation/           # [EXPERIMENTAL] Large-doc indexing and citation tools
│   ├── chat/               # Google Chat tools
│   ├── contacts/           # Google Contacts tools
│   ├── docs/               # Google Docs tools
│   ├── drive/              # Google Drive tools
│   ├── driveactivity/      # Drive Activity tools
│   ├── forms/              # Google Forms tools
│   ├── gmail/              # Gmail tools
│   ├── meet/               # Google Meet tools
│   ├── sheets/             # Google Sheets tools
│   ├── slides/             # Google Slides tools
│   └── tasks/              # Google Tasks tools
├── docs/
│   ├── AGENTS.md           # Full AI agent development guidelines
│   └── ROADMAP.md          # Development phases
├── .github/
│   └── copilot-instructions.md  # This file
├── README.md               # User documentation
└── INSTALLATION.md         # Getting started guide
```

## Build and Test Commands

```bash
go build -o gsuite-mcp ./cmd/gsuite-mcp   # Build the binary
go test ./...                              # Run all unit tests
go test -tags=e2e ./e2e/...               # Run E2E tests (requires .env)
go vet ./...                              # Check for issues
gofmt -w .                               # Format code

# CLI commands (after building)
./gsuite-mcp --help
./gsuite-mcp init          # Create default config
./gsuite-mcp auth          # Authenticate a Google account (opens browser)
./gsuite-mcp accounts      # List authenticated accounts
./gsuite-mcp check         # Verify setup (human-readable)
./gsuite-mcp check --json  # Verify setup (machine-readable, for cron)
```

## Configuration

```
~/.config/gsuite-mcp/
├── config.json             # Optional settings (oauth_port, drive_access, features)
├── credentials/
│   └── {email}.json        # OAuth tokens per account (keyed by email address)
└── client_secret.json      # Google OAuth app credentials
```

## Go MCP Patterns

### Tool Registration

All tools follow the same registration pattern. Always include `common.WithAccountParam()`:

```go
// register.go
func RegisterTools(s *server.MCPServer) {
    s.AddTool(mcp.NewTool("service_action",
        mcp.WithDescription("What this tool does"),
        mcp.WithString("param", mcp.Required(), mcp.Description("...")),
        common.WithAccountParam(),
    ), HandleServiceAction)
}
```

### Handler Pattern (WrapHandler + Testable)

Every tool uses a two-function pattern:
1. A `Testable*` function that accepts an explicit service dependency (testable without real APIs)
2. A `Handle*` var that wires in the real service via `common.WrapHandler`

```go
// service_tools.go — handle vars
var HandleServiceAction = common.WrapHandler[ServiceInterface](TestableServiceAction)

// testable_actions.go — testable implementation
func TestableServiceAction(
    ctx context.Context,
    request mcp.CallToolRequest,
    deps *ServiceHandlerDeps,
) (*mcp.CallToolResult, error) {
    svc, errResult, ok := ResolveServiceOrError(ctx, request, deps)
    if !ok {
        return errResult, nil
    }

    param, errResult := common.RequireStringArg(request.Params.Arguments, "param")
    if errResult != nil {
        return errResult, nil
    }

    result, err := svc.DoSomething(ctx, param)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("API error: %v", err)), nil
    }

    return common.MarshalToolResult(result)
}
```

### Argument Parsing Helpers

```go
// Required string — returns error result if missing
param, errResult := common.RequireStringArg(args, "key")
if errResult != nil { return errResult, nil }

// Optional string with default
val := common.ParseStringArg(args, "key", "default")

// Optional bool
flag := common.ParseBoolArg(args, "flag", false)

// Optional number (max results pattern)
maxResults := common.ParseMaxResults(args, common.GmailDefaultMaxResults, common.GmailMaxResultsLimit)
```

### Result Serialization

```go
// Serialize any JSON-marshalable value
return common.MarshalToolResult(myStruct)

// Simple text result
return mcp.NewToolResultText("success"), nil

// Error result (user-facing, not a Go error)
return mcp.NewToolResultError("message_id is required"), nil
```

## Google API Specifics

### Service Creation Pattern

```go
import (
    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
)

// client comes from common.GetDeps().AuthManager.GetClientOrAuthenticate(...)
srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
if err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("creating service: %v", err)), nil
}
```

### Parallel API Calls (errgroup pattern)

For batch operations fetching N items from Drive/Gmail, use bounded concurrency:

```go
import "golang.org/x/sync/errgroup"

results := make([]Result, len(ids))
g, gCtx := errgroup.WithContext(ctx)
g.SetLimit(5)  // max 5 concurrent API calls

for i, id := range ids {
    i, id := i, id  // capture loop vars
    g.Go(func() error {
        item, err := srv.Items.Get(id).Context(gCtx).Do()
        if err != nil {
            results[i] = Result{ID: id, Err: err}
            return nil  // partial failure — don't fail fast
        }
        results[i] = Result{ID: id, Item: item}
        return nil
    })
}
if err := g.Wait(); err != nil {
    return mcp.NewToolResultError(fmt.Sprintf("API error: %v", err)), nil
}
```

### OAuth Scope Requirements

Scopes are documented in `internal/auth/auth.go` → `ScopesByService`:

| Service   | Required Scopes                                                |
|-----------|----------------------------------------------------------------|
| Gmail     | `gmail.modify`, `gmail.compose`, `gmail.labels`, `gmail.settings.basic` |
| Calendar  | `calendar`, `calendar.events`                                  |
| Drive     | `drive`                                                        |
| Docs      | `documents`                                                    |
| Sheets    | `spreadsheets`                                                 |
| Slides    | `presentations`                                                |
| Forms     | `forms.body`, `forms.responses.readonly`                       |
| Tasks     | `tasks`                                                        |
| Contacts  | `contacts`                                                     |

All scopes are requested together in one consent screen (`DefaultScopes` in auth.go).

### Common Google API Errors

```go
import (
    "errors"
    "google.golang.org/api/googleapi"
)

var apiErr *googleapi.Error
if errors.As(err, &apiErr) {
    switch apiErr.Code {
    case 404:
        return mcp.NewToolResultError("not found"), nil
    case 403:
        // Could be quota, permission, or disabled API
        return mcp.NewToolResultError(fmt.Sprintf("permission denied: %v", apiErr.Message)), nil
    case 429:
        return mcp.NewToolResultError("rate limit exceeded — try again later"), nil
    }
}
```

## Error Handling Rules

1. **Never swallow errors** — if you catch and retry, log the original error first
2. **User errors → `mcp.NewToolResultError()`** — for invalid inputs, missing params, API errors
3. **Unexpected failures → `return nil, err`** — for framework-level issues only
4. **Wrap errors with context** — `fmt.Errorf("doing X: %w", err)`, never discard inner error
5. **Partial batch failures** — collect errors, succeed on successful items, report failures

```go
// Good: report partial success
var errs []string
for _, item := range items {
    if err := process(item); err != nil {
        errs = append(errs, fmt.Sprintf("%s: %v", item.ID, err))
    }
}
if len(errs) > 0 {
    result["errors"] = errs
}
return common.MarshalToolResult(result)
```

## Testing Patterns

### Unit Test Setup

Every service package has a mock implementing the service interface:

```go
// In your test file
func TestMyTool_Success(t *testing.T) {
    fixtures := NewGmailTestFixtures()  // or equivalent for other services

    // Seed test data
    fixtures.MockService.AddMessage(newTestMessage(...))

    // Build request
    request := mcp.CallToolRequest{
        Params: struct{ ... }{
            Arguments: map[string]any{
                "message_id": "msg123",
            },
        },
    }

    result, err := TestableGmailGetMessage(context.Background(), request, fixtures.Deps)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.IsError {
        t.Fatalf("tool error: %v", result.Content)
    }

    // Assert response fields
    response := extractResponse(t, result)
    if response["id"] != "msg123" {
        t.Errorf("expected id=msg123, got %v", response["id"])
    }
}
```

### Mock Error Injection

```go
fixtures.MockService.SetError("simulated API error")
// All subsequent calls return this error
```

### Table-Driven Tests

Prefer table-driven tests for argument validation and boundary conditions:

```go
cases := []struct {
    name    string
    args    map[string]any
    wantErr bool
}{
    {"missing_id", map[string]any{}, true},
    {"valid_id", map[string]any{"message_id": "msg1"}, false},
}
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        // ...
    })
}
```

## Common Mistakes to Avoid

1. **Never use `report.Accounts[i]` where `i` is from a filtered slice** — use a map to track indices
2. **Don't break multi-account support** — all tools must accept the `account` parameter
3. **Don't commit credentials** — never include tokens, emails, or client secrets
4. **Don't add unnecessary dependencies** — keep the binary lean
5. **Don't return Go errors for user errors** — use `mcp.NewToolResultError()` instead
6. **Don't use `base64.URLEncoding` to decode Gmail API data** that may not have padding — use `base64.StdEncoding` or `base64.RawURLEncoding` as appropriate, and check which the API actually returns
7. **Experimental features need `[EXPERIMENTAL]` prefix** in tool descriptions
8. **Capture loop variables** in goroutines: `i, id := i, id` before the `g.Go(func() {...})`

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

- [docs/AGENTS.md](../docs/AGENTS.md) — Full AI agent development guidelines
- [docs/ROADMAP.md](../docs/ROADMAP.md) — Development phases and priorities
- [Gmail API Reference](https://developers.google.com/gmail/api/reference/rest)
- [mcp-go Framework](https://github.com/mark3labs/mcp-go)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Google API Go Client](https://github.com/googleapis/google-api-go-client)
