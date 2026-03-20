package mcptest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/mcptest"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// TestMCPInitialize verifies the server responds with correct protocol version and capabilities.
func TestMCPInitialize(t *testing.T) {
	s := mcptest.NewTestServer()

	initMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {"name": "test", "version": "1.0"}
		}
	}`)

	resp := s.HandleMessage(context.Background(), initMsg)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var result struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  struct {
			ProtocolVersion string `json:"protocolVersion"`
			Capabilities    struct {
				Tools map[string]any `json:"tools"`
			} `json:"capabilities"`
			ServerInfo struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"serverInfo"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Result.ProtocolVersion != "2024-11-05" {
		t.Errorf("protocol version = %q, want %q", result.Result.ProtocolVersion, "2024-11-05")
	}
	if result.Result.ServerInfo.Name != mcptest.ServerName {
		t.Errorf("server name = %q, want %q", result.Result.ServerInfo.Name, mcptest.ServerName)
	}
	if result.Result.ServerInfo.Version != mcptest.ServerVersion {
		t.Errorf("server version = %q, want %q", result.Result.ServerInfo.Version, mcptest.ServerVersion)
	}
	if result.Result.Capabilities.Tools == nil {
		t.Error("capabilities.tools should not be nil")
	}
}

// TestMCPToolCount verifies the total number of registered tools hasn't changed unexpectedly.
// This is a regression guard — update mcptest.ServiceToolCounts when adding/removing tools.
func TestMCPToolCount(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)
	expected := mcptest.TotalToolCount()

	if len(tools) != expected {
		t.Errorf("tool count = %d, want %d", len(tools), expected)

		if len(tools) > expected {
			t.Log("New tools detected — update mcptest.ServiceToolCounts")
		} else {
			t.Log("Missing tools detected — a RegisterTools function may have a bug")
		}
	}
}

// TestMCPToolCountByService verifies each service registers the expected number of tools.
func TestMCPToolCountByService(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	// Group tools by service prefix
	counts := make(map[string]int)
	for _, tool := range tools {
		prefix := toolServicePrefix(tool.Name)
		counts[prefix]++
	}

	for service, expected := range mcptest.ServiceToolCounts {
		got := counts[service]
		if got != expected {
			t.Errorf("service %q: tool count = %d, want %d", service, got, expected)
		}
	}

	// Check for tools with unknown prefixes
	for prefix, count := range counts {
		if _, ok := mcptest.ServiceToolCounts[prefix]; !ok {
			t.Errorf("unknown tool prefix %q with %d tools — add to ServiceToolCounts", prefix, count)
		}
	}
}

// TestMCPToolsHaveDescriptions verifies every tool has a non-empty description.
func TestMCPToolsHaveDescriptions(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
}

// TestMCPToolsHaveAccountParam verifies every tool accepts the "account" parameter.
func TestMCPToolsHaveAccountParam(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	for _, tool := range tools {
		if _, hasAccount := tool.InputSchema.Properties["account"]; !hasAccount {
			t.Errorf("tool %q missing required 'account' parameter", tool.Name)
		}
	}
}

// TestMCPToolNamesAreUnique verifies no duplicate tool names.
func TestMCPToolNamesAreUnique(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	seen := make(map[string]bool)
	for _, tool := range tools {
		if seen[tool.Name] {
			t.Errorf("duplicate tool name: %q", tool.Name)
		}
		seen[tool.Name] = true
	}
}

// TestMCPToolNamesFollowConvention verifies all tool names use the service_action format.
func TestMCPToolNamesFollowConvention(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	validPrefixes := make(map[string]bool)
	for prefix := range mcptest.ServiceToolCounts {
		validPrefixes[prefix] = true
	}

	for _, tool := range tools {
		parts := strings.SplitN(tool.Name, "_", 2)
		if len(parts) < 2 {
			t.Errorf("tool %q does not follow service_action naming convention", tool.Name)
			continue
		}
		if !validPrefixes[parts[0]] {
			t.Errorf("tool %q has unknown service prefix %q", tool.Name, parts[0])
		}
	}
}

// TestMCPToolSchemasAreValid verifies tool schemas have correct JSON Schema structure.
func TestMCPToolSchemasAreValid(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	for _, tool := range tools {
		// Verify schema type is "object"
		if tool.InputSchema.Type != "object" {
			t.Errorf("tool %q: InputSchema.Type = %q, want %q", tool.Name, tool.InputSchema.Type, "object")
		}

		props := tool.InputSchema.Properties

		// Verify each property has a type or description
		for propName, propVal := range props {
			propMap, ok := propVal.(map[string]any)
			if !ok {
				t.Errorf("tool %q property %q: value is not a map", tool.Name, propName)
				continue
			}
			_, hasType := propMap["type"]
			_, hasDesc := propMap["description"]
			if !hasType && !hasDesc {
				t.Errorf("tool %q property %q: has neither type nor description", tool.Name, propName)
			}
		}

		// Verify required fields reference existing properties
		for _, req := range tool.InputSchema.Required {
			if _, ok := props[req]; !ok {
				t.Errorf("tool %q: required field %q not found in properties", tool.Name, req)
			}
		}
	}
}

// TestMCPToolCallWithoutCredentials verifies tools return clean MCP errors (not Go panics)
// when called without authentication configured.
func TestMCPToolCallWithoutCredentials(t *testing.T) {
	s := mcptest.NewTestServer()

	// Initialize the server first (required for tools/call)
	initMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {"name": "test", "version": "1.0"}
		}
	}`)
	s.HandleMessage(context.Background(), initMsg)

	// Test a representative tool from each service
	testCases := []struct {
		name string
		args string
	}{
		{"gmail_search", `{"query": "test", "account": "test@example.com"}`},
		{"calendar_list_calendars", `{"account": "test@example.com"}`},
		{"drive_list", `{"account": "test@example.com"}`},
		{"docs_get", `{"document_id": "abc", "account": "test@example.com"}`},
		{"sheets_get", `{"spreadsheet_id": "abc", "account": "test@example.com"}`},
		{"tasks_list_tasklists", `{"account": "test@example.com"}`},
		{"contacts_list", `{"account": "test@example.com"}`},
		{"slides_get_presentation", `{"presentation_id": "abc", "account": "test@example.com"}`},
		{"forms_get", `{"form_id": "abc", "account": "test@example.com"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callMsg := json.RawMessage(fmt.Sprintf(`{
				"jsonrpc": "2.0",
				"id": 2,
				"method": "tools/call",
				"params": {"name": %q, "arguments": %s}
			}`, tc.name, tc.args))

			resp := s.HandleMessage(context.Background(), callMsg)

			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatalf("failed to marshal response: %v", err)
			}

			// Should get a valid JSON-RPC response (not a panic or internal error)
			var rpcResp struct {
				JSONRPC string `json:"jsonrpc"`
				ID      int    `json:"id"`
				Result  *struct {
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
					IsError bool `json:"isError"`
				} `json:"result"`
				Error *struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}

			if err := json.Unmarshal(data, &rpcResp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Should be a tool result error (isError=true) not a JSON-RPC error
			if rpcResp.Error != nil {
				t.Errorf("got JSON-RPC error (code %d: %s), want tool result error", rpcResp.Error.Code, rpcResp.Error.Message)
				return
			}

			if rpcResp.Result == nil {
				t.Fatal("got nil result")
			}

			if !rpcResp.Result.IsError {
				t.Error("expected isError=true for unauthenticated call")
			}

			// Verify the error message is user-friendly
			if len(rpcResp.Result.Content) == 0 {
				t.Error("error result has no content")
				return
			}

			errText := rpcResp.Result.Content[0].Text
			if errText == "" {
				t.Error("error message is empty")
			}
		})
	}
}

// TestMCPInvalidToolCall verifies the server handles calls to non-existent tools gracefully.
func TestMCPInvalidToolCall(t *testing.T) {
	s := mcptest.NewTestServer()

	// Initialize first
	initMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {"name": "test", "version": "1.0"}
		}
	}`)
	s.HandleMessage(context.Background(), initMsg)

	callMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "tools/call",
		"params": {"name": "nonexistent_tool", "arguments": {}}
	}`)

	resp := s.HandleMessage(context.Background(), callMsg)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var rpcResp struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &rpcResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if rpcResp.Error == nil {
		t.Error("expected JSON-RPC error for nonexistent tool, got nil")
	}
}

// TestMCPToolListIsSorted verifies tools are returned in a consistent order.
func TestMCPToolListIsSorted(t *testing.T) {
	s := mcptest.NewTestServer()

	tools := listTools(t, s)

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}

	if !sort.StringsAreSorted(names) {
		// Not necessarily a bug, but good to verify consistency
		t.Log("tools are not alphabetically sorted (informational)")
	}

	// Verify tools within each service are grouped together
	lastPrefix := ""
	seenPrefixes := make(map[string]bool)
	for _, name := range names {
		prefix := toolServicePrefix(name)
		if prefix != lastPrefix {
			if seenPrefixes[prefix] {
				t.Errorf("tools for service %q are not grouped together", prefix)
			}
			seenPrefixes[prefix] = true
			lastPrefix = prefix
		}
	}
}

// listTools sends a tools/list request and returns the tools.
func listTools(t *testing.T, s *server.MCPServer) []mcp.Tool {
	t.Helper()

	// Initialize first
	initMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2024-11-05",
			"capabilities": {},
			"clientInfo": {"name": "test", "version": "1.0"}
		}
	}`)
	s.HandleMessage(context.Background(), initMsg)

	listMsg := json.RawMessage(`{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "tools/list",
		"params": {}
	}`)

	resp := s.HandleMessage(context.Background(), listMsg)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var result struct {
		Result struct {
			Tools []mcp.Tool `json:"tools"`
		} `json:"result"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal tools list: %v", err)
	}

	return result.Result.Tools
}

// toolServicePrefix extracts the service prefix from a tool name (e.g., "gmail" from "gmail_search").
func toolServicePrefix(name string) string {
	parts := strings.SplitN(name, "_", 2)
	return parts[0]
}
