//go:build e2e

package e2e

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/auth"
	"github.com/aliwatters/gsuite-mcp/internal/calendar"
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/aliwatters/gsuite-mcp/internal/drive"
	"github.com/aliwatters/gsuite-mcp/internal/gmail"
	"github.com/aliwatters/gsuite-mcp/internal/tasks"
	"github.com/mark3labs/mcp-go/mcp"
)

// testAccount is the email address loaded from E2E_TEST_ACCOUNT.
var testAccount string

// authManager is the real auth.Manager used for all E2E tests.
var authManager *auth.Manager

func init() {
	loadDotEnv()
	testAccount = os.Getenv("E2E_TEST_ACCOUNT")
}

// loadDotEnv loads .env from the project root (best-effort).
func loadDotEnv() {
	f, err := os.Open("../.env")
	if err != nil {
		// Also try current directory
		f, err = os.Open(".env")
		if err != nil {
			return
		}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip surrounding quotes
		val = strings.Trim(val, `"'`)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

// skipIfNoAccount skips the test if no E2E test account is configured.
func skipIfNoAccount(t *testing.T) {
	t.Helper()
	if testAccount == "" {
		t.Skip("E2E_TEST_ACCOUNT not set; skipping E2E test")
	}
}

// setupAuth initializes the auth manager and common.Deps for production use.
// Must be called once before any E2E test that needs Google API access.
func setupAuth(t *testing.T) {
	t.Helper()
	skipIfNoAccount(t)

	if authManager != nil {
		return
	}

	mgr, err := auth.NewManager()
	if err != nil {
		t.Fatalf("failed to create auth manager: %v", err)
	}
	authManager = mgr

	common.SetDeps(&common.Deps{
		AuthManager: authManager,
	})
}

// e2eServiceFactory creates real Google API services using the auth manager.
type e2eServiceFactory[S any] struct {
	constructor common.ServiceConstructor[S]
}

func (f *e2eServiceFactory[S]) CreateService(ctx context.Context, email string) (S, error) {
	client, err := authManager.GetClientOrAuthenticate(ctx, email, false)
	if err != nil {
		var zero S
		return zero, err
	}
	return f.constructor(ctx, client)
}

// newE2EDeps creates HandlerDeps that use real auth for E2E tests.
func newE2EDeps[S any](constructor common.ServiceConstructor[S]) *common.HandlerDeps[S] {
	return &common.HandlerDeps[S]{
		EmailResolver: func(request mcp.CallToolRequest) (string, error) {
			// Always resolve to the test account
			return testAccount, nil
		},
		ServiceFactory: &e2eServiceFactory[S]{constructor: constructor},
	}
}

// gmailDeps returns HandlerDeps for Gmail E2E tests.
func gmailDeps() *gmail.GmailHandlerDeps {
	return newE2EDeps(gmail.NewGmailService)
}

// calendarDeps returns HandlerDeps for Calendar E2E tests.
func calendarDeps() *calendar.CalendarHandlerDeps {
	return newE2EDeps(calendar.NewCalendarService)
}

// driveDeps returns HandlerDeps for Drive E2E tests.
func driveDeps() *drive.DriveHandlerDeps {
	return newE2EDeps(drive.NewDriveService)
}

// tasksDeps returns HandlerDeps for Tasks E2E tests.
func tasksDeps() *tasks.TasksHandlerDeps {
	return newE2EDeps(tasks.NewTasksService)
}

// makeRequest creates an MCP CallToolRequest with the given arguments.
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return common.CreateMCPRequest(args)
}

// callResult holds parsed MCP tool result data.
type callResult struct {
	IsError bool
	Text    string
	Data    map[string]any
}

// call invokes a testable handler and parses the result.
func call(t *testing.T, result *mcp.CallToolResult, err error) callResult {
	t.Helper()
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if result == nil {
		t.Fatal("handler returned nil result")
	}

	cr := callResult{IsError: result.IsError}

	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(mcp.TextContent); ok {
			cr.Text = tc.Text
			// Try to parse as JSON
			var data map[string]any
			if json.Unmarshal([]byte(tc.Text), &data) == nil {
				cr.Data = data
			}
		}
	}

	return cr
}

// requireSuccess calls a handler and asserts no error.
func requireSuccess(t *testing.T, result *mcp.CallToolResult, err error) callResult {
	t.Helper()
	cr := call(t, result, err)
	if cr.IsError {
		t.Fatalf("expected success, got error: %s", cr.Text)
	}
	return cr
}

// requireStringField extracts a string field from the result data.
func requireStringField(t *testing.T, cr callResult, key string) string {
	t.Helper()
	if cr.Data == nil {
		t.Fatalf("result has no data; wanted field %q", key)
	}
	val, ok := cr.Data[key].(string)
	if !ok {
		t.Fatalf("field %q not a string (got %T): %v", key, cr.Data[key], cr.Data[key])
	}
	return val
}

// requireArrayField extracts an array field from the result data.
func requireArrayField(t *testing.T, cr callResult, key string) []any {
	t.Helper()
	if cr.Data == nil {
		t.Fatalf("result has no data; wanted field %q", key)
	}
	val, ok := cr.Data[key].([]any)
	if !ok {
		t.Fatalf("field %q not an array (got %T): %v", key, cr.Data[key], cr.Data[key])
	}
	return val
}

// e2ePrefix returns a unique prefix for test artifacts to aid cleanup.
func e2ePrefix() string {
	return fmt.Sprintf("[e2e-test-%d]", os.Getpid())
}
