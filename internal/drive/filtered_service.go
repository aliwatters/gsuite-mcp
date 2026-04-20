package drive

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/drive/v3"
)

// FilteredDriveService wraps a DriveService and enforces drive access restrictions.
// Files in restricted drives are filtered from list results and blocked from direct access.
type FilteredDriveService struct {
	inner  DriveService
	filter *common.DriveAccessFilter
}

// NewFilteredDriveService creates a filtered service wrapping the given inner service.
func NewFilteredDriveService(inner DriveService, filter *common.DriveAccessFilter) *FilteredDriveService {
	return &FilteredDriveService{inner: inner, filter: filter}
}

// ensureResolved resolves drive names to IDs using the inner service's ListDrives.
func (f *FilteredDriveService) ensureResolved(ctx context.Context) {
	if f.filter.IsResolved() {
		return
	}

	var allDrives []common.DriveInfo
	pageToken := ""
	for {
		result, err := f.inner.ListDrives(ctx, 100, pageToken)
		if err != nil {
			return // fail open
		}
		for _, d := range result.Drives {
			allDrives = append(allDrives, common.DriveInfo{ID: d.Id, Name: d.Name})
		}
		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}
	f.filter.ResolveDriveNames(allDrives)
}

// checkFileAccess verifies that the file is in an allowed drive.
func (f *FilteredDriveService) checkFileAccess(ctx context.Context, fileID string) error {
	f.ensureResolved(ctx)
	file, err := f.inner.GetFile(ctx, fileID, "driveId")
	if err != nil {
		return nil // fail open
	}
	return f.filter.Check(file.DriveId)
}

// ensureDriveIDField ensures driveId is included in the requested fields.
func ensureDriveIDField(fields string) string {
	if fields == "" {
		return "driveId"
	}
	if strings.Contains(fields, "driveId") {
		return fields
	}
	return fields + ",driveId"
}

// === File Operations ===

// ListFiles returns results filtered to only include files from allowed drives.
func (f *FilteredDriveService) ListFiles(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
	f.ensureResolved(ctx)

	// Copy opts to avoid mutating the caller's struct when injecting driveId field
	innerOpts := opts
	if opts != nil && opts.Fields != "" {
		cp := *opts
		cp.Fields = ensureDriveIDField(opts.Fields)
		innerOpts = &cp
	}

	result, err := f.inner.ListFiles(ctx, innerOpts)
	if err != nil {
		return nil, fmt.Errorf("listing files: %w", err)
	}

	// Filter files by drive access
	filtered := make([]*drive.File, 0, len(result.Files))
	for _, file := range result.Files {
		if f.filter.Check(file.DriveId) == nil {
			filtered = append(filtered, file)
		}
	}
	result.Files = filtered

	return result, nil
}

// GetFile checks drive access after retrieving file metadata.
func (f *FilteredDriveService) GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error) {
	f.ensureResolved(ctx)
	file, err := f.inner.GetFile(ctx, fileID, ensureDriveIDField(fields))
	if err != nil {
		return nil, fmt.Errorf("getting file %s: %w", fileID, err)
	}
	if err := f.filter.Check(file.DriveId); err != nil {
		return nil, fmt.Errorf("GetFile drive access check: %w", err)
	}
	return file, nil
}

// CreateFile checks the destination parent's drive before creating.
func (f *FilteredDriveService) CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
	if len(file.Parents) > 0 {
		if err := f.checkFileAccess(ctx, file.Parents[0]); err != nil {
			return nil, fmt.Errorf("CreateFile parent access check: %w", err)
		}
	}
	return f.inner.CreateFile(ctx, file, content)
}

// UpdateFile checks drive access before updating.
func (f *FilteredDriveService) UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("UpdateFile access check: %w", err)
	}
	return f.inner.UpdateFile(ctx, fileID, file)
}

// MoveFile checks both source file and destination drive.
func (f *FilteredDriveService) MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("MoveFile source access check: %w", err)
	}
	if err := f.checkFileAccess(ctx, newParentID); err != nil {
		return nil, fmt.Errorf("MoveFile destination access check: %w", err)
	}
	return f.inner.MoveFile(ctx, fileID, newParentID, previousParents)
}

// CopyFile checks the source file's drive before copying.
func (f *FilteredDriveService) CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("CopyFile source access check: %w", err)
	}
	if len(file.Parents) > 0 {
		if err := f.checkFileAccess(ctx, file.Parents[0]); err != nil {
			return nil, fmt.Errorf("CopyFile destination access check: %w", err)
		}
	}
	return f.inner.CopyFile(ctx, fileID, file)
}

