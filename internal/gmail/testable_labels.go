package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailListLabels lists all labels.
func TestableGmailListLabels(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resp, err := svc.ListLabels(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	type labelInfo struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Type           string `json:"type"`
		MessagesTotal  int64  `json:"messages_total,omitempty"`
		MessagesUnread int64  `json:"messages_unread,omitempty"`
		ThreadsTotal   int64  `json:"threads_total,omitempty"`
		ThreadsUnread  int64  `json:"threads_unread,omitempty"`
		Color          any    `json:"color,omitempty"`
		LabelListVis   string `json:"label_list_visibility,omitempty"`
		MessageListVis string `json:"message_list_visibility,omitempty"`
	}

	labels := make([]labelInfo, 0, len(resp.Labels))
	for _, label := range resp.Labels {
		info := labelInfo{
			ID:             label.Id,
			Name:           label.Name,
			Type:           label.Type,
			MessagesTotal:  label.MessagesTotal,
			MessagesUnread: label.MessagesUnread,
			ThreadsTotal:   label.ThreadsTotal,
			ThreadsUnread:  label.ThreadsUnread,
			LabelListVis:   label.LabelListVisibility,
			MessageListVis: label.MessageListVisibility,
		}
		if label.Color != nil {
			info.Color = map[string]string{
				"text_color": label.Color.TextColor,
				"bg_color":   label.Color.BackgroundColor,
			}
		}
		labels = append(labels, info)
	}

	result := map[string]any{
		"labels": labels,
		"count":  len(labels),
	}

	return common.MarshalToolResult(result)
}

// TestableGmailCreateLabel creates a new label.
func TestableGmailCreateLabel(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	name := common.ParseStringArg(request.Params.Arguments, "name", "")
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	label := &gmail.Label{
		Name: name,
	}

	if vis := common.ParseStringArg(request.Params.Arguments, "label_list_visibility", ""); vis != "" {
		label.LabelListVisibility = vis
	}
	if vis := common.ParseStringArg(request.Params.Arguments, "message_list_visibility", ""); vis != "" {
		label.MessageListVisibility = vis
	}

	created, err := svc.CreateLabel(ctx, label)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":   created.Id,
		"name": created.Name,
		"type": created.Type,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailDeleteLabel deletes a label.
func TestableGmailDeleteLabel(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	labelID := common.ParseStringArg(request.Params.Arguments, "label_id", "")
	if labelID == "" {
		return mcp.NewToolResultError("label_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	err := svc.DeleteLabel(ctx, labelID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":      labelID,
		"deleted": true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailUpdateLabel updates a label.
func TestableGmailUpdateLabel(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	labelID := common.ParseStringArg(request.Params.Arguments, "label_id", "")
	if labelID == "" {
		return mcp.NewToolResultError("label_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	label := &gmail.Label{
		Id: labelID,
	}

	if name := common.ParseStringArg(request.Params.Arguments, "name", ""); name != "" {
		label.Name = name
	}
	if vis := common.ParseStringArg(request.Params.Arguments, "label_list_visibility", ""); vis != "" {
		label.LabelListVisibility = vis
	}
	if vis := common.ParseStringArg(request.Params.Arguments, "message_list_visibility", ""); vis != "" {
		label.MessageListVisibility = vis
	}

	updated, err := svc.UpdateLabel(ctx, labelID, label)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":   updated.Id,
		"name": updated.Name,
		"type": updated.Type,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailModifyMessage modifies labels on a message.
func TestableGmailModifyMessage(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.Params.Arguments, "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	var addLabels, removeLabels []string
	if add, ok := request.Params.Arguments["add_labels"].([]any); ok {
		for _, l := range add {
			if s, ok := l.(string); ok {
				addLabels = append(addLabels, s)
			}
		}
	}
	if remove, ok := request.Params.Arguments["remove_labels"].([]any); ok {
		for _, l := range remove {
			if s, ok := l.(string); ok {
				removeLabels = append(removeLabels, s)
			}
		}
	}

	return modifyMessageLabels(ctx, svc, request, addLabels, removeLabels)
}

// TestableGmailModifyThread modifies labels on a thread.
func TestableGmailModifyThread(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", "")
	if threadID == "" {
		return mcp.NewToolResultError("thread_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	var addLabels, removeLabels []string
	if add, ok := request.Params.Arguments["add_labels"].([]any); ok {
		for _, l := range add {
			if s, ok := l.(string); ok {
				addLabels = append(addLabels, s)
			}
		}
	}
	if remove, ok := request.Params.Arguments["remove_labels"].([]any); ok {
		for _, l := range remove {
			if s, ok := l.(string); ok {
				removeLabels = append(removeLabels, s)
			}
		}
	}

	modifyRequest := &gmail.ModifyThreadRequest{
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	thread, err := svc.ModifyThread(ctx, threadID, modifyRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return threadResult(thread)
}

// TestableGmailBatchModify modifies labels on multiple messages.
func TestableGmailBatchModify(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageIDsRaw, ok := request.Params.Arguments["message_ids"].([]any)
	if !ok || len(messageIDsRaw) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required (array)"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	var ids []string
	for _, id := range messageIDsRaw {
		if s, ok := id.(string); ok {
			ids = append(ids, s)
		}
	}

	var addLabels, removeLabels []string
	if add, ok := request.Params.Arguments["add_labels"].([]any); ok {
		for _, l := range add {
			if s, ok := l.(string); ok {
				addLabels = append(addLabels, s)
			}
		}
	}
	if remove, ok := request.Params.Arguments["remove_labels"].([]any); ok {
		for _, l := range remove {
			if s, ok := l.(string); ok {
				removeLabels = append(removeLabels, s)
			}
		}
	}

	req := &gmail.BatchModifyMessagesRequest{
		Ids:            ids,
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	err := svc.BatchModifyMessages(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"message_ids": ids,
		"success":     true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailBatchTrash moves multiple messages to trash.
func TestableGmailBatchTrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageIDsRaw, ok := request.Params.Arguments["message_ids"].([]any)
	if !ok || len(messageIDsRaw) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required (array)"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	var ids []string
	for _, id := range messageIDsRaw {
		if s, ok := id.(string); ok {
			ids = append(ids, s)
		}
	}

	req := &gmail.BatchModifyMessagesRequest{
		Ids:         ids,
		AddLabelIds: []string{"TRASH"},
	}

	err := svc.BatchModifyMessages(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"message_ids":   ids,
		"trashed":       true,
		"trashed_count": len(ids),
		"success":       true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailBatchArchive archives multiple messages.
func TestableGmailBatchArchive(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageIDsRaw, ok := request.Params.Arguments["message_ids"].([]any)
	if !ok || len(messageIDsRaw) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required (array)"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	var ids []string
	for _, id := range messageIDsRaw {
		if s, ok := id.(string); ok {
			ids = append(ids, s)
		}
	}

	req := &gmail.BatchModifyMessagesRequest{
		Ids:            ids,
		RemoveLabelIds: []string{"INBOX"},
	}

	err := svc.BatchModifyMessages(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"message_ids":    ids,
		"archived":       true,
		"archived_count": len(ids),
		"success":        true,
	}

	return common.MarshalToolResult(result)
}
