package calendar

import (
	"context"
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

	// Filter by event types (e.g., "focusTime", "outOfOffice")
	if eventTypesRaw, ok := request.Params.Arguments["event_types"].([]any); ok && len(eventTypesRaw) > 0 {
		eventTypes := make([]string, 0, len(eventTypesRaw))
		for _, et := range eventTypesRaw {
			if etStr, ok := et.(string); ok && etStr != "" {
				eventTypes = append(eventTypes, etStr)
			}
		}
		if len(eventTypes) > 0 {
			opts.EventTypes = eventTypes
		}
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

	eventID, errResult := common.RequireStringArg(request.Params.Arguments, "event_id")
	if errResult != nil {
		return errResult, nil
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

	summary, errResult := common.RequireStringArg(request.Params.Arguments, "summary")
	if errResult != nil {
		return errResult, nil
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

	// Set start/end times (required start_time, optional end_time/all_day/timezone)
	if errResult := setNewEventTimes(event, request.Params.Arguments); errResult != nil {
		return errResult, nil
	}

	// Attendees
	if attendees := parseAttendees(request.Params.Arguments); attendees != nil {
		event.Attendees = attendees
	}

	// Reminders
	if reminders := parseReminders(request.Params.Arguments); reminders != nil {
		event.Reminders = reminders
	}

	// Google Meet conferencing
	addConferencing := common.ParseBoolArg(request.Params.Arguments, "add_conferencing", false)

	if addConferencing {
		startTime := common.ParseStringArg(request.Params.Arguments, "start_time", "")
		event.ConferenceData = buildConferenceData(calendarID, startTime, summary)
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

	eventID, errResult := common.RequireStringArg(request.Params.Arguments, "event_id")
	if errResult != nil {
		return errResult, nil
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
	if errResult := updateEventTimes(event, request.Params.Arguments); errResult != nil {
		return errResult, nil
	}

	// Update attendees if provided
	if attendees := parseAttendees(request.Params.Arguments); attendees != nil {
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

	eventID, errResult := common.RequireStringArg(request.Params.Arguments, "event_id")
	if errResult != nil {
		return errResult, nil
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

	text, errResult := common.RequireStringArg(request.Params.Arguments, "text")
	if errResult != nil {
		return errResult, nil
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
