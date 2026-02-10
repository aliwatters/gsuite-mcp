# Claude Code Guidelines for gsuite-mcp

See **[docs/AGENTS.md](docs/AGENTS.md)** for full development guidelines, architecture, and patterns.

## Commands

| Command | Description |
|---------|-------------|
| `go build -o gsuite-mcp` | Build the binary |
| `go test ./...` | Run all tests |
| `go vet ./...` | Check for issues |
| `gofmt -w .` | Format code |

## Re-authenticating Accounts

When an account token expires or gets revoked:

```bash
./gsuite-mcp auth <label>
```

This opens a browser for Google OAuth. Replace `<label>` with the account label (e.g., `personal`, `work`). Run once per account that needs re-auth.

To check which accounts are authenticated:

```bash
./gsuite-mcp accounts
```

## Documentation

- [docs/AGENTS.md](docs/AGENTS.md) — Development guidelines and patterns
- [docs/ROADMAP.md](docs/ROADMAP.md) — Development phases and planning
- [README.md](README.md) — User documentation and tool reference
