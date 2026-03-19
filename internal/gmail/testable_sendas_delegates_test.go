package gmail

import (
	"context"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// getTextResult extracts the text from a tool result for error messages.
func getTextResult(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return ""
	}
	return tc.Text
}

func TestTestableGmailSendAs(t *testing.T) {
	fix := NewGmailTestFixtures()
	ctx := context.Background()

	// Seed send-as aliases
	fix.MockService.AddSendAs(&gmail.SendAs{
		SendAsEmail:        "primary@example.com",
		DisplayName:        "Primary User",
		IsPrimary:          true,
		IsDefault:          true,
		VerificationStatus: "accepted",
	})
	fix.MockService.AddSendAs(&gmail.SendAs{
		SendAsEmail:        "alias@example.com",
		DisplayName:        "Alias User",
		ReplyToAddress:     "reply@example.com",
		Signature:          "<b>My Signature</b>",
		VerificationStatus: "accepted",
	})

	t.Run("ListSendAs", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailListSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", getTextResult(result))
		}

		resp := extractResponse(t, result)
		if resp["send_as_count"] == nil {
			t.Error("expected send_as_count in response")
		}
		if !fix.MockService.WasMethodCalled("ListSendAs") {
			t.Error("expected ListSendAs to be called")
		}
	})

	t.Run("GetSendAs", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"send_as_email": "alias@example.com",
		})
		result, err := TestableGmailGetSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", getTextResult(result))
		}

		resp := extractResponse(t, result)
		if resp["send_as_email"] != "alias@example.com" {
			t.Errorf("expected alias@example.com, got %v", resp["send_as_email"])
		}
	})

	t.Run("GetSendAs_missing_email", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailGetSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for missing send_as_email")
		}
	})

	t.Run("GetSendAs_not_found", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"send_as_email": "notfound@example.com",
		})
		result, err := TestableGmailGetSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for non-existent send-as")
		}
	})

	t.Run("CreateSendAs", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"send_as_email":    "new@example.com",
			"display_name":     "New Alias",
			"reply_to_address": "reply-new@example.com",
		})
		result, err := TestableGmailCreateSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
		if !fix.MockService.WasMethodCalled("CreateSendAs") {
			t.Error("expected CreateSendAs to be called")
		}
		if _, ok := fix.MockService.SendAs["new@example.com"]; !ok {
			t.Error("expected send-as to be added to mock store")
		}
	})

	t.Run("UpdateSendAs", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"send_as_email": "alias@example.com",
			"display_name":  "Updated Alias",
		})
		result, err := TestableGmailUpdateSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
	})

	t.Run("DeleteSendAs", func(t *testing.T) {
		fix.MockService.AddSendAs(&gmail.SendAs{
			SendAsEmail: "todelete@example.com",
			DisplayName: "To Delete",
		})

		req := common.CreateMCPRequest(map[string]any{
			"send_as_email": "todelete@example.com",
		})
		result, err := TestableGmailDeleteSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
		if _, ok := fix.MockService.SendAs["todelete@example.com"]; ok {
			t.Error("expected send-as to be removed from mock store")
		}
	})

	t.Run("VerifySendAs", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"send_as_email": "alias@example.com",
		})
		result, err := TestableGmailVerifySendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
		if !fix.MockService.WasMethodCalled("VerifySendAs") {
			t.Error("expected VerifySendAs to be called")
		}
	})

	t.Run("API_error", func(t *testing.T) {
		fix.MockService.SetError("send-as API failure")
		defer func() { fix.MockService.Error = nil }()

		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailListSendAs(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result for API failure")
		}
	})
}

func TestTestableGmailDelegates(t *testing.T) {
	fix := NewGmailTestFixtures()
	ctx := context.Background()

	// Seed delegates
	fix.MockService.AddDelegate(&gmail.Delegate{
		DelegateEmail:      "delegate1@example.com",
		VerificationStatus: "accepted",
	})
	fix.MockService.AddDelegate(&gmail.Delegate{
		DelegateEmail:      "delegate2@example.com",
		VerificationStatus: "pending",
	})

	t.Run("ListDelegates", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailListDelegates(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("unexpected tool error: %s", getTextResult(result))
		}

		resp := extractResponse(t, result)
		if resp["delegate_count"] == nil {
			t.Error("expected delegate_count in response")
		}
		if !fix.MockService.WasMethodCalled("ListDelegates") {
			t.Error("expected ListDelegates to be called")
		}
	})

	t.Run("CreateDelegate", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"delegate_email": "newdelegate@example.com",
		})
		result, err := TestableGmailCreateDelegate(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
		if !fix.MockService.WasMethodCalled("CreateDelegate") {
			t.Error("expected CreateDelegate to be called")
		}
		if _, ok := fix.MockService.Delegates["newdelegate@example.com"]; !ok {
			t.Error("expected delegate to be added to mock store")
		}
	})

	t.Run("CreateDelegate_missing_email", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailCreateDelegate(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for missing delegate_email")
		}
	})

	t.Run("DeleteDelegate", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"delegate_email": "delegate1@example.com",
		})
		result, err := TestableGmailDeleteDelegate(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Errorf("unexpected error: %s", getTextResult(result))
		}
		if _, ok := fix.MockService.Delegates["delegate1@example.com"]; ok {
			t.Error("expected delegate to be removed from mock store")
		}
	})

	t.Run("DeleteDelegate_not_found", func(t *testing.T) {
		req := common.CreateMCPRequest(map[string]any{
			"delegate_email": "notfound@example.com",
		})
		result, err := TestableGmailDeleteDelegate(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error for non-existent delegate")
		}
	})

	t.Run("API_error", func(t *testing.T) {
		fix.MockService.SetError("delegate API failure")
		defer func() { fix.MockService.Error = nil }()

		req := common.CreateMCPRequest(map[string]any{})
		result, err := TestableGmailListDelegates(ctx, req, fix.Deps)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Error("expected error result for API failure")
		}
	})
}
