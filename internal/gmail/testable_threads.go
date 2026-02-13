package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailGetThread retrieves a thread using the provided service.
func TestableGmailGetThread(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", "")
	if threadID == "" {
		return mcp.NewToolResultError("thread_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.Params.Arguments, "format", "full")

	thread, err := svc.GetThread(ctx, threadID, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	opts := FormatMessageOptions{BodyFormat: parseBodyFormat(request.Params.Arguments)}
	messages := make([]map[string]any, 0, len(thread.Messages))
	for _, msg := range thread.Messages {
		messages = append(messages, FormatMessageWithOptions(msg, opts))
	}

	result := map[string]any{
		"thread_id": thread.Id,
		"messages":  messages,
		"count":     len(messages),
	}

	return common.MarshalToolResult(result)
}

// TestableGmailThreadArchive archives all messages in a thread (removes INBOX label).
func TestableGmailThreadArchive(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", "")
	if threadID == "" {
		return mcp.NewToolResultError("thread_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	modifyRequest := &gmail.ModifyThreadRequest{
		RemoveLabelIds: []string{"INBOX"},
	}

	thread, err := svc.ModifyThread(ctx, threadID, modifyRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return threadResult(thread)
}

// TestableGmailThreadTrash moves an entire thread to trash.
func TestableGmailThreadTrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", "")
	if threadID == "" {
		return mcp.NewToolResultError("thread_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	thread, err := svc.TrashThread(ctx, threadID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return threadResult(thread)
}

// TestableGmailThreadUntrash restores an entire thread from trash.
func TestableGmailThreadUntrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	threadID := common.ParseStringArg(request.Params.Arguments, "thread_id", "")
	if threadID == "" {
		return mcp.NewToolResultError("thread_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	thread, err := svc.UntrashThread(ctx, threadID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return threadResult(thread)
}
