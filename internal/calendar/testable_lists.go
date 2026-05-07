package calendar

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/calendar/v3"
)

// TestableCalendarListCalendars - List all calendars the user has access to
func TestableCalendarListCalendars(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	// Use field selection to reduce response payload size
	resp, err := srv.ListCalendars(ctx, CalendarListFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	calendars := make([]map[string]any, 0, len(resp.Items))
	for _, cal := range resp.Items {
		calInfo := map[string]any{
			"id":      cal.Id,
			"summary": cal.Summary,
			"primary": cal.Primary,
		}
		if cal.Description != "" {
			calInfo["description"] = cal.Description
		}
		if cal.BackgroundColor != "" {
			calInfo["background_color"] = cal.BackgroundColor
		}
		if cal.AccessRole != "" {
			calInfo["access_role"] = cal.AccessRole
		}
		if cal.TimeZone != "" {
			calInfo["timezone"] = cal.TimeZone
		}
		calendars = append(calendars, calInfo)
	}

	result := map[string]any{
		"calendars": calendars,
		"count":     len(calendars),
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarFreeBusy - Query free/busy information for calendars
func TestableCalendarFreeBusy(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	timeMin := common.ParseStringArg(request.GetArguments(), "time_min", "")
	if timeMin == "" {
		return mcp.NewToolResultError("time_min parameter is required (RFC3339 format)"), nil
	}

	timeMax := common.ParseStringArg(request.GetArguments(), "time_max", "")
	if timeMax == "" {
		return mcp.NewToolResultError("time_max parameter is required (RFC3339 format)"), nil
	}

	// Build calendar items
	var items []*calendar.FreeBusyRequestItem

	// Check for calendar_ids parameter (array)
	if calendarIDsRaw, ok := request.GetArguments()["calendar_ids"].([]any); ok && len(calendarIDsRaw) > 0 {
		for _, id := range calendarIDsRaw {
			if calID, ok := id.(string); ok && calID != "" {
				items = append(items, &calendar.FreeBusyRequestItem{Id: calID})
			}
		}
	} else {
		// Default to primary calendar
		items = append(items, &calendar.FreeBusyRequestItem{Id: common.DefaultCalendarID})
	}

	req := &calendar.FreeBusyRequest{
		TimeMin: timeMin,
		TimeMax: timeMax,
		Items:   items,
	}

	resp, err := srv.GetFreeBusy(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	// Format response
	calendars := make(map[string]any)
	for calID, calBusy := range resp.Calendars {
		busyPeriods := make([]map[string]string, 0)
		for _, period := range calBusy.Busy {
			busyPeriods = append(busyPeriods, map[string]string{
				"start": period.Start,
				"end":   period.End,
			})
		}
		calendars[calID] = map[string]any{
			"busy": busyPeriods,
		}
	}

	result := map[string]any{
		"time_min":  timeMin,
		"time_max":  timeMax,
		"calendars": calendars,
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarListInstances - List instances of a recurring event
func TestableCalendarListInstances(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	eventID, errResult := common.RequireStringArg(request.GetArguments(), "event_id")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := common.ParseStringArg(request.GetArguments(), "calendar_id", common.DefaultCalendarID)

	opts := &ListInstancesOptions{
		Fields: CalendarEventListFields,
	}

	if timeMin := common.ParseStringArg(request.GetArguments(), "time_min", ""); timeMin != "" {
		opts.TimeMin = timeMin
	}

	if timeMax := common.ParseStringArg(request.GetArguments(), "time_max", ""); timeMax != "" {
		opts.TimeMax = timeMax
	}

	opts.MaxResults = common.ParseMaxResults(request.GetArguments(), common.CalendarDefaultMaxResults, common.CalendarMaxResultsLimit)
	opts.PageToken = common.ParseStringArg(request.GetArguments(), "page_token", "")

	resp, err := srv.ListInstances(ctx, calendarID, eventID, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	instances := make([]map[string]any, 0, len(resp.Items))
	for _, event := range resp.Items {
		instances = append(instances, formatEvent(event))
	}

	result := map[string]any{
		"instances":       instances,
		"count":           len(instances),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableCalendarUpdateInstance - Update a single instance of a recurring event
func TestableCalendarUpdateInstance(ctx context.Context, request mcp.CallToolRequest, deps *CalendarHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCalendarServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	instanceID, errResult := common.RequireStringArg(request.GetArguments(), "instance_id")
	if errResult != nil {
		return errResult, nil
	}

	calendarID := common.ParseStringArg(request.GetArguments(), "calendar_id", common.DefaultCalendarID)

	// First, get the existing instance
	event, err := srv.GetEvent(ctx, calendarID, instanceID, "")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get instance: %v", err)), nil
	}

	// Update fields that are provided
	if summary := common.ParseStringArg(request.GetArguments(), "summary", ""); summary != "" {
		event.Summary = summary
	}
	if val, ok := request.GetArguments()["description"].(string); ok {
		event.Description = val
	}
	if val, ok := request.GetArguments()["location"].(string); ok {
		event.Location = val
	}

	// Update times if provided
	if errResult := updateEventTimes(event, request.GetArguments()); errResult != nil {
		return errResult, nil
	}

	updated, err := srv.UpdateEvent(ctx, calendarID, instanceID, event)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Calendar API error: %v", err)), nil
	}

	result := formatEvent(updated)
	result["html_link"] = updated.HtmlLink
	if updated.RecurringEventId != "" {
		result["recurring_event_id"] = updated.RecurringEventId
	}

	return common.MarshalToolResult(result)
}
