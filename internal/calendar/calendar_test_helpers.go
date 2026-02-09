package calendar

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/calendar/v3"
)

// CalendarTestFixtures provides pre-configured test data for Calendar tests.
type CalendarTestFixtures struct {
	DefaultEmail string
	MockService  *MockCalendarService
	Deps         *CalendarHandlerDeps
}

// NewCalendarTestFixtures creates a new test fixtures instance with default configuration.
func NewCalendarTestFixtures() *CalendarTestFixtures {
	ResetCalendarIDCounter()

	mockService := NewMockCalendarService()
	setupDefaultCalendarMockData(mockService)
	f := common.NewTestFixtures[CalendarService](mockService)

	return &CalendarTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}

// setupDefaultCalendarMockData populates the mock service with standard test data.
func setupDefaultCalendarMockData(mock *MockCalendarService) {
	// Add sample calendars
	mock.Calendars["primary"] = &calendar.CalendarListEntry{
		Id:          "primary",
		Summary:     "Primary Calendar",
		Description: "Main calendar",
		Primary:     true,
		AccessRole:  "owner",
	}
	mock.Calendars["work-calendar"] = &calendar.CalendarListEntry{
		Id:          "work-calendar",
		Summary:     "Work Calendar",
		Description: "Work events",
		Primary:     false,
		AccessRole:  "owner",
	}

	// Initialize events map for primary calendar
	mock.Events["primary"] = make(map[string]*calendar.Event)
	mock.Events["work-calendar"] = make(map[string]*calendar.Event)

	// Add sample events
	mock.Events["primary"]["event001"] = createTestEvent(
		"event001",
		"Team Meeting",
		"Weekly team sync",
		"2024-02-01T10:00:00-08:00",
		"2024-02-01T11:00:00-08:00",
		false,
	)
	mock.Events["primary"]["event002"] = createTestEvent(
		"event002",
		"Project Review",
		"Q1 project review",
		"2024-02-02T14:00:00-08:00",
		"2024-02-02T15:30:00-08:00",
		false,
	)
	mock.Events["primary"]["event003"] = createTestAllDayEvent(
		"event003",
		"Company Holiday",
		"Office closed",
		"2024-02-15",
		"2024-02-16",
	)
	mock.Events["work-calendar"]["event004"] = createTestEvent(
		"event004",
		"Client Call",
		"Call with ABC Corp",
		"2024-02-03T09:00:00-08:00",
		"2024-02-03T09:30:00-08:00",
		false,
	)
}

// createTestEvent creates a timed Calendar event with standard fields.
func createTestEvent(id, summary, description, startTime, endTime string, recurring bool) *calendar.Event {
	event := &calendar.Event{
		Id:          id,
		Summary:     summary,
		Description: description,
		Status:      "confirmed",
		HtmlLink:    "https://calendar.google.com/event?eid=" + id,
		Created:     "2024-01-01T12:00:00Z",
		Updated:     "2024-01-01T12:00:00Z",
		Start: &calendar.EventDateTime{
			DateTime: startTime,
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
			TimeZone: "America/Los_Angeles",
		},
		Creator: &calendar.EventCreator{
			Email: common.TestEmail,
			Self:  true,
		},
		Organizer: &calendar.EventOrganizer{
			Email: common.TestEmail,
			Self:  true,
		},
	}

	if recurring {
		event.RecurringEventId = "recurring-" + id
		event.Recurrence = []string{"RRULE:FREQ=WEEKLY;BYDAY=MO"}
	}

	return event
}

// createTestAllDayEvent creates an all-day Calendar event.
func createTestAllDayEvent(id, summary, description, startDate, endDate string) *calendar.Event {
	return &calendar.Event{
		Id:          id,
		Summary:     summary,
		Description: description,
		Status:      "confirmed",
		HtmlLink:    "https://calendar.google.com/event?eid=" + id,
		Created:     "2024-01-01T12:00:00Z",
		Updated:     "2024-01-01T12:00:00Z",
		Start: &calendar.EventDateTime{
			Date: startDate,
		},
		End: &calendar.EventDateTime{
			Date: endDate,
		},
		Creator: &calendar.EventCreator{
			Email: common.TestEmail,
			Self:  true,
		},
		Organizer: &calendar.EventOrganizer{
			Email: common.TestEmail,
			Self:  true,
		},
	}
}

// createTestEventWithAttendees creates an event with attendees.
func createTestEventWithAttendees(id, summary string, attendeeEmails []string) *calendar.Event {
	event := createTestEvent(id, summary, "", "2024-02-01T10:00:00-08:00", "2024-02-01T11:00:00-08:00", false)

	attendees := make([]*calendar.EventAttendee, len(attendeeEmails))
	for i, email := range attendeeEmails {
		attendees[i] = &calendar.EventAttendee{
			Email:          email,
			ResponseStatus: "needsAction",
		}
	}
	event.Attendees = attendees

	return event
}
