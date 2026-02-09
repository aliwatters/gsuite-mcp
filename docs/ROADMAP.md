# gsuite-mcp Roadmap

Go-based Gmail, Google Calendar, Google Docs, Google Tasks, and Google Drive MCP server with true multi-account support, designed to supersede the google-workspace MCP for Gmail operations.

## Why gsuite-mcp?

| Feature | google-workspace MCP | gsuite-mcp |
|---------|---------------------|------------|
| Multi-account | User-level only | Per-operation `account` param |
| JSON Schema | draft-07 (Claude errors) | draft 2020-12 |
| Message management | Limited | Full (archive, trash, labels) |
| Batch operations | Limited | Full support |
| Calendar support | None | Full CRUD support |
| Docs support | None | Create, read, modify |
| Tasks support | None | Full CRUD + completion |
| Runtime | Python + uv | Single Go binary |

---

## Phase 0: Multi-Account Foundation ✅

**Status**: Complete

**Goal**: Core infrastructure for managing multiple Gmail accounts.

### Completed
- [x] Config file loading and validation
- [x] Credential store (load/save OAuth tokens)
- [x] OAuth2 browser flow for new accounts
- [x] Token refresh handling
- [x] Account resolution (label → email → credentials)
- [x] Token import from google-workspace MCP
- [x] CLI commands: `init`, `auth`, `accounts`

### Config Structure
```
~/.config/gsuite-mcp/
├── config.json           # Account configuration
├── credentials/
│   ├── personal.json     # OAuth tokens
│   ├── support.json
│   └── ...
└── client_secret.json    # Google OAuth app credentials
```

---

## Phase 1: Fully Functional Gmail ✅

**Status**: Complete

**Goal**: Supersede google-workspace MCP server for Gmail operations.

### Core Operations

| Tool | Description | Status |
|------|-------------|--------|
| `gmail_search` | Search with Gmail query syntax | ✅ Complete |
| `gmail_get_message` | Get full message content | ✅ Complete |
| `gmail_get_messages` | Batch retrieve messages | ✅ Complete |
| `gmail_get_thread` | Get full conversation | ✅ Complete |
| `gmail_send` | Send emails with attachments | ✅ Complete |
| `gmail_reply` | Reply to email (keeps thread) | ✅ Complete |
| `gmail_draft` | Create draft | ✅ Complete |
| `gmail_list_labels` | List all labels | ✅ Complete |

### Management Operations

| Tool | Description | Status |
|------|-------------|--------|
| `gmail_modify_message` | Add/remove labels | ✅ Complete |
| `gmail_batch_modify` | Batch label operations | ✅ Complete |
| `gmail_trash` | Move to trash | ✅ Complete |
| `gmail_batch_trash` | Batch trash | ✅ Complete |
| `gmail_archive` | Remove from inbox | ✅ Complete |
| `gmail_batch_archive` | Batch archive | ✅ Complete |
| `gmail_mark_read` | Mark as read | ✅ Complete |
| `gmail_mark_unread` | Mark as unread | ✅ Complete |
| `gmail_untrash` | Restore from trash | ✅ Complete |
| `gmail_star` | Add star | ✅ Complete |
| `gmail_unstar` | Remove star | ✅ Complete |

---

## Phase 2: Extended Gmail Features ✅

**Status**: Complete

**Goal**: Power user features and advanced Gmail management.

### Label Management
- [x] `gmail_create_label` - Create new label
- [x] `gmail_update_label` - Update label name/settings
- [x] `gmail_delete_label` - Delete label

### Filter Management
- [x] `gmail_list_filters` - List message filters
- [x] `gmail_create_filter` - Create filter rule
- [x] `gmail_delete_filter` - Delete filter

### Attachment Handling
- [x] `gmail_get_attachment` - Download attachment

### Draft Management
- [x] `gmail_list_drafts` - List all drafts
- [x] `gmail_get_draft` - Get draft content
- [x] `gmail_update_draft` - Edit existing draft
- [x] `gmail_delete_draft` - Delete draft
- [x] `gmail_send_draft` - Send existing draft

### Thread Operations
- [x] `gmail_thread_archive` - Archive entire thread
- [x] `gmail_thread_trash` - Trash entire thread
- [x] `gmail_thread_untrash` - Restore thread from trash
- [x] `gmail_modify_thread` - Modify labels on thread

