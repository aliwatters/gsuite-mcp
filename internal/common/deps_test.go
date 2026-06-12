package common

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestResolveAccountFromRequestWithDepsRejectsInvalidAccount(t *testing.T) {
	tests := []struct {
		name    string
		account string
	}{
		{name: "traversal", account: "../user@example.com"},
		{name: "absolute path", account: filepath.Join(string(filepath.Separator), "tmp", "user@example.com")},
		{name: "null byte", account: "user\x00@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: map[string]any{"account": tt.account},
				},
			}

			email, err := ResolveAccountFromRequestWithDeps(request, nil)
			if err == nil {
				t.Fatal("expected invalid account error")
			}
			if email != "" {
				t.Errorf("expected empty email on error, got %q", email)
			}
			if !strings.Contains(err.Error(), "invalid account") {
				t.Errorf("expected invalid account error, got %v", err)
			}
		})
	}
}
