package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// emailContentType is the default Content-Type header for plain text emails.
const emailContentType = "Content-Type: text/plain; charset=\"UTF-8\""

// EmailMessage holds the components for building an RFC 2822 email message.
type EmailMessage struct {
	To           string
	Cc           string
	Bcc          string
	Subject      string
	Body         string
	ExtraHeaders map[string]string // Additional headers (e.g., In-Reply-To, References)
}

// buildEmailMessage constructs an RFC 2822 email message and returns it
// as a base64 URL-encoded string suitable for the Gmail API Raw field.
// Fields that are empty strings are omitted from the headers.
func buildEmailMessage(msg EmailMessage) string {
	var b strings.Builder

	if msg.To != "" {
		fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	}
	if msg.Cc != "" {
		fmt.Fprintf(&b, "Cc: %s\r\n", msg.Cc)
	}
	if msg.Bcc != "" {
		fmt.Fprintf(&b, "Bcc: %s\r\n", msg.Bcc)
	}
	if msg.Subject != "" {
		fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	}
	for key, value := range msg.ExtraHeaders {
		fmt.Fprintf(&b, "%s: %s\r\n", key, value)
	}
	b.WriteString(emailContentType + "\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)

	return base64.URLEncoding.EncodeToString([]byte(b.String()))
}

// modifyMessageLabels is a generic helper for single-message label modification
// operations (archive, star, mark read, spam, etc.). It extracts message_id from
// the request, applies the specified label changes, and returns a standard result.
func modifyMessageLabels(ctx context.Context, svc GmailService, request mcp.CallToolRequest, addLabels, removeLabels []string) (*mcp.CallToolResult, error) {
	messageID, _ := request.Params.Arguments["message_id"].(string)
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	modifyRequest := &gmail.ModifyMessageRequest{
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	msg, err := svc.ModifyMessage(ctx, messageID, modifyRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return messageResult(msg)
}

// batchModifyLabels is a generic helper for multi-message label modification
// operations (batch archive, batch trash). It extracts message_ids from the
// request, applies the specified label changes, and returns a standard result.
func batchModifyLabels(ctx context.Context, svc GmailService, request mcp.CallToolRequest, addLabels, removeLabels []string, countKey string) (*mcp.CallToolResult, error) {
	messageIDs := extractStringArray(request.Params.Arguments["message_ids"])
	if len(messageIDs) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required"), nil
	}

	batchRequest := &gmail.BatchModifyMessagesRequest{
		Ids:            messageIDs,
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	err := svc.BatchModifyMessages(ctx, batchRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		countKey:      len(messageIDs),
		"message_ids": messageIDs,
	}

	return common.MarshalToolResult(result)
}
