package gmail

import (
	"context"
	"testing"

	"google.golang.org/api/gmail/v1"
)

// === gmail_modify_message tests ===

func TestGmailModifyMessage_AddLabels(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
		"add_labels": []any{"STARRED", "IMPORTANT"},
	})

	result, err := TestableGmailModifyMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyMessage was called
	if !fixtures.MockService.WasMethodCalled("ModifyMessage") {
		t.Error("expected ModifyMessage to be called")
	}

	// Verify labels were added
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if !containsLabel(modifiedMsg.LabelIds, "STARRED") {
		t.Error("expected STARRED label to be added")
	}
	if !containsLabel(modifiedMsg.LabelIds, "IMPORTANT") {
		t.Error("expected IMPORTANT label to be added")
	}
}

func TestGmailModifyMessage_RemoveLabels(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX", "UNREAD", "STARRED"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id":    "msg123",
		"remove_labels": []any{"UNREAD", "STARRED"},
	})

	result, err := TestableGmailModifyMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify labels were removed
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if containsLabel(modifiedMsg.LabelIds, "UNREAD") {
		t.Error("expected UNREAD label to be removed")
	}
	if containsLabel(modifiedMsg.LabelIds, "STARRED") {
		t.Error("expected STARRED label to be removed")
	}
	if !containsLabel(modifiedMsg.LabelIds, "INBOX") {
		t.Error("INBOX label should still be present")
	}
}

func TestGmailModifyMessage_AddAndRemove(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX", "UNREAD"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id":    "msg123",
		"add_labels":    []any{"STARRED"},
		"remove_labels": []any{"UNREAD"},
	})

	result, err := TestableGmailModifyMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if !containsLabel(modifiedMsg.LabelIds, "STARRED") {
		t.Error("expected STARRED label to be added")
	}
	if containsLabel(modifiedMsg.LabelIds, "UNREAD") {
		t.Error("expected UNREAD label to be removed")
	}
}

func TestGmailModifyMessage_NotFound(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_id": "nonexistent",
		"add_labels": []any{"STARRED"},
	})

	result, err := TestableGmailModifyMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent message")
	}
}

// === gmail_batch_modify tests ===

func TestGmailBatchModify_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread1", "Subject 1", "a@example.com", "me@example.com", "Body 1", []string{"INBOX", "UNREAD"})
	msg2 := newTestMessage("msg2", "thread2", "Subject 2", "b@example.com", "me@example.com", "Body 2", []string{"INBOX", "UNREAD"})
	fixtures.MockService.AddMessage(msg1)
	fixtures.MockService.AddMessage(msg2)

	request := makeRequest(map[string]any{
		"message_ids":   []any{"msg1", "msg2"},
		"add_labels":    []any{"STARRED"},
		"remove_labels": []any{"UNREAD"},
	})

	result, err := TestableGmailBatchModify(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify BatchModifyMessages was called
	if !fixtures.MockService.WasMethodCalled("BatchModifyMessages") {
		t.Error("expected BatchModifyMessages to be called")
	}
}

func TestGmailBatchModify_EmptyIDs(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_ids": []any{},
		"add_labels":  []any{"STARRED"},
	})

	result, err := TestableGmailBatchModify(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty message_ids")
	}
}

// === gmail_trash tests ===

func TestGmailTrash_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailTrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify TrashMessage was called
	if !fixtures.MockService.WasMethodCalled("TrashMessage") {
		t.Error("expected TrashMessage to be called")
	}

	// Verify TRASH label was added
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if !containsLabel(modifiedMsg.LabelIds, "TRASH") {
		t.Error("expected TRASH label to be added")
	}
}

func TestGmailTrash_NotFound(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_id": "nonexistent",
	})

	result, err := TestableGmailTrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent message")
	}
}

// === gmail_untrash tests ===

func TestGmailUntrash_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"TRASH"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailUntrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UntrashMessage was called
	if !fixtures.MockService.WasMethodCalled("UntrashMessage") {
		t.Error("expected UntrashMessage to be called")
	}

	// Verify TRASH label was removed and INBOX added
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if containsLabel(modifiedMsg.LabelIds, "TRASH") {
		t.Error("expected TRASH label to be removed")
	}
	if !containsLabel(modifiedMsg.LabelIds, "INBOX") {
		t.Error("expected INBOX label to be added")
	}
}

// === gmail_archive tests ===

func TestGmailArchive_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX", "UNREAD"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailArchive(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyMessage was called (archive uses modify)
	if !fixtures.MockService.WasMethodCalled("ModifyMessage") {
		t.Error("expected ModifyMessage to be called")
	}

	// Verify INBOX label was removed
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if containsLabel(modifiedMsg.LabelIds, "INBOX") {
		t.Error("expected INBOX label to be removed")
	}
}

// === gmail_mark_read tests ===

func TestGmailMarkRead_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX", "UNREAD"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailMarkRead(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UNREAD label was removed
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if containsLabel(modifiedMsg.LabelIds, "UNREAD") {
		t.Error("expected UNREAD label to be removed")
	}
}

// === gmail_mark_unread tests ===

