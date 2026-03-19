package slides

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleSlidesGetPresentation(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get presentation successfully",
			args: map[string]any{
				"presentation_id": "test-pres-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "Test Presentation" {
					t.Errorf("expected title 'Test Presentation', got %v", result["title"])
				}
				if result["presentation_id"] != "test-pres-1" {
					t.Errorf("expected presentation_id 'test-pres-1', got %v", result["presentation_id"])
				}
				slideCount, ok := result["slide_count"].(float64)
				if !ok || slideCount != 2 {
					t.Errorf("expected slide_count 2, got %v", result["slide_count"])
				}
				url, ok := result["url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/presentation/d/") {
					t.Errorf("expected valid slides URL, got %v", result["url"])
				}
				slides, ok := result["slides"].([]any)
				if !ok || len(slides) != 2 {
					t.Errorf("expected 2 slides in summary, got %v", result["slides"])
				}
			},
		},
		{
			name: "get presentation with URL",
			args: map[string]any{
				"presentation_id": "https://docs.google.com/presentation/d/test-pres-1/edit",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["presentation_id"] != "test-pres-1" {
					t.Errorf("expected ID extraction from URL, got %v", result["presentation_id"])
				}
			},
		},
		{
			name:        "missing presentation_id",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "presentation_id parameter is required",
		},
		{
			name: "presentation not found",
			args: map[string]any{
				"presentation_id": "nonexistent",
			},
			wantErr:     true,
			errContains: "Slides API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableSlidesGetPresentation(context.Background(), request, fixtures.Deps)
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

func TestHandleSlidesGetPage(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get page successfully",
			args: map[string]any{
				"presentation_id": "test-pres-1",
				"page_id":         "slide-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["page_id"] != "slide-1" {
					t.Errorf("expected page_id 'slide-1', got %v", result["page_id"])
				}
				if result["page_type"] != "SLIDE" {
					t.Errorf("expected page_type 'SLIDE', got %v", result["page_type"])
				}
				elements, ok := result["elements"].([]any)
				if !ok || len(elements) != 1 {
					t.Errorf("expected 1 element, got %v", result["elements"])
				}
			},
		},
		{
			name: "missing page_id",
			args: map[string]any{
				"presentation_id": "test-pres-1",
			},
			wantErr:     true,
			errContains: "page_id parameter is required",
		},
		{
			name: "page not found",
			args: map[string]any{
				"presentation_id": "test-pres-1",
				"page_id":         "nonexistent",
			},
			wantErr:     true,
			errContains: "Slides API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableSlidesGetPage(context.Background(), request, fixtures.Deps)
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

func TestHandleSlidesGetThumbnail(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get thumbnail successfully",
			args: map[string]any{
				"presentation_id": "test-pres-1",
				"page_id":         "slide-1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				contentURL, ok := result["content_url"].(string)
				if !ok || contentURL == "" {
					t.Error("expected non-empty content_url")
				}
				width, ok := result["width"].(float64)
				if !ok || width <= 0 {
					t.Errorf("expected positive width, got %v", result["width"])
				}
			},
		},
		{
			name: "page not found for thumbnail",
			args: map[string]any{
				"presentation_id": "test-pres-1",
				"page_id":         "nonexistent",
			},
			wantErr:     true,
			errContains: "Slides API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableSlidesGetThumbnail(context.Background(), request, fixtures.Deps)
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

func TestHandleSlidesCreate(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "create presentation successfully",
			args: map[string]any{
				"title": "My New Presentation",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["title"] != "My New Presentation" {
					t.Errorf("expected title 'My New Presentation', got %v", result["title"])
				}
				if result["presentation_id"] == nil || result["presentation_id"] == "" {
					t.Error("expected presentation_id to be set")
				}
				url, ok := result["url"].(string)
				if !ok || !strings.Contains(url, "docs.google.com/presentation/d/") {
					t.Errorf("expected valid slides URL, got %v", result["url"])
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
			result, err := testableSlidesCreate(context.Background(), request, fixtures.Deps)
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

func TestHandleSlidesBatchUpdate(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

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
				"presentation_id": "test-pres-1",
				"requests":        `[{"createSlide": {"slideLayoutReference": {"predefinedLayout": "BLANK"}}}]`,
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["presentation_id"] != "test-pres-1" {
					t.Errorf("expected presentation_id 'test-pres-1', got %v", result["presentation_id"])
				}
			},
		},
		{
			name: "invalid JSON requests",
			args: map[string]any{
				"presentation_id": "test-pres-1",
				"requests":        "not valid json",
			},
			wantErr:     true,
			errContains: "Invalid requests JSON",
		},
		{
			name: "missing requests parameter",
			args: map[string]any{
				"presentation_id": "test-pres-1",
			},
			wantErr:     true,
			errContains: "requests parameter is required",
		},
		{
			name: "presentation not found",
			args: map[string]any{
				"presentation_id": "nonexistent",
				"requests":        `[{"createSlide": {}}]`,
			},
			wantErr:     true,
			errContains: "Slides API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableSlidesBatchUpdate(context.Background(), request, fixtures.Deps)
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

func TestSlidesServiceErrors(t *testing.T) {
	fixtures := NewSlidesTestFixtures()

	t.Run("API error on get presentation", func(t *testing.T) {
		fixtures.MockService.Errors.GetPresentation = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.GetPresentation = nil }()

		request := common.CreateMCPRequest(map[string]any{"presentation_id": "test-pres-1"})
		result, err := testableSlidesGetPresentation(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
		text := getTextContent(result)
		if !strings.Contains(text, "Slides API error") {
			t.Errorf("expected Slides API error, got: %s", text)
		}
	})

	t.Run("API error on create", func(t *testing.T) {
		fixtures.MockService.Errors.Create = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.Create = nil }()

		request := common.CreateMCPRequest(map[string]any{"title": "Test"})
		result, err := testableSlidesCreate(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}
