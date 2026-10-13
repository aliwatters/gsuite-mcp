package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestExecuteCallWritesOnlyJSONToStdout(t *testing.T) {
	registry := server.NewMCPServer("test", "test")
	registry.AddTool(mcp.NewTool("example"), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if req.GetArguments()["account"] != "user@example.com" {
			t.Fatalf("account argument = %v", req.GetArguments()["account"])
		}
		return mcp.NewToolResultText(`{"ok":true}`), nil
	})
	var stdout, stderr bytes.Buffer
	code := executeCall(registry, "example", "user@example.com", map[string]any{"account": "user@example.com"}, &stdout, &stderr)
	if code != 0 || stdout.String() != "{\"ok\":true}\n" || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCallReturnsToolError(t *testing.T) {
	registry := server.NewMCPServer("test", "test")
	registry.AddTool(mcp.NewTool("example"), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("bad arguments"), nil
	})
	var stdout, stderr bytes.Buffer
	code := executeCall(registry, "example", "user@example.com", nil, &stdout, &stderr)
	if code != callExitError || stdout.Len() != 0 || !strings.Contains(stderr.String(), "example account=user@example.com: bad arguments") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestExecuteCallClassifiesAuthenticationRequired(t *testing.T) {
	registry := server.NewMCPServer("test", "test")
	registry.AddTool(mcp.NewTool("example"), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("auth_expired: run 'gsuite-mcp auth user@example.com' to re-authenticate"), nil
	})
	var stdout, stderr bytes.Buffer
	code := executeCall(registry, "example", "user@example.com", nil, &stdout, &stderr)
	if code != callExitAuthRequired || stdout.Len() != 0 || !strings.Contains(stderr.String(), "gsuite-mcp auth user@example.com") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunCallArgumentErrorsAndArgsFile(t *testing.T) {
	t.Run("invalid JSON", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := runCall([]string{"example", "--account", "user@example.com", "--args", "{"}, &stdout, &stderr)
		if code != callExitError || stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid JSON arguments") {
			t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
		}
	})
	t.Run("args file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "args.json")
		if err := os.WriteFile(path, []byte(`{"folder_id":"root"}`), 0600); err != nil {
			t.Fatal(err)
		}
		// File decoding succeeds before app initialization; use a missing tool to stop before credentials are needed.
		var stdout, stderr bytes.Buffer
		code := runCall([]string{"missing", "--account", "user@example.com", "--args-file", path}, &stdout, &stderr)
		if code != callExitError || !strings.Contains(stderr.String(), "unknown tool") {
			t.Fatalf("code=%d stderr=%q", code, stderr.String())
		}
	})
}

func TestRunCallMissingCredentialsDoesNotStartAuthServer(t *testing.T) {
	previous := config.DefaultConfigDir()
	configDir := t.TempDir()
	config.SetConfigDir(configDir)
	t.Cleanup(func() { config.SetConfigDir(previous) })
	if err := os.WriteFile(filepath.Join(configDir, "client_secret.json"), []byte(`{"installed":{"client_id":"test-client","client_secret":"test-secret","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","redirect_uris":["http://localhost"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runCall([]string{"drive_get", "--account", "user@example.com", "--args", `{"file_id":"file-id"}`}, &stdout, &stderr)
	if code != callExitAuthRequired || stdout.Len() != 0 || !strings.Contains(stderr.String(), "gsuite-mcp auth user@example.com") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(configDir, "oauth-port.pid")); !os.IsNotExist(err) {
		t.Fatalf("auth server pidfile exists or stat failed: %v", err)
	}
}
