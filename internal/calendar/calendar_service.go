package calendar

import (
	"context"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

// CalendarService defines the interface for Calendar API operations.
// This interface enables dependency injection and testing with mocks.
type CalendarService interface {
	// Calendars
	ListCalendars(ctx context.Context, fields string) (*calendar.CalendarList, error)

	// Events
	ListEvents(ctx context.Context, calendarID string, opts *ListEventsOptions) (*calendar.Events, error)
	GetEvent(ctx context.Context, calendarID string, eventID string, fields string) (*calendar.Event, error)
	CreateEvent(ctx context.Context, calendarID string, event *calendar.Event, conferenceDataVersion int) (*calendar.Event, error)
	UpdateEvent(ctx context.Context, calendarID string, eventID string, event *calendar.Event) (*calendar.Event, error)
	DeleteEvent(ctx context.Context, calendarID string, eventID string) error
	QuickAddEvent(ctx context.Context, calendarID string, text string) (*calendar.Event, error)

	// Recurring Events
	ListInstances(ctx context.Context, calendarID string, eventID string, opts *ListInstancesOptions) (*calendar.Events, error)

	// FreeBusy
	GetFreeBusy(ctx context.Context, req *calendar.FreeBusyRequest) (*calendar.FreeBusyResponse, error)
}

// ListEventsOptions contains optional parameters for listing events.
type ListEventsOptions struct {
	MaxResults   int64
	PageToken    string
	TimeMin      string // RFC3339 timestamp
	TimeMax      string // RFC3339 timestamp
	SingleEvents bool   // Expand recurring events
	OrderBy      string // "startTime" or "updated"
	Query        string // Free text search
	ShowDeleted  bool
	UpdatedMin   string // RFC3339 timestamp
	Fields       string // Selector specifying which fields to include in a partial response
}

// ListInstancesOptions contains optional parameters for listing recurring event instances.
type ListInstancesOptions struct {
	MaxResults int64
	PageToken  string
	TimeMin    string // RFC3339 timestamp
	TimeMax    string // RFC3339 timestamp
	Fields     string // Selector specifying which fields to include in a partial response
}

// RealCalendarService wraps the Calendar API client and implements CalendarService.
type RealCalendarService struct {
	service *calendar.Service
}

// NewRealCalendarService creates a new RealCalendarService wrapping the given Calendar API service.
func NewRealCalendarService(service *calendar.Service) *RealCalendarService {
	return &RealCalendarService{service: service}
}

// ListCalendars lists all calendars the user has access to.
func (s *RealCalendarService) ListCalendars(ctx context.Context, fields string) (*calendar.CalendarList, error) {
	call := s.service.CalendarList.List().Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// ListEvents lists events from a calendar with optional filters.
func (s *RealCalendarService) ListEvents(ctx context.Context, calendarID string, opts *ListEventsOptions) (*calendar.Events, error) {
	call := s.service.Events.List(calendarID).Context(ctx)

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.TimeMin != "" {
			call = call.TimeMin(opts.TimeMin)
		}
		if opts.TimeMax != "" {
			call = call.TimeMax(opts.TimeMax)
		}
		if opts.SingleEvents {
			call = call.SingleEvents(true)
		}
		if opts.OrderBy != "" {
			call = call.OrderBy(opts.OrderBy)
		}
		if opts.Query != "" {
			call = call.Q(opts.Query)
		}
		if opts.ShowDeleted {
			call = call.ShowDeleted(true)
		}
		if opts.UpdatedMin != "" {
			call = call.UpdatedMin(opts.UpdatedMin)
		}
		if opts.Fields != "" {
			call = call.Fields(googleapi.Field(opts.Fields))
		}
	}

	return call.Do()
}

// GetEvent retrieves a single event by ID.
func (s *RealCalendarService) GetEvent(ctx context.Context, calendarID string, eventID string, fields string) (*calendar.Event, error) {
	call := s.service.Events.Get(calendarID, eventID).Context(ctx)
	if fields != "" {
		call = call.Fields(googleapi.Field(fields))
	}
	return call.Do()
}

// CreateEvent creates a new calendar event.
func (s *RealCalendarService) CreateEvent(ctx context.Context, calendarID string, event *calendar.Event, conferenceDataVersion int) (*calendar.Event, error) {
	call := s.service.Events.Insert(calendarID, event).Context(ctx)
	if conferenceDataVersion > 0 {
		call = call.ConferenceDataVersion(int64(conferenceDataVersion))
	}
	return call.Do()
}

// UpdateEvent updates an existing calendar event.
func (s *RealCalendarService) UpdateEvent(ctx context.Context, calendarID string, eventID string, event *calendar.Event) (*calendar.Event, error) {
	return s.service.Events.Update(calendarID, eventID, event).Context(ctx).Do()
}

// DeleteEvent deletes a calendar event.
func (s *RealCalendarService) DeleteEvent(ctx context.Context, calendarID string, eventID string) error {
	return s.service.Events.Delete(calendarID, eventID).Context(ctx).Do()
}

// QuickAddEvent creates an event from a natural language string.
func (s *RealCalendarService) QuickAddEvent(ctx context.Context, calendarID string, text string) (*calendar.Event, error) {
	return s.service.Events.QuickAdd(calendarID, text).Context(ctx).Do()
}

// ListInstances lists instances of a recurring event.
func (s *RealCalendarService) ListInstances(ctx context.Context, calendarID string, eventID string, opts *ListInstancesOptions) (*calendar.Events, error) {
	call := s.service.Events.Instances(calendarID, eventID).Context(ctx)

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.TimeMin != "" {
			call = call.TimeMin(opts.TimeMin)
		}
		if opts.TimeMax != "" {
			call = call.TimeMax(opts.TimeMax)
		}
		if opts.Fields != "" {
			call = call.Fields(googleapi.Field(opts.Fields))
		}
	}

	return call.Do()
}

// GetFreeBusy queries free/busy information for a set of calendars.
func (s *RealCalendarService) GetFreeBusy(ctx context.Context, req *calendar.FreeBusyRequest) (*calendar.FreeBusyResponse, error) {
	return s.service.Freebusy.Query(req).Context(ctx).Do()
}
