package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// Helper function to create a CallToolRequest from a map
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Arguments: args,
		},
	}
}

// Helper function to extract JSON response from result
func extractResponse(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}

	// mcp.Content is an interface, we need to type assert to TextContent
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return response
}

// Test fixtures
func newTestMessage(id, threadID, subject, from, to, body string, labelIDs []string) *gmail.Message {
	return &gmail.Message{
		Id:           id,
		ThreadId:     threadID,
		LabelIds:     labelIDs,
		Snippet:      body[:min(len(body), 100)],
		InternalDate: 1704067200000, // 2024-01-01 00:00:00 UTC
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: subject},
				{Name: "From", Value: from},
				{Name: "To", Value: to},
				{Name: "Date", Value: "Mon, 1 Jan 2024 00:00:00 +0000"},
			},
			MimeType: "text/plain",
			Body: &gmail.MessagePartBody{
				Data: encodeBase64(body),
			},
		},
	}
}

func newTestThread(id string, messages []*gmail.Message) *gmail.Thread {
	return &gmail.Thread{
		Id:       id,
		Messages: messages,
	}
}

func newTestLabel(id, name, labelType string, messagesTotal, messagesUnread int64) *gmail.Label {
	return &gmail.Label{
		Id:                    id,
		Name:                  name,
		Type:                  labelType,
		MessagesTotal:         messagesTotal,
		MessagesUnread:        messagesUnread,
		LabelListVisibility:   "labelShow",
		MessageListVisibility: "show",
	}
}

// encodeBase64 encodes a string to URL-safe base64
func encodeBase64(s string) string {
	// URL-safe base64 encoding without padding
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var result []byte
	for i := 0; i < len(s); i += 3 {
		var n uint32
		remaining := len(s) - i
		if remaining >= 3 {
			n = uint32(s[i])<<16 | uint32(s[i+1])<<8 | uint32(s[i+2])
			result = append(result, alphabet[n>>18&63], alphabet[n>>12&63], alphabet[n>>6&63], alphabet[n&63])
		} else if remaining == 2 {
			n = uint32(s[i])<<16 | uint32(s[i+1])<<8
			result = append(result, alphabet[n>>18&63], alphabet[n>>12&63], alphabet[n>>6&63])
		} else {
			n = uint32(s[i]) << 16
			result = append(result, alphabet[n>>18&63], alphabet[n>>12&63])
		}
	}
	return string(result)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// === gmail_search tests ===

func TestGmailSearch_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add test messages
	msg1 := newTestMessage("msg1", "thread1", "Test Subject 1", "sender@example.com", "me@example.com", "Test body 1", []string{"INBOX", "UNREAD"})
	msg2 := newTestMessage("msg2", "thread2", "Test Subject 2", "other@example.com", "me@example.com", "Test body 2", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg1)
	fixtures.MockService.AddMessage(msg2)

	request := makeRequest(map[string]any{
		"query": "is:unread",
	})

	result, err := TestableGmailSearch(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify mock was called correctly
	if !fixtures.MockService.WasMethodCalled("ListMessages") {
		t.Error("expected ListMessages to be called")
	}

	// Verify response structure
	response := extractResponse(t, result)
	messages, ok := response["messages"].([]any)
	if !ok {
		t.Fatal("expected messages array in response")
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestGmailSearch_EmptyQuery(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"query": "",
	})

	result, err := TestableGmailSearch(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty query should return error
	if !result.IsError {
		t.Error("expected error for empty query")
	}
}

func TestGmailSearch_WithPagination(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add many messages
	for i := 0; i < 30; i++ {
		msg := newTestMessage(
			fmt.Sprintf("msg%d", i),
			fmt.Sprintf("thread%d", i),
			"Subject",
			"sender@example.com",
			"me@example.com",
			"Body",
			[]string{"INBOX"},
		)
		fixtures.MockService.AddMessage(msg)
	}

	request := makeRequest(map[string]any{
		"query":       "in:inbox",
		"max_results": float64(10),
	})

	result, err := TestableGmailSearch(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify max_results was passed
	lastCall := fixtures.MockService.GetLastCall()
	if lastCall == nil {
		t.Fatal("ListMessages was not called")
	}
	if lastCall.Method != "ListMessages" {
		t.Fatalf("expected ListMessages to be last call, got %v", lastCall.Method)
	}
	// Args[1] is maxResults (Args[0] is query, Args[1] is maxResults, Args[2] is pageToken)
	if len(lastCall.Args) > 1 {
		if lastCall.Args[1] != int64(10) {
			t.Errorf("expected maxResults=10, got %v", lastCall.Args[1])
		}
	}
}

// === gmail_get_message tests ===

func TestGmailGetMessage_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Important Email", "boss@company.com", "me@example.com", "Please review the attached document.", []string{"INBOX", "IMPORTANT"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailGetMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify response contains message data
	response := extractResponse(t, result)
	if response["id"] != "msg123" {
		t.Errorf("expected id=msg123, got %v", response["id"])
	}
	// Subject is in headers map
	headers, ok := response["headers"].(map[string]any)
	if !ok {
		t.Fatal("expected headers map in response")
	}
	if headers["subject"] != "Important Email" {
		t.Errorf("expected subject='Important Email', got %v", headers["subject"])
	}
}

func TestGmailGetMessage_NotFound(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_id": "nonexistent",
	})

	result, err := TestableGmailGetMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent message")
	}
}

