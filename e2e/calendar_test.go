//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/calendar"
)

// TestCalendarListCalendars verifies that the test account can list calendars.
func TestCalendarListCalendars(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := calendarDeps()

	result, err := calendar.TestableCalendarListCalendars(ctx, makeRequest(nil), deps)
	cr := requireSuccess(t, result, err)

	calendars := requireArrayField(t, cr, "calendars")
	if len(calendars) == 0 {
		t.Fatal("expected at least one calendar")
	}

	// Verify primary calendar is present
	found := false
	for _, c := range calendars {
		cal, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if cal["primary"] == true {
			found = true
			t.Logf("primary calendar: %v", cal["summary"])
			break
		}
	}
	if !found {
		t.Error("primary calendar not found")
	}

	t.Logf("found %d calendars", len(calendars))
}

// TestCalendarCreateUpdateDeleteEvent tests the full event lifecycle.
func TestCalendarCreateUpdateDeleteEvent(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := calendarDeps()

	prefix := e2ePrefix()
	summary := fmt.Sprintf("%s test event %d", prefix, time.Now().UnixMilli())

	// Create an event 1 hour from now
	startTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(2 * time.Hour).Format(time.RFC3339)

	t.Log("creating calendar event")
	result, err := calendar.TestableCalendarCreateEvent(ctx, makeRequest(map[string]any{
		"summary":    summary,
		"start_time": startTime,
		"end_time":   endTime,
		"location":   "E2E Test Location",
	}), deps)
	createResult := requireSuccess(t, result, err)

	eventID := requireStringField(t, createResult, "id")
	t.Logf("created event id=%s", eventID)

	defer func() {
		t.Log("cleanup: deleting test event")
		_, _ = calendar.TestableCalendarDeleteEvent(ctx, makeRequest(map[string]any{
			"event_id": eventID,
		}), deps)
	}()

	// Get event
	t.Log("getting event")
	result, err = calendar.TestableCalendarGetEvent(ctx, makeRequest(map[string]any{
		"event_id": eventID,
	}), deps)
	getResult := requireSuccess(t, result, err)

	gotSummary := requireStringField(t, getResult, "summary")
	if gotSummary != summary {
		t.Errorf("expected summary %q, got %q", summary, gotSummary)
	}

	// Update event
	updatedSummary := summary + " (updated)"
	t.Log("updating event")
	result, err = calendar.TestableCalendarUpdateEvent(ctx, makeRequest(map[string]any{
		"event_id": eventID,
		"summary":  updatedSummary,
	}), deps)
	requireSuccess(t, result, err)

	// Verify update
	t.Log("verifying update")
	result, err = calendar.TestableCalendarGetEvent(ctx, makeRequest(map[string]any{
		"event_id": eventID,
	}), deps)
	verifyResult := requireSuccess(t, result, err)

	gotSummary = requireStringField(t, verifyResult, "summary")
	if gotSummary != updatedSummary {
		t.Errorf("expected updated summary %q, got %q", updatedSummary, gotSummary)
	}

	// List events and verify ours appears
	t.Log("listing events")
	result, err = calendar.TestableCalendarListEvents(ctx, makeRequest(map[string]any{
		"time_min":      time.Now().Format(time.RFC3339),
		"time_max":      time.Now().Add(3 * time.Hour).Format(time.RFC3339),
		"single_events": true,
	}), deps)
	listResult := requireSuccess(t, result, err)

	events := requireArrayField(t, listResult, "events")
	found := false
	for _, e := range events {
		evt, ok := e.(map[string]any)
		if !ok {
			continue
		}
		if evt["id"] == eventID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created event %s not found in event list", eventID)
	}

	// Delete event
	t.Log("deleting event")
	result, err = calendar.TestableCalendarDeleteEvent(ctx, makeRequest(map[string]any{
		"event_id": eventID,
	}), deps)
	requireSuccess(t, result, err)
}