func TestGmailMarkUnread_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailMarkUnread(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UNREAD label was added
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if !containsLabel(modifiedMsg.LabelIds, "UNREAD") {
		t.Error("expected UNREAD label to be added")
	}
}

// === gmail_star tests ===

func TestGmailStar_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailStar(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify STARRED label was added
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if !containsLabel(modifiedMsg.LabelIds, "STARRED") {
		t.Error("expected STARRED label to be added")
	}
}

// === gmail_unstar tests ===

func TestGmailUnstar_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Body", []string{"INBOX", "STARRED"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailUnstar(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify STARRED label was removed
	modifiedMsg := fixtures.MockService.Messages["msg123"]
	if containsLabel(modifiedMsg.LabelIds, "STARRED") {
		t.Error("expected STARRED label to be removed")
	}
}

// === gmail_batch_archive tests ===

func TestGmailBatchArchive_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread1", "Subject 1", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	msg2 := newTestMessage("msg2", "thread2", "Subject 2", "b@example.com", "me@example.com", "Body 2", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg1)
	fixtures.MockService.AddMessage(msg2)

	request := makeRequest(map[string]any{
		"message_ids": []any{"msg1", "msg2"},
	})

	result, err := TestableGmailBatchArchive(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify BatchModifyMessages was called
	if !fixtures.MockService.WasMethodCalled("BatchModifyMessages") {
		t.Error("expected BatchModifyMessages to be called")
	}

	// Verify response
	response := extractResponse(t, result)
	if response["archived_count"] != float64(2) {
		t.Errorf("expected archived_count=2, got %v", response["archived_count"])
	}
}

// === gmail_batch_trash tests ===

func TestGmailBatchTrash_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread1", "Subject 1", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	msg2 := newTestMessage("msg2", "thread2", "Subject 2", "b@example.com", "me@example.com", "Body 2", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg1)
	fixtures.MockService.AddMessage(msg2)

	request := makeRequest(map[string]any{
		"message_ids": []any{"msg1", "msg2"},
	})

	result, err := TestableGmailBatchTrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify BatchModifyMessages was called
	if !fixtures.MockService.WasMethodCalled("BatchModifyMessages") {
		t.Error("expected BatchModifyMessages to be called")
	}

	// Verify response
	response := extractResponse(t, result)
	if response["trashed_count"] != float64(2) {
		t.Errorf("expected trashed_count=2, got %v", response["trashed_count"])
	}
}

// === Thread operation tests ===

func TestGmailThreadArchive_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread123", "Subject", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	msg2 := newTestMessage("msg2", "thread123", "Re: Subject", "me@example.com", "a@example.com", "Body 2", []string{"SENT"})
	thread := &gmail.Thread{Id: "thread123", Messages: []*gmail.Message{msg1, msg2}}
	fixtures.MockService.AddThread(thread)

	request := makeRequest(map[string]any{
		"thread_id": "thread123",
	})

	result, err := TestableGmailThreadArchive(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyThread was called
	if !fixtures.MockService.WasMethodCalled("ModifyThread") {
		t.Error("expected ModifyThread to be called")
	}
}

func TestGmailThreadTrash_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread123", "Subject", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	thread := &gmail.Thread{Id: "thread123", Messages: []*gmail.Message{msg1}}
	fixtures.MockService.AddThread(thread)

	request := makeRequest(map[string]any{
		"thread_id": "thread123",
	})

	result, err := TestableGmailThreadTrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify TrashThread was called
	if !fixtures.MockService.WasMethodCalled("TrashThread") {
		t.Error("expected TrashThread to be called")
	}
}

func TestGmailThreadUntrash_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread123", "Subject", "a@example.com", "me@example.com", "Body 1", []string{"TRASH"})
	thread := &gmail.Thread{Id: "thread123", Messages: []*gmail.Message{msg1}}
	fixtures.MockService.AddThread(thread)

	request := makeRequest(map[string]any{
		"thread_id": "thread123",
	})

	result, err := TestableGmailThreadUntrash(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UntrashThread was called
	if !fixtures.MockService.WasMethodCalled("UntrashThread") {
		t.Error("expected UntrashThread to be called")
	}
}

func TestGmailModifyThread_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread123", "Subject", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	thread := &gmail.Thread{Id: "thread123", Messages: []*gmail.Message{msg1}}
	fixtures.MockService.AddThread(thread)

	request := makeRequest(map[string]any{
		"thread_id":  "thread123",
		"add_labels": []any{"IMPORTANT"},
	})

	result, err := TestableGmailModifyThread(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyThread was called
	if !fixtures.MockService.WasMethodCalled("ModifyThread") {
		t.Error("expected ModifyThread to be called")
	}
}

// === Error handling tests ===

func TestGmailModifyMessage_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"message_id": "msg123",
		"add_labels": []any{"STARRED"},
	})

	result, err := TestableGmailModifyMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

func TestGmailBatchModify_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"message_ids": []any{"msg1", "msg2"},
		"add_labels":  []any{"STARRED"},
	})

	result, err := TestableGmailBatchModify(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

// Helper function to check if a label exists in slice
func containsLabel(labels []string, target string) bool {
	for _, l := range labels {
		if l == target {
			return true
		}
	}
	return false
}
