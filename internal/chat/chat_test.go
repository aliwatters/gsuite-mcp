package chat

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleChatListSpaces(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list spaces successfully",
			args: map[string]any{},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["space_count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected space_count 2, got %v", result["space_count"])
				}
				spaces, ok := result["spaces"].([]any)
				if !ok || len(spaces) != 2 {
					t.Errorf("expected 2 spaces, got %v", result["spaces"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatListSpaces(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatGetSpace(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get space successfully",
			args: map[string]any{
				"space_name": "spaces/test-space-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["name"] != "spaces/test-space-1" {
					t.Errorf("expected name 'spaces/test-space-1', got %v", result["name"])
				}
				if result["display_name"] != "Test Space" {
					t.Errorf("expected display_name 'Test Space', got %v", result["display_name"])
				}
			},
		},
		{
			name:        "missing space_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "space_name parameter is required",
		},
		{
			name: "space not found",
			args: map[string]any{
				"space_name": "spaces/nonexistent",
			},
			wantErr:     true,
			errContains: "Chat API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatGetSpace(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatCreateSpace(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "create space successfully",
			args: map[string]any{
				"display_name": "New Space",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["display_name"] != "New Space" {
					t.Errorf("expected display_name 'New Space', got %v", result["display_name"])
				}
				if result["name"] == nil || result["name"] == "" {
					t.Error("expected name to be set")
				}
			},
		},
		{
			name: "create space with type",
			args: map[string]any{
				"display_name": "Group Chat",
				"space_type":   "GROUP_CHAT",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["space_type"] != "GROUP_CHAT" {
					t.Errorf("expected space_type 'GROUP_CHAT', got %v", result["space_type"])
				}
			},
		},
		{
			name:        "missing display_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "display_name parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatCreateSpace(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatListMessages(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list messages successfully",
			args: map[string]any{
				"space_name": "spaces/test-space-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["message_count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected message_count 2, got %v", result["message_count"])
				}
			},
		},
		{
			name:        "missing space_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "space_name parameter is required",
		},
		{
			name: "space not found",
			args: map[string]any{
				"space_name": "spaces/nonexistent",
			},
			wantErr:     true,
			errContains: "Chat API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatListMessages(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatGetMessage(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get message successfully",
			args: map[string]any{
				"message_name": "spaces/test-space-1/messages/msg-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["text"] != "Hello, world!" {
					t.Errorf("expected text 'Hello, world!', got %v", result["text"])
				}
				sender, ok := result["sender"].(map[string]any)
				if !ok || sender["display_name"] != "Alice" {
					t.Errorf("expected sender 'Alice', got %v", result["sender"])
				}
			},
		},
		{
			name:        "missing message_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "message_name parameter is required",
		},
		{
			name: "message not found",
			args: map[string]any{
				"message_name": "spaces/test-space-1/messages/nonexistent",
			},
			wantErr:     true,
			errContains: "Chat API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatGetMessage(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatSendMessage(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "send message successfully",
			args: map[string]any{
				"space_name": "spaces/test-space-1",
				"text":       "Hello from test!",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["text"] != "Hello from test!" {
					t.Errorf("expected text 'Hello from test!', got %v", result["text"])
				}
				if result["name"] == nil || result["name"] == "" {
					t.Error("expected name to be set")
				}
			},
		},
		{
			name: "send threaded message",
			args: map[string]any{
				"space_name":  "spaces/test-space-1",
				"text":        "Reply in thread",
				"thread_name": "spaces/test-space-1/threads/thread-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["thread_name"] != "spaces/test-space-1/threads/thread-1" {
					t.Errorf("expected thread_name, got %v", result["thread_name"])
				}
			},
		},
		{
			name: "missing space_name",
			args: map[string]any{
				"text": "Hello",
			},
			wantErr:     true,
			errContains: "space_name parameter is required",
		},
		{
			name: "missing text",
			args: map[string]any{
				"space_name": "spaces/test-space-1",
			},
			wantErr:     true,
			errContains: "text parameter is required",
		},
		{
			name: "space not found",
			args: map[string]any{
				"space_name": "spaces/nonexistent",
				"text":       "Hello",
			},
			wantErr:     true,
			errContains: "Chat API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatSendMessage(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatCreateReaction(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "create reaction successfully",
			args: map[string]any{
				"message_name": "spaces/test-space-1/messages/msg-1",
				"emoji":        "\U0001F44D",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["reaction_name"] == nil || result["reaction_name"] == "" {
					t.Error("expected reaction_name to be set")
				}
				if result["emoji"] != "\U0001F44D" {
					t.Errorf("expected emoji thumbs up, got %v", result["emoji"])
				}
			},
		},
		{
			name: "missing message_name",
			args: map[string]any{
				"emoji": "\U0001F44D",
			},
			wantErr:     true,
			errContains: "message_name parameter is required",
		},
		{
			name: "missing emoji",
			args: map[string]any{
				"message_name": "spaces/test-space-1/messages/msg-1",
			},
			wantErr:     true,
			errContains: "emoji parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatCreateReaction(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatDeleteReaction(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "delete reaction successfully",
			args: map[string]any{
				"reaction_name": "spaces/test-space-1/messages/msg-1/reactions/react-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["deleted"] != true {
					t.Errorf("expected deleted true, got %v", result["deleted"])
				}
			},
		},
		{
			name:        "missing reaction_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "reaction_name parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatDeleteReaction(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestHandleChatListMembers(t *testing.T) {
	fixtures := NewChatTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list members successfully",
			args: map[string]any{
				"space_name": "spaces/test-space-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["member_count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected member_count 2, got %v", result["member_count"])
				}
				members, ok := result["members"].([]any)
				if !ok || len(members) != 2 {
					t.Errorf("expected 2 members, got %v", result["members"])
				}
			},
		},
		{
			name:        "missing space_name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "space_name parameter is required",
		},
		{
			name: "space not found",
			args: map[string]any{
				"space_name": "spaces/nonexistent",
			},
			wantErr:     true,
			errContains: "Chat API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableChatListMembers(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}
			text := getTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got: %s", text)
				}
				if tt.errContains != "" && !strings.Contains(text, tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, text)
				}
			} else {
				if result.IsError {
					t.Fatalf("unexpected error: %s", text)
				}
				var resultData map[string]any
				if err := json.Unmarshal([]byte(text), &resultData); err != nil {
					t.Fatalf("failed to parse result: %v", err)
				}
				if tt.checkResult != nil {
					tt.checkResult(t, resultData)
				}
			}
		})
	}
}

func TestChatServiceErrors(t *testing.T) {
	fixtures := NewChatTestFixtures()

	t.Run("API error on list spaces", func(t *testing.T) {
		fixtures.MockService.Errors.ListSpaces = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListSpaces = nil }()

		request := common.CreateMCPRequest(map[string]any{})
		result, err := testableChatListSpaces(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
		text := getTextContent(result)
		if !strings.Contains(text, "Chat API error") {
			t.Errorf("expected Chat API error, got: %s", text)
		}
	})

	t.Run("API error on send message", func(t *testing.T) {
		fixtures.MockService.Errors.SendMessage = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.SendMessage = nil }()

		request := common.CreateMCPRequest(map[string]any{"space_name": "spaces/test-space-1", "text": "Hello"})
		result, err := testableChatSendMessage(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on create reaction", func(t *testing.T) {
		fixtures.MockService.Errors.CreateReaction = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.CreateReaction = nil }()

		request := common.CreateMCPRequest(map[string]any{
			"message_name": "spaces/test-space-1/messages/msg-1",
			"emoji":        "\U0001F44D",
		})
		result, err := testableChatCreateReaction(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}