func TestGmailGetMessage_MissingID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{})

	result, err := TestableGmailGetMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing message_id")
	}
}

// === gmail_get_messages tests ===

func TestGmailGetMessages_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread1", "Subject 1", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	msg2 := newTestMessage("msg2", "thread2", "Subject 2", "b@example.com", "me@example.com", "Body 2", []string{"INBOX"})
	msg3 := newTestMessage("msg3", "thread3", "Subject 3", "c@example.com", "me@example.com", "Body 3", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg1)
	fixtures.MockService.AddMessage(msg2)
	fixtures.MockService.AddMessage(msg3)

	request := makeRequest(map[string]any{
		"message_ids": []any{"msg1", "msg2", "msg3"},
	})

	result, err := TestableGmailGetMessages(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify response
	response := extractResponse(t, result)
	messages, ok := response["messages"].([]any)
	if !ok {
		t.Fatal("expected messages array in response")
	}
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}
}

func TestGmailGetMessages_PartialSuccess(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread1", "Subject 1", "a@example.com", "me@example.com", "Body 1", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg1)

	request := makeRequest(map[string]any{
		"message_ids": []any{"msg1", "nonexistent"},
	})

	result, err := TestableGmailGetMessages(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should succeed with partial results
	if result.IsError {
		t.Fatalf("expected success with partial results, got error: %v", result.Content)
	}
}

func TestGmailGetMessages_TooMany(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Create array of 30 message IDs (exceeds max of 25)
	ids := make([]any, 30)
	for i := range ids {
		ids[i] = fmt.Sprintf("msg%d", i)
	}

	request := makeRequest(map[string]any{
		"message_ids": ids,
	})

	result, err := TestableGmailGetMessages(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for too many message IDs")
	}
}

// === gmail_get_thread tests ===

func TestGmailGetThread_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg1 := newTestMessage("msg1", "thread123", "Original Subject", "alice@example.com", "me@example.com", "Original message", []string{"INBOX"})
	msg2 := newTestMessage("msg2", "thread123", "Re: Original Subject", "me@example.com", "alice@example.com", "Reply message", []string{"SENT"})
	thread := newTestThread("thread123", []*gmail.Message{msg1, msg2})
	fixtures.MockService.AddThread(thread)

	request := makeRequest(map[string]any{
		"thread_id": "thread123",
	})

	result, err := TestableGmailGetThread(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify response
	response := extractResponse(t, result)
	if response["thread_id"] != "thread123" {
		t.Errorf("expected thread id=thread123, got %v", response["thread_id"])
	}
	messages, ok := response["messages"].([]any)
	if !ok {
		t.Fatal("expected messages array in response")
	}
	if len(messages) != 2 {
		t.Errorf("expected 2 messages in thread, got %d", len(messages))
	}
}

func TestGmailGetThread_NotFound(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"thread_id": "nonexistent",
	})

	result, err := TestableGmailGetThread(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent thread")
	}
}

// === gmail_send tests ===

func TestGmailSend_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"to":      "recipient@example.com",
		"subject": "Test Email",
		"body":    "This is a test email.",
	})

	result, err := TestableGmailSend(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify SendMessage was called
	if !fixtures.MockService.WasMethodCalled("SendMessage") {
		t.Error("expected SendMessage to be called")
	}

	// Verify response has message id (indicating success)
	response := extractResponse(t, result)
	if response["id"] == nil || response["id"] == "" {
		t.Errorf("expected message id in response, got %v", response["id"])
	}
}

