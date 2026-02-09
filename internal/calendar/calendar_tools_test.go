package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/calendar/v3"
)

// CreateMCPRequest is an alias to the common function for backward compatibility.
var CreateMCPRequest = common.CreateMCPRequest

// getTextContent extracts text content from an MCP CallToolResult.
func getCalendarTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	if textContent, ok := result.Content[0].(mcp.TextContent); ok {
		return textContent.Text
	}
	return ""
}

// ============================================================================
// calendar_list_events tests
// ============================================================================

func TestCalendarListEvents(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "list events from primary calendar",
			args: map[string]any{},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				events, ok := data["events"].([]any)
				if !ok {
					t.Error("expected events array in response")
				}
				if len(events) == 0 {
					t.Error("expected at least one event")
				}
			},
		},
		{
			name: "list events from specific calendar",
			args: map[string]any{
				"calendar_id": "work-calendar",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				events, ok := data["events"].([]any)
				if !ok {
					t.Error("expected events array in response")
				}
				if len(events) == 0 {
					t.Error("expected at least one event in work calendar")
				}
			},
		},
		{
			name: "list events with max_results",
			args: map[string]any{
				"max_results": float64(5),
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				count, ok := data["count"].(float64)
				if !ok {
					t.Error("expected count in response")
				}
				if count < 0 {
					t.Error("count should be non-negative")
				}
			},
		},
		{
			name: "api error",
			args: map[string]any{},
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
			result, err := TestableCalendarListEvents(context.Background(), request, fixtures.Deps)
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
// calendar_get_event tests
// ============================================================================

func TestCalendarGetEvent(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "get event successfully",
			args: map[string]any{
				"event_id": "event001",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["id"] != "event001" {
					t.Errorf("expected event id event001, got %v", data["id"])
				}
				if data["summary"] != "Team Meeting" {
					t.Errorf("expected summary 'Team Meeting', got %v", data["summary"])
				}
			},
		},
		{
			name: "get event from specific calendar",
			args: map[string]any{
				"event_id":    "event004",
				"calendar_id": "work-calendar",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["id"] != "event004" {
					t.Errorf("expected event id event004, got %v", data["id"])
				}
			},
		},
		{
			name:       "missing event_id",
			args:       map[string]any{},
			wantErr:    true,
			errContain: "event_id parameter is required",
		},
		{
			name: "event not found",
			args: map[string]any{
				"event_id": "nonexistent",
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
			result, err := TestableCalendarGetEvent(context.Background(), request, fixtures.Deps)
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
// calendar_create_event tests
// ============================================================================

func TestCalendarCreateEvent(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "create timed event successfully",
			args: map[string]any{
				"summary":    "New Meeting",
				"start_time": "2024-03-01T10:00:00-08:00",
				"end_time":   "2024-03-01T11:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "New Meeting" {
					t.Errorf("expected summary 'New Meeting', got %v", data["summary"])
				}
				if data["id"] == nil {
					t.Error("expected event id in response")
				}
				if data["html_link"] == nil {
					t.Error("expected html_link in response")
				}
			},
		},
		{
			name: "create event with default 1 hour duration",
			args: map[string]any{
				"summary":    "Quick Meeting",
				"start_time": "2024-03-01T14:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Quick Meeting" {
					t.Errorf("expected summary 'Quick Meeting', got %v", data["summary"])
				}
			},
		},
		{
			name: "create all-day event",
			args: map[string]any{
				"summary":    "Holiday",
				"start_time": "2024-03-15",
				"all_day":    true,
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Holiday" {
					t.Errorf("expected summary 'Holiday', got %v", data["summary"])
				}
			},
		},
		{
			name: "create event with attendees",
			args: map[string]any{
				"summary":    "Team Sync",
				"start_time": "2024-03-01T10:00:00-08:00",
				"attendees":  []any{"alice@example.com", "bob@example.com"},
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Team Sync" {
					t.Errorf("expected summary 'Team Sync', got %v", data["summary"])
				}
			},
		},
		{
			name: "create event with description and location",
			args: map[string]any{
				"summary":     "Office Meeting",
				"description": "Quarterly review",
				"location":    "Conference Room A",
				"start_time":  "2024-03-01T10:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["location"] != "Conference Room A" {
					t.Errorf("expected location 'Conference Room A', got %v", data["location"])
				}
			},
		},
		{
			name:       "missing summary",
			args:       map[string]any{"start_time": "2024-03-01T10:00:00-08:00"},
			wantErr:    true,
			errContain: "summary parameter is required",
		},
		{
			name:       "missing start_time",
			args:       map[string]any{"summary": "Test Event"},
			wantErr:    true,
			errContain: "start_time parameter is required",
		},
		{
			name: "api error",
			args: map[string]any{
				"summary":    "Test Event",
				"start_time": "2024-03-01T10:00:00-08:00",
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
			result, err := TestableCalendarCreateEvent(context.Background(), request, fixtures.Deps)
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
// calendar_update_event tests
// ============================================================================

func TestCalendarUpdateEvent(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "update event summary",
			args: map[string]any{
				"event_id": "event001",
				"summary":  "Updated Meeting Title",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["summary"] != "Updated Meeting Title" {
					t.Errorf("expected summary 'Updated Meeting Title', got %v", data["summary"])
				}
			},
		},
		{
			name: "update event times",
			args: map[string]any{
				"event_id":   "event001",
				"start_time": "2024-02-01T11:00:00-08:00",
				"end_time":   "2024-02-01T12:00:00-08:00",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["start"] != "2024-02-01T11:00:00-08:00" {
					t.Errorf("expected start '2024-02-01T11:00:00-08:00', got %v", data["start"])
				}
			},
		},
		{
			name: "update event location",
			args: map[string]any{
				"event_id": "event001",
				"location": "New Conference Room",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["location"] != "New Conference Room" {
					t.Errorf("expected location 'New Conference Room', got %v", data["location"])
				}
			},
		},
		{
			name:       "missing event_id",
			args:       map[string]any{"summary": "New Title"},
			wantErr:    true,
			errContain: "event_id parameter is required",
		},
		{
			name: "event not found",
			args: map[string]any{
				"event_id": "nonexistent",
				"summary":  "New Title",
			},
			wantErr:    true,
			errContain: "Failed to get event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixtures := NewCalendarTestFixtures()
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := CreateMCPRequest(tt.args)
			result, err := TestableCalendarUpdateEvent(context.Background(), request, fixtures.Deps)
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
// calendar_delete_event tests
// ============================================================================

func TestCalendarDeleteEvent(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]any
		setupMock  func(*MockCalendarService)
		wantErr    bool
		errContain string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "delete event successfully",
			args: map[string]any{
				"event_id": "event001",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["success"] != true {
					t.Error("expected success to be true")
				}
				if data["event_id"] != "event001" {
					t.Errorf("expected event_id 'event001', got %v", data["event_id"])
				}
				if data["message"] != "Event deleted successfully" {
					t.Errorf("expected delete success message, got %v", data["message"])
				}
			},
		},
		{
			name: "delete event from specific calendar",
			args: map[string]any{
				"event_id":    "event004",
				"calendar_id": "work-calendar",
			},
			validate: func(t *testing.T, result string) {
				var data map[string]any
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if data["success"] != true {
					t.Error("expected success to be true")
				}
				if data["calendar_id"] != "work-calendar" {
					t.Errorf("expected calendar_id 'work-calendar', got %v", data["calendar_id"])
				}
			},
		},
		{
			name:       "missing event_id",
			args:       map[string]any{},
			wantErr:    true,
			errContain: "event_id parameter is required",
		},
		{
			name: "event not found",
			args: map[string]any{
				"event_id": "nonexistent",
			},
			wantErr:    true,
			errContain: "Calendar API error",
		},
		{
			name: "api error",
			args: map[string]any{
				"event_id": "event001",
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
			result, err := TestableCalendarDeleteEvent(context.Background(), request, fixtures.Deps)
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
// Helper function tests
// ============================================================================

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    func() *calendar.Event
		validate func(t *testing.T, result map[string]any)
	}{
		{
			name: "format timed event",
			event: func() *calendar.Event {
				return createTestEvent("test1", "Test Event", "Description", "2024-02-01T10:00:00-08:00", "2024-02-01T11:00:00-08:00", false)
			},
			validate: func(t *testing.T, result map[string]any) {
				if result["id"] != "test1" {
					t.Errorf("expected id 'test1', got %v", result["id"])
				}
				if result["summary"] != "Test Event" {
					t.Errorf("expected summary 'Test Event', got %v", result["summary"])
				}
				if result["start"] != "2024-02-01T10:00:00-08:00" {
					t.Errorf("expected start time, got %v", result["start"])
				}
				if result["all_day"] != nil {
					t.Error("timed event should not have all_day field")
				}
			},
		},
		{
			name: "format all-day event",
			event: func() *calendar.Event {
				return createTestAllDayEvent("test2", "Holiday", "Day off", "2024-02-15", "2024-02-16")
			},
			validate: func(t *testing.T, result map[string]any) {
				if result["id"] != "test2" {
					t.Errorf("expected id 'test2', got %v", result["id"])
				}
				if result["start"] != "2024-02-15" {
					t.Errorf("expected start date '2024-02-15', got %v", result["start"])
				}
				if result["all_day"] != true {
					t.Error("all-day event should have all_day = true")
				}
			},
		},
		{
			name: "format event with location",
			event: func() *calendar.Event {
				e := createTestEvent("test3", "Meeting", "", "2024-02-01T10:00:00-08:00", "2024-02-01T11:00:00-08:00", false)
				e.Location = "Room 123"
				return e
			},
			validate: func(t *testing.T, result map[string]any) {
				if result["location"] != "Room 123" {
					t.Errorf("expected location 'Room 123', got %v", result["location"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.event()
			result := formatEvent(event)
			tt.validate(t, result)
		})
	}
}

func TestFormatEventFull(t *testing.T) {
	tests := []struct {
		name     string
		event    func() *calendar.Event
		validate func(t *testing.T, result map[string]any)
	}{
		{
			name: "full format includes all fields",
			event: func() *calendar.Event {
				e := createTestEvent("test1", "Test Event", "Test Description", "2024-02-01T10:00:00-08:00", "2024-02-01T11:00:00-08:00", false)
				e.HtmlLink = "https://calendar.google.com/event?eid=test1"
				return e
			},
			validate: func(t *testing.T, result map[string]any) {
				if result["html_link"] != "https://calendar.google.com/event?eid=test1" {
					t.Errorf("expected html_link, got %v", result["html_link"])
				}
				if result["description"] != "Test Description" {
					t.Errorf("expected description 'Test Description', got %v", result["description"])
				}
				if result["created"] == nil {
					t.Error("expected created timestamp")
				}
			},
		},
		{
			name: "full format includes attendees",
			event: func() *calendar.Event {
				return createTestEventWithAttendees("test2", "Team Meeting", []string{"alice@example.com", "bob@example.com"})
			},
			validate: func(t *testing.T, result map[string]any) {
				attendees, ok := result["attendees"].([]map[string]any)
				if !ok {
					t.Error("expected attendees array")
					return
				}
				if len(attendees) != 2 {
					t.Errorf("expected 2 attendees, got %d", len(attendees))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.event()
			result := formatEventFull(event)
			tt.validate(t, result)
		})
	}
}

// ============================================================================
// Mock service tests
// ============================================================================

func TestMockCalendarService(t *testing.T) {
	t.Run("tracks method calls", func(t *testing.T) {
		mock := NewMockCalendarService()
		mock.Events["primary"] = make(map[string]*calendar.Event)

		_, _ = mock.ListEvents(context.Background(), "primary", nil)
		_, _ = mock.GetEvent(context.Background(), "primary", "event1", "")

		if len(mock.MethodCalls) != 2 {
			t.Errorf("expected 2 method calls, got %d", len(mock.MethodCalls))
		}
		if mock.MethodCalls[0].Method != "ListEvents" {
			t.Errorf("expected first call to be ListEvents, got %s", mock.MethodCalls[0].Method)
		}
		if mock.MethodCalls[1].Method != "GetEvent" {
			t.Errorf("expected second call to be GetEvent, got %s", mock.MethodCalls[1].Method)
		}
	})

	t.Run("returns injected error", func(t *testing.T) {
		mock := NewMockCalendarService()
		mock.Error = errors.New("injected error")

		_, err := mock.ListEvents(context.Background(), "primary", nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "injected error" {
			t.Errorf("expected 'injected error', got %v", err)
		}
	})

	t.Run("creates events with generated IDs", func(t *testing.T) {
		ResetCalendarIDCounter()
		mock := NewMockCalendarService()

		event1, _ := mock.CreateEvent(context.Background(), "primary", &calendar.Event{Summary: "Event 1"}, 0)
		event2, _ := mock.CreateEvent(context.Background(), "primary", &calendar.Event{Summary: "Event 2"}, 0)

		if event1.Id == event2.Id {
			t.Error("expected different IDs for different events")
		}
	})
}
