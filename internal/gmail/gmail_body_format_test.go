package gmail

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

func TestParseBodyFormatDefaultsToText(t *testing.T) {
	for _, args := range []map[string]any{
		nil,
		{},
		{"body_format": nil},
		{"body_format": ""},
	} {
		bodyFormat, errResult := parseBodyFormat(args)
		if errResult != nil {
			t.Fatalf("unexpected tool error for arguments %#v: %v", args, errResult.Content)
		}
		if bodyFormat != BodyFormatText {
			t.Errorf("parseBodyFormat(%#v) = %q, want %q", args, bodyFormat, BodyFormatText)
		}
	}
}

func TestBodyFormatValidationOnReadSurfaces(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		validValue string
		args       func(any) map[string]any
		setup      func(*MockGmailService)
		call       func(context.Context, mcp.CallToolRequest, *GmailHandlerDeps) (*mcp.CallToolResult, error)
	}{
		{
			name:       "message",
			method:     "GetMessage",
			validValue: "text",
			args: func(bodyFormat any) map[string]any {
				return map[string]any{"message_id": "msg1", "body_format": bodyFormat}
			},
			setup: func(service *MockGmailService) {
				service.AddMessage(&gmail.Message{Id: "msg1"})
			},
			call: TestableGmailGetMessage,
		},
		{
			name:       "messages",
			method:     "GetMessage",
			validValue: "html",
			args: func(bodyFormat any) map[string]any {
				return map[string]any{"message_ids": []any{"msg1"}, "body_format": bodyFormat}
			},
			setup: func(service *MockGmailService) {
				service.AddMessage(&gmail.Message{Id: "msg1"})
			},
			call: TestableGmailGetMessages,
		},
		{
			name:       "thread",
			method:     "GetThread",
			validValue: "full",
			args: func(bodyFormat any) map[string]any {
				return map[string]any{"thread_id": "thread1", "body_format": bodyFormat}
			},
			setup: func(service *MockGmailService) {
				service.AddThread(&gmail.Thread{Id: "thread1"})
			},
			call: TestableGmailGetThread,
		},
		{
			name:       "draft",
			method:     "GetDraft",
			validValue: "html",
			args: func(bodyFormat any) map[string]any {
				return map[string]any{"draft_id": "draft1", "body_format": bodyFormat}
			},
			setup: func(service *MockGmailService) {
				service.AddDraft(&gmail.Draft{Id: "draft1", Message: &gmail.Message{Id: "msg1"}})
			},
			call: TestableGmailGetDraft,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("valid explicit value", func(t *testing.T) {
				fixtures := NewGmailTestFixtures()
				tt.setup(fixtures.MockService)

				result, err := tt.call(t.Context(), makeBodyFormatRequest(tt.args(tt.validValue)), fixtures.Deps)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result.IsError {
					t.Fatalf("unexpected tool error: %v", result.Content)
				}
				if !fixtures.MockService.WasMethodCalled(tt.method) {
					t.Errorf("expected %s to be called for valid body_format", tt.method)
				}
			})

			for _, invalid := range []struct {
				name            string
				value           any
				errorSubstrings []string
			}{
				{name: "invalid string", value: "plain", errorSubstrings: []string{"plain", "text", "html", "full"}},
				{name: "non-string", value: 123, errorSubstrings: []string{"int", "text", "html", "full"}},
			} {
				t.Run(invalid.name, func(t *testing.T) {
					fixtures := NewGmailTestFixtures()
					tt.setup(fixtures.MockService)

					result, err := tt.call(t.Context(), makeBodyFormatRequest(tt.args(invalid.value)), fixtures.Deps)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !result.IsError {
						t.Fatalf("expected tool error for body_format %#v", invalid.value)
					}
					if fixtures.MockService.WasMethodCalled(tt.method) {
						t.Errorf("expected %s not to be called for invalid body_format", tt.method)
					}
					errText := fmt.Sprintf("%v", result.Content)
					for _, substring := range invalid.errorSubstrings {
						if !strings.Contains(errText, substring) {
							t.Errorf("expected error to contain %q, got: %s", substring, errText)
						}
					}
				})
			}
		})
	}
}

func makeBodyFormatRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: args},
	}
}
