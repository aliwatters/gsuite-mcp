# What gsuite-mcp is for

**gsuite-mcp is a locally run Go Model Context Protocol server that gives MCP clients authenticated access to Google Workspace operations.**

## Why this document exists

This settles whether the repository is a standalone Google productivity application or an integration layer. Its executable starts an MCP server, registers service-specific tools, and connects them to authenticated Google APIs. The accompanying CLI exists to set up and diagnose that server rather than to provide an interactive mailbox, calendar, or document interface.

## What it does

- `cmd/gsuite-mcp/main.go` creates an `mcp-go` server and serves it through `server.ServeStdio`.
- It registers tools from `internal/gmail`, `calendar`, `docs`, `tasks`, `drive`, `sheets`, `slides`, `forms`, `contacts`, `meet`, `driveactivity`, and `chat`. The optional `internal/citation` tools are registered only when `large_doc_indexing` is enabled.
- The MCP surface includes concrete operations such as `gmail_search`, `gmail_send`, `calendar_list_events`, `calendar_create_event`, `drive_search`, and the matching tool families in the registered service packages.
- The `init`, `auth`, `accounts`, and `check` commands create local configuration, authenticate Google accounts, report account token status, and validate configured API access.
- `internal/auth` and `internal/config` use a local configuration directory and per-email credentials to establish Google OAuth clients for tool calls.

## What it is not for

- It is not a browser or desktop Google Workspace client. The primary executable surface is a stdio MCP server; its named CLI commands are setup and diagnostic commands.
- It is not a general HTTP API or hosted web application. The only HTTP listener started by the executable is a localhost OAuth re-authentication helper; the MCP server itself is served over stdio.
- It is not an OAuth credential issuer or Google account-provisioning system. `init` instructs the operator to provide a Google OAuth client secret, and `auth` obtains credentials for an account the user signs into.
- It is not limited to read-only retrieval. Registered tools intentionally include mutations such as `gmail_send`, Gmail label/trash actions, and calendar creation, updates, and deletion.

## How to tell it is working

- `gsuite-mcp accounts` lists locally authenticated accounts and their token health; `--json` exposes the same status for automation.
- `gsuite-mcp check` validates the OAuth configuration, authenticated tokens, and API access, returning a structured report with `--json`.
- An MCP client that launches the binary can enumerate the registered service tools and receives MCP-formatted results or authentication errors rather than a Go process error.
- After an account is authenticated and the relevant Google API is enabled, a registered tool such as `gmail_search`, `calendar_list_events`, or `drive_search` can use that account to make its corresponding API request.

## Where it fits

MCP clients consume the `gsuite-mcp` binary as a stdio server. The binary depends on Google OAuth credentials and the Google APIs selected by its registered services. Releases are built from `cmd/gsuite-mcp` as archives for Linux, macOS, and Windows, while `install.sh` supports local source installation.