### Profile & Settings
- [x] `gmail_get_profile` - Get email/message counts
- [x] `gmail_get_vacation` - Get vacation responder
- [x] `gmail_set_vacation` - Enable/disable auto-reply
- [x] `gmail_spam` / `gmail_not_spam` - Spam management

---

## Phase 3: Google Calendar ✅

**Status**: Complete

**Goal**: Core calendar operations for event management.

### Calendar Core
| Tool | Description | Status |
|------|-------------|--------|
| `calendar_list_events` | List events with filtering | ✅ Complete |
| `calendar_get_event` | Get full event details | ✅ Complete |
| `calendar_create_event` | Create new event | ✅ Complete |
| `calendar_update_event` | Update existing event | ✅ Complete |
| `calendar_delete_event` | Delete event | ✅ Complete |

### Features
- Multi-account support via `account` parameter
- Calendar selection via `calendar_id` parameter
- All-day and timed events
- Attendee management
- Reminder configuration
- Recurring event support
- Query search within events
- Pagination for large calendars

---

## Phase 4: Google Docs ✅

**Status**: Complete

**Goal**: Core Google Docs operations for document creation and editing.

### Docs Core
| Tool | Description | Status |
|------|-------------|--------|
| `docs_create` | Create new document | ✅ Complete |
| `docs_get` | Get document content | ✅ Complete |
| `docs_get_metadata` | Get document metadata | ✅ Complete |
| `docs_append_text` | Append text to document | ✅ Complete |
| `docs_insert_text` | Insert text at position | ✅ Complete |

### Features
- Multi-account support via `account` parameter
- Document ID extraction from URLs
- Plain text content extraction (including tables)
- Word and character count in metadata

---

## Phase 5: Google Tasks ✅

**Status**: Complete

**Goal**: Core Google Tasks operations for task management.

### Tasks Core
| Tool | Description | Status |
|------|-------------|--------|
| `tasks_list_tasklists` | List all task lists | ✅ Complete |
| `tasks_list` | List tasks in a task list | ✅ Complete |
| `tasks_get` | Get task details by ID | ✅ Complete |
| `tasks_create` | Create new task | ✅ Complete |
| `tasks_update` | Update task fields | ✅ Complete |
| `tasks_complete` | Mark task as completed | ✅ Complete |
| `tasks_delete` | Delete task | ✅ Complete |

### Features
- Multi-account support via `account` parameter
- Default task list (`@default`) for quick access
- Due date management
- Notes/description support
- Task completion status
- Parent task (subtask) support
- Pagination for large task lists

---

## Phase 6: Extended Calendar ✅

**Status**: Complete

**Goal**: Advanced calendar features.

### Calendar Management
- [x] `calendar_list_calendars` - List all calendars
- [x] `calendar_quick_add` - Create event from natural language
- [x] `calendar_free_busy` - Query free/busy information

### Recurring Events
- [x] `calendar_list_instances` - List recurring event instances
- [x] `calendar_update_instance` - Update single instance

---

## Phase 7: Google Drive ✅

**Status**: Complete

**Goal**: File management and storage integration.

### Core Operations
| Tool | Description | Status |
|------|-------------|--------|
| `drive_search` | Search files with query syntax | ✅ Complete |
| `drive_get` | Get file metadata | ✅ Complete |
| `drive_download` | Download file content | ✅ Complete |
| `drive_upload` | Upload new file | ✅ Complete |

### File Management
| Tool | Description | Status |
|------|-------------|--------|
| `drive_list` | List files in folder | ✅ Complete |
| `drive_create_folder` | Create new folder | ✅ Complete |
| `drive_move` | Move file to different folder | ✅ Complete |
| `drive_copy` | Copy a file | ✅ Complete |
| `drive_trash` | Move file to trash | ✅ Complete |
| `drive_delete` | Permanently delete file | ✅ Complete |

### Sharing
| Tool | Description | Status |
|------|-------------|--------|
| `drive_share` | Share file with users | ✅ Complete |
| `drive_get_permissions` | Get file permissions | ✅ Complete |

### Features
- Multi-account support via `account` parameter
- File ID extraction from Google Drive URLs
- Google Workspace file export (Docs→text, Sheets→CSV)
- File size limit for downloads (10MB)
- Base64 encoding for binary files

---

