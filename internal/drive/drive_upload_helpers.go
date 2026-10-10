package drive

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/googleapi"
)

// uploadChunkSize is the resumable-upload chunk size. Files larger than one
// chunk are sent with Drive's resumable protocol, so memory use stays bounded.
const uploadChunkSize = 8 * 1024 * 1024

// DriveFileUploadResultFields are requested after an upload or content update
// so the caller can verify size and md5Checksum against the local file.
const DriveFileUploadResultFields = "id,name,mimeType,size,md5Checksum,createdTime,modifiedTime,parents,webViewLink"

// openRegularFile validates that path is an absolute path to a regular file
// (not a directory, symlink, device, or missing path) and opens it.
func openRegularFile(path string) (*os.File, int64, error) {
	if !filepath.IsAbs(path) {
		return nil, 0, fmt.Errorf("file_path %q must be an absolute path", path)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, 0, fmt.Errorf("file_path %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, 0, fmt.Errorf("file_path %q is a symlink; pass the real file path", path)
	}
	if !info.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("file_path %q is not a regular file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("file_path %q: %w", path, err)
	}
	return f, info.Size(), nil
}

// driveAPIErrorResult logs a Drive API failure with context and converts it to
// a tool error. For 403 and 429 responses Google's error reason is surfaced
// together with the file name. Nothing is retried.
func driveAPIErrorResult(op, fileName string, err error) *mcp.CallToolResult {
	log.Printf("drive: %s failed for file %q: %v", op, fileName, err)

	var gerr *googleapi.Error
	if errors.As(err, &gerr) && (gerr.Code == http.StatusForbidden || gerr.Code == http.StatusTooManyRequests) {
		reasons := make([]string, 0, len(gerr.Errors))
		for _, e := range gerr.Errors {
			if e.Reason != "" {
				reasons = append(reasons, e.Reason)
			}
		}
		reason := "unknown"
		if len(reasons) > 0 {
			reason = strings.Join(reasons, ",")
		}
		return mcp.NewToolResultError(fmt.Sprintf(
			"Drive API error during %s of file %q: HTTP %d, reason: %s, message: %s (not retried)",
			op, fileName, gerr.Code, reason, gerr.Message))
	}
	return mcp.NewToolResultError(fmt.Sprintf("Drive API error during %s of file %q: %v", op, fileName, err))
}

// escapeDriveQueryValue escapes a value for use inside a single-quoted Drive query string.
func escapeDriveQueryValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	return strings.ReplaceAll(v, "'", `\'`)
}
