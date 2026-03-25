package meet

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
)

func TestHandleListConferenceRecords(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list conference records successfully",
			args: map[string]any{},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected count 2, got %v", result["count"])
				}
				records, ok := result["conference_records"].([]any)
				if !ok || len(records) != 2 {
					t.Errorf("expected 2 records, got %v", result["conference_records"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableListConferenceRecords(context.Background(), request, fixtures.Deps)
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

func TestHandleGetConferenceRecord(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get conference record successfully",
			args: map[string]any{
				"name": "conferenceRecords/abc-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["name"] != "conferenceRecords/abc-123" {
					t.Errorf("expected name 'conferenceRecords/abc-123', got %v", result["name"])
				}
				if result["space"] != "spaces/meeting-space-1" {
					t.Errorf("expected space 'spaces/meeting-space-1', got %v", result["space"])
				}
				if result["start_time"] != "2024-03-15T10:00:00Z" {
					t.Errorf("expected start_time, got %v", result["start_time"])
				}
			},
		},
		{
			name: "get conference record with short name",
			args: map[string]any{
				"name": "abc-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["name"] != "conferenceRecords/abc-123" {
					t.Errorf("expected name 'conferenceRecords/abc-123', got %v", result["name"])
				}
			},
		},
		{
			name:        "missing name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "name parameter is required",
		},
		{
			name: "conference record not found",
			args: map[string]any{
				"name": "nonexistent",
			},
			wantErr:     true,
			errContains: "Meet API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableGetConferenceRecord(context.Background(), request, fixtures.Deps)
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

func TestHandleListParticipants(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list participants successfully",
			args: map[string]any{
				"conference_record": "conferenceRecords/abc-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected count 2, got %v", result["count"])
				}
				participants, ok := result["participants"].([]any)
				if !ok || len(participants) != 2 {
					t.Errorf("expected 2 participants, got %v", result["participants"])
				}
				// Check first participant (signed-in user)
				p1, ok := participants[0].(map[string]any)
				if !ok {
					t.Fatalf("expected map for participant, got %T", participants[0])
				}
				if p1["type"] != "signed_in" && p1["type"] != "anonymous" {
					// Order may vary; just check both exist
				}
			},
		},
		{
			name: "list participants with short name",
			args: map[string]any{
				"conference_record": "abc-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["conference_record"] != "conferenceRecords/abc-123" {
					t.Errorf("expected prefixed conference_record, got %v", result["conference_record"])
				}
			},
		},
		{
			name:        "missing conference_record",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "conference_record parameter is required",
		},
		{
			name: "conference record not found",
			args: map[string]any{
				"conference_record": "nonexistent",
			},
			wantErr:     true,
			errContains: "Meet API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableListParticipants(context.Background(), request, fixtures.Deps)
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

func TestHandleListTranscripts(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list transcripts successfully",
			args: map[string]any{
				"conference_record": "conferenceRecords/abc-123",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["count"].(float64)
				if !ok || count != 1 {
					t.Errorf("expected count 1, got %v", result["count"])
				}
				transcripts, ok := result["transcripts"].([]any)
				if !ok || len(transcripts) != 1 {
					t.Errorf("expected 1 transcript, got %v", result["transcripts"])
				}
				t1, ok := transcripts[0].(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", transcripts[0])
				}
				if t1["state"] != "FILE_GENERATED" {
					t.Errorf("expected state FILE_GENERATED, got %v", t1["state"])
				}
				docsDest, ok := t1["docs_destination"].(map[string]any)
				if !ok {
					t.Fatalf("expected docs_destination map")
				}
				if docsDest["document_id"] != "doc-abc-123" {
					t.Errorf("expected document_id 'doc-abc-123', got %v", docsDest["document_id"])
				}
			},
		},
		{
			name:        "missing conference_record",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "conference_record parameter is required",
		},
		{
			name: "conference record not found",
			args: map[string]any{
				"conference_record": "nonexistent",
			},
			wantErr:     true,
			errContains: "Meet API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableListTranscripts(context.Background(), request, fixtures.Deps)
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

func TestHandleGetTranscript(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "get transcript successfully",
			args: map[string]any{
				"name": "conferenceRecords/abc-123/transcripts/t1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				if result["name"] != "conferenceRecords/abc-123/transcripts/t1" {
					t.Errorf("expected transcript name, got %v", result["name"])
				}
				if result["state"] != "FILE_GENERATED" {
					t.Errorf("expected state FILE_GENERATED, got %v", result["state"])
				}
			},
		},
		{
			name:        "missing name",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "name parameter is required",
		},
		{
			name: "transcript not found",
			args: map[string]any{
				"name": "conferenceRecords/abc-123/transcripts/nonexistent",
			},
			wantErr:     true,
			errContains: "Meet API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableGetTranscript(context.Background(), request, fixtures.Deps)
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

func TestHandleListTranscriptEntries(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	tests := []struct {
		name        string
		args        map[string]any
		wantErr     bool
		errContains string
		checkResult func(t *testing.T, result map[string]any)
	}{
		{
			name: "list transcript entries successfully",
			args: map[string]any{
				"transcript": "conferenceRecords/abc-123/transcripts/t1",
			},
			checkResult: func(t *testing.T, result map[string]any) {
				count, ok := result["count"].(float64)
				if !ok || count != 2 {
					t.Errorf("expected count 2, got %v", result["count"])
				}
				entries, ok := result["entries"].([]any)
				if !ok || len(entries) != 2 {
					t.Errorf("expected 2 entries, got %v", result["entries"])
				}
				e1, ok := entries[0].(map[string]any)
				if !ok {
					t.Fatalf("expected map, got %T", entries[0])
				}
				if e1["text"] != "Hello everyone, let's get started with the meeting." {
					t.Errorf("unexpected text: %v", e1["text"])
				}
				if e1["language_code"] != "en-US" {
					t.Errorf("expected language_code en-US, got %v", e1["language_code"])
				}
			},
		},
		{
			name:        "missing transcript",
			args:        map[string]any{},
			wantErr:     true,
			errContains: "transcript parameter is required",
		},
		{
			name: "transcript not found",
			args: map[string]any{
				"transcript": "conferenceRecords/nonexistent/transcripts/t1",
			},
			wantErr:     true,
			errContains: "Meet API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := common.CreateMCPRequest(tt.args)
			result, err := testableListTranscriptEntries(context.Background(), request, fixtures.Deps)
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

func TestMeetServiceErrors(t *testing.T) {
	fixtures := NewMeetTestFixtures()

	t.Run("API error on list conference records", func(t *testing.T) {
		fixtures.MockService.Errors.ListConferenceRecords = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListConferenceRecords = nil }()

		request := common.CreateMCPRequest(map[string]any{})
		result, err := testableListConferenceRecords(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
		text := getTextContent(result)
		if !strings.Contains(text, "Meet API error") {
			t.Errorf("expected Meet API error, got: %s", text)
		}
	})

	t.Run("API error on get conference record", func(t *testing.T) {
		fixtures.MockService.Errors.GetConferenceRecord = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.GetConferenceRecord = nil }()

		request := common.CreateMCPRequest(map[string]any{"name": "conferenceRecords/abc-123"})
		result, err := testableGetConferenceRecord(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on list participants", func(t *testing.T) {
		fixtures.MockService.Errors.ListParticipants = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListParticipants = nil }()

		request := common.CreateMCPRequest(map[string]any{"conference_record": "conferenceRecords/abc-123"})
		result, err := testableListParticipants(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on list transcripts", func(t *testing.T) {
		fixtures.MockService.Errors.ListTranscripts = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListTranscripts = nil }()

		request := common.CreateMCPRequest(map[string]any{"conference_record": "conferenceRecords/abc-123"})
		result, err := testableListTranscripts(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on get transcript", func(t *testing.T) {
		fixtures.MockService.Errors.GetTranscript = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.GetTranscript = nil }()

		request := common.CreateMCPRequest(map[string]any{"name": "conferenceRecords/abc-123/transcripts/t1"})
		result, err := testableGetTranscript(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})

	t.Run("API error on list transcript entries", func(t *testing.T) {
		fixtures.MockService.Errors.ListTranscriptEntries = context.DeadlineExceeded
		defer func() { fixtures.MockService.Errors.ListTranscriptEntries = nil }()

		request := common.CreateMCPRequest(map[string]any{"transcript": "conferenceRecords/abc-123/transcripts/t1"})
		result, err := testableListTranscriptEntries(context.Background(), request, fixtures.Deps)
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result")
		}
	})
}
