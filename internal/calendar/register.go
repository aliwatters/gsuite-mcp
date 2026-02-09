package calendar

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Calendar tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Calendar Core (Phase 1) ===

	// calendar_list_events - List upcoming events
	s.AddTool(mcp.NewTool("calendar_list_events",
		mcp.WithDescription("List upcoming calendar events with optional date/time filtering."),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary' for main calendar)")),
		mcp.WithString("time_min", mcp.Description("Start of time range (RFC3339, e.g., 2024-01-15T00:00:00Z). Defaults to now.")),
		mcp.WithString("time_max", mcp.Description("End of time range (RFC3339)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum events to return (1-250, default 25)")),
		common.WithPageToken(),
		mcp.WithString("query", mcp.Description("Free text search query")),
		mcp.WithBoolean("single_events", mcp.Description("Expand recurring events into instances (default: false)")),
		common.WithAccountParam(),
	), HandleCalendarListEvents)

	// calendar_get_event - Get full event details
	s.AddTool(mcp.NewTool("calendar_get_event",
		mcp.WithDescription("Get full details for a calendar event."),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("Calendar event ID")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarGetEvent)

	// calendar_create_event - Create new calendar event
	s.AddTool(mcp.NewTool("calendar_create_event",
		mcp.WithDescription("Create a new calendar event."),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Event title")),
		mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time (RFC3339, e.g., 2024-01-15T09:00:00-08:00)")),
		mcp.WithString("end_time", mcp.Description("End time (RFC3339). Defaults to 1 hour after start.")),
		mcp.WithString("description", mcp.Description("Event description")),
		mcp.WithString("location", mcp.Description("Event location")),
		mcp.WithString("timezone", mcp.Description("Timezone (e.g., America/Los_Angeles)")),
		mcp.WithBoolean("all_day", mcp.Description("Create all-day event (use YYYY-MM-DD for start/end)")),
		mcp.WithBoolean("add_conferencing", mcp.Description("Add Google Meet video conferencing to the event")),
		mcp.WithArray("attendees", mcp.Description("List of attendee email addresses")),
		mcp.WithArray("reminders", mcp.Description("List of reminder times in minutes before event")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarCreateEvent)

	// calendar_update_event - Modify existing event
	s.AddTool(mcp.NewTool("calendar_update_event",
		mcp.WithDescription("Update an existing calendar event. Only provided fields are updated."),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("Calendar event ID")),
		mcp.WithString("summary", mcp.Description("Event title")),
		mcp.WithString("start_time", mcp.Description("Start time (RFC3339)")),
		mcp.WithString("end_time", mcp.Description("End time (RFC3339)")),
		mcp.WithString("description", mcp.Description("Event description")),
		mcp.WithString("location", mcp.Description("Event location")),
		mcp.WithArray("attendees", mcp.Description("List of attendee email addresses (replaces existing)")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarUpdateEvent)

	// calendar_delete_event - Remove calendar event
	s.AddTool(mcp.NewTool("calendar_delete_event",
		mcp.WithDescription("Delete a calendar event."),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("Calendar event ID")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarDeleteEvent)

	// === Calendar Extended (Phase 2) ===

	// calendar_list_calendars - List all calendars
	s.AddTool(mcp.NewTool("calendar_list_calendars",
		mcp.WithDescription("List all calendars the user has access to."),
		common.WithAccountParam(),
	), HandleCalendarListCalendars)

	// calendar_quick_add - Create event from natural language
	s.AddTool(mcp.NewTool("calendar_quick_add",
		mcp.WithDescription("Create a calendar event from a natural language string (e.g., 'Lunch with Bob tomorrow at noon')."),
		mcp.WithString("text", mcp.Required(), mcp.Description("Natural language event description")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarQuickAdd)

	// calendar_free_busy - Query free/busy information
	s.AddTool(mcp.NewTool("calendar_free_busy",
		mcp.WithDescription("Query free/busy information for one or more calendars."),
		mcp.WithString("time_min", mcp.Required(), mcp.Description("Start of time range (RFC3339)")),
		mcp.WithString("time_max", mcp.Required(), mcp.Description("End of time range (RFC3339)")),
		mcp.WithArray("calendar_ids", mcp.Description("Calendar IDs to query (default: ['primary'])")),
		common.WithAccountParam(),
	), HandleCalendarFreeBusy)

	// calendar_list_instances - List recurring event instances
	s.AddTool(mcp.NewTool("calendar_list_instances",
		mcp.WithDescription("List instances of a recurring event."),
		mcp.WithString("event_id", mcp.Required(), mcp.Description("Recurring event ID")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		mcp.WithString("time_min", mcp.Description("Start of time range (RFC3339)")),
		mcp.WithString("time_max", mcp.Description("End of time range (RFC3339)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum instances to return (1-250, default 25)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleCalendarListInstances)

	// calendar_update_instance - Update single recurring event instance
	s.AddTool(mcp.NewTool("calendar_update_instance",
		mcp.WithDescription("Update a single instance of a recurring event."),
		mcp.WithString("instance_id", mcp.Required(), mcp.Description("Instance ID (from calendar_list_instances)")),
		mcp.WithString("summary", mcp.Description("Event title")),
		mcp.WithString("start_time", mcp.Description("Start time (RFC3339)")),
		mcp.WithString("end_time", mcp.Description("End time (RFC3339)")),
		mcp.WithString("description", mcp.Description("Event description")),
		mcp.WithString("location", mcp.Description("Event location")),
		mcp.WithString("calendar_id", mcp.Description("Calendar ID (default: 'primary')")),
		common.WithAccountParam(),
	), HandleCalendarUpdateInstance)
}
