package drive

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// === Drive sub-interfaces (domain-scoped, ≤5 methods each) ===

// DriveFileReader provides read access to Drive file metadata and content.
type DriveFileReader interface {
	ListFiles(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error)
	GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error)
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)
	ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error)
}

// DriveFileWriter provides write access to Drive file content.
type DriveFileWriter interface {
	CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error)
	UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error)
	CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	DeleteFile(ctx context.Context, fileID string) error
}

// DrivePermissionService manages Drive file permissions.
type DrivePermissionService interface {
	ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error)
	CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error)
	DeletePermission(ctx context.Context, fileID string, permissionID string) error
}

// DriveCommentService manages Drive file comments.
type DriveCommentService interface {
	ListComments(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.CommentList, error)
	GetComment(ctx context.Context, fileID string, commentID string, fields string, includeDeleted bool) (*drive.Comment, error)
	CreateComment(ctx context.Context, fileID string, comment *drive.Comment, fields string) (*drive.Comment, error)
	UpdateComment(ctx context.Context, fileID string, commentID string, comment *drive.Comment, fields string) (*drive.Comment, error)
	DeleteComment(ctx context.Context, fileID string, commentID string) error
}

// DriveReplyService manages Drive comment replies.
type DriveReplyService interface {
	ListReplies(ctx context.Context, fileID string, commentID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.ReplyList, error)
	CreateReply(ctx context.Context, fileID string, commentID string, reply *drive.Reply, fields string) (*drive.Reply, error)
}

// DriveRevisionService manages Drive file revisions.
type DriveRevisionService interface {
	ListRevisions(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string) (*drive.RevisionList, error)
	GetRevision(ctx context.Context, fileID string, revisionID string, fields string) (*drive.Revision, error)
	DownloadRevision(ctx context.Context, fileID string, revisionID string) (io.ReadCloser, error)
}

// DriveDriveService provides access to shared drive metadata.
type DriveDriveService interface {
	GetDrive(ctx context.Context, driveID string) (*drive.Drive, error)
	ListDrives(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error)
}

// DriveService defines the complete interface for Google Drive API operations.
// It is composed from focused sub-interfaces, each covering a single domain.
// This interface enables dependency injection and testing with mocks.
type DriveService interface {
	DriveFileReader
	DriveFileWriter
	DrivePermissionService
	DriveCommentService
	DriveReplyService
	DriveRevisionService
	DriveDriveService
}

// ListFilesOptions contains optional parameters for listing files.
type ListFilesOptions struct {
	Query     string
	PageSize  int64
	PageToken string
	OrderBy   string
	Fields    string
	Corpora   string
}

// RealDriveService wraps the Drive API client and implements DriveService.
type RealDriveService struct {
	service *drive.Service
}

// NewRealDriveService creates a new RealDriveService wrapping the given Drive API service.
func NewRealDriveService(service *drive.Service) *RealDriveService {
	return &RealDriveService{service: service}
}

// ListFiles lists files matching the query.
func (s *RealDriveService) ListFiles(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
	call := s.service.Files.List().Context(ctx).
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true)

	if opts != nil {
		if opts.Query != "" {
			call = call.Q(opts.Query)
		}
		if opts.PageSize > 0 {
			call = call.PageSize(opts.PageSize)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.OrderBy != "" {
			call = call.OrderBy(opts.OrderBy)
		}
		if opts.Fields != "" {
			call = call.Fields(googleapi.Field(opts.Fields))
		}
		if opts.Corpora != "" {
			call = call.Corpora(opts.Corpora)
		}
	}

	return call.Do()
}

// GetFile gets a file by ID.
func (s *RealDriveService) GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error) {
	call := s.service.Files.Get(fileID).Context(ctx).
		SupportsAllDrives(true)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// CreateFile creates a new file.
func (s *RealDriveService) CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
	call := s.service.Files.Create(file).Context(ctx).
		SupportsAllDrives(true)
	if content != nil {
		call = call.Media(content)
	}
	return call.Do()
}

// UpdateFile updates an existing file's metadata.
func (s *RealDriveService) UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	return s.service.Files.Update(fileID, file).Context(ctx).
		SupportsAllDrives(true).Do()
}

// MoveFile moves a file to a new folder.
func (s *RealDriveService) MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error) {
	return s.service.Files.Update(fileID, nil).
		Context(ctx).
		SupportsAllDrives(true).
		AddParents(newParentID).
		RemoveParents(previousParents).
		Fields("id, name, parents, webViewLink").
		Do()
}

