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
go build -o gsuite-mcp ./cmd/gsuite-mcp/

# Move to your PATH
mv gsuite-mcp ~/.local/bin/
# or: sudo mv gsuite-mcp /usr/local/bin/
```

### Verify Installation

```bash
gsuite-mcp --help
```

You should see the help output with available commands.

## 2. Create a GCP Project (5 minutes)

> **New to GCP?** See the [detailed GCP Setup Guide](docs/GCP_SETUP.md) with step-by-step instructions, exact URLs, and troubleshooting for common pitfalls.

### Step 1: Create or select a project

- Visit [console.cloud.google.com](https://console.cloud.google.com) and sign in
- Click the project dropdown at the top → "New Project"
- Name it something like "gsuite-mcp" → "Create"

### Step 2: Configure the OAuth consent screen

Go to **APIs & Services → OAuth consent screen**.

**This has two independent settings that both must be correct:**

| Setting | Google Workspace orgs | Personal Gmail |
|---------|----------------------|----------------|
| **User type** | Internal (no verification needed) | External |
| **Publishing status** | Testing or In production | Testing or In production |
| **Test users** | Not required for Internal | Add your email (required in Testing mode) |

- **User type**: Controls who can authenticate. "Internal" restricts to your org; "External" allows any Google account.
- **Publishing status**: "Testing" works fine for personal use — just add yourself as a test user. Note that tokens in Testing mode expire every 7 days, requiring periodic re-auth. Switch to "In production" if that gets annoying.

Fill in the required fields:
- App name: "gsuite-mcp"
- User support email: your email
- Developer contact email: your email

If using External + Testing mode, add yourself under "Test users" → "Add users".

### Step 3: Enable the APIs

Go to **APIs & Services → Library** and enable the required APIs:

- Gmail API
- Google Calendar API
- Google Drive API
- Google Docs API
- Google Sheets API
- Google Slides API
- Google Tasks API
- People API (for Contacts)
- Google Forms API
- Google Meet API

> **Tip**: You can enable just what you need now. `gsuite-mcp check` will tell you which are missing later.

### Step 4: Create OAuth credentials

- Go to **APIs & Services → Credentials**
- Click **"Create Credentials" → "OAuth client ID"**
- Application type: **"Desktop app"**
- Name: "gsuite-mcp"
- Click "Create"

> **Why Desktop app, not Web application?** Desktop app credentials implement RFC 8252 loopback redirect — Google accepts any available localhost port automatically. Web application credentials require every port to be pre-registered in GCP Console, and changes take minutes to hours to propagate. This matters in practice: the MCP server and the standalone `gsuite-mcp auth` command both need the OAuth callback port, and if there is a conflict, a Web client traps you in a `redirect_uri_mismatch` loop with no easy escape. Desktop app eliminates this entire failure class. If you accidentally created a Web application credential, see [Migrating from a Web application OAuth client](#migrating-from-a-web-application-oauth-client) below.

### Step 5: Download and install the credentials

- Click the download icon next to your new credential
- Save as `client_secret.json`
- Move to the config directory:

```bash
mkdir -p ~/.config/gsuite-mcp
mv ~/Downloads/client_secret.json ~/.config/gsuite-mcp/
```

> **How to verify you're in the right project**: The OAuth client ID starts with a number (e.g., `305192952884-...`). This is your GCP project number. If it doesn't match the project where you enabled APIs and configured the consent screen, things won't work.

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

## 5. Verify Your Setup

Run the preflight check to validate everything in one command:

```bash
gsuite-mcp check
```

This checks:
- `client_secret.json` exists and parses correctly
- All authenticated account tokens are still valid
- All required Google APIs are enabled for your GCP project

Any issues include actionable fix instructions (re-auth commands, API enable links).

Then start a new Claude conversation and ask:

> "What's in my inbox?"

or

> "What meetings do I have today?"

Claude should use gsuite-mcp to fetch your real data.

## Troubleshooting

### Quick Reference

| Error | Cause | Fix |
|-------|-------|-----|
| `403: access_denied` "not completed verification" | Not listed as test user in Testing mode | Add your email as a test user, or set publishing status to "In production" |
| `400: redirect_uri_mismatch` | Web application credential used instead of Desktop app | [Migrate to Desktop app type](#migrating-from-a-web-application-oauth-client) |
| `403: SERVICE_DISABLED` | API not enabled in GCP project | Run `gsuite-mcp check` for enable links |
| `invalid_grant` | Token expired or revoked | Run `gsuite-mcp auth` |
| `401: invalid_client` | Wrong or corrupted client_secret.json | Re-download from GCP Console |
| `port XXXX is in use` | OAuth callback port conflict | Set `oauth_port` in config.json or `GSUITE_MCP_OAUTH_PORT` env var |

### Token Expired or Invalid Credentials

Re-run the auth command:

```bash
gsuite-mcp auth
```

Or, if gsuite-mcp is running as an MCP server, open `http://localhost:38917/auth` in your browser to re-authenticate without restarting the server. Error messages from tools will include this URL when the auth server is running.