## Future Opportunities (Evaluated)

Research was conducted on potential expansions beyond the core Gmail + Calendar + Docs + Tasks stack.

### Google Drive ✅ Recommended

Natural extension of Docs support. Same OAuth credentials work with added Drive scope.
See Phase 7 above.

### Gemini API ⏸️ Use Separate MCP

Gemini integration is better served by dedicated MCP servers:
- [gemini-mcp](https://github.com/RLabs-Inc/gemini-mcp) - Deep research, image gen, YouTube analysis
- [mcp-server-gemini](https://github.com/aliargun/mcp-server-gemini) - Gemini 2.5 thinking, vision

**Rationale**: Gemini is a separate AI service, not a Google Workspace productivity tool. Use alongside gsuite-mcp rather than integrating.

### Google Photos ❌ Blocked

As of March 2025, Google Photos API is limited to **app-created content only**.
Cannot access user's existing photo library. Google recommends Photos Picker API for full access, which requires user interaction.

**Rationale**: API restrictions make meaningful integration impossible.

### Google Keep ❌ Blocked

No official Google Keep API exists. Existing solutions use:
- [gkeepapi](https://github.com/kiwiz/gkeepapi) - Unofficial Python client
- Requires Google Master Token (security risk - full account access)

**Rationale**: Unofficial API with dangerous authentication requirements.

### AdWords / AdSense / Analytics ❌ Out of Scope

Business/marketing tools, not personal productivity.
Different audience and use cases than gsuite-mcp.

**Rationale**: Should be separate MCP server(s) if needed.

---

## Token Reuse from google-workspace MCP

Automatic token import is supported. The auth manager checks:

**Location**: `~/.google_workspace_mcp/credentials/{email}.json`

**Format** (compatible):
```json
{
  "token": "ya29...",
  "refresh_token": "1//...",
  "token_uri": "https://oauth2.googleapis.com/token",
  "client_id": "...",
  "client_secret": "...",
  "scopes": ["https://www.googleapis.com/auth/gmail.modify", ...],
  "expiry": "2024-01-31T12:00:00.000000"
}
```

When credentials are not found for an account, gsuite-mcp automatically attempts to import from the legacy location.

---

## Required Scopes

### Gmail
```
https://www.googleapis.com/auth/gmail.modify    # Read, send, modify
https://www.googleapis.com/auth/gmail.labels    # Label management
https://www.googleapis.com/auth/gmail.settings.basic  # Filters
```

### Calendar
```
https://www.googleapis.com/auth/calendar         # Full calendar access
https://www.googleapis.com/auth/calendar.events  # Event management
```

### Docs
```
https://www.googleapis.com/auth/documents        # Full access to Google Docs
```

### Tasks
```
https://www.googleapis.com/auth/tasks            # Full access to Google Tasks
```

### Drive
```
https://www.googleapis.com/auth/drive            # Full access to Google Drive
```

---

## Related Issues

- [#14](https://github.com/aliwatters/gsuite-mcp/issues/14) - AGENTS.md documentation
- [#15](https://github.com/aliwatters/gsuite-mcp/issues/15) - CLAUDE.md documentation
- [Upstream feature request](https://github.com/taylorwilsdon/google_workspace_mcp/issues/410)

## References

- [Gmail API Reference](https://developers.google.com/gmail/api/reference/rest)
- [messages.modify](https://developers.google.com/gmail/api/reference/rest/v1/users.messages/modify)
- [messages.batchModify](https://developers.google.com/gmail/api/reference/rest/v1/users.messages/batchModify)
- [Calendar API Reference](https://developers.google.com/calendar/api/v3/reference)
- [events.list](https://developers.google.com/calendar/api/v3/reference/events/list)
- [events.insert](https://developers.google.com/calendar/api/v3/reference/events/insert)
- [Docs API Reference](https://developers.google.com/docs/api/reference/rest)
- [documents.get](https://developers.google.com/docs/api/reference/rest/v1/documents/get)
- [documents.batchUpdate](https://developers.google.com/docs/api/reference/rest/v1/documents/batchUpdate)
- [Tasks API Reference](https://developers.google.com/tasks/reference/rest)
- [tasks.list](https://developers.google.com/tasks/reference/rest/v1/tasks/list)
- [tasks.insert](https://developers.google.com/tasks/reference/rest/v1/tasks/insert)
