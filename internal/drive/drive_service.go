package drive

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// DriveService defines the interface for Google Drive API operations.
// This interface enables dependency injection and testing with mocks.
type DriveService interface {
	// Files
	ListFiles(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error)
	GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error)
	CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error)
	UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error)
	CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	DeleteFile(ctx context.Context, fileID string) error
	DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error)
	ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error)

	// Permissions
	ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error)
	CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error)
	DeletePermission(ctx context.Context, fileID string, permissionID string) error
}

// ListFilesOptions contains optional parameters for listing files.
type ListFilesOptions struct {
	Query     string
	PageSize  int64
	PageToken string
	OrderBy   string
	Fields    string
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
	call := s.service.Files.List().Context(ctx)

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
	}

	return call.Do()
}

// GetFile gets a file by ID.
func (s *RealDriveService) GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error) {
	call := s.service.Files.Get(fileID).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// CreateFile creates a new file.
func (s *RealDriveService) CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
	call := s.service.Files.Create(file).Context(ctx)
	if content != nil {
		call = call.Media(content)
	}
	return call.Do()
}

// UpdateFile updates an existing file's metadata.
func (s *RealDriveService) UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	return s.service.Files.Update(fileID, file).Context(ctx).Do()
}

// MoveFile moves a file to a new folder.
func (s *RealDriveService) MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error) {
	return s.service.Files.Update(fileID, nil).
		Context(ctx).
		AddParents(newParentID).
		RemoveParents(previousParents).
		Fields("id, name, parents, webViewLink").
		Do()
}

// CopyFile creates a copy of a file.
func (s *RealDriveService) CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	return s.service.Files.Copy(fileID, file).Context(ctx).Do()
}

// DeleteFile permanently deletes a file.
func (s *RealDriveService) DeleteFile(ctx context.Context, fileID string) error {
	return s.service.Files.Delete(fileID).Context(ctx).Do()
}

// DownloadFile downloads a file's content.
func (s *RealDriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	resp, err := s.service.Files.Get(fileID).Context(ctx).Download()
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

// ListPermissions lists a file's permissions.
func (s *RealDriveService) ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error) {
	return s.service.Permissions.List(fileID).Context(ctx).Do()
}

// CreatePermission creates a permission for a file.
func (s *RealDriveService) CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error) {
	return s.service.Permissions.Create(fileID, permission).
		Context(ctx).
		SendNotificationEmail(sendNotification).
		Do()
}

// DeletePermission deletes a permission from a file.
func (s *RealDriveService) DeletePermission(ctx context.Context, fileID string, permissionID string) error {
	return s.service.Permissions.Delete(fileID, permissionID).Context(ctx).Do()
}