### API Not Enabled

Run `gsuite-mcp check` to see which APIs are disabled — it provides direct enable links.

Or go to Google Cloud Console → APIs & Services → Library and enable the required API.

### Access Denied or Insufficient Scopes

Your token may have been created before you enabled all APIs. Delete your token and re-authenticate:

```bash
rm ~/.config/gsuite-mcp/credentials/you@example.com.json
gsuite-mcp auth
```

### OAuth Consent Screen Not Configured

Complete the OAuth consent screen setup in Google Cloud Console before creating credentials. See [Step 2](#step-2-configure-the-oauth-consent-screen) above.

### "Not completed verification" / access_denied

This means you're not listed as a test user for an app in "Testing" mode. Either:

1. **Add yourself as a test user**: Google Cloud Console → APIs & Services → OAuth consent screen → "Test users" → "Add users" → add your email
2. **Or switch to "In production"**: Click "Publish App" on the OAuth consent screen (does not require Google verification for personal use)

### User Not in Test Users List

If using an External consent screen in "Testing" mode, add your email as a test user:
- Google Cloud Console → APIs & Services → OAuth consent screen
- Under "Test users", click "Add users"
- Add your email address

Or switch to "In production" mode to avoid this entirely.

### Re-authentication

You may need to re-authenticate when:
- A token expires or gets revoked
- You enable additional API scopes
- You switch GCP projects or OAuth clients

For a clean re-auth:

```bash
rm ~/.config/gsuite-mcp/credentials/you@example.com.json
gsuite-mcp auth
```

To check which accounts are authenticated:

```bash
gsuite-mcp accounts
```

### Migrating from a Web application OAuth client

If you created a **Web application** credential instead of a **Desktop app** credential, you will hit `redirect_uri_mismatch` errors whenever the OAuth callback port changes — which happens whenever the MCP server is already running on port 38917 and you try to run `gsuite-mcp auth` manually.

**Why this happens with Web credentials**: Google requires exact port matching for Web application redirect URIs. You must pre-register every port in GCP Console, and changes take minutes to hours to propagate through Google's cache. The MCP server and the standalone auth command share the same port, creating a race. Desktop app credentials avoid all of this — they accept any localhost port automatically per RFC 8252.

**How to tell which type you have**: Open `~/.config/gsuite-mcp/client_secret.json`. The top-level key is either `"installed"` (Desktop app — correct) or `"web"` (Web application — needs migration).

**Migration steps**:

1. Go to [GCP Console → APIs & Services → Credentials](https://console.cloud.google.com/apis/credentials)
2. Click **Create Credentials → OAuth client ID**
3. Choose **Desktop app** as the application type
4. Name it `gsuite-mcp-desktop` (or any name you like)
5. Click **Create**, then download the JSON file
6. Back up your existing credential and replace it:

```bash
# Back up the old Web credential (optional)
cp ~/.config/gsuite-mcp/client_secret.json ~/.config/gsuite-mcp/client_secret.web.bak

# Install the new Desktop credential
mv ~/Downloads/client_secret_*.json ~/.config/gsuite-mcp/client_secret.json
```

7. Re-authenticate (the old token is no longer valid with the new credential):

```bash
rm ~/.config/gsuite-mcp/credentials/you@example.com.json
gsuite-mcp auth
```

The new Desktop credential will accept any free port — no port registration required, no propagation wait.

> **Clean up**: After confirming the new credential works, you can delete the old Web application credential from GCP Console to avoid confusion.

### Still Stuck?

Open an issue at [github.com/aliwatters/gsuite-mcp/issues](https://github.com/aliwatters/gsuite-mcp/issues) with:
- The error message
- Which step failed
- Output of `gsuite-mcp check`

## Configuration Reference

All configuration lives in `~/.config/gsuite-mcp/`:

```
~/.config/gsuite-mcp/
├── client_secret.json            # Your OAuth app credentials
├── config.json                   # Settings (optional — created by `gsuite-mcp init`)
└── credentials/
    ├── alice@gmail.com.json      # Token for alice@gmail.com
    └── bob@company.com.json      # Token for bob@company.com
```

### config.json

Optional configuration file created by `gsuite-mcp init`. Settings:

| Key | Default | Description |
|-----|---------|-------------|
| `oauth_port` | `38917` | Port for the OAuth callback server during `gsuite-mcp auth` |

Override `oauth_port` via the `GSUITE_MCP_OAUTH_PORT` environment variable.

## Next Steps

- Read the [README](README.md) for the full tool reference
- See the [GCP Setup Guide](docs/GCP_SETUP.md) for detailed setup instructions with troubleshooting
- Check [docs/AGENTS.md](docs/AGENTS.md) if you're building integrations
- See [docs/ROADMAP.md](docs/ROADMAP.md) for upcoming features
