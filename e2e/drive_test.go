//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/drive"
)

// TestDriveUploadSearchTrash tests uploading a file, searching for it, and trashing it.
func TestDriveUploadSearchTrash(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := driveDeps()

	prefix := e2ePrefix()
	fileName := fmt.Sprintf("%s-test-%d.txt", prefix, time.Now().UnixMilli())
	content := "This is an E2E test file. Safe to delete."

	// Upload a text file
	t.Logf("uploading file %q", fileName)
	result, err := drive.TestableDriveUpload(ctx, makeRequest(map[string]any{
		"name":    fileName,
		"content": content,
	}), deps)
	uploadResult := requireSuccess(t, result, err)

	fileID := requireStringField(t, uploadResult, "id")
	t.Logf("uploaded file id=%s", fileID)

	defer func() {
		t.Log("cleanup: permanently deleting test file")
		_, _ = drive.TestableDriveDelete(ctx, makeRequest(map[string]any{
			"file_id": fileID,
		}), deps)
	}()

	// Search for the file
	t.Log("searching for uploaded file")
	result, err = drive.TestableDriveSearch(ctx, makeRequest(map[string]any{
		"query": fmt.Sprintf("name = '%s'", fileName),
	}), deps)
	searchResult := requireSuccess(t, result, err)

	files := requireArrayField(t, searchResult, "files")
	found := false
	for _, f := range files {
		file, ok := f.(map[string]any)
		if !ok {
			continue
		}
		if file["id"] == fileID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("uploaded file %s not found in search results", fileID)
	}

	// Get file metadata
	t.Log("getting file metadata")
	result, err = drive.TestableDriveGet(ctx, makeRequest(map[string]any{
		"file_id": fileID,
	}), deps)
	getResult := requireSuccess(t, result, err)

	gotName := requireStringField(t, getResult, "name")
	if gotName != fileName {
		t.Errorf("expected name %q, got %q", fileName, gotName)
	}

	// Download file content
	t.Log("downloading file content")
	result, err = drive.TestableDriveDownload(ctx, makeRequest(map[string]any{
		"file_id": fileID,
	}), deps)
	downloadResult := requireSuccess(t, result, err)

	downloadedContent := requireStringField(t, downloadResult, "content")
	if downloadedContent != content {
		t.Errorf("expected content %q, got %q", content, downloadedContent)
	}

	// Trash the file
	t.Log("trashing file")
	result, err = drive.TestableDriveTrash(ctx, makeRequest(map[string]any{
		"file_id": fileID,
	}), deps)
	requireSuccess(t, result, err)
}

// TestDriveCreateFolder tests creating and deleting a folder.
func TestDriveCreateFolder(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := driveDeps()

	prefix := e2ePrefix()
	folderName := fmt.Sprintf("%s-folder-%d", prefix, time.Now().UnixMilli())

	t.Logf("creating folder %q", folderName)
	result, err := drive.TestableDriveCreateFolder(ctx, makeRequest(map[string]any{
		"name": folderName,
	}), deps)
	createResult := requireSuccess(t, result, err)

	folderID := requireStringField(t, createResult, "id")
	t.Logf("created folder id=%s", folderID)

	defer func() {
		t.Log("cleanup: permanently deleting test folder")
		_, _ = drive.TestableDriveDelete(ctx, makeRequest(map[string]any{
			"file_id": folderID,
		}), deps)
	}()

	// Verify folder exists
	t.Log("getting folder metadata")
	result, err = drive.TestableDriveGet(ctx, makeRequest(map[string]any{
		"file_id": folderID,
	}), deps)
	getResult := requireSuccess(t, result, err)

	gotName := requireStringField(t, getResult, "name")
	if gotName != folderName {
		t.Errorf("expected folder name %q, got %q", folderName, gotName)
	}

	// Delete folder
	t.Log("deleting folder")
	result, err = drive.TestableDriveDelete(ctx, makeRequest(map[string]any{
		"file_id": folderID,
	}), deps)
	requireSuccess(t, result, err)
}
