package drive

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// TestDriveServiceCreateFile tests the CreateFile method of the DriveService interface.
func TestDriveServiceCreateFile(t *testing.T) {
	tests := []struct {
		name      string
		file      *drive.File
		content   string
		wantID    string
		wantName  string
		wantSize  int64
		wantErr   bool
		setupMock func(mock *MockDriveService)
	}{
		{
			name: "create file with content",
			file: &drive.File{
				Name:     "test.txt",
				MimeType: "text/plain",
			},
			content:  "Hello, World!",
			wantID:   "newfile001",
			wantName: "test.txt",
			wantSize: 13,
		},
		{
			name: "create file in folder",
			file: &drive.File{
				Name:    "document.txt",
				Parents: []string{"folder123"},
			},
			content:  "File content",
			wantID:   "newfile001",
			wantName: "document.txt",
			wantSize: 12,
		},
		{
			name: "create empty file",
			file: &drive.File{
				Name: "empty.txt",
			},
			content:  "",
			wantID:   "newfile001",
			wantName: "empty.txt",
			wantSize: 0,
		},
		{
			name: "create file with description",
			file: &drive.File{
				Name:        "described.txt",
				Description: "A test file with description",
			},
			content:  "Content",
			wantID:   "newfile001",
			wantName: "described.txt",
			wantSize: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			var contentReader io.Reader
			if tt.content != "" {
				contentReader = strings.NewReader(tt.content)
			}

			result, err := fixtures.MockService.CreateFile(context.Background(), tt.file, contentReader)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Id != tt.wantID {
				t.Errorf("got id %q, want %q", result.Id, tt.wantID)
			}

			if result.Name != tt.wantName {
				t.Errorf("got name %q, want %q", result.Name, tt.wantName)
			}

			if result.Size != tt.wantSize {
				t.Errorf("got size %d, want %d", result.Size, tt.wantSize)
			}
		})
	}
}

// TestDriveServiceListFiles tests the ListFiles method of the DriveService interface.
func TestDriveServiceListFiles(t *testing.T) {
	tests := []struct {
		name      string
		opts      *ListFilesOptions
		wantCount int
		wantErr   bool
		setupMock func(mock *MockDriveService)
	}{
		{
			name:      "list all files",
			opts:      nil,
			wantCount: 3,
		},
		{
			name: "list with page size",
			opts: &ListFilesOptions{
				PageSize: 10,
			},
			wantCount: 3,
		},
		{
			name: "list with query",
			opts: &ListFilesOptions{
				Query: "name contains 'Document'",
			},
			wantCount: 3, // Mock returns same data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			result, err := fixtures.MockService.ListFiles(context.Background(), tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Files) != tt.wantCount {
				t.Errorf("got %d files, want %d", len(result.Files), tt.wantCount)
			}
		})
	}
}

// TestDriveServiceGetFile tests the GetFile method of the DriveService interface.
func TestDriveServiceGetFile(t *testing.T) {
	tests := []struct {
		name     string
		fileID   string
		wantName string
		wantErr  bool
	}{
		{
			name:     "get document",
			fileID:   "file001",
			wantName: "Document.docx",
		},
		{
			name:     "get spreadsheet",
			fileID:   "file002",
			wantName: "Spreadsheet.xlsx",
		},
		{
			name:     "get folder",
			fileID:   "folder001",
			wantName: "My Folder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			result, err := fixtures.MockService.GetFile(context.Background(), tt.fileID, "")

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Name != tt.wantName {
				t.Errorf("got name %q, want %q", result.Name, tt.wantName)
			}
		})
	}
}

// TestDriveServiceCopyFile tests the CopyFile method of the DriveService interface.
func TestDriveServiceCopyFile(t *testing.T) {
	tests := []struct {
		name    string
		fileID  string
		newFile *drive.File
		wantID  string
		wantErr bool
	}{
		{
			name:   "copy file",
			fileID: "file001",
			newFile: &drive.File{
				Name: "Document Copy.docx",
			},
			wantID: "copy_file001",
		},
		{
			name:   "copy file to folder",
			fileID: "file002",
			newFile: &drive.File{
				Name:    "Spreadsheet Copy.xlsx",
				Parents: []string{"folder001"},
			},
			wantID: "copy_file002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			result, err := fixtures.MockService.CopyFile(context.Background(), tt.fileID, tt.newFile)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Id != tt.wantID {
				t.Errorf("got id %q, want %q", result.Id, tt.wantID)
			}
		})
	}
}

