package calendar

import (
	"google.golang.org/api/calendar/v3"
)

// formatEvent extracts useful fields from a Calendar event (compact format)
func formatEvent(event *calendar.Event) map[string]any {
	result := map[string]any{
		"id":      event.Id,
		"summary": event.Summary,
		"status":  event.Status,
	}

	// Handle start time (can be date or dateTime)
	if event.Start != nil {
		if event.Start.DateTime != "" {
			result["start"] = event.Start.DateTime
		} else if event.Start.Date != "" {
			result["start"] = event.Start.Date
			result["all_day"] = true
		}
	}

	// Handle end time
	if event.End != nil {
		if event.End.DateTime != "" {
			result["end"] = event.End.DateTime
		} else if event.End.Date != "" {
			result["end"] = event.End.Date
		}
	}

	if event.Location != "" {
		result["location"] = event.Location
	}

	return result
}

// formatEventFull extracts all useful fields from a Calendar event
func formatEventFull(event *calendar.Event) map[string]any {
	result := formatEvent(event)

	// Add additional details for full view
	result["html_link"] = event.HtmlLink
	result["created"] = event.Created
	result["updated"] = event.Updated

	if event.Description != "" {
		result["description"] = event.Description
	}

	if event.Creator != nil {
		result["creator"] = map[string]any{
			"email": event.Creator.Email,
			"self":  event.Creator.Self,
		}
	}

	if event.Organizer != nil {
		result["organizer"] = map[string]any{
			"email": event.Organizer.Email,
			"self":  event.Organizer.Self,
		}
	}

	if len(event.Attendees) > 0 {
		attendees := make([]map[string]any, 0, len(event.Attendees))
		for _, a := range event.Attendees {
			attendee := map[string]any{
				"email":           a.Email,
				"response_status": a.ResponseStatus,
			}
			if a.DisplayName != "" {
				attendee["display_name"] = a.DisplayName
			}
			if a.Organizer {
				attendee["organizer"] = true
			}
			if a.Self {
				attendee["self"] = true
			}
			attendees = append(attendees, attendee)
		}
		result["attendees"] = attendees
	}

	if event.Reminders != nil {
		reminders := map[string]any{
			"use_default": event.Reminders.UseDefault,
		}
		if len(event.Reminders.Overrides) > 0 {
			overrides := make([]map[string]any, 0, len(event.Reminders.Overrides))
			for _, r := range event.Reminders.Overrides {
				overrides = append(overrides, map[string]any{
					"method":  r.Method,
					"minutes": r.Minutes,
				})
			}
			reminders["overrides"] = overrides
		}
		result["reminders"] = reminders
	}

	if event.RecurringEventId != "" {
		result["recurring_event_id"] = event.RecurringEventId
	}
	if len(event.Recurrence) > 0 {
		result["recurrence"] = event.Recurrence
	}

	if event.HangoutLink != "" {
		result["hangout_link"] = event.HangoutLink
	}

	if event.ConferenceData != nil && event.ConferenceData.EntryPoints != nil {
		for _, ep := range event.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				result["video_link"] = ep.Uri
				break
			}
		}
	}

	return result
}
