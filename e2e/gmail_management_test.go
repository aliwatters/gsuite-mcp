//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/gmail"
)

// TestGmailLabels tests creating, updating, listing, and deleting labels.
func TestGmailLabels(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	prefix := e2ePrefix()
	labelName := fmt.Sprintf("%s-label-%d", prefix, time.Now().UnixMilli())

	// Create label
	t.Logf("creating label %q", labelName)
	result, err := gmail.TestableGmailCreateLabel(ctx, makeRequest(map[string]any{
		"name": labelName,
	}), deps)
	createResult := requireSuccess(t, result, err)

	labelID := requireStringField(t, createResult, "id")
	t.Logf("created label id=%s", labelID)

	defer func() {
		t.Log("cleanup: deleting test label")
		_, _ = gmail.TestableGmailDeleteLabel(ctx, makeRequest(map[string]any{
			"label_id": labelID,
		}), deps)
	}()

	// Update label
	updatedName := labelName + "-updated"
	t.Logf("updating label to %q", updatedName)
	result, err = gmail.TestableGmailUpdateLabel(ctx, makeRequest(map[string]any{
		"label_id": labelID,
		"name":     updatedName,
	}), deps)
	requireSuccess(t, result, err)

	// List labels and verify ours appears
	t.Log("listing labels")
	result, err = gmail.TestableGmailListLabels(ctx, makeRequest(nil), deps)
	listResult := requireSuccess(t, result, err)

	labels := requireArrayField(t, listResult, "labels")
	found := false
	for _, l := range labels {
		label, ok := l.(map[string]any)
		if !ok {
			continue
		}
		if label["id"] == labelID {
			if label["name"] != updatedName {
				t.Errorf("expected label name %q, got %q", updatedName, label["name"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created label %s not found in label list", labelID)
	}

	// Delete label (also handled by defer, but test the operation explicitly)
	t.Log("deleting label")
	result, err = gmail.TestableGmailDeleteLabel(ctx, makeRequest(map[string]any{
		"label_id": labelID,
	}), deps)
	requireSuccess(t, result, err)
}

// TestGmailMessageOperations tests archive, trash, untrash, star, unstar, mark read/unread.
func TestGmailMessageOperations(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	prefix := e2ePrefix()
	subject := fmt.Sprintf("%s ops test %d", prefix, time.Now().UnixMilli())

	// Send a test message
	t.Log("sending test message")
	result, err := gmail.TestableGmailSend(ctx, makeRequest(map[string]any{
		"to":      testAccount,
		"subject": subject,
		"body":    "Message for testing operations. Safe to delete.",
	}), deps)
	sendResult := requireSuccess(t, result, err)

	messageID := requireStringField(t, sendResult, "id")
	threadID := requireStringField(t, sendResult, "thread_id")

	defer func() {
		t.Log("cleanup: trashing test thread")
		_, _ = gmail.TestableGmailThreadTrash(ctx, makeRequest(map[string]any{
			"thread_id": threadID,
		}), deps)
	}()

	time.Sleep(2 * time.Second)

	// Star the message
	t.Log("starring message")
	result, err = gmail.TestableGmailStar(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Unstar
	t.Log("unstarring message")
	result, err = gmail.TestableGmailUnstar(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Mark read
	t.Log("marking message as read")
	result, err = gmail.TestableGmailMarkRead(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Mark unread
	t.Log("marking message as unread")
	result, err = gmail.TestableGmailMarkUnread(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Archive (remove from inbox)
	t.Log("archiving message")
	result, err = gmail.TestableGmailArchive(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Trash
	t.Log("trashing message")
	result, err = gmail.TestableGmailTrash(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	// Untrash
	t.Log("untrashing message")
	result, err = gmail.TestableGmailUntrash(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	requireSuccess(t, result, err)

	t.Log("all message operations succeeded")
}

// TestGmailDrafts tests creating, listing, updating, and deleting drafts.
func TestGmailDrafts(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	prefix := e2ePrefix()
	subject := fmt.Sprintf("%s draft test %d", prefix, time.Now().UnixMilli())

	// Create draft
	t.Log("creating draft")
	result, err := gmail.TestableGmailDraft(ctx, makeRequest(map[string]any{
		"to":      testAccount,
		"subject": subject,
		"body":    "Draft body. Safe to delete.",
	}), deps)
	createResult := requireSuccess(t, result, err)

	draftID := requireStringField(t, createResult, "id")
	t.Logf("created draft id=%s", draftID)

	defer func() {
		t.Log("cleanup: deleting test draft")
		_, _ = gmail.TestableGmailDeleteDraft(ctx, makeRequest(map[string]any{
			"draft_id": draftID,
		}), deps)
	}()

	// List drafts — verify ours is present
	t.Log("listing drafts")
	result, err = gmail.TestableGmailListDrafts(ctx, makeRequest(map[string]any{
		"max_results": float64(20),
	}), deps)
	listResult := requireSuccess(t, result, err)

	drafts := requireArrayField(t, listResult, "drafts")
	found := false
	for _, d := range drafts {
		draft, ok := d.(map[string]any)
		if !ok {
			continue
		}
		if draft["id"] == draftID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("draft %s not found in drafts list", draftID)
	}

	// Get draft
	t.Log("getting draft")
	result, err = gmail.TestableGmailGetDraft(ctx, makeRequest(map[string]any{
		"draft_id": draftID,
	}), deps)
	requireSuccess(t, result, err)

	// Update draft
	t.Log("updating draft")
	result, err = gmail.TestableGmailUpdateDraft(ctx, makeRequest(map[string]any{
		"draft_id": draftID,
		"subject":  subject + " (updated)",
		"body":     "Updated draft body.",
		"to":       testAccount,
	}), deps)
	requireSuccess(t, result, err)

	// Delete draft
	t.Log("deleting draft")
	result, err = gmail.TestableGmailDeleteDraft(ctx, makeRequest(map[string]any{
		"draft_id": draftID,
	}), deps)
	requireSuccess(t, result, err)
}

// TestGmailFilters tests creating, listing, and deleting filters.
func TestGmailFilters(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	// Create a filter matching a unique subject
	prefix := e2ePrefix()
	filterSubject := fmt.Sprintf("%s-filter-%d", prefix, time.Now().UnixMilli())

	t.Logf("creating filter for subject %q", filterSubject)
	result, err := gmail.TestableGmailCreateFilter(ctx, makeRequest(map[string]any{
		"subject": filterSubject,
	}), deps)
	createResult := requireSuccess(t, result, err)

	filterID := requireStringField(t, createResult, "id")
	t.Logf("created filter id=%s", filterID)

	defer func() {
		t.Log("cleanup: deleting test filter")
		_, _ = gmail.TestableGmailDeleteFilter(ctx, makeRequest(map[string]any{
			"filter_id": filterID,
		}), deps)
	}()

	// List filters
	t.Log("listing filters")
	result, err = gmail.TestableGmailListFilters(ctx, makeRequest(nil), deps)
	listResult := requireSuccess(t, result, err)

	filters := requireArrayField(t, listResult, "filters")
	found := false
	for _, f := range filters {
		filter, ok := f.(map[string]any)
		if !ok {
			continue
		}
		if filter["id"] == filterID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("filter %s not found in filters list", filterID)
	}

	// Delete filter
	t.Log("deleting filter")
	result, err = gmail.TestableGmailDeleteFilter(ctx, makeRequest(map[string]any{
		"filter_id": filterID,
	}), deps)
	requireSuccess(t, result, err)
}
