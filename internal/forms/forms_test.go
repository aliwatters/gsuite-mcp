package forms

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleFormsGet(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get form successfully",
			args: map[string]any{
				"form_id": "test-form-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Test Form" {
					t.Errorf("expected title 'Test Form', got %v", result["title"])
				}
				if result["form_id"] != "test-form-1" {
					t.Errorf("expected form_id 'test-form-1', got %v", result["form_id"])
				}
				if result["description"] != "A test form" {
					t.Errorf("expected description 'A test form', got %v", result["description"])
				}
				itemCount, ok := result["item_count"].(float64)
				if !ok || itemCount != 3 {
					t.Errorf("expected item_count 3, got %v", result["item_count"])
				}
				url, ok := result["edit_url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/forms/d/") {
					t.Errorf("expected valid forms URL, got %v", result["edit_url"])
				}
				items, ok := result["items"].([]any)
				if !ok || len(items) != 3 {
					t.Errorf("expected 3 items, got %v", result["items"])
				}
			},
		},
		{
			name: "get form with URL",
			args: map[string]any{
				"form_id": "https://docs.google.com/forms/d/test-form-1/edit",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["form_id"] != "test-form-1" {
					t.Errorf("expected ID extraction from URL, got %v", result["form_id"])
				}
			},
		},
		{
			name:        "missing form_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "form_id parameter is required",
		},
		{
			name: "form not found",
			args: map[string]any{
				"form_id": "nonexistent",
			},
			wantErr:     true,
			errContains: "Forms API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableFormsGet(context.Background(), request, fixtures.Deps)
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

func TestHandleFormsCreate(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "create form successfully",
			args: map[string]any{
				"title": "My New Form",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "My New Form" {
					t.Errorf("expected title 'My New Form', got %v", result["title"])
				}
				if result["form_id"] == nil || result["form_id"] == "" {
					t.Error("expected form_id to be set")
				}
				url, ok := result["edit_url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/forms/d/") {
					t.Errorf("expected valid forms URL, got %v", result["edit_url"])
				}
			},
		},
		{
			name:        "missing title parameter",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "title parameter is required",
		},
		{
			name: "empty title parameter",
			args: map[string]any{
				"title": "",
			},
			wantErr:     true,
			errContains: "title parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableFormsCreate(context.Background(), request, fixtures.Deps)
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

func TestHandleFormsBatchUpdate(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "batch update successfully",
			args: map[string]any{
				"form_id":  "test-form-1",
				"requests": `[{"createItem": {"item": {"title": "New Question", "questionItem": {"question": {"textQuestion": {}}}}, "location": {"index": 0}}}]`,
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["form_id"] != "test-form-1" {
					t.Errorf("expected form_id 'test-form-1', got %v", result["form_id"])
				}
			},
		},
		{
			name: "invalid JSON requests",
			args: map[string]any{
				"form_id":  "test-form-1",
				"requests": "not valid json",
			},
			wantErr:     true,
			errContains: "Invalid requests JSON",
		},
		{
			name: "missing requests parameter",
			args: map[string]any{
				"form_id": "test-form-1",
			},
			wantErr:     true,
			errContains: "requests parameter is required",
		},
		{
			name: "form not found",
			args: map[string]any{
				"form_id":  "nonexistent",
				"requests": `[{"createItem": {"item": {"title": "Q"}, "location": {"index": 0}}}]`,
			},
			wantErr:     true,
			errContains: "Forms API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableFormsBatchUpdate(context.Background(), request, fixtures.Deps)
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

func TestHandleFormsListResponses(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list responses successfully",
			args: map[string]any{
				"form_id": "test-form-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["form_id"] != "test-form-1" {
					t.Errorf("expected form_id 'test-form-1', got %v", result["form_id"])
				}
				count, ok := result["response_count"].(float64)
				if !ok || count != 1 {
					t.Errorf("expected response_count 1, got %v", result["response_count"])
				}
				responses, ok := result["responses"].([]any)
				if !ok || len(responses) != 1 {
					t.Errorf("expected 1 response, got %v", result["responses"])
				}
			},
		},
		{
			name:        "missing form_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "form_id parameter is required",
		},
		{
			name: "form not found",
			args: map[string]any{
				"form_id": "nonexistent",
			},
			wantErr:     true,
			errContains: "Forms API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableFormsListResponses(context.Background(), request, fixtures.Deps)
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

func TestHandleFormsGetResponse(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get response successfully",
			args: map[string]any{
				"form_id":     "test-form-1",
				"response_id": "resp-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["response_id"] != "resp-1" {
					t.Errorf("expected response_id 'resp-1', got %v", result["response_id"])
				}
				if result["respondent_email"] != "user@example.com" {
					t.Errorf("expected respondent_email 'user@example.com', got %v", result["respondent_email"])
				}
				if result["form_id"] != "test-form-1" {
					t.Errorf("expected form_id 'test-form-1', got %v", result["form_id"])
				}
			},
		},
		{
			name: "missing response_id",
			args: map[string]any{
				"form_id": "test-form-1",
			},
			wantErr:     true,
			errContains: "response_id parameter is required",
		},
		{
			name: "response not found",
			args: map[string]any{
				"form_id":     "test-form-1",
				"response_id": "nonexistent",
			},
			wantErr:     true,
			errContains: "Forms API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableFormsGetResponse(context.Background(), request, fixtures.Deps)
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

func TestFormsServiceErrors(t *testing.T) {
	fixtures := NewFormsTestFixtures()

	t.Run("API error on get form", func(t *testing.T) {
		fixtures.MockService.Errors.GetForm = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.GetForm = nil }()

		request := common.CreateMCPRequest(map[string]any{"form_id": "test-form-1"})
		result, err := testableFormsGet(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
		text := getTextContent(result)
		if !strings.Contains(text, "Forms API error") {
			t.Errorf("expected Forms API error, got: %s", text)
		}
	})

	t.Run("API error on create", func(t *testing.T) {
		fixtures.MockService.Errors.Create = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.Create = nil }()

		request := common.CreateMCPRequest(map[string]any{"title": "Test"})
		result, err := testableFormsCreate(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on list responses", func(t *testing.T) {
		fixtures.MockService.Errors.ListResponses = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListResponses = nil }()

		request := common.CreateMCPRequest(map[string]any{"form_id": "test-form-1"})
		result, err := testableFormsListResponses(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}
