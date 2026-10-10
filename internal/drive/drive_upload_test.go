package drive

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func resultText(t *testing.T, r *mcp.CallToolResult) string {
	t.Helper()
	return r.Content[0].(mcp.TextContent).Text
}

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "episode.mp3")
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDriveUploadFilePathStreamsAndReportsSizeMD5(t *testing.T) {
	data := []byte(strings.Repeat("abc123", 1000))
	sum := md5.Sum(data)
	wantMD5 := hex.EncodeToString(sum[:])
	path := writeTemp(t, data)

	f := NewDriveTestFixtures()
	var got []byte
	var gotParents []string
	f.MockService.CreateFileFunc = func(_ context.Context, file *drive.File, content io.Reader) (*drive.File, error) {
		got, _ = io.ReadAll(content)
		gotParents = file.Parents
		s := md5.Sum(got)
		return &drive.File{Id: "id1", Name: file.Name, Size: int64(len(got)), Md5Checksum: hex.EncodeToString(s[:])}, nil
	}

	req := common.CreateMCPRequest(map[string]any{"name": "ep.mp3", "file_path": path, "parent_id": "folderA"})
	res, err := TestableDriveUpload(context.Background(), req, f.Deps)
	if err != nil || res.IsError {
		t.Fatalf("unexpected failure: %v %v", err, res)
	}
	if string(got) != string(data) {
		t.Error("uploaded body differs from local file")
	}
	if len(gotParents) != 1 || gotParents[0] != "folderA" {
		t.Errorf("parents = %v", gotParents)
	}
	text := resultText(t, res)
	for _, want := range []string{`"id":"id1"`, `"size":6000`, wantMD5, "md5Checksum", `"name":"ep.mp3"`} {
		if !strings.Contains(text, want) {
			t.Errorf("result missing %q: %s", want, text)
		}
	}
}

func TestDriveUploadInlineContentStillWorks(t *testing.T) {
	f := NewDriveTestFixtures()
	req := common.CreateMCPRequest(map[string]any{"name": "a.txt", "content": "hello"})
	res, _ := TestableDriveUpload(context.Background(), req, f.Deps)
	if res.IsError || !strings.Contains(resultText(t, res), `"size":5`) {
		t.Errorf("unexpected: %s", resultText(t, res))
	}
}

func TestDriveUploadSourceValidation(t *testing.T) {
	dir := t.TempDir()
	real := writeTemp(t, []byte("x"))
	link := filepath.Join(dir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("symlinks unsupported")
	}

	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"neither", map[string]any{"name": "n"}, "exactly one of content or file_path"},
		{"both", map[string]any{"name": "n", "content": "c", "file_path": real}, "not both"},
		{"missing", map[string]any{"name": "n", "file_path": filepath.Join(dir, "nope")}, "no such file"},
		{"directory", map[string]any{"name": "n", "file_path": dir}, "not a regular file"},
		{"symlink", map[string]any{"name": "n", "file_path": link}, "symlink"},
		{"relative", map[string]any{"name": "n", "file_path": "rel/x.mp3"}, "absolute"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewDriveTestFixtures()
			called := false
			f.MockService.CreateFileFunc = func(context.Context, *drive.File, io.Reader) (*drive.File, error) {
				called = true
				return &drive.File{}, nil
			}
			res, _ := TestableDriveUpload(context.Background(), common.CreateMCPRequest(tt.args), f.Deps)
			if !res.IsError || !strings.Contains(resultText(t, res), tt.want) {
				t.Errorf("want error containing %q, got %v %s", tt.want, res.IsError, resultText(t, res))
			}
			if called {
				t.Error("CreateFile must not be called on invalid source")
			}
		})
	}
}

func rateLimitErr(code int, reason string) error {
	return &googleapi.Error{Code: code, Message: "quota hit", Errors: []googleapi.ErrorItem{{Reason: reason, Message: "quota hit"}}}
}

func TestDriveUploadSurfacesRateLimitWithoutRetry(t *testing.T) {
	for _, tc := range []struct {
		code   int
		reason string
	}{{403, "rateLimitExceeded"}, {429, "userRateLimitExceeded"}} {
		f := NewDriveTestFixtures()
		calls := 0
		f.MockService.CreateFileFunc = func(context.Context, *drive.File, io.Reader) (*drive.File, error) {
			calls++
			return nil, rateLimitErr(tc.code, tc.reason)
		}
		path := writeTemp(t, []byte("data"))
		res, _ := TestableDriveUpload(context.Background(), common.CreateMCPRequest(map[string]any{"name": "ep.mp3", "file_path": path}), f.Deps)
		text := resultText(t, res)
		if !res.IsError || !strings.Contains(text, tc.reason) || !strings.Contains(text, "ep.mp3") {
			t.Errorf("%d: error must include reason and file name: %s", tc.code, text)
		}
		if calls != 1 {
			t.Errorf("%d: expected exactly 1 call (no retry), got %d", tc.code, calls)
		}
	}
}

