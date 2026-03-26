//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/gmail"
)

// TestGmailGetProfile verifies that the test account can authenticate
// and that gmail_get_profile returns the expected email address.
func TestGmailGetProfile(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	result, err := gmail.TestableGmailGetProfile(ctx, makeRequest(nil), deps)
	cr := requireSuccess(t, result, err)

	email := requireStringField(t, cr, "email")
	if email != testAccount {
		t.Errorf("expected email %q, got %q", testAccount, email)
	}

	// Profile should have message and thread counts
	if cr.Data["messages_total"] == nil {
		t.Error("expected messages_total in profile")
	}
	if cr.Data["threads_total"] == nil {
		t.Error("expected threads_total in profile")
	}

	t.Logf("profile: email=%s messages=%.0f threads=%.0f",
		email,
		cr.Data["messages_total"],
		cr.Data["threads_total"],
	)
}
