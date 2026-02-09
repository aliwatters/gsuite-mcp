package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailListDrafts lists all drafts.
func TestableGmailListDrafts(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.GmailDefaultMaxResults, common.GmailMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	resp, err := svc.ListDrafts(ctx, maxResults, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	type draftInfo struct {
		ID        string `json:"id"`
		MessageID string `json:"message_id,omitempty"`
		ThreadID  string `json:"thread_id,omitempty"`
		Snippet   string `json:"snippet,omitempty"`
	}

	drafts := make([]draftInfo, 0, len(resp.Drafts))
	for _, draft := range resp.Drafts {
		info := draftInfo{ID: draft.Id}
		if draft.Message != nil {
			info.MessageID = draft.Message.Id
			info.ThreadID = draft.Message.ThreadId
			info.Snippet = draft.Message.Snippet
		}
		drafts = append(drafts, info)
	}

	result := map[string]any{
		"drafts":          drafts,
		"count":           len(drafts),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetDraft retrieves a draft by ID.
func TestableGmailGetDraft(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	draftID := common.ParseStringArg(request.Params.Arguments, "draft_id", "")
	if draftID == "" {
		return mcp.NewToolResultError("draft_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.Params.Arguments, "format", "full")

	draft, err := svc.GetDraft(ctx, draftID, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id": draft.Id,
	}

	if draft.Message != nil {
		result["message"] = FormatMessage(draft.Message)
	}

	return common.MarshalToolResult(result)
}

// TestableGmailUpdateDraft updates an existing draft.
func TestableGmailUpdateDraft(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	draftID := common.ParseStringArg(request.Params.Arguments, "draft_id", "")
	if draftID == "" {
		return mcp.NewToolResultError("draft_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	to := common.ParseStringArg(request.Params.Arguments, "to", "")
	subject := common.ParseStringArg(request.Params.Arguments, "subject", "")
	body := common.ParseStringArg(request.Params.Arguments, "body", "")
	cc := common.ParseStringArg(request.Params.Arguments, "cc", "")
	bcc := common.ParseStringArg(request.Params.Arguments, "bcc", "")

	raw := buildEmailMessage(EmailMessage{
		To:      to,
		Cc:      cc,
		Bcc:     bcc,
		Subject: subject,
		Body:    body,
	})

	message := &gmail.Message{
		Raw: raw,
	}

	draft := &gmail.Draft{
		Message: message,
	}

	updated, err := svc.UpdateDraft(ctx, draftID, draft)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":         updated.Id,
		"message_id": "",
	}
	if updated.Message != nil {
		result["message_id"] = updated.Message.Id
	}

	return common.MarshalToolResult(result)
}

// TestableGmailDeleteDraft deletes a draft.
func TestableGmailDeleteDraft(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	draftID := common.ParseStringArg(request.Params.Arguments, "draft_id", "")
	if draftID == "" {
		return mcp.NewToolResultError("draft_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	err := svc.DeleteDraft(ctx, draftID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":      draftID,
		"deleted": true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSendDraft sends a draft.
func TestableGmailSendDraft(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	draftID := common.ParseStringArg(request.Params.Arguments, "draft_id", "")
	if draftID == "" {
		return mcp.NewToolResultError("draft_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	draft := &gmail.Draft{
		Id: draftID,
	}

	sent, err := svc.SendDraft(ctx, draft)
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

// TestableGmailDraft creates a new draft.
func TestableGmailDraft(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	to := common.ParseStringArg(request.Params.Arguments, "to", "")
	subject := common.ParseStringArg(request.Params.Arguments, "subject", "")
	body := common.ParseStringArg(request.Params.Arguments, "body", "")
	cc := common.ParseStringArg(request.Params.Arguments, "cc", "")
	bcc := common.ParseStringArg(request.Params.Arguments, "bcc", "")

	raw := buildEmailMessage(EmailMessage{
		To:      to,
		Cc:      cc,
		Bcc:     bcc,
		Subject: subject,
		Body:    body,
	})

	message := &gmail.Message{
		Raw: raw,
	}

	// Support creating draft as reply
	if threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", ""); threadID != "" {
		message.ThreadId = threadID
	}

	draft := &gmail.Draft{
		Message: message,
	}

	created, err := svc.CreateDraft(ctx, draft)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":         created.Id,
		"message_id": "",
	}
	if created.Message != nil {
		result["message_id"] = created.Message.Id
	}

	return common.MarshalToolResult(result)
}
