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

	// Drives
	GetDriveFunc func(ctx context.Context, driveID string) (*drive.Drive, error)

	// Permissions
	ListPermissionsFunc  func(ctx context.Context, fileID string) (*drive.PermissionList, error)
	CreatePermissionFunc func(ctx context.Context, fileID string, permission *drive.Permission, sendNotification bool) (*drive.Permission, error)
	DeletePermissionFunc func(ctx context.Context, fileID string, permissionID string) error

	// Comments
	ListCommentsFunc  func(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.CommentList, error)
	GetCommentFunc    func(ctx context.Context, fileID string, commentID string, fields string, includeDeleted bool) (*drive.Comment, error)
	CreateCommentFunc func(ctx context.Context, fileID string, comment *drive.Comment, fields string) (*drive.Comment, error)
	UpdateCommentFunc func(ctx context.Context, fileID string, commentID string, comment *drive.Comment, fields string) (*drive.Comment, error)
	DeleteCommentFunc func(ctx context.Context, fileID string, commentID string) error

	// Replies
	ListRepliesFunc func(ctx context.Context, fileID string, commentID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.ReplyList, error)
	CreateReplyFunc func(ctx context.Context, fileID string, commentID string, reply *drive.Reply, fields string) (*drive.Reply, error)

	// Revisions
	ListRevisionsFunc    func(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string) (*drive.RevisionList, error)
	GetRevisionFunc      func(ctx context.Context, fileID string, revisionID string, fields string) (*drive.Revision, error)
	DownloadRevisionFunc func(ctx context.Context, fileID string, revisionID string) (io.ReadCloser, error)
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

// Drive methods

func (m *MockDriveService) GetDrive(ctx context.Context, driveID string) (*drive.Drive, error) {
	if m.GetDriveFunc != nil {
		return m.GetDriveFunc(ctx, driveID)
	}
	return &drive.Drive{}, nil
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

// Comment methods

func (m *MockDriveService) ListComments(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.CommentList, error) {
	if m.ListCommentsFunc != nil {
		return m.ListCommentsFunc(ctx, fileID, fields, pageSize, pageToken, includeDeleted)
	}
	return &drive.CommentList{}, nil
}

func (m *MockDriveService) GetComment(ctx context.Context, fileID string, commentID string, fields string, includeDeleted bool) (*drive.Comment, error) {
	if m.GetCommentFunc != nil {
		return m.GetCommentFunc(ctx, fileID, commentID, fields, includeDeleted)
	}
	return &drive.Comment{}, nil
}

func (m *MockDriveService) CreateComment(ctx context.Context, fileID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	if m.CreateCommentFunc != nil {
		return m.CreateCommentFunc(ctx, fileID, comment, fields)
	}
	return comment, nil
}

func (m *MockDriveService) UpdateComment(ctx context.Context, fileID string, commentID string, comment *drive.Comment, fields string) (*drive.Comment, error) {
	if m.UpdateCommentFunc != nil {
		return m.UpdateCommentFunc(ctx, fileID, commentID, comment, fields)
	}
	return comment, nil
}

func (m *MockDriveService) DeleteComment(ctx context.Context, fileID string, commentID string) error {
	if m.DeleteCommentFunc != nil {
		return m.DeleteCommentFunc(ctx, fileID, commentID)
	}
	return nil
}

// Reply methods

func (m *MockDriveService) ListReplies(ctx context.Context, fileID string, commentID string, fields string, pageSize int64, pageToken string, includeDeleted bool) (*drive.ReplyList, error) {
	if m.ListRepliesFunc != nil {
		return m.ListRepliesFunc(ctx, fileID, commentID, fields, pageSize, pageToken, includeDeleted)
	}
	return &drive.ReplyList{}, nil
}

func (m *MockDriveService) CreateReply(ctx context.Context, fileID string, commentID string, reply *drive.Reply, fields string) (*drive.Reply, error) {
	if m.CreateReplyFunc != nil {
		return m.CreateReplyFunc(ctx, fileID, commentID, reply, fields)
	}
	return reply, nil
}

// Revision methods

func (m *MockDriveService) ListRevisions(ctx context.Context, fileID string, fields string, pageSize int64, pageToken string) (*drive.RevisionList, error) {
	if m.ListRevisionsFunc != nil {
		return m.ListRevisionsFunc(ctx, fileID, fields, pageSize, pageToken)
	}
	return &drive.RevisionList{}, nil
}

func (m *MockDriveService) GetRevision(ctx context.Context, fileID string, revisionID string, fields string) (*drive.Revision, error) {
	if m.GetRevisionFunc != nil {
		return m.GetRevisionFunc(ctx, fileID, revisionID, fields)
	}
	return &drive.Revision{}, nil
}

func (m *MockDriveService) DownloadRevision(ctx context.Context, fileID string, revisionID string) (io.ReadCloser, error) {
	if m.DownloadRevisionFunc != nil {
		return m.DownloadRevisionFunc(ctx, fileID, revisionID)
	}
	return nil, nil
}
