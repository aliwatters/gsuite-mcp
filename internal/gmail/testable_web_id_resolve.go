package gmail

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestableGmailResolveWebID resolves a Gmail web-UI message/thread ID
// (as found in a browser URL) to the Gmail API message/thread IDs.
//
// It accepts a full Gmail URL or any of the documented web ID forms:
//   - FMfcg… (current, base64url-encoded)
//   - thread-f:<decimal> (legacy thread form)
//   - msg-f:<decimal>    (legacy message form)
//   - Already-API hex ID (passed through unchanged)
//
// On success it returns message_id, thread_id, id_kind, and source_id.
// On failure it returns a clear error — never a silent wrong-ID fallback.
func TestableGmailResolveWebID(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	input, errResult := common.RequireStringArg(request.GetArguments(), "id")
	if errResult != nil {
		return errResult, nil
	}

	resolved, err := ParseGmailWebID(strings.TrimSpace(input))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("cannot resolve Gmail web ID: %v", err)), nil
	}

	if resolved.Kind == WebIDKindFMfcg {
		svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
		if !ok {
			return errResult, nil
		}
		if err := validateFMfcgResolution(ctx, svc, resolved); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
	}

	result := map[string]any{
		"id_kind":   string(resolved.Kind),
		"source_id": input,
	}

	if resolved.ThreadID != "" {
		result["thread_id"] = resolved.ThreadID
	}
	if resolved.MessageID != "" {
		result["message_id"] = resolved.MessageID
	}

	// Provide a hint about which API calls to use
	switch resolved.Kind {
	case WebIDKindThreadF:
		result["hint"] = fmt.Sprintf("use thread_id %q with gmail_get_thread", resolved.ThreadID)
	case WebIDKindMsgF:
		result["hint"] = fmt.Sprintf("use message_id %q with gmail_get_message", resolved.MessageID)
	case WebIDKindFMfcg:
		result["hint"] = fmt.Sprintf("use thread_id %q with gmail_get_thread, or message_id %q with gmail_get_message", resolved.ThreadID, resolved.MessageID)
	case WebIDKindAPIID:
		result["hint"] = "input already appears to be a Gmail API ID; passed through unchanged"
	}

	return common.MarshalToolResult(result)
}

func validateFMfcgResolution(ctx context.Context, svc GmailService, resolved *ResolvedWebID) error {
	if resolved.MessageID != "" {
		if _, err := svc.GetMessage(ctx, resolved.MessageID, "metadata"); err != nil {
			return fmt.Errorf("cannot resolve Gmail FMfcg web ID for this account: decoded message_id %q is not fetchable (%w); use gmail_search with subject, sender, or date terms to find the message", resolved.MessageID, err)
		}
	}

	if resolved.ThreadID != "" {
		if _, err := svc.GetThread(ctx, resolved.ThreadID, "metadata"); err != nil {
			return fmt.Errorf("cannot resolve Gmail FMfcg web ID for this account: decoded thread_id %q is not fetchable (%w); use gmail_search with subject, sender, or date terms to find the message", resolved.ThreadID, err)
		}
	}

	return nil
}
