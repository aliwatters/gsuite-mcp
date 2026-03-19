package drive

import (
	"context"
	"io"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/aliwatters/gsuite-mcp/internal/config"
	"google.golang.org/api/drive/v3"
)

func newTestFilter(allowed, blocked []string) *common.DriveAccessFilter {
	f := common.NewDriveAccessFilter(&config.DriveAccess{
		Allowed: allowed,
		Blocked: blocked,
	})
	f.ResolveDriveNames([]common.DriveInfo{
		{ID: "drive-marketing", Name: "Marketing"},
		{ID: "drive-sensitive", Name: "SENSITIVE"},
		{ID: "drive-hr", Name: "HR"},
	})
	return f
}

func TestFilteredDriveService_ListFiles_Allowlist(t *testing.T) {
	filter := newTestFilter([]string{"Marketing"}, nil)
	mock := &MockDriveService{
		ListFilesFunc: func(ctx context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
			return &drive.FileList{
				Files: []*drive.File{
					{Id: "1", Name: "allowed.doc", DriveId: "drive-marketing"},
					{Id: "2", Name: "blocked.doc", DriveId: "drive-sensitive"},
					{Id: "3", Name: "mydrive.doc", DriveId: ""},
				},
			}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	result, err := srv.ListFiles(context.Background(), &ListFilesOptions{Fields: "files(id,name)"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}
	if result.Files[0].Id != "1" {
		t.Errorf("expected first file id=1, got %s", result.Files[0].Id)
	}
	if result.Files[1].Id != "3" {
		t.Errorf("expected second file id=3, got %s", result.Files[1].Id)
	}
}

func TestFilteredDriveService_GetFile_Blocked(t *testing.T) {
	filter := newTestFilter([]string{"Marketing"}, nil)
	mock := &MockDriveService{
		GetFileFunc: func(ctx context.Context, fileID string, fields string) (*drive.File, error) {
			return &drive.File{Id: fileID, DriveId: "drive-sensitive"}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	_, err := srv.GetFile(context.Background(), "file1", "id,name")
	if err == nil {
		t.Fatal("expected error for file in blocked drive")
	}
}

func TestFilteredDriveService_GetFile_Allowed(t *testing.T) {
	filter := newTestFilter([]string{"Marketing"}, nil)
	mock := &MockDriveService{
		GetFileFunc: func(ctx context.Context, fileID string, fields string) (*drive.File, error) {
			return &drive.File{Id: fileID, DriveId: "drive-marketing"}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	file, err := srv.GetFile(context.Background(), "file1", "id,name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Id != "file1" {
		t.Errorf("expected file1, got %s", file.Id)
	}
}

func TestFilteredDriveService_DeleteFile_Blocked(t *testing.T) {
	filter := newTestFilter(nil, []string{"SENSITIVE"})
	mock := &MockDriveService{
		GetFileFunc: func(ctx context.Context, fileID string, fields string) (*drive.File, error) {
			return &drive.File{Id: fileID, DriveId: "drive-sensitive"}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
		DeleteFileFunc: func(ctx context.Context, fileID string) error {
			t.Fatal("DeleteFile should not be called for blocked file")
			return nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	err := srv.DeleteFile(context.Background(), "file-in-sensitive")
	if err == nil {
		t.Fatal("expected error for file in blocked drive")
	}
}

func TestFilteredDriveService_DownloadFile_MyDrive(t *testing.T) {
	filter := newTestFilter([]string{"Marketing"}, nil)
	mock := &MockDriveService{
		GetFileFunc: func(ctx context.Context, fileID string, fields string) (*drive.File, error) {
			return &drive.File{Id: fileID, DriveId: ""}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
		DownloadFileFunc: func(ctx context.Context, fileID string) (io.ReadCloser, error) {
			return nil, io.ErrUnexpectedEOF
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	// My Drive files should pass the access check
	_, err := srv.DownloadFile(context.Background(), "my-file")
	// The error should come from the download (not access denied)
	if err == nil || err != io.ErrUnexpectedEOF {
		t.Fatalf("expected download to proceed for My Drive file, got: %v", err)
	}
}

func TestFilteredDriveService_GetDrive_Blocked(t *testing.T) {
	filter := newTestFilter(nil, []string{"SENSITIVE"})
	mock := &MockDriveService{
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	_, err := srv.GetDrive(context.Background(), "drive-sensitive")
	if err == nil {
		t.Fatal("expected error for blocked drive")
	}
}

func TestFilteredDriveService_MoveFile_BlockedDestination(t *testing.T) {
	filter := newTestFilter(nil, []string{"SENSITIVE"})
	mock := &MockDriveService{
		GetFileFunc: func(ctx context.Context, fileID string, fields string) (*drive.File, error) {
			if fileID == "src-file" {
				return &drive.File{Id: fileID, DriveId: ""}, nil
			}
			return &drive.File{Id: fileID, DriveId: "drive-sensitive"}, nil
		},
		ListDrivesFunc: func(ctx context.Context, pageSize int64, pageToken string) (*drive.DriveList, error) {
			return &drive.DriveList{}, nil
		},
	}

	srv := NewFilteredDriveService(mock, filter)
	_, err := srv.MoveFile(context.Background(), "src-file", "dest-in-sensitive", "old-parent")
	if err == nil {
		t.Fatal("expected error when moving to blocked drive")
	}
}
