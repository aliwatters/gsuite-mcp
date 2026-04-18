package common

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveInfo holds minimal shared drive metadata for name resolution.
type DriveInfo struct {
	ID   string
	Name string
}

// DriveAccessFilter enforces allowlist/blocklist access control on shared drives.
// It is safe for concurrent use.
type DriveAccessFilter struct {
	config     *config.DriveAccess
	allowedIDs map[string]bool
	blockedIDs map[string]bool
	resolved   bool
	mu         sync.RWMutex
}

// NewDriveAccessFilter creates a filter from configuration.
// Returns nil if cfg is nil or has no restrictions.
func NewDriveAccessFilter(cfg *config.DriveAccess) *DriveAccessFilter {
	if cfg == nil {
		return nil
	}
	if len(cfg.Allowed) == 0 && len(cfg.Blocked) == 0 {
		return nil
	}
	return &DriveAccessFilter{config: cfg}
}

// IsActive returns true if the filter has any restrictions configured.
func (f *DriveAccessFilter) IsActive() bool {
	if f == nil || f.config == nil {
		return false
	}
	return len(f.config.Allowed) > 0 || len(f.config.Blocked) > 0
}

// ResolveDriveNames maps configured drive names to IDs using the provided drive list.
// Call once after listing shared drives. Safe to call multiple times (no-op after first).
func (f *DriveAccessFilter) ResolveDriveNames(drives []DriveInfo) {
	if f == nil {
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.resolved {
		return
	}

	nameToID := make(map[string]string)
	idSet := make(map[string]bool)
	for _, d := range drives {
		nameToID[strings.ToLower(d.Name)] = d.ID
		idSet[d.ID] = true
	}

	resolve := func(names []string) map[string]bool {
		ids := make(map[string]bool)
		for _, name := range names {
			lower := strings.ToLower(name)
			if id, ok := nameToID[lower]; ok {
				ids[id] = true
			} else if idSet[name] {
				// Config value is already an ID
				ids[name] = true
			}
			// Unresolved names are silently ignored — may be drives
			// the user doesn't have access to with this account.
		}
		return ids
	}

	if len(f.config.Allowed) > 0 {
		f.allowedIDs = resolve(f.config.Allowed)
	}
	if len(f.config.Blocked) > 0 {
		f.blockedIDs = resolve(f.config.Blocked)
	}

	f.resolved = true
}

// IsResolved returns true if drive names have been resolved to IDs.
func (f *DriveAccessFilter) IsResolved() bool {
	if f == nil {
		return false
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.resolved
}

// Check returns an error if the given driveID is not permitted.
// Empty driveID (My Drive) is always allowed.
func (f *DriveAccessFilter) Check(driveID string) error {
	if f == nil || !f.IsActive() {
		return nil
	}

	// My Drive is always allowed
	if driveID == "" {
		return nil
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	if !f.resolved {
		// Names not yet resolved — fail open to avoid blocking
		return nil
	}

	if f.allowedIDs != nil {
		if !f.allowedIDs[driveID] {
			return fmt.Errorf("access denied: file is in a shared drive not in the allowed list; check drive_access.allowed in config.json")
		}
		return nil
	}

	if f.blockedIDs != nil {
		if f.blockedIDs[driveID] {
			return fmt.Errorf("access denied: file is in a blocked shared drive; check drive_access.blocked in config.json")
		}
	}

	return nil
}

// ensureResolved resolves drive names to IDs if not already done.
func (f *DriveAccessFilter) ensureResolved(ctx context.Context, driveSrv *drive.Service) {
	f.mu.RLock()
	if f.resolved {
		f.mu.RUnlock()
		return
	}
	f.mu.RUnlock()

	var allDrives []DriveInfo
	pageToken := ""
	for {
		call := driveSrv.Drives.List().PageSize(100).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		result, err := call.Do()
		if err != nil {
			return // fail open
		}
		for _, d := range result.Drives {
			allDrives = append(allDrives, DriveInfo{ID: d.Id, Name: d.Name})
		}
		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	f.ResolveDriveNames(allDrives)
}

// CheckFileAccess checks whether a file is in an allowed drive using the Drive API.
// Fails open on API errors (does not block if check cannot be performed).
func (f *DriveAccessFilter) CheckFileAccess(ctx context.Context, client *http.Client, fileID string) error {
	if f == nil || !f.IsActive() {
		return nil
	}

	driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil // fail open
	}

	f.ensureResolved(ctx, driveSrv)

	file, err := driveSrv.Files.Get(fileID).
		Fields("driveId").
		SupportsAllDrives(true).
		Context(ctx).Do()
	if err != nil {
		return nil // fail open
	}

	return f.Check(file.DriveId)
}

// WithDriveAccessCheck wraps an MCP handler to check drive access before execution.
// paramName is the request parameter containing the file/document/spreadsheet ID.
// Use this to protect Docs, Sheets, and other services that access Drive files.
// Pass explicit appDeps to avoid reading from the global singleton; omit to fall
// back to GetDeps() for backward compatibility.
func WithDriveAccessCheck(handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), paramName string, appDeps ...*Deps) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var injected *Deps
	if len(appDeps) > 0 {
		injected = appDeps[0]
	}
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		d := injected
		if d == nil {
			d = GetDeps()
		}
		if d == nil || d.DriveAccessFilter == nil || !d.DriveAccessFilter.IsActive() {
			return handler(ctx, request)
		}

		fileID := ParseStringArg(request.Params.Arguments, paramName, "")
		if fileID == "" {
			return handler(ctx, request)
		}
		fileID = ExtractGoogleResourceID(fileID)

		email, err := ResolveAccountFromRequest(request)
		if err != nil {
			// Let the handler deal with auth errors
			return handler(ctx, request)
		}

		client, err := d.AuthManager.GetClientOrAuthenticate(ctx, email, false)
		if err != nil {
			return handler(ctx, request)
		}

		if err := d.DriveAccessFilter.CheckFileAccess(ctx, client, fileID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return handler(ctx, request)
	}
}
