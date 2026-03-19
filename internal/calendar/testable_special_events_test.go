package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"google.golang.org/api/calendar/v3"
)

// ============================================================================
// calendar_create_focus_time tests
// ============================================================================

func TestCalendarCreateFocusTime(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "create focus time with defaults",
			args: map[string]any{
				"start_time": "2024-03-01T09:00:00-08:00",
				"end_time":   "2024-03-01T11:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Focus Time" {
					t.Errorf("expected default summary 'Focus Time', got %v", data["summary"])
				}
				if data["event_type"] != "focusTime" {
					t.Errorf("expected event_type 'focusTime', got %v", data["event_type"])
				}
				if data["html_link"] == nil {
					t.Error("expected html_link in response")
				}
			},
		},
		{
			name: "create focus time with custom summary",
			args: map[string]any{
				"summary":    "Deep Work",
				"start_time": "2024-03-01T09:00:00-08:00",
				"end_time":   "2024-03-01T12:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Deep Work" {
					t.Errorf("expected summary 'Deep Work', got %v", data["summary"])
				}
			},
		},
		{
			name: "create focus time with auto_decline disabled",
			args: map[string]any{
				"start_time":   "2024-03-01T09:00:00-08:00",
				"auto_decline": false,
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["event_type"] != "focusTime" {
					t.Errorf("expected event_type 'focusTime', got %v", data["event_type"])
				}
			},
		},
		{
			name: "create recurring focus time",
			args: map[string]any{
				"start_time": "2024-03-01T09:00:00-08:00",
				"end_time":   "2024-03-01T11:00:00-08:00",
				"recurrence": "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["event_type"] != "focusTime" {
					t.Errorf("expected event_type 'focusTime', got %v", data["event_type"])
				}
			},
		},
		{
			name:       "missing start_time",
			args:       map[string]any{},
			wantErr:    true,
			errContain: "start_time parameter is required",
		},
		{
			name: "api error",
			args: map[string]any{
				"start_time": "2024-03-01T09:00:00-08:00",
			},
			setupMock: func(m *MockCalendarService) {
				m.Error = errors.New("API error")
			},
			wantErr:    true,
			errContain: "Calendar API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewCalendarTestFixtures()
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := CreateMCPRequest(tt.args)
			result, err := TestableCalendarCreateFocusTime(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content := getCalendarTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got success")
				}
				if tt.errContain != "" && !strings.Contains(content, tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, content)
				}
				return
			}

			if result.IsError {
				t.Errorf("unexpected error: %s", content)
				return
			}

			if tt.validate != nil {
				tt.validate(t, content)
			}
		})
	}
}

// ============================================================================
// calendar_create_out_of_office tests
// ============================================================================

func TestCalendarCreateOutOfOffice(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "create OOO with defaults",
			args: map[string]any{
				"start_time": "2024-03-15T00:00:00Z",
				"end_time":   "2024-03-16T00:00:00Z",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Out of office" {
					t.Errorf("expected default summary 'Out of office', got %v", data["summary"])
				}
				if data["event_type"] != "outOfOffice" {
					t.Errorf("expected event_type 'outOfOffice', got %v", data["event_type"])
				}
			},
		},
		{
			name: "create OOO with custom summary and message",
			args: map[string]any{
				"summary":         "Vacation",
				"start_time":      "2024-03-15T00:00:00Z",
				"end_time":        "2024-03-22T00:00:00Z",
				"decline_message": "On vacation, back March 22.",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Vacation" {
					t.Errorf("expected summary 'Vacation', got %v", data["summary"])
				}
				if data["event_type"] != "outOfOffice" {
					t.Errorf("expected event_type 'outOfOffice', got %v", data["event_type"])
				}
			},
		},
		{
			name: "create all-day OOO",
			args: map[string]any{
				"start_time": "2024-03-15",
				"end_time":   "2024-03-16",
				"all_day":    true,
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["event_type"] != "outOfOffice" {
					t.Errorf("expected event_type 'outOfOffice', got %v", data["event_type"])
				}
			},
		},
		{
			name:       "missing start_time",
			args:       map[string]any{},
			wantErr:    true,
			errContain: "start_time parameter is required",
		},
		{
			name: "api error",
			args: map[string]any{
				"start_time": "2024-03-15T00:00:00Z",
			},
			setupMock: func(m *MockCalendarService) {
				m.Error = errors.New("API error")
			},
			wantErr:    true,
			errContain: "Calendar API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewCalendarTestFixtures()
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := CreateMCPRequest(tt.args)
			result, err := TestableCalendarCreateOutOfOffice(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content := getCalendarTextContent(result)
			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result, got success")
				}
				if tt.errContain != "" && !strings.Contains(content, tt.errContain) {
					t.Errorf("expected error containing %q, got %q", tt.errContain, content)
				}
				return
			}

			if result.IsError {
				t.Errorf("unexpected error: %s", content)
				return
			}

			if tt.validate != nil {
				tt.validate(t, content)
			}
		})
	}
}

