//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/gmail"
)

// TestGmailSendSearchReadReply tests the core Gmail flow:
// send a message to self, search for it, read it, reply to it, verify thread grouping.
// Cleans up by trashing the thread afterward.
func TestGmailSendSearchReadReply(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	prefix := e2ePrefix()
	subject := fmt.Sprintf("%s core test %d", prefix, time.Now().UnixMilli())
	body := "This is an automated E2E test message. Safe to delete."

	// Step 1: Send a message to self
	t.Log("sending test email to self")
	result, err := gmail.TestableGmailSend(ctx, makeRequest(map[string]any{
		"to":      testAccount,
		"subject": subject,
		"body":    body,
	}), deps)
	sendResult := requireSuccess(t, result, err)

	messageID := requireStringField(t, sendResult, "id")
	threadID := requireStringField(t, sendResult, "thread_id")
	t.Logf("sent message_id=%s thread_id=%s", messageID, threadID)

	// Cleanup: trash the thread when done
	defer func() {
		t.Log("cleanup: trashing test thread")
		_, _ = gmail.TestableGmailThreadTrash(ctx, makeRequest(map[string]any{
			"thread_id": threadID,
		}), deps)
	}()

	// Step 2: Wait briefly for message to appear in search index
	time.Sleep(3 * time.Second)

	// Step 3: Search for the message by subject
	t.Log("searching for test email")
	result, err = gmail.TestableGmailSearch(ctx, makeRequest(map[string]any{
		"query":       fmt.Sprintf("subject:\"%s\"", subject),
		"max_results": float64(5),
	}), deps)
	searchResult := requireSuccess(t, result, err)

	messages := requireArrayField(t, searchResult, "messages")
	if len(messages) == 0 {
		t.Fatal("search returned no results; message may not have been indexed yet")
	}

	// Verify the sent message appears in results
	found := false
	for _, m := range messages {
		msg, ok := m.(map[string]any)
		if !ok {
			continue
		}
		if msg["id"] == messageID {
			found = true
			break
		}
	}
	if !found {
		t.Logf("search returned %d results but sent message %s not found", len(messages), messageID)
	}

	// Step 4: Read the message back
	t.Log("reading test email")
	result, err = gmail.TestableGmailGetMessage(ctx, makeRequest(map[string]any{
		"message_id": messageID,
	}), deps)
	readResult := requireSuccess(t, result, err)

	readSubject := requireStringField(t, readResult, "subject")
	if readSubject != subject {
		t.Errorf("expected subject %q, got %q", subject, readSubject)
	}

	// Step 5: Reply to the message
	t.Log("replying to test email")
	result, err = gmail.TestableGmailReply(ctx, makeRequest(map[string]any{
		"message_id": messageID,
		"body":       "This is an automated reply. Safe to delete.",
	}), deps)
	replyResult := requireSuccess(t, result, err)

	replyThreadID := requireStringField(t, replyResult, "thread_id")
	if replyThreadID != threadID {
		t.Errorf("reply thread_id %q differs from original %q — thread grouping broken", replyThreadID, threadID)
	}

	// Step 6: Verify thread grouping - get the thread and check it has 2+ messages
	t.Log("verifying thread grouping")
	result, err = gmail.TestableGmailGetThread(ctx, makeRequest(map[string]any{
		"thread_id": threadID,
	}), deps)
	threadResult := requireSuccess(t, result, err)

	threadMessages := requireArrayField(t, threadResult, "messages")
	if len(threadMessages) < 2 {
		t.Errorf("expected at least 2 messages in thread, got %d", len(threadMessages))
	}

	t.Logf("thread contains %d messages", len(threadMessages))
}
