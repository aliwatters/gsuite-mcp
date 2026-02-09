package calendar

import (
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/calendar/v3"
)

// Calendar API field constants for optimized responses.
// These reduce response payload size by only requesting needed fields.
const (
	// CalendarEventListFields contains fields for event listings (compact format)
	CalendarEventListFields = "nextPageToken,items(id,summary,status,start,end,location,htmlLink,hangoutLink)"
	// CalendarEventGetFields contains fields for single event retrieval (full format)
	CalendarEventGetFields = "id,summary,status,description,location,start,end,htmlLink,created,updated,creator,organizer,attendees,reminders,recurrence,recurringEventId,hangoutLink,conferenceData"
	// CalendarListFields contains fields for calendar list
	CalendarListFields = "items(id,summary,description,primary,backgroundColor,accessRole,timeZone)"
)

// extractDateFromDateTime safely extracts the date portion (YYYY-MM-DD) from a datetime string.
// Returns an error if the string is too short.
func extractDateFromDateTime(dateTime string) (string, error) {
	if len(dateTime) < 10 {
		return "", fmt.Errorf("date string too short: expected at least 10 characters (YYYY-MM-DD), got %d", len(dateTime))
	}
	return dateTime[:10], nil
}

// === Handle functions - generated via WrapHandler ===

var (
	HandleCalendarListEvents     = common.WrapHandler[CalendarService](TestableCalendarListEvents)
	HandleCalendarGetEvent       = common.WrapHandler[CalendarService](TestableCalendarGetEvent)
	HandleCalendarCreateEvent    = common.WrapHandler[CalendarService](TestableCalendarCreateEvent)
	HandleCalendarUpdateEvent    = common.WrapHandler[CalendarService](TestableCalendarUpdateEvent)
	HandleCalendarDeleteEvent    = common.WrapHandler[CalendarService](TestableCalendarDeleteEvent)
	HandleCalendarListCalendars  = common.WrapHandler[CalendarService](TestableCalendarListCalendars)
	HandleCalendarQuickAdd       = common.WrapHandler[CalendarService](TestableCalendarQuickAdd)
	HandleCalendarFreeBusy       = common.WrapHandler[CalendarService](TestableCalendarFreeBusy)
	HandleCalendarListInstances  = common.WrapHandler[CalendarService](TestableCalendarListInstances)
	HandleCalendarUpdateInstance = common.WrapHandler[CalendarService](TestableCalendarUpdateInstance)
)

// Suppressing unused import/variable warnings
var _ = calendar.Event{}
