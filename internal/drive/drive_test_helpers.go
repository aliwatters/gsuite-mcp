package drive

import (
	"context"
	"io"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/drive/v3"
)

// DriveTestFixtures provides pre-configured test data for Drive tests.
type DriveTestFixtures struct {
	DefaultEmail string
	MockService  *MockDriveService
	Deps         *DriveHandlerDeps
}

// NewDriveTestFixtures creates a new test fixtures instance with default configuration.
func NewDriveTestFixtures() *DriveTestFixtures {
	mockService := &MockDriveService{}
	setupDefaultDriveMockData(mockService)
	f := common.NewTestFixtures[DriveService](mockService)

	return &DriveTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}

// setupDefaultDriveMockData populates the mock service with standard test data.
func setupDefaultDriveMockData(mock *MockDriveService) {
	// Set up ListFiles to return sample files
	mock.ListFilesFunc = func(_ context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
		return &drive.FileList{
			Files: []*drive.File{
				createTestFile("file001", "Document.docx", "application/vnd.google-apps.document", 1024),
				createTestFile("file002", "Spreadsheet.xlsx", "application/vnd.google-apps.spreadsheet", 2048),
				createTestFile("folder001", "My Folder", "application/vnd.google-apps.folder", 0),
			},
		}, nil
	}

	// Set up GetFile to return a file by ID
	mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
		switch fileID {
		case "file001":
			return createTestFile("file001", "Document.docx", "application/vnd.google-apps.document", 1024), nil
		case "file002":
			return createTestFile("file002", "Spreadsheet.xlsx", "application/vnd.google-apps.spreadsheet", 2048), nil
		case "folder001":
			return createTestFile("folder001", "My Folder", "application/vnd.google-apps.folder", 0), nil
		default:
			return createTestFile(fileID, "Unknown File", "application/octet-stream", 0), nil
		}
	}

	// Set up CreateFile to return a new file
	mock.CreateFileFunc = func(_ context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
		file.Id = "newfile001"
		file.CreatedTime = "2024-02-01T12:00:00Z"
		file.WebViewLink = "https://drive.google.com/file/d/newfile001/view"
		if file.Size == 0 && content != nil {
			// Read content to simulate getting size
			data, _ := io.ReadAll(content)
			file.Size = int64(len(data))
		}
		return file, nil
	}

	// Set up UpdateFile
	mock.UpdateFileFunc = func(_ context.Context, fileID string, file *drive.File) (*drive.File, error) {
		file.Id = fileID
		file.ModifiedTime = "2024-02-01T14:00:00Z"
		return file, nil
	}

	// Set up CopyFile
	mock.CopyFileFunc = func(_ context.Context, fileID string, file *drive.File) (*drive.File, error) {
		file.Id = "copy_" + fileID
		file.CreatedTime = "2024-02-01T12:00:00Z"
		file.WebViewLink = "https://drive.google.com/file/d/copy_" + fileID + "/view"
		return file, nil
	}

	// Set up DeleteFile
	mock.DeleteFileFunc = func(_ context.Context, fileID string) error {
		return nil
	}

	// Set up DownloadFile to return content
	mock.DownloadFileFunc = func(_ context.Context, fileID string) (io.ReadCloser, error) {
		content := "Test file content for " + fileID
		return io.NopCloser(strings.NewReader(content)), nil
	}

	// Set up ExportFile to return exported content
	mock.ExportFileFunc = func(_ context.Context, fileID string, mimeType string) (io.ReadCloser, error) {
		content := "Exported content for " + fileID + " as " + mimeType
		return io.NopCloser(strings.NewReader(content)), nil
	}

	// Set up ListPermissions
	mock.ListPermissionsFunc = func(_ context.Context, fileID string) (*drive.PermissionList, error) {
		return &drive.PermissionList{
			Permissions: []*drive.Permission{
				{
					Id:           "perm001",
					Type:         "user",
					Role:         "owner",
					EmailAddress: "owner@example.com",
				},
				{
					Id:           "perm002",
					Type:         "user",
					Role:         "writer",
					EmailAddress: "editor@example.com",
				},
			},
		}, nil
	}

	// Set up CreatePermission
	mock.CreatePermissionFunc = func(_ context.Context, fileID string, permission *drive.Permission, _ bool) (*drive.Permission, error) {
		permission.Id = "newperm001"
		return permission, nil
	}

	// Set up DeletePermission
	mock.DeletePermissionFunc = func(_ context.Context, fileID string, permissionID string) error {
		return nil
	}
}

// createTestFile creates a File with standard fields.
func createTestFile(id, name, mimeType string, size int64) *drive.File {
	file := &drive.File{
		Id:           id,
		Name:         name,
		MimeType:     mimeType,
		Size:         size,
		CreatedTime:  "2024-02-01T12:00:00Z",
		ModifiedTime: "2024-02-01T12:00:00Z",
		WebViewLink:  "https://drive.google.com/file/d/" + id + "/view",
	}

	if mimeType == "application/vnd.google-apps.folder" {
		file.WebViewLink = "https://drive.google.com/drive/folders/" + id
	}

	return file
}

// createTestFileWithParent creates a File in a specific folder.
func createTestFileWithParent(id, name, mimeType, parentID string) *drive.File {
	file := createTestFile(id, name, mimeType, 1024)
	file.Parents = []string{parentID}
	return file
}