// CopyFile creates a copy of a file.
func (s *RealDriveService) CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	return s.service.Files.Copy(fileID, file).Context(ctx).
		SupportsAllDrives(true).Do()
}

// DeleteFile permanently deletes a file.
func (s *RealDriveService) DeleteFile(ctx context.Context, fileID string) error {
	return s.service.Files.Delete(fileID).Context(ctx).
		SupportsAllDrives(true).Do()
}

// DownloadFile downloads a file's content.
func (s *RealDriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := s.service.Files.Get(fileID).Context(ctx).
		SupportsAllDrives(true).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ExportFile exports a Google Workspace document to the specified format.
func (s *RealDriveService) ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error) {
	resp, err := s.service.Files.Export(fileID, mimeType).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// GetDrive gets a shared drive by ID.
func (s *RealDriveService) GetDrive(ctx context.Context, driveID string) (*drive.Drive, error) {
	return s.service.Drives.Get(driveID).Context(ctx).Fields("id,name").Do()
}

// ListDrives lists shared drives accessible to the user.
func (s *RealDriveService) ListDrives(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
	call := s.service.Drives.List().Context(ctx).PageSize(pageSize)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Do()
}

// ListPermissions lists a file's permissions.
func (s *RealDriveService) ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error) {
	return s.service.Permissions.List(fileID).Context(ctx).
		SupportsAllDrives(true).Do()
}

// CreatePermission creates a permission for a file.
func (s *RealDriveService) CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error) {
	return s.service.Permissions.Create(fileID, permission).
		Context(ctx).
		SupportsAllDrives(true).
		SendNotificationEmail(sendNotification).
		Do()
}

// DeletePermission deletes a permission from a file.
func (s *RealDriveService) DeletePermission(ctx context.Context, fileID string, permissionID string) error {
	return s.service.Permissions.Delete(fileID, permissionID).Context(ctx).
		SupportsAllDrives(true).Do()
}

// ListComments lists comments on a file.
func (s *RealDriveService) ListComments(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.CommentList, error) {
	call := s.service.Comments.List(fileID).Context(ctx).
		IncludeDeleted(includeDeleted)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Do()
}

// GetComment gets a comment by ID.
func (s *RealDriveService) GetComment(ctx context.Context, fileID string, commentID string, fields string, includeDeleted bool) (*drive.Comment, error) {
	call := s.service.Comments.Get(fileID, commentID).Context(ctx).
		IncludeDeleted(includeDeleted)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// CreateComment creates a comment on a file.
func (s *RealDriveService) CreateComment(ctx context.Context, fileID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	call := s.service.Comments.Create(fileID, comment).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// UpdateComment updates a comment.
func (s *RealDriveService) UpdateComment(ctx context.Context, fileID string, commentID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	call := s.service.Comments.Update(fileID, commentID, comment).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// DeleteComment deletes a comment.
func (s *RealDriveService) DeleteComment(ctx context.Context, fileID string, commentID string) error {
	return s.service.Comments.Delete(fileID, commentID).Context(ctx).Do()
}

// ListReplies lists replies on a comment.
func (s *RealDriveService) ListReplies(ctx context.Context, fileID string, commentID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.ReplyList, error) {
	call := s.service.Replies.List(fileID, commentID).Context(ctx).
		IncludeDeleted(includeDeleted)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Do()
}

// CreateReply creates a reply on a comment.
func (s *RealDriveService) CreateReply(ctx context.Context, fileID string, commentID string, reply *drive.Reply, fields string) (*drive.Reply, error) {
	call := s.service.Replies.Create(fileID, commentID, reply).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// ListRevisions lists revisions of a file.
func (s *RealDriveService) ListRevisions(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string) (*drive.RevisionList, error) {
	call := s.service.Revisions.List(fileID).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Do()
}

// GetRevision gets a revision by ID.
func (s *RealDriveService) GetRevision(ctx context.Context, fileID string, revisionID string, fields string) (*drive.Revision, error) {
	call := s.service.Revisions.Get(fileID, revisionID).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// DownloadRevision downloads the content of a specific revision.
func (s *RealDriveService) DownloadRevision(ctx context.Context, fileID string, revisionID string) (io.ReadCloser, error) {
	resp, err := s.service.Revisions.Get(fileID, revisionID).Context(ctx).
		AcknowledgeAbuse(true).Download()
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
