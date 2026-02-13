package gmail

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailSearch performs a Gmail search using the provided service.
func TestableGmailSearch(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	query, _ := request.Params.Arguments["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.GmailDefaultMaxResults, common.GmailMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	resp, err := svc.ListMessages(ctx, query, maxResults, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	type messageInfo struct {
		ID       string `json:"id"`
		ThreadID string `json:"thread_id"`
	}

	messages := make([]messageInfo, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, messageInfo{
			ID:       msg.Id,
			ThreadID: msg.ThreadId,
		})
	}

	result := map[string]any{
		"messages":        messages,
		"result_size":     resp.ResultSizeEstimate,
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetMessage retrieves a single message using the provided service.
func TestableGmailGetMessage(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.Params.Arguments, "format", "full")

	msg, err := svc.GetMessage(ctx, messageID, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := FormatMessageWithOptions(msg, FormatMessageOptions{BodyFormat: parseBodyFormat(request.Params.Arguments)})
	return common.MarshalToolResult(result)
}

// TestableGmailGetMessages retrieves multiple messages using the provided service.
func TestableGmailGetMessages(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageIDsRaw, ok := request.Params.Arguments["message_ids"].([]any)
	if !ok || len(messageIDsRaw) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required (array of message IDs)"), nil
	}

	if len(messageIDsRaw) > common.GmailMaxBatchMessages {
		return mcp.NewToolResultError(fmt.Sprintf("maximum %d messages per request", common.GmailMaxBatchMessages)), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.Params.Arguments, "format", "full")

	messages := make([]map[string]any, 0, len(messageIDsRaw))
	var errors []string

	opts := FormatMessageOptions{BodyFormat: parseBodyFormat(request.Params.Arguments)}
	for _, idRaw := range messageIDsRaw {
		messageID, ok := idRaw.(string)
		if !ok {
			continue
		}

		msg, err := svc.GetMessage(ctx, messageID, format)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
			continue
		}

		messages = append(messages, FormatMessageWithOptions(msg, opts))
	}

	result := map[string]any{
		"messages": messages,
		"count":    len(messages),
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSend sends a new email using the provided service.
func TestableGmailSend(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	to := common.ParseStringArg(request.Params.Arguments, "to", "")
	if to == "" {
		return mcp.NewToolResultError("to parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	message := buildMessageFromArgs(request.Params.Arguments)

	sent, err := svc.SendMessage(ctx, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":        sent.Id,
		"thread_id": sent.ThreadId,
		"labels":    sent.LabelIds,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailReply replies to an email using the provided service.
func TestableGmailReply(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required (the message you're replying to)"), nil
	}

	body := common.ParseStringArg(request.Params.Arguments, "body", "")
	if body == "" {
		return mcp.NewToolResultError("body parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	// Get the original message to extract headers
	origMsg, err := svc.GetMessage(ctx, messageID, "metadata")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get original message: %v", err)), nil
	}

	var origFrom, origTo, origCc, origSubject, origMessageID, origReferences string
	if origMsg.Payload != nil {
		for _, h := range origMsg.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "from":
				origFrom = h.Value
			case "to":
				origTo = h.Value
			case "cc":
				origCc = h.Value
			case "subject":
				origSubject = h.Value
			case "message-id":
				origMessageID = h.Value
			case "references":
				origReferences = h.Value
			}
		}
	}

	replyTo := origFrom
	if to := common.ParseStringArg(request.Params.Arguments, "to", ""); to != "" {
		replyTo = to
	}

	replyAll := common.ParseBoolArg(request.Params.Arguments, "reply_all", false)

	var cc string
	if replyAll {
		var ccList []string
		if origTo != "" {
			ccList = append(ccList, origTo)
		}
		if origCc != "" {
			ccList = append(ccList, origCc)
		}
		cc = strings.Join(ccList, ", ")
	}
	if ccOverride := common.ParseStringArg(request.Params.Arguments, "cc", ""); ccOverride != "" {
		cc = ccOverride
	}

	subject := origSubject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	references := origMessageID
	if origReferences != "" {
		references = origReferences + " " + origMessageID
	}

	emailMsg := EmailMessage{
		To:      replyTo,
		Cc:      cc,
		Subject: subject,
		Body:    body,
		ExtraHeaders: map[string]string{
			"References":  references,
			"In-Reply-To": origMessageID,
		},
	}

	raw := buildEmailMessage(emailMsg)

	message := &gmail.Message{
		Raw:      raw,
		ThreadId: origMsg.ThreadId,
	}

	sent, err := svc.SendMessage(ctx, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":        sent.Id,
		"thread_id": sent.ThreadId,
		"labels":    sent.LabelIds,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSpam marks a message as spam.
func TestableGmailSpam(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"SPAM"}, []string{"INBOX"})
}

// TestableGmailNotSpam removes a message from spam and moves to inbox.
func TestableGmailNotSpam(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"INBOX"}, []string{"SPAM"})
}

// TestableGmailTrash moves a message to trash.
func TestableGmailTrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.TrashMessage(ctx, messageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":       msg.Id,
		"labelIds": msg.LabelIds,
		"success":  true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailUntrash removes a message from trash.
func TestableGmailUntrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.UntrashMessage(ctx, messageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":       msg.Id,
		"labelIds": msg.LabelIds,
		"success":  true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetAttachment downloads an attachment.
func TestableGmailGetAttachment(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	attachmentID := common.ParseStringArg(request.Params.Arguments, "attachment_id", "")
	if attachmentID == "" {
		return mcp.NewToolResultError("attachment_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	attach, err := svc.GetAttachment(ctx, messageID, attachmentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"attachment_id": attachmentID,
		"message_id":    messageID,
		"size":          attach.Size,
		"data":          attach.Data,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailArchive archives a message.
func TestableGmailArchive(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"INBOX"})
}

// TestableGmailMarkRead marks a message as read.
func TestableGmailMarkRead(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"UNREAD"})
}

// TestableGmailMarkUnread marks a message as unread.
func TestableGmailMarkUnread(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"UNREAD"}, nil)
}

// TestableGmailStar stars a message.
func TestableGmailStar(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"STARRED"}, nil)
}

// TestableGmailUnstar unstars a message.
func TestableGmailUnstar(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"STARRED"})
}
