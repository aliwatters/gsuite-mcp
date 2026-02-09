package calendar

import (
	"context"
	"errors"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/calendar/v3"
)

// MethodCall is an alias to the common.MethodCall for backward compatibility in tests
type MethodCall = common.MethodCall

// MockCalendarService is a mock implementation of CalendarService for testing.
type MockCalendarService struct {
	// Storage
	Calendars map[string]*calendar.CalendarListEntry
	Events    map[string]map[string]*calendar.Event // calendarID -> eventID -> event

	// Error injection for testing error handling
	Error error

	// Track method calls for verification
	MethodCalls []MethodCall
}

// NewMockCalendarService creates a new mock Calendar service with empty storage.
func NewMockCalendarService() *MockCalendarService {
	return &MockCalendarService{
		Calendars: make(map[string]*calendar.CalendarListEntry),
		Events:    make(map[string]map[string]*calendar.Event),
	}
}

func (m *MockCalendarService) recordCall(method string, args map[string]any) {
	m.MethodCalls = append(m.MethodCalls, MethodCall{Method: method, Args: args})
}

// ListCalendars returns all calendars.
func (m *MockCalendarService) ListCalendars(ctx context.Context, fields string) (*calendar.CalendarList, error) {
	m.recordCall("ListCalendars", map[string]any{"fields": fields})
	if m.Error != nil {
		return nil, m.Error
	}

	items := make([]*calendar.CalendarListEntry, 0, len(m.Calendars))
	for _, cal := range m.Calendars {
		items = append(items, cal)
	}

	return &calendar.CalendarList{Items: items}, nil
}

// ListEvents returns events from a calendar.
func (m *MockCalendarService) ListEvents(ctx context.Context, calendarID string, opts *ListEventsOptions) (*calendar.Events, error) {
	m.recordCall("ListEvents", map[string]any{"calendarID": calendarID, "opts": opts})
	if m.Error != nil {
		return nil, m.Error
	}

	calEvents, ok := m.Events[calendarID]
	if !ok {
		calEvents = make(map[string]*calendar.Event)
	}

	items := make([]*calendar.Event, 0, len(calEvents))
	for _, event := range calEvents {
		items = append(items, event)
	}

	return &calendar.Events{Items: items}, nil
}

// GetEvent returns an event by ID.
func (m *MockCalendarService) GetEvent(ctx context.Context, calendarID string, eventID string, fields string) (*calendar.Event, error) {
	m.recordCall("GetEvent", map[string]any{"calendarID": calendarID, "eventID": eventID, "fields": fields})
	if m.Error != nil {
		return nil, m.Error
	}

	calEvents, ok := m.Events[calendarID]
	if !ok {
		return nil, errors.New("event not found")
	}

	event, ok := calEvents[eventID]
	if !ok {
		return nil, errors.New("event not found")
	}

	return event, nil
}

// CreateEvent creates a new event.
func (m *MockCalendarService) CreateEvent(ctx context.Context, calendarID string, event *calendar.Event, conferenceDataVersion int) (*calendar.Event, error) {
	m.recordCall("CreateEvent", map[string]any{"calendarID": calendarID, "summary": event.Summary, "conferenceDataVersion": conferenceDataVersion})
	if m.Error != nil {
		return nil, m.Error
	}

	if m.Events[calendarID] == nil {
		m.Events[calendarID] = make(map[string]*calendar.Event)
	}

	// Generate an ID
	eventIDCounter++
	id := "event-" + string(rune('a'+eventIDCounter-1)) + "123"
	event.Id = id
	event.HtmlLink = "https://calendar.google.com/event?eid=" + id
	event.Created = "2024-01-01T12:00:00Z"
	event.Updated = "2024-01-01T12:00:00Z"

	m.Events[calendarID][id] = event
	return event, nil
}

// UpdateEvent updates an existing event.
func (m *MockCalendarService) UpdateEvent(ctx context.Context, calendarID string, eventID string, event *calendar.Event) (*calendar.Event, error) {
	m.recordCall("UpdateEvent", map[string]any{"calendarID": calendarID, "eventID": eventID})
	if m.Error != nil {
		return nil, m.Error
	}

	calEvents, ok := m.Events[calendarID]
	if !ok {
		return nil, errors.New("event not found")
	}

	if _, ok := calEvents[eventID]; !ok {
		return nil, errors.New("event not found")
	}

	event.Id = eventID
	event.Updated = "2024-01-01T13:00:00Z"
	event.HtmlLink = "https://calendar.google.com/event?eid=" + eventID
	m.Events[calendarID][eventID] = event

	return event, nil
}

// DeleteEvent deletes an event.
func (m *MockCalendarService) DeleteEvent(ctx context.Context, calendarID string, eventID string) error {
	m.recordCall("DeleteEvent", map[string]any{"calendarID": calendarID, "eventID": eventID})
	if m.Error != nil {
		return m.Error
	}

	calEvents, ok := m.Events[calendarID]
	if !ok {
		return errors.New("event not found")
	}

	if _, ok := calEvents[eventID]; !ok {
		return errors.New("event not found")
	}

	delete(m.Events[calendarID], eventID)
	return nil
}

// QuickAddEvent creates an event from text.
func (m *MockCalendarService) QuickAddEvent(ctx context.Context, calendarID string, text string) (*calendar.Event, error) {
	m.recordCall("QuickAddEvent", map[string]any{"calendarID": calendarID, "text": text})
	if m.Error != nil {
		return nil, m.Error
	}

	event := &calendar.Event{
		Summary: text,
	}

	return m.CreateEvent(ctx, calendarID, event, 0)
}

// ListInstances lists instances of a recurring event.
func (m *MockCalendarService) ListInstances(ctx context.Context, calendarID string, eventID string, opts *ListInstancesOptions) (*calendar.Events, error) {
	m.recordCall("ListInstances", map[string]any{"calendarID": calendarID, "eventID": eventID, "opts": opts})
	if m.Error != nil {
		return nil, m.Error
	}

	calEvents, ok := m.Events[calendarID]
	if !ok {
		return nil, errors.New("event not found")
	}

	event, ok := calEvents[eventID]
	if !ok {
		return nil, errors.New("event not found")
	}

	// For mock purposes, return the parent event as an instance
	items := []*calendar.Event{event}

	return &calendar.Events{Items: items}, nil
}

// GetFreeBusy queries free/busy information.
func (m *MockCalendarService) GetFreeBusy(ctx context.Context, req *calendar.FreeBusyRequest) (*calendar.FreeBusyResponse, error) {
	m.recordCall("GetFreeBusy", map[string]any{"timeMin": req.TimeMin, "timeMax": req.TimeMax})
	if m.Error != nil {
		return nil, m.Error
	}

	calendars := make(map[string]calendar.FreeBusyCalendar)
	for _, item := range req.Items {
		calendars[item.Id] = calendar.FreeBusyCalendar{
			Busy: []*calendar.TimePeriod{},
		}
	}

	return &calendar.FreeBusyResponse{
		Calendars: calendars,
	}, nil
}

// eventIDCounter for generating unique event IDs in tests.
var eventIDCounter int

// ResetCalendarIDCounter resets the event ID counter for predictable IDs in tests.
func ResetCalendarIDCounter() {
	eventIDCounter = 0
}