// ============================================================================
// boolToDeclineMode tests
// ============================================================================

func TestBoolToDeclineMode(t *testing.T) {
	if got := boolToDeclineMode(true); got != "declineAllConflictingInvitations" {
		t.Errorf("expected declineAllConflictingInvitations, got %s", got)
	}
	if got := boolToDeclineMode(false); got != "declineNone" {
		t.Errorf("expected declineNone, got %s", got)
	}
}

// ============================================================================
// formatEvent with event_type tests
// ============================================================================

func TestFormatEventWithEventType(t *testing.T) {
	t.Run("focusTime event includes event_type", func(t *testing.T) {
		event := createTestEvent("ft1", "Focus Time", "", "2024-03-01T09:00:00-08:00", "2024-03-01T11:00:00-08:00", false)
		event.EventType = "focusTime"
		result := formatEvent(event)
		if result["event_type"] != "focusTime" {
			t.Errorf("expected event_type 'focusTime', got %v", result["event_type"])
		}
	})

	t.Run("outOfOffice event includes event_type", func(t *testing.T) {
		event := createTestEvent("ooo1", "Out of office", "", "2024-03-15T00:00:00Z", "2024-03-16T00:00:00Z", false)
		event.EventType = "outOfOffice"
		result := formatEvent(event)
		if result["event_type"] != "outOfOffice" {
			t.Errorf("expected event_type 'outOfOffice', got %v", result["event_type"])
		}
	})

	t.Run("default event omits event_type", func(t *testing.T) {
		event := createTestEvent("e1", "Meeting", "", "2024-03-01T10:00:00-08:00", "2024-03-01T11:00:00-08:00", false)
		event.EventType = "default"
		result := formatEvent(event)
		if _, exists := result["event_type"]; exists {
			t.Error("default event should not include event_type field")
		}
	})
}

// ============================================================================
// formatEventFull with conference data tests
// ============================================================================

func TestFormatEventFullWithConferenceData(t *testing.T) {
	t.Run("includes conference data entry points", func(t *testing.T) {
		event := createTestEvent("conf1", "Meeting", "", "2024-03-01T10:00:00-08:00", "2024-03-01T11:00:00-08:00", false)
		event.ConferenceData = &calendar.ConferenceData{
			ConferenceId: "abc-defg-hij",
			ConferenceSolution: &calendar.ConferenceSolution{
				Name: "Google Meet",
			},
			EntryPoints: []*calendar.EntryPoint{
				{
					EntryPointType: "video",
					Uri:            "https://meet.google.com/abc-defg-hij",
					Label:          "meet.google.com/abc-defg-hij",
				},
				{
					EntryPointType: "phone",
					Uri:            "tel:+1-234-567-8901",
				},
			},
		}

		result := formatEventFull(event)

		if result["video_link"] != "https://meet.google.com/abc-defg-hij" {
			t.Errorf("expected video_link, got %v", result["video_link"])
		}

		confData, ok := result["conference_data"].(map[string]any)
		if !ok {
			t.Fatal("expected conference_data map")
		}
		if confData["solution"] != "Google Meet" {
			t.Errorf("expected solution 'Google Meet', got %v", confData["solution"])
		}
		if confData["conference_id"] != "abc-defg-hij" {
			t.Errorf("expected conference_id, got %v", confData["conference_id"])
		}

		entryPoints, ok := confData["entry_points"].([]map[string]any)
		if !ok {
			t.Fatal("expected entry_points array")
		}
		if len(entryPoints) != 2 {
			t.Errorf("expected 2 entry points, got %d", len(entryPoints))
		}
	})
}
