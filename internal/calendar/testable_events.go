package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/calendar/v3"
)

// TestableCalendarListEvents is the testable version of handleCalendarListEvents.
func TestableCalendarListEvents(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", "primary")

	opts := &ListEventsOptions{
		Fields: CalendarEventListFields,
	}

	// Default to showing future events if no time range specified
	if timeMin := common.ParseStringArg(request.Params.Arguments, "time_min", ""); timeMin != "" {
		opts.TimeMin = timeMin
	} else {
		// Default to now
		opts.TimeMin = time.Now().Format(time.RFC3339)
	}

	if timeMax := common.ParseStringArg(request.Params.Arguments, "time_max", ""); timeMax != "" {
		opts.TimeMax = timeMax
	}

	opts.MaxResults = common.ParseMaxResults(request.Params.Arguments, common.CalendarDefaultMaxResults, common.CalendarMaxResultsLimit)

	if pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", ""); pageToken != "" {
		opts.PageToken = pageToken
	}

	if query := common.ParseStringArg(request.Params.Arguments, "query", ""); query != "" {
		opts.Query = query
	}

	// Expand recurring events into individual instances
	if common.ParseBoolArg(request.Params.Arguments, "single_events", false) {
		opts.SingleEvents = true
		opts.OrderBy = "startTime"
	}

	resp, err := srv.ListEvents(ctx, calendarID, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	events := make([]map[string]any, 0, len(resp.Items))
	for _, event := range resp.Items {
		events = append(events, formatEvent(event))
	}

	result := map[string]any{
		"events":          events,
		"count":           len(events),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarGetEvent is the testable version of handleCalendarGetEvent.
func TestableCalendarGetEvent(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	eventID := common.ParseStringArg(request.Params.Arguments, "event_id", "")
	if eventID == "" {
		return mcp.NewToolResultError("event_id parameter is required"), nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	event, err := srv.GetEvent(ctx, calendarID, eventID, CalendarEventGetFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEventFull(event)

	return common.MarshalToolResult(result)
}

// TestableCalendarCreateEvent is the testable version of handleCalendarCreateEvent.
func TestableCalendarCreateEvent(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	summary := common.ParseStringArg(request.Params.Arguments, "summary", "")
	if summary == "" {
		return mcp.NewToolResultError("summary parameter is required"), nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	event := &calendar.Event{
		Summary: summary,
	}

	if desc := common.ParseStringArg(request.Params.Arguments, "description", ""); desc != "" {
		event.Description = desc
	}

	if loc := common.ParseStringArg(request.Params.Arguments, "location", ""); loc != "" {
		event.Location = loc
	}

	// Start time - required
	startTime := common.ParseStringArg(request.Params.Arguments, "start_time", "")
	if startTime == "" {
		return mcp.NewToolResultError("start_time parameter is required (RFC3339 format, e.g., 2024-01-15T09:00:00-08:00)"), nil
	}

	// Check if all-day event (date only, no time component)
	allDay := common.ParseBoolArg(request.Params.Arguments, "all_day", false)

	if allDay {
		// All-day events use Date field (YYYY-MM-DD format)
		startDate, err := extractDateFromDateTime(startTime)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time for all-day event: %v", err)), nil
		}
		event.Start = &calendar.EventDateTime{Date: startDate}

		endTime := common.ParseStringArg(request.Params.Arguments, "end_time", "")
		if endTime != "" {
			endDate, err := extractDateFromDateTime(endTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time for all-day event: %v", err)), nil
			}
			event.End = &calendar.EventDateTime{Date: endDate}
		} else {
			// Default to same day
			event.End = &calendar.EventDateTime{Date: startDate}
		}
	} else {
		// Timed events use DateTime field
		event.Start = &calendar.EventDateTime{DateTime: startTime}
		endTime := common.ParseStringArg(request.Params.Arguments, "end_time", "")
		if endTime != "" {
			event.End = &calendar.EventDateTime{DateTime: endTime}
		} else {
			// Default to 1 hour duration
			t, err := time.Parse(time.RFC3339, startTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time format: %v", err)), nil
			}
			event.End = &calendar.EventDateTime{DateTime: t.Add(time.Hour).Format(time.RFC3339)}
		}
	}

	// Time zone
	if tz := common.ParseStringArg(request.Params.Arguments, "timezone", ""); tz != "" {
		event.Start.TimeZone = tz
		event.End.TimeZone = tz
	}

	// Attendees
	if attendeesRaw, ok := request.Params.Arguments["attendees"].([]any); ok {
		attendees := make([]*calendar.EventAttendee, 0, len(attendeesRaw))
		for _, a := range attendeesRaw {
			if email, ok := a.(string); ok && email != "" {
				attendees = append(attendees, &calendar.EventAttendee{Email: email})
			}
		}
		if len(attendees) > 0 {
			event.Attendees = attendees
		}
	}

	// Reminders
	if reminders, ok := request.Params.Arguments["reminders"].([]any); ok && len(reminders) > 0 {
		overrides := make([]*calendar.EventReminder, 0, len(reminders))
		for _, r := range reminders {
			if minutes, ok := r.(float64); ok {
				overrides = append(overrides, &calendar.EventReminder{
					Method:  "popup",
					Minutes: int64(minutes),
				})
			}
		}
		if len(overrides) > 0 {
			event.Reminders = &calendar.EventReminders{
				UseDefault: false,
				Overrides:  overrides,
			}
		}
	}

	// Google Meet conferencing
	addConferencing := common.ParseBoolArg(request.Params.Arguments, "add_conferencing", false)

	if addConferencing {
		// Generate deterministic request ID for idempotency
		hashInput := fmt.Sprintf("%s|%s|%s", calendarID, startTime, summary)
		hash := sha256.Sum256([]byte(hashInput))
		requestID := hex.EncodeToString(hash[:16])

		event.ConferenceData = &calendar.ConferenceData{
			CreateRequest: &calendar.CreateConferenceRequest{
				RequestId: requestID,
				ConferenceSolutionKey: &calendar.ConferenceSolutionKey{
					Type: "hangoutsMeet",
				},
			},
		}
	}

	confVersion := 0
	if addConferencing {
		confVersion = 1
	}

	created, err := srv.CreateEvent(ctx, calendarID, event, confVersion)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(created)
	result["html_link"] = created.HtmlLink

	// Include Meet link if conferencing was added
	if created.HangoutLink != "" {
		result["meet_link"] = created.HangoutLink
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarUpdateEvent is the testable version of handleCalendarUpdateEvent.
func TestableCalendarUpdateEvent(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	eventID := common.ParseStringArg(request.Params.Arguments, "event_id", "")
	if eventID == "" {
		return mcp.NewToolResultError("event_id parameter is required"), nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	// First, get the existing event
	event, err := srv.GetEvent(ctx, calendarID, eventID, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get event: %v", err)), nil
	}

	// Update fields that are provided
	if summary := common.ParseStringArg(request.Params.Arguments, "summary", ""); summary != "" {
		event.Summary = summary
	}
	if val, ok := request.Params.Arguments["description"].(string); ok {
		event.Description = val
	}
	if val, ok := request.Params.Arguments["location"].(string); ok {
		event.Location = val
	}

	// Update times if provided
	if startTime := common.ParseStringArg(request.Params.Arguments, "start_time", ""); startTime != "" {
		if event.Start.Date != "" {
			// All-day event
			startDate, err := extractDateFromDateTime(startTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time for all-day event: %v", err)), nil
			}
			event.Start = &calendar.EventDateTime{Date: startDate}
		} else {
			event.Start = &calendar.EventDateTime{DateTime: startTime}
		}
	}
	if endTime := common.ParseStringArg(request.Params.Arguments, "end_time", ""); endTime != "" {
		if event.End.Date != "" {
			// All-day event
			endDate, err := extractDateFromDateTime(endTime)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time for all-day event: %v", err)), nil
			}
			event.End = &calendar.EventDateTime{Date: endDate}
		} else {
			event.End = &calendar.EventDateTime{DateTime: endTime}
		}
	}

	// Update attendees if provided
	if attendeesRaw, ok := request.Params.Arguments["attendees"].([]any); ok {
		attendees := make([]*calendar.EventAttendee, 0, len(attendeesRaw))
		for _, a := range attendeesRaw {
			if email, ok := a.(string); ok && email != "" {
				attendees = append(attendees, &calendar.EventAttendee{Email: email})
			}
		}
		event.Attendees = attendees
	}

	updated, err := srv.UpdateEvent(ctx, calendarID, eventID, event)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(updated)
	result["html_link"] = updated.HtmlLink

	return common.MarshalToolResult(result)
}

// TestableCalendarDeleteEvent is the testable version of handleCalendarDeleteEvent.
func TestableCalendarDeleteEvent(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	eventID := common.ParseStringArg(request.Params.Arguments, "event_id", "")
	if eventID == "" {
		return mcp.NewToolResultError("event_id parameter is required"), nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	err := srv.DeleteEvent(ctx, calendarID, eventID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"event_id":    eventID,
		"calendar_id": calendarID,
		"message":     "Event deleted successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarQuickAdd creates an event from natural language string.
func TestableCalendarQuickAdd(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	text := common.ParseStringArg(request.Params.Arguments, "text", "")
	if text == "" {
		return mcp.NewToolResultError("text parameter is required"), nil
	}

	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	event, err := srv.QuickAddEvent(ctx, calendarID, text)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(event)
	result["html_link"] = event.HtmlLink

	return common.MarshalToolResult(result)
}
