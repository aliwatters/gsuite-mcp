//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/gmail"
)

// TestGmailListSendAs verifies that listing send-as aliases works
// and that the primary address is included.
func TestGmailListSendAs(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	result, err := gmail.TestableGmailListSendAs(ctx, makeRequest(nil), deps)
	cr := requireSuccess(t, result, err)

	aliases := requireArrayField(t, cr, "send_as")
	if len(aliases) == 0 {
		t.Fatal("expected at least one send-as alias (the primary address)")
	}

	// Verify the primary address is present
	found := false
	for _, a := range aliases {
		alias, ok := a.(map[string]any)
		if !ok {
			continue
		}
		if alias["send_as_email"] == testAccount || alias["is_primary"] == true {
			found = true
			break
		}
	}
	if !found {
		t.Logf("send-as aliases: %v", aliases)
		t.Error("primary send-as alias not found")
	}

	t.Logf("found %d send-as aliases", len(aliases))
}

// TestGmailGetSendAs verifies getting details of the primary send-as address.
func TestGmailGetSendAs(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	result, err := gmail.TestableGmailGetSendAs(ctx, makeRequest(map[string]any{
		"send_as_email": testAccount,
	}), deps)
	cr := requireSuccess(t, result, err)

	email := requireStringField(t, cr, "send_as_email")
	if email != testAccount {
		t.Errorf("expected send_as_email %q, got %q", testAccount, email)
	}
}

// TestGmailVacationSettings verifies get/set vacation settings.
// It reads the current settings, writes a disabled setting, then restores.
func TestGmailVacationSettings(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := gmailDeps()

	// Get current vacation settings
	t.Log("getting vacation settings")
	result, err := gmail.TestableGmailGetVacation(ctx, makeRequest(nil), deps)
	cr := requireSuccess(t, result, err)

	// Note the current enabled state for restore
	wasEnabled := false
	if cr.Data != nil {
		wasEnabled, _ = cr.Data["enabled"].(bool)
	}

	// Set vacation to disabled (safe operation)
	t.Log("setting vacation to disabled")
	result, err = gmail.TestableGmailSetVacation(ctx, makeRequest(map[string]any{
		"enabled": false,
	}), deps)
	requireSuccess(t, result, err)

	// Verify it was set
	t.Log("verifying vacation is disabled")
	result, err = gmail.TestableGmailGetVacation(ctx, makeRequest(nil), deps)
	verifyResult := requireSuccess(t, result, err)
	if verifyResult.Data != nil {
		enabled, _ := verifyResult.Data["enabled"].(bool)
		if enabled {
			t.Error("vacation should be disabled after set")
		}
	}

	// Restore if it was previously enabled
	if wasEnabled {
		t.Log("restoring vacation to enabled")
		_, _ = gmail.TestableGmailSetVacation(ctx, makeRequest(map[string]any{
			"enabled": true,
		}), deps)
	}
}