// DeleteFile checks drive access before deleting.
func (f *FilteredDriveService) DeleteFile(ctx context.Context, fileID string) error {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return fmt.Errorf("DeleteFile access check: %w", err)
	}
	return f.inner.DeleteFile(ctx, fileID)
}

// DownloadFile checks drive access before downloading.
func (f *FilteredDriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("DownloadFile access check: %w", err)
	}
	return f.inner.DownloadFile(ctx, fileID)
}

// ExportFile checks drive access before exporting.
func (f *FilteredDriveService) ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("ExportFile access check: %w", err)
	}
	return f.inner.ExportFile(ctx, fileID, mimeType)
}

// === Drives ===

// GetDrive checks if the requested drive is accessible.
func (f *FilteredDriveService) GetDrive(ctx context.Context, driveID string) (*drive.Drive, error) {
	f.ensureResolved(ctx)
	if err := f.filter.Check(driveID); err != nil {
		return nil, fmt.Errorf("GetDrive access check: %w", err)
	}
	return f.inner.GetDrive(ctx, driveID)
}

// ListDrives passes through to inner service (needed for name resolution).
func (f *FilteredDriveService) ListDrives(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
	return f.inner.ListDrives(ctx, pageSize, pageToken)
}

// === Permissions ===

func (f *FilteredDriveService) ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("ListPermissions access check: %w", err)
	}
	return f.inner.ListPermissions(ctx, fileID)
}

func (f *FilteredDriveService) CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("CreatePermission access check: %w", err)
	}
	return f.inner.CreatePermission(ctx, fileID, permission, sendNotification)
}

func (f *FilteredDriveService) DeletePermission(ctx context.Context, fileID string, permissionID string) error {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return fmt.Errorf("DeletePermission access check: %w", err)
	}
	return f.inner.DeletePermission(ctx, fileID, permissionID)
}

// === Comments ===

func (f *FilteredDriveService) ListComments(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.CommentList, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("ListComments access check: %w", err)
	}
	return f.inner.ListComments(ctx, fileID, fields, pageSize, pageToken, includeDeleted)
}

func (f *FilteredDriveService) GetComment(ctx context.Context, fileID string, commentID string, fields string, includeDeleted bool) (*drive.Comment, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("GetComment access check: %w", err)
	}
	return f.inner.GetComment(ctx, fileID, commentID, fields, includeDeleted)
}

func (f *FilteredDriveService) CreateComment(ctx context.Context, fileID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("CreateComment access check: %w", err)
	}
	return f.inner.CreateComment(ctx, fileID, comment, fields)
}

func (f *FilteredDriveService) UpdateComment(ctx context.Context, fileID string, commentID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("UpdateComment access check: %w", err)
	}
	return f.inner.UpdateComment(ctx, fileID, commentID, comment, fields)
}

func (f *FilteredDriveService) DeleteComment(ctx context.Context, fileID string, commentID string) error {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return fmt.Errorf("DeleteComment access check: %w", err)
	}
	return f.inner.DeleteComment(ctx, fileID, commentID)
}

// === Replies ===

func (f *FilteredDriveService) ListReplies(ctx context.Context, fileID string, commentID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.ReplyList, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("ListReplies access check: %w", err)
	}
	return f.inner.ListReplies(ctx, fileID, commentID, fields, pageSize, pageToken, includeDeleted)
}

func (f *FilteredDriveService) CreateReply(ctx context.Context, fileID string, commentID string, reply *drive.Reply, fields string) (*drive.Reply, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("CreateReply access check: %w", err)
	}
	return f.inner.CreateReply(ctx, fileID, commentID, reply, fields)
}

// === Revisions ===

func (f *FilteredDriveService) ListRevisions(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string) (*drive.RevisionList, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("ListRevisions access check: %w", err)
	}
	return f.inner.ListRevisions(ctx, fileID, fields, pageSize, pageToken)
}

func (f *FilteredDriveService) GetRevision(ctx context.Context, fileID string, revisionID string, fields string) (*drive.Revision, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("GetRevision access check: %w", err)
	}
	return f.inner.GetRevision(ctx, fileID, revisionID, fields)
}

func (f *FilteredDriveService) DownloadRevision(ctx context.Context, fileID string, revisionID string) (io.ReadCloser, error) {
	if err := f.checkFileAccess(ctx, fileID); err != nil {
		return nil, fmt.Errorf("DownloadRevision access check: %w", err)
	}
	return f.inner.DownloadRevision(ctx, fileID, revisionID)
}