// TestDriveServiceDeleteFile tests the DeleteFile method of the DriveService interface.
func TestDriveServiceDeleteFile(t *testing.T) {
	tests := []struct {
		name    string
		fileID  string
		wantErr bool
	}{
		{
			name:   "delete file",
			fileID: "file001",
		},
		{
			name:   "delete folder",
			fileID: "folder001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			err := fixtures.MockService.DeleteFile(context.Background(), tt.fileID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestDriveServiceDownloadFile tests the DownloadFile method of the DriveService interface.
func TestDriveServiceDownloadFile(t *testing.T) {
	tests := []struct {
		name        string
		fileID      string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "download file",
			fileID:      "file001",
			wantContent: "Test file content for file001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			result, err := fixtures.MockService.DownloadFile(context.Background(), tt.fileID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			defer result.Close()
			content, _ := io.ReadAll(result)

			if string(content) != tt.wantContent {
				t.Errorf("got content %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

// TestDriveServiceListPermissions tests the ListPermissions method of the DriveService interface.
func TestDriveServiceListPermissions(t *testing.T) {
	tests := []struct {
		name      string
		fileID    string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list permissions",
			fileID:    "file001",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			result, err := fixtures.MockService.ListPermissions(context.Background(), tt.fileID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Permissions) != tt.wantCount {
				t.Errorf("got %d permissions, want %d", len(result.Permissions), tt.wantCount)
			}
		})
	}
}

// TestDriveServiceCreatePermission tests the CreatePermission method of the DriveService interface.
func TestDriveServiceCreatePermission(t *testing.T) {
	tests := []struct {
		name             string
		fileID           string
		permission       *drive.Permission
		sendNotification bool
		wantID           string
		wantErr          bool
	}{
		{
			name:   "share with user as writer",
			fileID: "file001",
			permission: &drive.Permission{
				Type:         "user",
				Role:         "writer",
				EmailAddress: "user@example.com",
			},
			sendNotification: true,
			wantID:           "newperm001",
		},
		{
			name:   "share with anyone as reader",
			fileID: "file001",
			permission: &drive.Permission{
				Type: "anyone",
				Role: "reader",
			},
			sendNotification: false,
			wantID:           "newperm001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()

			result, err := fixtures.MockService.CreatePermission(context.Background(), tt.fileID, tt.permission, tt.sendNotification)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Id != tt.wantID {
				t.Errorf("got id %q, want %q", result.Id, tt.wantID)
			}
		})
	}
}

// TestDriveSearchCorpora tests that TestableDriveSearch passes the corpora parameter correctly.
func TestDriveSearchCorpora(t *testing.T) {
	tests := []struct {
		name        string
		args        map[string]any
		wantCorpora string
	}{
		{
			name:        "default corpora is allDrives",
			args:        map[string]any{"query": "test"},
			wantCorpora: "allDrives",
		},
		{
			name:        "explicit corpora user",
			args:        map[string]any{"query": "test", "corpora": "user"},
			wantCorpora: "user",
		},
		{
			name:        "explicit corpora domain",
			args:        map[string]any{"query": "test", "corpora": "domain"},
			wantCorpora: "domain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()
			var capturedOpts *ListFilesOptions
			fixtures.MockService.ListFilesFunc = func(_ context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
				capturedOpts = opts
				return &drive.FileList{}, nil
			}

			request := common.CreateMCPRequest(tt.args)
			_, err := TestableDriveSearch(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if capturedOpts == nil {
				t.Fatal("ListFiles was not called")
			}
			if capturedOpts.Corpora != tt.wantCorpora {
				t.Errorf("got corpora %q, want %q", capturedOpts.Corpora, tt.wantCorpora)
			}
		})
	}
}

// TestDriveListCorpora tests that TestableDriveList passes corpora for shared drive support.
func TestDriveListCorpora(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	var capturedOpts *ListFilesOptions
	fixtures.MockService.ListFilesFunc = func(_ context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
		capturedOpts = opts
		return &drive.FileList{}, nil
	}

	request := common.CreateMCPRequest(map[string]any{})
	_, err := TestableDriveList(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedOpts == nil {
		t.Fatal("ListFiles was not called")
	}
	if capturedOpts.Corpora != "allDrives" {
		t.Errorf("got corpora %q, want %q", capturedOpts.Corpora, "allDrives")
	}
}

// TestExtractFileID tests the common.ExtractGoogleResourceID helper function (formerly extractFileID).
func TestExtractFileID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain id",
			input: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "file view url",
			input: "https://drive.google.com/file/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/view",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "file open url",
			input: "https://drive.google.com/open?id=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "docs url",
			input: "https://docs.google.com/document/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:  "sheets url",
			input: "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit",
			want:  "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ExtractGoogleResourceID(tt.input)
			if got != tt.want {
				t.Errorf("ExtractGoogleResourceID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestFormatFile tests the formatFile helper function.
func TestFormatFile(t *testing.T) {
	file := &drive.File{
		Id:          "test123",
		Name:        "Test File.txt",
		MimeType:    "text/plain",
		Size:        1024,
		CreatedTime: "2024-02-01T12:00:00Z",
	}

	result := formatFile(file)

	if result["id"] != "test123" {
		t.Errorf("got id %v, want test123", result["id"])
	}
	if result["name"] != "Test File.txt" {
		t.Errorf("got name %v, want 'Test File.txt'", result["name"])
	}
	if result["mime_type"] != "text/plain" {
		t.Errorf("got mime_type %v, want text/plain", result["mime_type"])
	}
	if result["size"] != int64(1024) {
		t.Errorf("got size %v, want 1024", result["size"])
	}
}

// TestFormatFileFull tests the formatFileFull helper function.
func TestFormatFileFull(t *testing.T) {
	file := &drive.File{
		Id:             "test123",
		Name:           "Test File.txt",
		MimeType:       "text/plain",
		Size:           1024,
		CreatedTime:    "2024-02-01T12:00:00Z",
		ModifiedTime:   "2024-02-01T14:00:00Z",
		WebViewLink:    "https://drive.google.com/file/d/test123/view",
		WebContentLink: "https://drive.google.com/uc?id=test123",
		Description:    "A test file",
		Parents:        []string{"folder001"},
		Starred:        true,
		Trashed:        false,
	}

	result := formatFileFull(file)

	if result["id"] != "test123" {
		t.Errorf("got id %v, want test123", result["id"])
	}
	if result["description"] != "A test file" {
		t.Errorf("got description %v, want 'A test file'", result["description"])
	}
	// The function stores web_view_link as "url" in formatFile
	if result["url"] != "https://drive.google.com/file/d/test123/view" {
		t.Errorf("got url %v, want 'https://drive.google.com/file/d/test123/view'", result["url"])
	}
	// download_url comes from WebContentLink
	if result["download_url"] != "https://drive.google.com/uc?id=test123" {
		t.Errorf("got download_url %v, want 'https://drive.google.com/uc?id=test123'", result["download_url"])
	}
	// starred is only included if true
	if result["starred"] != true {
		t.Errorf("got starred %v, want true", result["starred"])
	}

	parents := result["parents"].([]string)
	if len(parents) != 1 || parents[0] != "folder001" {
		t.Errorf("got parents %v, want [folder001]", parents)
	}
}

// TestIsTextMimeType tests the isTextMimeType helper function.
func TestIsTextMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{"text/plain", true},
		{"text/html", true},
		{"text/csv", true},
		{"application/json", true},
		{"application/xml", true},
		{"application/javascript", true},
		{"application/pdf", false},
		{"image/png", false},
		{"application/octet-stream", false},
		{"application/vnd.google-apps.document", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := isTextMimeType(tt.mimeType)
			if got != tt.want {
				t.Errorf("isTextMimeType(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

// TestDriveSearchFileType tests the file_type filter in TestableDriveSearch.
func TestDriveSearchFileType(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		wantQuery string
	}{
		{
			name:      "friendly name doc",
			args:      map[string]any{"query": "test", "file_type": "doc"},
			wantQuery: "(test) and mimeType = 'application/vnd.google-apps.document'",
		},
		{
			name:      "friendly name sheet",
			args:      map[string]any{"query": "test", "file_type": "sheet"},
			wantQuery: "(test) and mimeType = 'application/vnd.google-apps.spreadsheet'",
		},
		{
			name:      "friendly name slides",
			args:      map[string]any{"query": "test", "file_type": "slides"},
			wantQuery: "(test) and mimeType = 'application/vnd.google-apps.presentation'",
		},
		{
			name:      "friendly name pdf",
			args:      map[string]any{"query": "test", "file_type": "pdf"},
			wantQuery: "(test) and mimeType = 'application/pdf'",
		},
		{
			name:      "friendly name folder",
			args:      map[string]any{"query": "test", "file_type": "folder"},
			wantQuery: "(test) and mimeType = 'application/vnd.google-apps.folder'",
		},
		{
			name:      "friendly name image uses contains",
			args:      map[string]any{"query": "test", "file_type": "image"},
			wantQuery: "(test) and mimeType contains 'image/'",
		},
		{
			name:      "raw mimeType passthrough",
			args:      map[string]any{"query": "test", "file_type": "application/zip"},
			wantQuery: "(test) and mimeType = 'application/zip'",
		},
		{
			name:      "no file_type leaves query unchanged",
			args:      map[string]any{"query": "test"},
			wantQuery: "test",
		},
		{
			name:      "case insensitive friendly name",
			args:      map[string]any{"query": "test", "file_type": "PDF"},
			wantQuery: "(test) and mimeType = 'application/pdf'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewDriveTestFixtures()
			var capturedOpts *ListFilesOptions
			fixtures.MockService.ListFilesFunc = func(_ context.Context, opts *ListFilesOptions) (*drive.FileList, error) {
				capturedOpts = opts
				return &drive.FileList{}, nil
			}

			request := common.CreateMCPRequest(tt.args)
			_, err := TestableDriveSearch(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if capturedOpts == nil {
				t.Fatal("ListFiles was not called")
			}
			if capturedOpts.Query != tt.wantQuery {
				t.Errorf("got query %q, want %q", capturedOpts.Query, tt.wantQuery)
			}
		})
	}
}

// TestDriveGetShareableLink tests TestableDriveGetShareableLink.
func TestDriveGetShareableLink(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.GetFileFunc = func(_ context.Context, fileID string, fields string) (*drive.File, error) {
		return &drive.File{
			Id:          fileID,
			Name:        "Test.docx",
			MimeType:    "application/vnd.google-apps.document",
			WebViewLink: "https://docs.google.com/document/d/" + fileID + "/edit",
			Permissions: []*drive.Permission{
				{Id: "perm1", Type: "user", Role: "owner", EmailAddress: "owner@example.com"},
				{Id: "perm2", Type: "anyone", Role: "reader"},
			},
		}, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123"})
	result, err := TestableDriveGetShareableLink(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "file123") {
		t.Errorf("result should contain file_id, got: %s", text)
	}
	if !strings.Contains(text, "https://docs.google.com/document/d/file123/edit") {
		t.Errorf("result should contain URL, got: %s", text)
	}
	if !strings.Contains(text, "owner@example.com") {
		t.Errorf("result should contain permissions, got: %s", text)
	}
	if !strings.Contains(text, `"permission_count":2`) {
		t.Errorf("result should contain permission_count of 2, got: %s", text)
	}
}

// TestDriveListComments tests TestableDriveListComments.
func TestDriveListComments(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.ListCommentsFunc = func(_ context.Context, fileID string, _ string, _ int64, _ string, _ bool) (*drive.CommentList, error) {
		return &drive.CommentList{
			Comments: []*drive.Comment{
				{
					Id:          "comment1",
					Content:     "This is a comment",
					CreatedTime: "2024-01-01T00:00:00Z",
					Author:      &drive.User{DisplayName: "Alice", EmailAddress: "alice@example.com"},
				},
				{
					Id:          "comment2",
					Content:     "Another comment",
					Resolved:    true,
					CreatedTime: "2024-01-02T00:00:00Z",
					Author:      &drive.User{DisplayName: "Bob", EmailAddress: "bob@example.com"},
				},
			},
		}, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123"})
	result, err := TestableDriveListComments(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "comment1") {
		t.Errorf("result should contain comment1, got: %s", text)
	}
	if !strings.Contains(text, `"count":2`) {
		t.Errorf("result should contain count of 2, got: %s", text)
	}
}

// TestDriveCreateComment tests TestableDriveCreateComment.
func TestDriveCreateComment(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.CreateCommentFunc = func(_ context.Context, fileID string, comment *drive.Comment, _ string) (*drive.Comment, error) {
		comment.Id = "newcomment1"
		comment.CreatedTime = "2024-01-01T00:00:00Z"
		comment.Author = &drive.User{DisplayName: "Test User", EmailAddress: "test@example.com"}
		return comment, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123", "content": "Great work!"})
	result, err := TestableDriveCreateComment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "newcomment1") {
		t.Errorf("result should contain new comment ID, got: %s", text)
	}
	if !strings.Contains(text, "Great work!") {
		t.Errorf("result should contain content, got: %s", text)
	}
}

// TestDriveDeleteComment tests TestableDriveDeleteComment.
func TestDriveDeleteComment(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	var deletedFileID, deletedCommentID string
	fixtures.MockService.DeleteCommentFunc = func(_ context.Context, fileID string, commentID string) error {
		deletedFileID = fileID
		deletedCommentID = commentID
		return nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123", "comment_id": "comment1"})
	result, err := TestableDriveDeleteComment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"success":true`) {
		t.Errorf("result should contain success, got: %s", text)
	}
	if deletedFileID != "file123" {
		t.Errorf("got file_id %q, want %q", deletedFileID, "file123")
	}
	if deletedCommentID != "comment1" {
		t.Errorf("got comment_id %q, want %q", deletedCommentID, "comment1")
	}
}

// TestDriveCreateReply tests TestableDriveCreateReply.
func TestDriveCreateReply(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.CreateReplyFunc = func(_ context.Context, fileID string, commentID string, reply *drive.Reply, _ string) (*drive.Reply, error) {
		reply.Id = "reply1"
		reply.CreatedTime = "2024-01-01T00:00:00Z"
		reply.Author = &drive.User{DisplayName: "Test User", EmailAddress: "test@example.com"}
		return reply, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123", "comment_id": "comment1", "content": "I agree"})
	result, err := TestableDriveCreateReply(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "reply1") {
		t.Errorf("result should contain reply ID, got: %s", text)
	}
}

// TestDriveListRevisions tests TestableDriveListRevisions.
func TestDriveListRevisions(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.ListRevisionsFunc = func(_ context.Context, fileID string, _ string, _ int64, _ string) (*drive.RevisionList, error) {
		return &drive.RevisionList{
			Revisions: []*drive.Revision{
				{
					Id:           "rev1",
					ModifiedTime: "2024-01-01T00:00:00Z",
					Size:         1024,
					LastModifyingUser: &drive.User{
						DisplayName:  "Alice",
						EmailAddress: "alice@example.com",
					},
				},
				{
					Id:           "rev2",
					ModifiedTime: "2024-01-02T00:00:00Z",
					Size:         2048,
					LastModifyingUser: &drive.User{
						DisplayName:  "Bob",
						EmailAddress: "bob@example.com",
					},
				},
			},
		}, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123"})
	result, err := TestableDriveListRevisions(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "rev1") {
		t.Errorf("result should contain rev1, got: %s", text)
	}
	if !strings.Contains(text, `"count":2`) {
		t.Errorf("result should contain count of 2, got: %s", text)
	}
}

// TestDriveGetRevision tests TestableDriveGetRevision.
func TestDriveGetRevision(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.GetRevisionFunc = func(_ context.Context, fileID string, revisionID string, _ string) (*drive.Revision, error) {
		return &drive.Revision{
			Id:           revisionID,
			ModifiedTime: "2024-01-01T00:00:00Z",
			Size:         1024,
			MimeType:     "application/pdf",
			LastModifyingUser: &drive.User{
				DisplayName:  "Alice",
				EmailAddress: "alice@example.com",
			},
		}, nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123", "revision_id": "rev1"})
	result, err := TestableDriveGetRevision(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "rev1") {
		t.Errorf("result should contain revision ID, got: %s", text)
	}
	if !strings.Contains(text, "alice@example.com") {
		t.Errorf("result should contain author, got: %s", text)
	}
}

// TestDriveDownloadRevision tests TestableDriveDownloadRevision.
func TestDriveDownloadRevision(t *testing.T) {
	fixtures := NewDriveTestFixtures()
	fixtures.MockService.GetFileFunc = func(_ context.Context, fileID string, _ string) (*drive.File, error) {
		return &drive.File{
			Id:       fileID,
			Name:     "test.txt",
			MimeType: "text/plain",
			Size:     100,
		}, nil
	}
	fixtures.MockService.DownloadRevisionFunc = func(_ context.Context, fileID string, revisionID string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("revision content here")), nil
	}

	request := common.CreateMCPRequest(map[string]any{"file_id": "file123", "revision_id": "rev1"})
	result, err := TestableDriveDownloadRevision(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "revision content here") {
		t.Errorf("result should contain revision content, got: %s", text)
	}
	if !strings.Contains(text, "utf-8") {
		t.Errorf("result should indicate utf-8 encoding for text, got: %s", text)
	}
}

// TestCreateFileWithBytesReader verifies that CreateFile works correctly with bytes.NewReader.
// This tests the fix for review comment #2750034117.
func TestCreateFileWithBytesReader(t *testing.T) {
	fixtures := NewDriveTestFixtures()

	// Test that bytes.NewReader is properly handled
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	file := &drive.File{
		Name:     "binary.bin",
		MimeType: "application/octet-stream",
	}

	// Use bytes.NewReader like the fixed code does
	result, err := fixtures.MockService.CreateFile(context.Background(), file, bytes.NewReader(binaryData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file was created
	if result.Id == "" {
		t.Error("expected file to have an ID")
	}

	// The mock reads the content and sets Size
	if result.Size != int64(len(binaryData)) {
		t.Errorf("got size %d, want %d", result.Size, len(binaryData))
	}
}
