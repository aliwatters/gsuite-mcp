# Installation

Get from zero to working in under 10 minutes.

## Prerequisites

- **Go 1.23+** (to build from source) or download a [pre-built binary](https://github.com/aliwatters/gsuite-mcp/releases)
- **A Google account** (personal or Workspace)

## 1. Install gsuite-mcp (30 seconds)

### From Source (Recommended)

```bash
# Clone and build
git clone https://github.com/aliwatters/gsuite-mcp.git
cd gsuite-mcp
go build -o gsuite-mcp

# Move to your PATH
mv gsuite-mcp ~/.local/bin/
# or: sudo mv gsuite-mcp /usr/local/bin/
```

### Verify Installation

```bash
gsuite-mcp --help
```

You should see the help output with available commands.

## 2. Get Google OAuth Credentials (5 minutes)

OAuth is just "Login with Google" for apps. You're creating credentials for *your own app* that will access *your own data*.

### Step-by-Step

1. **Go to Google Cloud Console**
   - Visit [console.cloud.google.com](https://console.cloud.google.com)
   - Sign in with your Google account

2. **Create a new project** (or select existing)
   - Click the project dropdown at the top
   - Click "New Project"
   - Name it something like "gsuite-mcp"
   - Click "Create"

3. **Enable the APIs**
   - Go to "APIs & Services" → "Library"
   - Search for and enable each API you need:
     - Gmail API
     - Google Calendar API
     - Google Docs API
     - Google Sheets API
     - Google Tasks API
     - People API (for Contacts)

   > **Tip**: You can enable just what you need. Start with Gmail and Calendar.

4. **Create OAuth credentials**
   - Go to "APIs & Services" → "Credentials"
   - Click "Create Credentials" → "OAuth client ID"
   - If prompted, configure the OAuth consent screen first:
     - Choose "External" (unless you're in a Workspace org)
     - Fill in app name: "gsuite-mcp"
     - Add your email as test user
     - Save
   - Back to credentials:
     - Application type: "Desktop app"
     - Name: "gsuite-mcp"
     - Click "Create"

5. **Download the credentials**
   - Click the download icon next to your new credential
   - Save as `client_secret.json`
   - Move to the config directory:

   ```bash
   mkdir -p ~/.config/gsuite-mcp
   mv ~/Downloads/client_secret.json ~/.config/gsuite-mcp/
   ```

## 3. Authenticate Your Account (2 minutes)

```bash
gsuite-mcp auth
```

This will:
1. Open your browser
2. Ask you to sign in to Google
3. Request permission for the APIs you enabled
4. Save your token locally (keyed by your email address)

### Multiple Accounts

Run `gsuite-mcp auth` again while signed into a different Google account. Each account's credentials are stored separately by email address.

Then use the `account` parameter in tool calls to specify which account to use (by email or partial match).

## 4. Add to Claude (1 minute)

### Claude Code (CLI)

Add to `~/.claude/claude_desktop_config.json` (create if it doesn't exist):

```json
{
  "mcpServers": {
    "gsuite": {
      "command": "/path/to/gsuite-mcp"
    }
  }
}
```

Replace `/path/to/gsuite-mcp` with the actual path (e.g., `/Users/you/.local/bin/gsuite-mcp`).

### Claude Desktop

Same configuration — edit Claude Desktop's config file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

### Cursor / Other MCP Clients

Most MCP clients use similar configuration. Check your client's documentation for the config file location.

## Verify It Works

Start a new Claude conversation and ask:

> "What's in my inbox?"

or

> "What meetings do I have today?"

Claude should use gsuite-mcp to fetch your real data.

## Troubleshooting

### "Token expired" or "Invalid credentials"

Re-run the auth command:

```bash
gsuite-mcp auth
```

### "API not enabled"

Go back to Google Cloud Console → APIs & Services → Library and enable the required API.

### "Access denied" or "Insufficient scopes"

Your token may have been created before you enabled all APIs. Delete your token and re-authenticate:

```bash
rm ~/.config/gsuite-mcp/credentials/<your-email>.json
gsuite-mcp auth
```

### "OAuth consent screen not configured"

Complete the OAuth consent screen setup in Google Cloud Console before creating credentials.

### "User not in test users list"

If using an external OAuth consent screen in "Testing" mode, add your email as a test user:
- Google Cloud Console → APIs & Services → OAuth consent screen
- Under "Test users", click "Add users"
- Add your email address

### Still stuck?

Open an issue at [github.com/aliwatters/gsuite-mcp/issues](https://github.com/aliwatters/gsuite-mcp/issues) with:
- The error message
- Which step failed
- Your OS and Go version

## Configuration Reference

All configuration lives in `~/.config/gsuite-mcp/`:

```
~/.config/gsuite-mcp/
├── client_secret.json      # Your OAuth app credentials
└── credentials/
    ├── alice@gmail.com.json      # Token for alice@gmail.com
    └── bob@company.com.json      # Token for bob@company.com
```

## Next Steps

- Read the [README](README.md) for the full tool reference
- Check [docs/AGENTS.md](docs/AGENTS.md) if you're building integrations
- See [docs/ROADMAP.md](docs/ROADMAP.md) for upcoming features
