# Releasing gsuite-mcp

This document describes the release process and versioning policy for gsuite-mcp.

## Versioning Policy

gsuite-mcp uses a **patch-first** versioning policy. This is an intentional, documented
deviation from strict Semantic Versioning (which would treat any new feature as a minor bump).

| Change type | Bump | Example |
|---|---|---|
| Bug fix | **patch** | Fix a nil-pointer panic in Gmail search → `v0.3.3 → v0.3.4` |
| Small / low-risk additive feature (e.g. a new optional tool parameter, a new optional field in a response) | **patch** | Add optional `color_id` param to `calendar_create_event` → `v0.3.3 → v0.3.4` |
| Larger or notable feature set (multiple new tools, new service area) | **minor** | Add Google Chat support (9 tools) → `v0.3.x → v0.4.0` |
| Breaking change (removed or renamed tool/parameter, changed required parameter type, behavior change that breaks existing callers) | **major** | Rename `gmail_get` → `gmail_get_message` (breaking) → `v0.x.y → v1.0.0` |

**Guiding principle**: patch is the default for bug fixes and small/low-risk additive changes.
Reserve minor for releases that add meaningful new surface area (a new service, a batch of new
tools, or a notable capability). Major is for breaking changes only.

Strict SemVer would call any new feature a minor bump. For this project's cadence — where individual
optional parameters and small quality-of-life additions ship frequently — that would produce a lot
of minor bumps for changes that pose no compatibility risk. Patch-first keeps the version history
aligned with how the project actually evolves.

## Release Process

1. **Update `CHANGELOG.md`** — add a new `## [X.Y.Z] - YYYY-MM-DD` section with changes grouped
   under `### Added`, `### Changed`, `### Fixed`, `### Removed`.

2. **Tag the release**:
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```

3. **GoReleaser** — the `.goreleaser.yml` and `.github/workflows/release.yml` build and publish
   the release artifacts automatically when the tag is pushed.

4. **Verify** — confirm the GitHub release page shows the correct binaries and the CHANGELOG
   entry is accurate.

## Changelog Format

Follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) conventions:

- **Added** — new features (tools, parameters, flags)
- **Changed** — changes to existing behavior
- **Fixed** — bug fixes
- **Removed** — removed features or parameters
- **Security** — security fixes

Reference the GitHub issue or PR number in parentheses where applicable: `(#123)`.