func TestGmailSend_WithCCAndBCC(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"to":      "recipient@example.com",
		"cc":      "cc@example.com",
		"bcc":     "bcc@example.com",
		"subject": "Test Email",
		"body":    "This is a test email.",
	})

	result, err := TestableGmailSend(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestGmailSend_MissingTo(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"subject": "Test Email",
		"body":    "This is a test email.",
	})

	result, err := TestableGmailSend(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing 'to' field")
	}
}

func TestGmailSend_NoSubjectAllowed(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Gmail allows sending emails without a subject
	request := makeRequest(map[string]any{
		"to":   "recipient@example.com",
		"body": "This is a test email.",
	})

	result, err := TestableGmailSend(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should succeed - subject is optional
	if result.IsError {
		t.Errorf("expected success for email without subject, got error: %v", result.Content)
	}
}

// === gmail_reply tests ===

func TestGmailReply_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add original message to reply to
	originalMsg := newTestMessage("msg123", "thread123", "Original Subject", "sender@example.com", "me@example.com", "Original body", []string{"INBOX"})
	fixtures.MockService.AddMessage(originalMsg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
		"body":       "Thanks for your email!",
	})

	result, err := TestableGmailReply(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify GetMessage was called to fetch original
	if !fixtures.MockService.WasMethodCalled("GetMessage") {
		t.Error("expected GetMessage to be called")
	}

	// Verify SendMessage was called
	if !fixtures.MockService.WasMethodCalled("SendMessage") {
		t.Error("expected SendMessage to be called")
	}
}

func TestGmailReply_ReplyAll(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	originalMsg := newTestMessage("msg123", "thread123", "Original Subject", "sender@example.com", "me@example.com, other@example.com", "Original body", []string{"INBOX"})
	fixtures.MockService.AddMessage(originalMsg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
		"body":       "Reply to all!",
		"reply_all":  true,
	})

	result, err := TestableGmailReply(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestGmailReply_MessageNotFound(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_id": "nonexistent",
		"body":       "Reply body",
	})

	result, err := TestableGmailReply(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for nonexistent message")
	}
}

// === gmail_draft tests ===

func TestGmailDraft_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"to":      "recipient@example.com",
		"subject": "Draft Subject",
		"body":    "Draft body content",
	})

	result, err := TestableGmailDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify CreateDraft was called
	if !fixtures.MockService.WasMethodCalled("CreateDraft") {
		t.Error("expected CreateDraft to be called")
	}

	// Verify response contains draft ID
	response := extractResponse(t, result)
	if response["id"] == nil {
		t.Error("expected draft id in response")
	}
}

func TestGmailDraft_ForThread(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"to":        "recipient@example.com",
		"subject":   "Re: Original Subject",
		"body":      "Draft reply",
		"thread_id": "thread123",
	})

	result, err := TestableGmailDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

// === gmail_list_labels tests ===

func TestGmailListLabels_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add test labels
	fixtures.MockService.AddLabel(newTestLabel("INBOX", "INBOX", "system", 100, 10))
	fixtures.MockService.AddLabel(newTestLabel("SENT", "SENT", "system", 50, 0))
	fixtures.MockService.AddLabel(newTestLabel("Label_1", "Work", "user", 25, 5))

	request := makeRequest(map[string]any{})

	result, err := TestableGmailListLabels(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ListLabels was called
	if !fixtures.MockService.WasMethodCalled("ListLabels") {
		t.Error("expected ListLabels to be called")
	}

	// Verify response
	response := extractResponse(t, result)
	labels, ok := response["labels"].([]any)
	if !ok {
		t.Fatal("expected labels array in response")
	}
	if len(labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(labels))
	}
}

func TestGmailListLabels_Empty(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{})

	result, err := TestableGmailListLabels(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify response has empty labels array
	response := extractResponse(t, result)
	labels, ok := response["labels"].([]any)
	if !ok {
		t.Fatal("expected labels array in response")
	}
	if len(labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(labels))
	}
}

// === Error handling tests ===

func TestGmailSearch_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"query": "test",
	})

	result, err := TestableGmailSearch(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

func TestGmailGetMessage_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailGetMessage(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

func TestGmailSend_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"to":      "recipient@example.com",
		"subject": "Test",
		"body":    "Body",
	})

	result, err := TestableGmailSend(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}
