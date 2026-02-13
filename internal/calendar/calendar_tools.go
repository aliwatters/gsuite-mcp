package calendar

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
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

// updateEventTimes updates start/end times on an existing event from request arguments.
// It handles both all-day and timed events. Returns a tool error result if parsing fails, or nil on success.
func updateEventTimes(event *calendar.Event, args map[string]any) *mcp.CallToolResult {
	if startTime := common.ParseStringArg(args, "start_time", ""); startTime != "" {
		if event.Start.Date != "" {
			startDate, err := extractDateFromDateTime(startTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time for all-day event: %v", err))
			}
			event.Start = &calendar.EventDateTime{Date: startDate}
		} else {
			event.Start = &calendar.EventDateTime{DateTime: startTime}
		}
	}
	if endTime := common.ParseStringArg(args, "end_time", ""); endTime != "" {
		if event.End.Date != "" {
			endDate, err := extractDateFromDateTime(endTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time for all-day event: %v", err))
			}
			event.End = &calendar.EventDateTime{Date: endDate}
		} else {
			event.End = &calendar.EventDateTime{DateTime: endTime}
		}
	}
	return nil
}

// parseAttendees extracts attendee emails from request arguments and returns calendar attendees.
func parseAttendees(args map[string]any) []*calendar.EventAttendee {
	attendeesRaw, ok := args["attendees"].([]any)
	if !ok {
		return nil
	}
	attendees := make([]*calendar.EventAttendee, 0, len(attendeesRaw))
	for _, a := range attendeesRaw {
		if email, ok := a.(string); ok && email != "" {
			attendees = append(attendees, &calendar.EventAttendee{Email: email})
		}
	}
	if len(attendees) == 0 {
		return nil
	}
	return attendees
}

// parseReminders extracts reminder overrides from request arguments.
func parseReminders(args map[string]any) *calendar.EventReminders {
	reminders, ok := args["reminders"].([]any)
	if !ok || len(reminders) == 0 {
		return nil
	}
	overrides := make([]*calendar.EventReminder, 0, len(reminders))
	for _, r := range reminders {
		if minutes, ok := r.(float64); ok {
			overrides = append(overrides, &calendar.EventReminder{
				Method:  "popup",
				Minutes: int64(minutes),
			})
		}
	}
	if len(overrides) == 0 {
		return nil
	}
	return &calendar.EventReminders{
		UseDefault: false,
		Overrides:  overrides,
	}
}

// buildConferenceData creates Google Meet conference data for an event.
func buildConferenceData(calendarID, startTime, summary string) *calendar.ConferenceData {
	hashInput := fmt.Sprintf("%s|%s|%s", calendarID, startTime, summary)
	hash := sha256.Sum256([]byte(hashInput))
	requestID := hex.EncodeToString(hash[:16])

	return &calendar.ConferenceData{
		CreateRequest: &calendar.CreateConferenceRequest{
			RequestId: requestID,
			ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
				Type: "hangoutsMeet",
			},
		},
	}
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
