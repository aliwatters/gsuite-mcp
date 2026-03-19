package calendar

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/calendar/v3"
)

// TestableCalendarCreateFocusTime creates a Focus Time event with auto-decline.
func TestableCalendarCreateFocusTime(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	summary := common.ParseStringArg(request.Params.Arguments, "summary", "Focus Time")
	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	event := &calendar.Event{
		Summary:   summary,
		EventType: "focusTime",
	}

	if desc := common.ParseStringArg(request.Params.Arguments, "description", ""); desc != "" {
		event.Description = desc
	}

	// Set start/end times
	if errResult := setNewEventTimes(event, request.Params.Arguments); errResult != nil {
		return errResult, nil
	}

	// Focus Time uses FocusTimeProperties for auto-decline
	autoDecline := common.ParseBoolArg(request.Params.Arguments, "auto_decline", true)
	declineMessage := common.ParseStringArg(request.Params.Arguments, "decline_message", "Declined because I am in focus time.")
	chatStatus := common.ParseStringArg(request.Params.Arguments, "chat_status", "doNotDisturb")

	event.FocusTimeProperties = &calendar.EventFocusTimeProperties{
		AutoDeclineMode: boolToDeclineMode(autoDecline),
		DeclineMessage:  declineMessage,
		ChatStatus:      chatStatus,
	}

	// Recurrence
	if rrule := common.ParseStringArg(request.Params.Arguments, "recurrence", ""); rrule != "" {
		event.Recurrence = []string{rrule}
	}

	created, err := srv.CreateEvent(ctx, calendarID, event, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(created)
	result["html_link"] = created.HtmlLink
	result["event_type"] = "focusTime"

	return common.MarshalToolResult(result)
}

// TestableCalendarCreateOutOfOffice creates an Out of Office event.
func TestableCalendarCreateOutOfOffice(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	summary := common.ParseStringArg(request.Params.Arguments, "summary", "Out of office")
	calendarID := common.ParseStringArg(request.Params.Arguments, "calendar_id", common.DefaultCalendarID)

	event := &calendar.Event{
		Summary:   summary,
		EventType: "outOfOffice",
	}

	if desc := common.ParseStringArg(request.Params.Arguments, "description", ""); desc != "" {
		event.Description = desc
	}

	// Set start/end times
	if errResult := setNewEventTimes(event, request.Params.Arguments); errResult != nil {
		return errResult, nil
	}

	// Out of Office uses OutOfOfficeProperties for auto-decline
	autoDecline := common.ParseBoolArg(request.Params.Arguments, "auto_decline", true)
	declineMessage := common.ParseStringArg(request.Params.Arguments, "decline_message", "I am out of office and unable to attend.")

	event.OutOfOfficeProperties = &calendar.EventOutOfOfficeProperties{
		AutoDeclineMode: boolToDeclineMode(autoDecline),
		DeclineMessage:  declineMessage,
	}

	// Recurrence
	if rrule := common.ParseStringArg(request.Params.Arguments, "recurrence", ""); rrule != "" {
		event.Recurrence = []string{rrule}
	}

	created, err := srv.CreateEvent(ctx, calendarID, event, 0)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(created)
	result["html_link"] = created.HtmlLink
	result["event_type"] = "outOfOffice"

	return common.MarshalToolResult(result)
}

// boolToDeclineMode converts a boolean auto_decline flag to the Google Calendar API decline mode string.
func boolToDeclineMode(autoDecline bool) string {
	if autoDecline {
		return "declineAllConflictingInvitations"
	}
	return "declineNone"
}
