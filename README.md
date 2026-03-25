# gsuite-mcp

[![Go 1.23+](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![MCP](https://img.shields.io/badge/MCP-compatible-blue)](https://modelcontextprotocol.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Complete Google Workspace MCP server — Gmail, Calendar, Drive, Docs, Tasks, Sheets, Slides, Forms, and Contacts. Single Go binary with true multi-account support.

## Quick Start

**[→ Full Installation Guide](INSTALLATION.md)** — Get running in under 10 minutes

```bash
# Build
go build -o ~/.local/bin/gsuite-mcp

# Authenticate
gsuite-mcp auth personal

# Add to your MCP client config
```

## Why gsuite-mcp?

- **Multi-account**: Switch accounts per-operation with `"account": "work"`
- **Complete coverage**: Gmail, Calendar, Drive, Docs, Tasks, Sheets, Slides, Forms, Contacts
- **Single binary**: No Python, no Node, no runtime dependencies
- **MCP native**: JSON Schema 2020-12, proper tool descriptions

<details>
<summary>Comparison with google-workspace MCP</summary>

| Feature | google-workspace | gsuite-mcp |
|---------|------------------|------------|
| Multi-account | User-level | Per-operation |
| JSON Schema | draft-07 (Claude errors) | 2020-12 |
| Services | Gmail only | Full Workspace |
| Runtime | Python + uv | Single Go binary |
| Batch operations | Limited | Full support |

</details>

## Tools Overview

### Gmail (45 tools)
Full inbox management: search, read, send, reply, archive, trash, labels, filters, drafts, threads, batch operations, vacation responder, send-as aliases, delegation.

### Calendar (12 tools)
Complete calendar control: list events, create/update/delete, recurring events, free/busy queries, Google Meet integration.

### Drive (23 tools)
File management with shared drive support: search (with friendly file type filter), upload, download, list, create folders, move, copy, trash, delete, share, permissions, shareable links, comments & replies, version history (revisions).

### Docs (25 tools)
Document creation and editing: create, read, structure, append, insert, replace, delete, formatting (bold, italic, headings, format-by-find), lists, tables, images, headers/footers, markdown export, PDF export, import.

### Tasks (10 tools)
Task management: lists, tasks, subtasks, due dates, completion, reordering.

### Sheets (16 tools)
Spreadsheet operations: read, write, append, batch operations, create, cell formatting, conditional formatting, data validation, charts, pivot tables.

### Slides (5 tools)
Presentation operations: get presentation metadata and structure, get individual slide details, slide thumbnails, create presentations, batch update (add/modify/delete slides, text, shapes, images).

### Forms (5 tools)
Form management: get form structure and questions, create forms, batch update (add/update/delete questions, settings), list responses, get individual responses.

### Contacts (12 tools)
Contact management: list, search, create, update, delete, contact groups.

<details>
<summary>Full tool reference</summary>

#### Gmail Core
| Tool | Description |
|------|-------------|
| `gmail_search` | Search messages with Gmail query syntax |
| `gmail_get` | Get single message with full content |
| `gmail_get_message` | Alias for `gmail_get` |
| `gmail_get_messages` | Batch get messages (max 25) |
| `gmail_get_thread` | Get full conversation thread |
| `gmail_send` | Send new email |
| `gmail_reply` | Reply to existing thread |
| `gmail_draft` | Create draft |
| `gmail_list_labels` | List all labels with counts |

#### Gmail Management
| Tool | Description |
|------|-------------|
| `gmail_modify_message` | Add/remove labels |
| `gmail_archive` | Remove from inbox |
| `gmail_trash` / `gmail_untrash` | Move to/from trash |
| `gmail_mark_read` / `gmail_mark_unread` | Mark read status |
| `gmail_star` / `gmail_unstar` | Star/unstar |
| `gmail_spam` / `gmail_not_spam` | Spam management |

#### Gmail Batch
| Tool | Description |
|------|-------------|
| `gmail_batch_modify` | Batch label changes |
| `gmail_batch_archive` | Archive multiple |
| `gmail_batch_trash` | Trash multiple |

#### Gmail Extended
| Tool | Description |
|------|-------------|
| `gmail_get_attachment` | Download attachment |
| `gmail_list_filters` / `gmail_create_filter` / `gmail_delete_filter` | Filter management |
| `gmail_create_label` / `gmail_update_label` / `gmail_delete_label` | Label management |
| `gmail_list_drafts` / `gmail_get_draft` / `gmail_update_draft` / `gmail_delete_draft` / `gmail_send_draft` | Draft management |
| `gmail_thread_archive` / `gmail_thread_trash` / `gmail_thread_untrash` / `gmail_modify_thread` | Thread operations |
| `gmail_get_profile` | Account info |
| `gmail_get_vacation` / `gmail_set_vacation` | Vacation responder |
| `gmail_list_send_as` / `gmail_get_send_as` | List/get send-as aliases |
| `gmail_create_send_as` / `gmail_update_send_as` / `gmail_delete_send_as` | Manage send-as aliases |
| `gmail_verify_send_as` | Verify external send-as alias |
| `gmail_list_delegates` / `gmail_create_delegate` / `gmail_delete_delegate` | Delegation management |

#### Calendar
| Tool | Description |
|------|-------------|
| `calendar_list_events` | List events with filtering (supports event_types filter) |
| `calendar_get_event` | Get event details (includes conference data) |
| `calendar_create_event` | Create event (with optional Google Meet) |
| `calendar_update_event` | Update event |
| `calendar_delete_event` | Delete event |
| `calendar_list_calendars` | List available calendars |
| `calendar_quick_add` | Create from natural language |
| `calendar_free_busy` | Query availability across calendars |
| `calendar_list_instances` | List recurring event instances |
| `calendar_update_instance` | Update single recurrence |
| `calendar_create_focus_time` | Create Focus Time with auto-decline |
| `calendar_create_out_of_office` | Create Out of Office with auto-decline |

#### Drive
| Tool | Description |
|------|-------------|
| `drive_search` | Search files with query syntax (includes shared drives) |
| `drive_get` | Get file metadata |
| `drive_download` | Download file content (text or base64) |
| `drive_upload` | Upload new file |
| `drive_list` | List files in folder |
| `drive_create_folder` | Create folder |
| `drive_move` | Move file to different folder |
| `drive_copy` | Copy a file |
| `drive_trash` | Move file to trash |
| `drive_delete` | Permanently delete file |
| `drive_share` | Share file with users |
| `drive_get_permissions` | Get file permissions |
| `drive_get_shareable_link` | Get shareable URL with sharing status |
| `drive_list_comments` | List comments on a file |
| `drive_get_comment` | Get a specific comment |
| `drive_create_comment` | Create a comment |
| `drive_update_comment` | Update a comment |
| `drive_delete_comment` | Delete a comment |
| `drive_list_replies` | List replies on a comment |
| `drive_create_reply` | Reply to a comment |
| `drive_list_revisions` | List file version history |
| `drive_get_revision` | Get revision metadata |
| `drive_download_revision` | Download a specific revision |

#### Docs
| Tool | Description |
|------|-------------|
| `docs_create` | Create new document |
| `docs_get` | Get content as plain text |
| `docs_get_metadata` | Get title, word count |
| `docs_get_structure` | Get document structure with character indices |
| `docs_append_text` | Append to end |
| `docs_insert_text` | Insert at position |
| `docs_replace_text` | Find and replace |
| `docs_delete_text` | Delete range |
| `docs_insert_table` | Insert table |
| `docs_insert_link` | Insert hyperlink |
| `docs_format_text` | Bold, italic, color, font |
| `docs_format_by_find` | Find text and apply formatting (no position math) |
| `docs_clear_formatting` | Remove formatting |
| `docs_set_paragraph_style` | Headings, alignment |
| `docs_create_list` / `docs_remove_list` | Bullet/numbered lists |
| `docs_insert_page_break` | Page breaks |
| `docs_insert_image` | Insert image from URL |
| `docs_create_header` / `docs_create_footer` | Headers/footers |
| `docs_batch_update` | Raw API access |
| `docs_get_as_markdown` | Get content as markdown |
| `docs_find_and_replace` | Find and replace text |
| `docs_export_to_pdf` | Export Doc/Sheet/Slides to PDF |
| `docs_import_to_google_doc` | Import text/HTML/markdown as Google Doc |

#### Tasks
| Tool | Description |
|------|-------------|
| `tasks_list_tasklists` | List all task lists |
| `tasks_list` | List tasks in a list |
| `tasks_get` | Get task details |
| `tasks_create` | Create task |
| `tasks_update` | Update task |
| `tasks_complete` | Mark complete |
| `tasks_delete` | Delete task |
| `tasks_create_tasklist` / `tasks_update_tasklist` / `tasks_delete_tasklist` | Task list management |
| `tasks_move` | Reorder or make subtask |
| `tasks_clear_completed` | Clear completed |

#### Sheets
| Tool | Description |
|------|-------------|
| `sheets_get` | Get spreadsheet metadata |
| `sheets_read` | Read cell range |
| `sheets_write` | Write to cells |
| `sheets_append` | Append rows |
| `sheets_create` | Create spreadsheet |
| `sheets_batch_read` | Read multiple ranges |
| `sheets_batch_write` | Write multiple ranges |
| `sheets_clear` | Clear cell range |
| `sheets_format_cells` | Background, font, bold, alignment, number format |
| `sheets_add_conditional_format` | Boolean or gradient formatting rules |
| `sheets_add_data_validation` | Dropdowns, number constraints, custom formulas |
| `sheets_create_chart` | Create embedded chart (bar, line, pie, etc.) |
| `sheets_update_chart` | Update chart title or type |
| `sheets_delete_chart` | Delete embedded chart |
| `sheets_create_pivot_table` | Create pivot table for data summarization |
| `sheets_batch_update` | Raw API access for advanced operations |

#### Contacts
| Tool | Description |
|------|-------------|
| `contacts_list` | List all contacts |
| `contacts_get` | Get contact details |
| `contacts_search` | Search contacts |
| `contacts_create` | Create contact |
| `contacts_update` | Update contact |
| `contacts_delete` | Delete contact |
| `contacts_list_groups` | List contact groups |
| `contacts_get_group` | Get group with members |
| `contacts_create_group` / `contacts_update_group` / `contacts_delete_group` | Group management |
| `contacts_modify_group_members` | Add/remove from group |

#### Slides
| Tool | Description |
|------|-------------|
| `slides_get_presentation` | Get presentation metadata, slide list with text previews |
| `slides_get_page` | Get full details of a single slide (shapes, images, tables, text) |
| `slides_get_thumbnail` | Get slide thumbnail image URL |
| `slides_create` | Create new presentation |
| `slides_batch_update` | Batch update (add/modify/delete slides, text, shapes, images) |

#### Forms
| Tool | Description |
|------|-------------|
| `forms_get` | Get form metadata, questions, and structure |
| `forms_create` | Create new form |
| `forms_batch_update` | Batch update (add/update/delete questions, settings) |
| `forms_list_responses` | List all form responses |
| `forms_get_response` | Get a single form response |

</details>

## Usage Examples

### Multi-Account

All tools accept an optional `account` parameter:

```json
{"query": "is:unread", "account": "work"}
{"message_id": "abc123", "account": "personal"}
{"query": "from:amazon"}  // uses default account
```

### Common Workflows

<details>
<summary>Email triage</summary>

```
1. gmail_search({"query": "is:unread", "account": "support"})
2. gmail_get({"message_id": "...", "account": "support"})
3. gmail_reply({"message_id": "...", "body": "Thanks for reaching out..."})
4. gmail_batch_archive({"message_ids": [...]})
```

</details>

<details>
<summary>Schedule a meeting</summary>

```
calendar_create_event({
  "summary": "Team Standup",
  "start_time": "2024-02-05T09:00:00-08:00",
  "end_time": "2024-02-05T09:30:00-08:00",
  "attendees": ["colleague@example.com"],
  "add_conferencing": true
})
```

</details>

<details>
<summary>Create a formatted document</summary>

```
1. docs_create({"title": "Project Report"})
2. docs_append_text({"document_id": "...", "text": "Project Report\n\nIntroduction..."})
3. docs_set_paragraph_style({"document_id": "...", "start_index": 1, "end_index": 15, "named_style_type": "TITLE"})
```

</details>

## For AI Agents

If you're an AI agent or building integrations, see [docs/AGENTS.md](docs/AGENTS.md) for:
- Project architecture and patterns
- Tool registration patterns
- Error handling conventions
- Testing guidelines

### Natural Language Prompts

These prompts work well with gsuite-mcp:

- "Show my unread emails"
- "What meetings do I have today?"
- "Send an email to X about Y"
- "Create a task to review the PR by Friday"
- "Add a new contact for John Smith"
- "What's on my work calendar this week?" (uses account parameter)

## Configuration

See [INSTALLATION.md](INSTALLATION.md) for full setup instructions.

**Config location**: `~/.config/gsuite-mcp/`

```
~/.config/gsuite-mcp/
├── client_secret.json      # Your OAuth app credentials
├── config.json             # Settings (optional — created by `gsuite-mcp init`)
└── credentials/
    └── {email}.json        # Per-account tokens
```

### OAuth Callback Port

The default OAuth callback port is **38917**. Override it in `config.json`:

```json
{ "oauth_port": 9000 }
```

Or via environment variable: `GSUITE_MCP_OAUTH_PORT=9000`

### Drive Access Filtering

Restrict which shared drives are accessible via MCP tools. Add `drive_access` to `config.json`:

```json
{
  "oauth_port": 38917,
  "drive_access": {
    "allowed": ["Marketing", "Engineering"]
  }
}
```

**Modes** (choose one):
- **Allowlist**: `"allowed": ["Drive A", "Drive B"]` — only these shared drives + My Drive
- **Blocklist**: `"blocked": ["SENSITIVE", "HR"]` — everything except these drives

My Drive is always accessible. Setting both `allowed` and `blocked` is an error. Drives can be specified by name (case-insensitive) or ID. The filter applies to Drive, Docs, and Sheets tools.

### HTTP Auth Endpoint (MCP Server Mode)

When running as an MCP server, gsuite-mcp starts a persistent HTTP server on the OAuth port so agents and users can trigger re-authentication from a browser:

- **`http://localhost:38917/auth`** — starts OAuth flow (opens Google consent screen)
- **`http://localhost:38917/auth?account=user@gmail.com`** — pre-selects the Google account

When a tool encounters missing credentials, the error message includes a clickable auth URL. If the port is unavailable, the MCP server continues without the auth endpoint.

## Development

```bash
go build -o gsuite-mcp
go test ./...
go vet ./...
```

See [docs/AGENTS.md](docs/AGENTS.md) for contribution guidelines.

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for planned features and development phases.

## License

MIT
