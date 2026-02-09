package gmail

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// messageResult creates a standard MCP result from a Gmail message response.
// Used by single-message operations (modify, trash, archive, star, etc.)
func messageResult(msg *gmail.Message) (*mcp.CallToolResult, error) {
	result := map[string]any{
		"id":        msg.Id,
		"thread_id": msg.ThreadId,
		"labels":    msg.LabelIds,
	}

	return common.MarshalToolResult(result)
}

// === Handle functions - generated via WrapHandler ===

var (
	HandleGmailModifyMessage = common.WrapHandler[GmailService](TestableGmailModifyMessage)
	HandleGmailBatchModify   = common.WrapHandler[GmailService](TestableGmailBatchModify)
	HandleGmailTrash         = common.WrapHandler[GmailService](TestableGmailTrash)
	HandleGmailArchive       = common.WrapHandler[GmailService](TestableGmailArchive)
	HandleGmailMarkRead      = common.WrapHandler[GmailService](TestableGmailMarkRead)
	HandleGmailBatchArchive  = common.WrapHandler[GmailService](TestableGmailBatchArchive)
	HandleGmailBatchTrash    = common.WrapHandler[GmailService](TestableGmailBatchTrash)
	HandleGmailUntrash       = common.WrapHandler[GmailService](TestableGmailUntrash)
	HandleGmailMarkUnread    = common.WrapHandler[GmailService](TestableGmailMarkUnread)
	HandleGmailStar          = common.WrapHandler[GmailService](TestableGmailStar)
	HandleGmailUnstar        = common.WrapHandler[GmailService](TestableGmailUnstar)
)

// === Utility functions ===

// extractStringArray converts an any to []string
func extractStringArray(v any) []string {
	if v == nil {
		return nil
	}

	arr, ok := v.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
