package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// parseResult unmarshals the JSON text from a tool result into a map.
func parseResult(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	return data
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name      string
		parents   []string
		setupMock func(mock *MockDriveService)
		want      string
	}{
		{
			name:    "no parents returns empty",
			parents: nil,
			want:    "",
		},
		{
			name:    "file in My Drive root",
			parents: []string{"root1"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					if fileID == "root1" {
						return &drive.File{Id: "root1", Name: "My Drive"}, nil
					}
					return nil, fmt.Errorf("unexpected file: %s", fileID)
				}
			},
			want: "My Drive",
		},
		{
			name:    "nested folders in My Drive",
			parents: []string{"folder2"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					switch fileID {
					case "folder2":
						return &drive.File{Id: "folder2", Name: "2025", Parents: []string{"folder1"}}, nil
					case "folder1":
						return &drive.File{Id: "folder1", Name: "Projects", Parents: []string{"root1"}}, nil
					case "root1":
						return &drive.File{Id: "root1", Name: "My Drive"}, nil
					default:
						return nil, fmt.Errorf("unexpected file: %s", fileID)
					}
				}
			},
			want: "My Drive/Projects/2025",
		},
		{
			name:    "shared drive root folder",
			parents: []string{"sdroot"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					if fileID == "sdroot" {
						return &drive.File{Id: "sdroot", Name: "sdroot", DriveId: "drive1"}, nil
					}
					return nil, fmt.Errorf("unexpected file: %s", fileID)
				}
				mock.GetDriveFunc = func(_ context.Context, driveID string) (*drive.Drive, error) {
					return &drive.Drive{Id: "drive1", Name: "Team Drive"}, nil
				}
			},
			want: "Team Drive",
		},
		{
			name:    "shared drive nested folder",
			parents: []string{"subfolder"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					switch fileID {
					case "subfolder":
						return &drive.File{Id: "subfolder", Name: "Folder", Parents: []string{"sdroot"}, DriveId: "drive1"}, nil
					case "sdroot":
						return &drive.File{Id: "sdroot", Name: "sdroot", DriveId: "drive1"}, nil
					default:
						return nil, fmt.Errorf("unexpected file: %s", fileID)
					}
				}
				mock.GetDriveFunc = func(_ context.Context, driveID string) (*drive.Drive, error) {
					return &drive.Drive{Id: "drive1", Name: "Team Drive"}, nil
				}
			},
			want: "Team Drive/Folder",
		},
		{
			name:    "API error returns empty",
			parents: []string{"folder1"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					return nil, fmt.Errorf("API error")
				}
			},
			want: "",
		},
		{
			name:    "GetDrive error falls back to drive ID",
			parents: []string{"sdroot"},
			setupMock: func(mock *MockDriveService) {
				mock.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
					if fileID == "sdroot" {
						return &drive.File{Id: "sdroot", Name: "sdroot", DriveId: "drive1"}, nil
					}
					return nil, fmt.Errorf("unexpected file: %s", fileID)
				}
				mock.GetDriveFunc = func(_ context.Context, driveID string) (*drive.Drive, error) {
					return nil, fmt.Errorf("forbidden")
				}
			},
			want: "drive1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockDriveService{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			resolver := NewPathResolver(mock)
			got := resolver.ResolvePath(context.Background(), tt.parents)
			if got != tt.want {
				t.Errorf("ResolvePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPathResolverCache(t *testing.T) {
	getFileCalls := 0
	mock := &MockDriveService{
		GetFileFunc: func(_ context.Context, fileID string, _ string) (*drive.File, error) {
			getFileCalls++
			switch fileID {
			case "shared_parent":
				return &drive.File{Id: "shared_parent", Name: "Shared Folder", Parents: []string{"root1"}}, nil
			case "root1":
				return &drive.File{Id: "root1", Name: "My Drive"}, nil
			default:
				return nil, fmt.Errorf("unexpected file: %s", fileID)
			}
		},
	}

	resolver := NewPathResolver(mock)
	ctx := context.Background()

	got1 := resolver.ResolvePath(ctx, []string{"shared_parent"})
	if got1 != "My Drive/Shared Folder" {
		t.Errorf("first resolve = %q, want %q", got1, "My Drive/Shared Folder")
	}
	firstCallCount := getFileCalls

	// Second resolve with same parent should hit cache
	got2 := resolver.ResolvePath(ctx, []string{"shared_parent"})
	if got2 != "My Drive/Shared Folder" {
		t.Errorf("second resolve = %q, want %q", got2, "My Drive/Shared Folder")
	}

	if getFileCalls != firstCallCount {
		t.Errorf("expected cache hit: GetFile called %d times after second resolve, want %d", getFileCalls, firstCallCount)
	}
}

func TestPathResolverMaxDepth(t *testing.T) {
	callCount := 0
	mock := &MockDriveService{
		GetFileFunc: func(_ context.Context, fileID string, _ string) (*drive.File, error) {
			callCount++
			return &drive.File{
				Id:      fileID,
				Name:    "Folder",
				Parents: []string{"parent_of_" + fileID},
			}, nil
		},
	}

	resolver := NewPathResolver(mock)
	got := resolver.ResolvePath(context.Background(), []string{"start"})

	if callCount > maxPathDepth {
		t.Errorf("GetFile called %d times, want at most %d", callCount, maxPathDepth)
	}

	if got != "" {
		segments := strings.Count(got, "/") + 1
		if segments > maxPathDepth {
			t.Errorf("path has %d segments, want at most %d", segments, maxPathDepth)
		}
	}
}

// TestDriveGetIncludesPath verifies TestableDriveGet includes the path field.
func TestDriveGetIncludesPath(t *testing.T) {
	fixtures := NewDriveTestFixtures()

	fixtures.MockService.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
		switch fileID {
		case "file001":
			f := createTestFile("file001", "Document.docx", "application/vnd.google-apps.document", 1024)
			f.Parents = []string{"folder001"}
			return f, nil
		case "folder001":
			return &drive.File{Id: "folder001", Name: "My Drive"}, nil
		default:
			return createTestFile(fileID, "Unknown", "application/octet-stream", 0), nil
		}
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file001"})
	result, err := TestableDriveGet(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := parseResult(t, result)
	path, ok := data["path"]
	if !ok {
		t.Fatal("expected 'path' field in result, got none")
	}
	if path != "My Drive" {
		t.Errorf("got path %q, want %q", path, "My Drive")
	}
}

// TestDriveSearchIncludesPath verifies TestableDriveSearch includes path per file.
func TestDriveSearchIncludesPath(t *testing.T) {
	fixtures := NewDriveTestFixtures()

	fixtures.MockService.ListFilesFunc = func(_ context.Context, _ *ListFilesOptions) (*drive.FileList, error) {
		f := createTestFile("file001", "Report.docx", "application/vnd.google-apps.document", 1024)
		f.Parents = []string{"folder001"}
		return &drive.FileList{Files: []*drive.File{f}}, nil
	}
	fixtures.MockService.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
		if fileID == "folder001" {
			return &drive.File{Id: "folder001", Name: "My Drive"}, nil
		}
		return nil, fmt.Errorf("unexpected file: %s", fileID)
	}

	request := common.CreateMCPRequest(map[string]any{"query": "name contains 'Report'"})
	result, err := TestableDriveSearch(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := parseResult(t, result)
	files := data["files"].([]any)
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}

	file := files[0].(map[string]any)
	path, ok := file["path"]
	if !ok {
		t.Fatal("expected 'path' field in search result file, got none")
	}
	if path != "My Drive" {
		t.Errorf("got path %q, want %q", path, "My Drive")
	}
}

// TestDriveListIncludesPath verifies TestableDriveList includes path per file.
func TestDriveListIncludesPath(t *testing.T) {
	fixtures := NewDriveTestFixtures()

	fixtures.MockService.ListFilesFunc = func(_ context.Context, _ *ListFilesOptions) (*drive.FileList, error) {
		f := createTestFile("file001", "Notes.txt", "text/plain", 512)
		f.Parents = []string{"subfolder"}
		return &drive.FileList{Files: []*drive.File{f}}, nil
	}
	fixtures.MockService.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
		switch fileID {
		case "subfolder":
			return &drive.File{Id: "subfolder", Name: "Projects", Parents: []string{"root1"}}, nil
		case "root1":
			return &drive.File{Id: "root1", Name: "My Drive"}, nil
		default:
			return nil, fmt.Errorf("unexpected file: %s", fileID)
		}
	}

	request := common.CreateMCPRequest(map[string]any{"folder_id": "subfolder"})
	result, err := TestableDriveList(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := parseResult(t, result)
	files := data["files"].([]any)
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}

	file := files[0].(map[string]any)
	path, ok := file["path"]
	if !ok {
		t.Fatal("expected 'path' field in list result file, got none")
	}
	if path != "My Drive/Projects" {
		t.Errorf("got path %q, want %q", path, "My Drive/Projects")
	}
}
