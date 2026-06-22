# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses a patch-first versioning policy — see [RELEASING.md](RELEASING.md) for the full policy.

## [0.4.5] - 2026-06-22

### Changed

- Updated Go dependencies: `golang.org/x/sync` to `0.21.0`, `google.golang.org/api` to `0.284.0`, and `modernc.org/sqlite` to `1.52.0` (#184)

## [0.4.4] - 2026-06-22

### Changed

- Updated GitHub Actions workflows to use `actions/checkout@v7` (#186)

## [0.4.3] - 2026-06-22

### Changed

- Gmail message `headers` is now a curated convenience map again, widened with high-value routing, threading, list, content-type, and trust headers such as `reply-to`, `in-reply-to`, `references`, `list-unsubscribe`, and `authentication-results` (#188)

### Fixed

- Gmail message `headers` no longer includes verbose `Received` traces, arbitrary headers, or full `DKIM-Signature` values; full ordered headers remain available via `payload_headers` (#188)

## [0.4.2] - 2026-06-22

### Added

- Gmail message reads now include `payload_headers`, preserving Gmail's ordered `payload.headers[]` list and repeated headers such as `Received` (#187)

### Changed

- Gmail message `headers` now includes all returned message headers as lowercased lookup keys instead of a curated subset (#187)

### Fixed

- `gmail_get` with `format: "raw"` now returns Gmail's raw RFC822 base64url payload when provided by the API (#187)

## [0.2.0] - 2026-03-20

### Added

- **Multi-installation support**: `--config-dir` flag and `GSUITE_MCP_CONFIG_DIR` env var to run multiple instances with different GCP projects (#85)
- **Citation tools** (11 tools): Large document indexing with chunking, FTS5 search, concept extraction, and citation formatting (feature-flagged via `large_doc_indexing`)
- **Google Forms** (5 tools): Get form structure, create forms, batch update, list/get responses (#78)
- **Google Slides** (5 tools): Read presentations, get slides, thumbnails, create, batch update (#77)
- **Gmail send-as aliases and delegation** (9 tools): Manage send-as identities and delegate access (#79)
- **Drive access filtering**: Configurable allowlist/blocklist for shared drives (#76)
- **Docs enhancements**: Markdown export, find-replace, PDF export, doc import (#74)
- **Docs styling and Sheets formatting**: Paragraph styles, charts, pivot tables (#73)
- **Calendar enhancements**: Focus Time, OOO events, conference data, free/busy (#72)
- **Drive enhancements**: Shareable links, file type filter, comments, revisions (#71)
- Protocol-level MCP tests for regression prevention (#83)

### Changed

- Default OAuth port changed from `8100` to `38917` to avoid conflicts with common dev servers

## [0.1.0] - 2026-02-09

### Added

- **Gmail** (36 tools): Search, read, send, reply, archive, trash, labels, filters, drafts, threads, batch operations, vacation responder
- **Calendar** (10 tools): List events, create/update/delete, recurring events, free/busy queries, Google Meet integration
- **Docs** (16 tools): Create, read, append, insert, replace, delete, formatting, lists, tables, images, headers/footers
- **Tasks** (10 tools): Lists, tasks, subtasks, due dates, completion, reordering
- **Sheets** (8 tools): Read, write, append, batch operations, create, clear
- **Contacts** (12 tools): List, search, create, update, delete, contact groups
- **Drive** (13 tools): Search, get, download, upload, list, create folder, move, copy, trash, delete, share, permissions
- Multi-account support with per-operation `account` parameter
- Dynamic OAuth2 credential discovery
- CLI subcommands: `init`, `auth`, `accounts`