func TestDriveUpdateFileRenameOnly(t *testing.T) {
	f := NewDriveTestFixtures()
	contentCalled := false
	f.MockService.UpdateFileContentFunc = func(context.Context, string, *drive.File, io.Reader) (*drive.File, error) {
		contentCalled = true
		return &drive.File{}, nil
	}
	var gotName string
	f.MockService.UpdateFileFunc = func(_ context.Context, id string, file *drive.File) (*drive.File, error) {
		gotName = file.Name
		return &drive.File{Id: id, Name: file.Name, Size: 7, Md5Checksum: "abc"}, nil
	}
	res, _ := TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "name": "old-ep.mp3"}), f.Deps)
	if res.IsError || gotName != "old-ep.mp3" || contentCalled {
		t.Fatalf("rename failed: %v %s content=%v", gotName, resultText(t, res), contentCalled)
	}
}

func TestDriveUpdateFileReplaceBodyAndRename(t *testing.T) {
	path := writeTemp(t, []byte("new body"))
	f := NewDriveTestFixtures()
	var gotBody []byte
	var gotName string
	f.MockService.UpdateFileContentFunc = func(_ context.Context, id string, file *drive.File, c io.Reader) (*drive.File, error) {
		gotBody, _ = io.ReadAll(c)
		gotName = file.Name
		return &drive.File{Id: id, Name: "n.mp3", Size: int64(len(gotBody)), Md5Checksum: "m5"}, nil
	}
	res, _ := TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "file_path": path, "name": "n.mp3"}), f.Deps)
	text := resultText(t, res)
	if res.IsError || string(gotBody) != "new body" || gotName != "n.mp3" || !strings.Contains(text, "m5") {
		t.Fatalf("unexpected: %s body=%q name=%q", text, gotBody, gotName)
	}
}

func TestDriveUpdateFileErrors(t *testing.T) {
	f := NewDriveTestFixtures()
	res, _ := TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1"}), f.Deps)
	if !res.IsError || !strings.Contains(resultText(t, res), "nothing to update") {
		t.Errorf("expected nothing-to-update error: %s", resultText(t, res))
	}

	res, _ = TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "content": "c", "file_path": "/x"}), f.Deps)
	if !res.IsError || !strings.Contains(resultText(t, res), "not both") {
		t.Errorf("expected both-sources error: %s", resultText(t, res))
	}

	calls := 0
	f.MockService.UpdateFileFunc = func(context.Context, string, *drive.File) (*drive.File, error) {
		calls++
		return nil, rateLimitErr(403, "rateLimitExceeded")
	}
	res, _ = TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "name": "x.mp3"}), f.Deps)
	text := resultText(t, res)
	if !res.IsError || !strings.Contains(text, "rateLimitExceeded") || !strings.Contains(text, "x.mp3") || calls != 1 {
		t.Errorf("403 handling wrong (calls=%d): %s", calls, text)
	}

	res, _ = TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "file_path": "/definitely/not/here"}), f.Deps)
	if !res.IsError {
		t.Error("expected error for missing file_path")
	}
}

func TestDriveUpdateFileBodyErrorUsesExistingName(t *testing.T) {
	f := NewDriveTestFixtures()
	f.MockService.UpdateFileContentFunc = func(context.Context, string, *drive.File, io.Reader) (*drive.File, error) {
		return nil, rateLimitErr(429, "userRateLimitExceeded")
	}
	f.MockService.GetFileFunc = func(context.Context, string, string) (*drive.File, error) {
		return &drive.File{Name: "existing.mp3"}, nil
	}

	res, _ := TestableDriveUpdateFile(context.Background(), common.CreateMCPRequest(map[string]any{"file_id": "f1", "content": "new body"}), f.Deps)
	text := resultText(t, res)
	if !res.IsError || !strings.Contains(text, "existing.mp3") || !strings.Contains(text, "userRateLimitExceeded") {
		t.Fatalf("body update error must include existing name and reason: %s", text)
	}
}

func TestDriveListAndSearchExactNameFilter(t *testing.T) {
	f := NewDriveTestFixtures()
	var q string
	f.MockService.ListFilesFunc = func(_ context.Context, o *ListFilesOptions) (*drive.FileList, error) {
		q = o.Query
		return &drive.FileList{Files: []*drive.File{{Id: "1", Name: "it's.mp3", Size: 9, Md5Checksum: "deadbeef"}}}, nil
	}

	res, _ := TestableDriveList(context.Background(), common.CreateMCPRequest(map[string]any{"folder_id": "fold", "name": "it's.mp3"}), f.Deps)
	if res.IsError || !strings.Contains(q, `'fold' in parents`) || !strings.Contains(q, `name = 'it\'s.mp3'`) {
		t.Errorf("list query = %q", q)
	}
	text := resultText(t, res)
	if !strings.Contains(text, "deadbeef") || !strings.Contains(text, `"size":9`) {
		t.Errorf("list result lacks size/md5: %s", text)
	}

	res, _ = TestableDriveSearch(context.Background(), common.CreateMCPRequest(map[string]any{"name": "a.mp3", "parent_id": "fold"}), f.Deps)
	if res.IsError || !strings.Contains(q, "name = 'a.mp3'") || !strings.Contains(q, "'fold' in parents") {
		t.Errorf("search query = %q", q)
	}
	if !strings.Contains(resultText(t, res), "deadbeef") {
		t.Error("search result lacks md5Checksum")
	}

	res, _ = TestableDriveSearch(context.Background(), common.CreateMCPRequest(map[string]any{}), f.Deps)
	if !res.IsError {
		t.Error("search with no criteria must error")
	}
}
