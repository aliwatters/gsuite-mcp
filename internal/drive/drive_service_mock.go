package drive

import (
	"context"
	"io"

	"google.golang.org/api/drive/v3"
)

// MockDriveService implements DriveService for testing.
type MockDriveService struct {
	// Files
	ListFilesFunc    func(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error)
	GetFileFunc      func(ctx context.Context, fileID string, fields string) (*drive.File, error)
	CreateFileFunc   func(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error)
	UpdateFileFunc   func(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	MoveFileFunc     func(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error)
	CopyFileFunc     func(ctx context.Context, fileID string, file *drive.File) (*drive.File, error)
	DeleteFileFunc   func(ctx context.Context, fileID string) error
	DownloadFileFunc func(ctx context.Context, fileID string) (io.ReadCloser, error)
	ExportFileFunc   func(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error)

	// Permissions
	ListPermissionsFunc  func(ctx context.Context, fileID string) (*drive.PermissionList, error)
	CreatePermissionFunc func(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error)
	DeletePermissionFunc func(ctx context.Context, fileID string, permissionID string) error
}

// File methods

func (m *MockDriveService) ListFiles(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
	if m.ListFilesFunc != nil {
		return m.ListFilesFunc(ctx, opts)
	}
	return &drive.FileList{}, nil
}

func (m *MockDriveService) GetFile(ctx context.Context, fileID string, fields string) (*drive.File, error) {
	if m.GetFileFunc != nil {
		return m.GetFileFunc(ctx, fileID, fields)
	}
	return &drive.File{}, nil
}

func (m *MockDriveService) CreateFile(ctx context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
	if m.CreateFileFunc != nil {
		return m.CreateFileFunc(ctx, file, content)
	}
	return file, nil
}

func (m *MockDriveService) UpdateFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	if m.UpdateFileFunc != nil {
		return m.UpdateFileFunc(ctx, fileID, file)
	}
	return file, nil
}

func (m *MockDriveService) MoveFile(ctx context.Context, fileID string, newParentID string, previousParents string) (*drive.File, error) {
	if m.MoveFileFunc != nil {
		return m.MoveFileFunc(ctx, fileID, newParentID, previousParents)
	}
	return &drive.File{}, nil
}

func (m *MockDriveService) CopyFile(ctx context.Context, fileID string, file *drive.File) (*drive.File, error) {
	if m.CopyFileFunc != nil {
		return m.CopyFileFunc(ctx, fileID, file)
	}
	return file, nil
}

func (m *MockDriveService) DeleteFile(ctx context.Context, fileID string) error {
	if m.DeleteFileFunc != nil {
		return m.DeleteFileFunc(ctx, fileID)
	}
	return nil
}

func (m *MockDriveService) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	if m.DownloadFileFunc != nil {
		return m.DownloadFileFunc(ctx, fileID)
	}
	return nil, nil
}

func (m *MockDriveService) ExportFile(ctx context.Context, fileID string, mimeType string) (io.ReadCloser, error) {
	if m.ExportFileFunc != nil {
		return m.ExportFileFunc(ctx, fileID, mimeType)
	}
	return nil, nil
}

// Permission methods

func (m *MockDriveService) ListPermissions(ctx context.Context, fileID string) (*drive.PermissionList, error) {
	if m.ListPermissionsFunc != nil {
		return m.ListPermissionsFunc(ctx, fileID)
	}
	return &drive.PermissionList{}, nil
}

func (m *MockDriveService) CreatePermission(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error) {
	if m.CreatePermissionFunc != nil {
		return m.CreatePermissionFunc(ctx, fileID, permission, sendNotification)
	}
	return permission, nil
}

func (m *MockDriveService) DeletePermission(ctx context.Context, fileID string, permissionID string) error {
	if m.DeletePermissionFunc != nil {
		return m.DeletePermissionFunc(ctx, fileID, permissionID)
	}
	return nil
}
