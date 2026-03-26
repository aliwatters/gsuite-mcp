package chat

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	chatapi "google.golang.org/api/chat/v1"
)

// TestableChatListSpaces lists Chat spaces the user is a member of.
func TestableChatListSpaces(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	pageSize := common.ParseMaxResults(request.Params.Arguments, 20, 100)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	spaces, nextPageToken, err := srv.ListSpaces(ctx, pageSize, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	formattedSpaces := make([]map[string]any, 0, len(spaces))
	for _, s := range spaces {
		formattedSpaces = append(formattedSpaces, formatSpace(s))
	}

	result := map[string]any{
		"space_count": len(formattedSpaces),
		"spaces":      formattedSpaces,
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// TestableChatGetSpace gets details about a Chat space.
func TestableChatGetSpace(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spaceName, errResult := common.RequireStringArg(request.Params.Arguments, "space_name")
	if errResult != nil {
		return errResult, nil
	}

	spaceName = normalizeSpaceName(spaceName)

	space, err := srv.GetSpace(ctx, spaceName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatSpace(space))
}

// TestableChatCreateSpace creates a new Chat space.
func TestableChatCreateSpace(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	displayName, errResult := common.RequireStringArg(request.Params.Arguments, "display_name")
	if errResult != nil {
		return errResult, nil
	}

	spaceType := common.ParseStringArg(request.Params.Arguments, "space_type", "SPACE")

	space, err := srv.CreateSpace(ctx, displayName, spaceType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatSpace(space))
}

// TestableChatListMessages lists messages in a Chat space.
func TestableChatListMessages(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spaceName, errResult := common.RequireStringArg(request.Params.Arguments, "space_name")
	if errResult != nil {
		return errResult, nil
	}

	spaceName = normalizeSpaceName(spaceName)
	pageSize := common.ParseMaxResults(request.Params.Arguments, 25, 100)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	filter := common.ParseStringArg(request.Params.Arguments, "filter", "")

	messages, nextPageToken, err := srv.ListMessages(ctx, spaceName, pageSize, pageToken, filter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	formattedMessages := make([]map[string]any, 0, len(messages))
	for _, msg := range messages {
		formattedMessages = append(formattedMessages, formatMessage(msg))
	}

	result := map[string]any{
		"message_count": len(formattedMessages),
		"messages":      formattedMessages,
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// TestableChatGetMessage gets a specific message.
func TestableChatGetMessage(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	messageName, errResult := common.RequireStringArg(request.Params.Arguments, "message_name")
	if errResult != nil {
		return errResult, nil
	}

	msg, err := srv.GetMessage(ctx, messageName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatMessage(msg))
}

// TestableChatSendMessage sends a message to a Chat space.
func TestableChatSendMessage(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spaceName, errResult := common.RequireStringArg(request.Params.Arguments, "space_name")
	if errResult != nil {
		return errResult, nil
	}

	text, errResult := common.RequireStringArg(request.Params.Arguments, "text")
	if errResult != nil {
		return errResult, nil
	}

	spaceName = normalizeSpaceName(spaceName)
	threadName := common.ParseStringArg(request.Params.Arguments, "thread_name", "")

	msg, err := srv.SendMessage(ctx, spaceName, text, threadName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatMessage(msg))
}

// TestableChatCreateReaction adds a reaction to a message.
func TestableChatCreateReaction(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	messageName, errResult := common.RequireStringArg(request.Params.Arguments, "message_name")
	if errResult != nil {
		return errResult, nil
	}

	emoji, errResult := common.RequireStringArg(request.Params.Arguments, "emoji")
	if errResult != nil {
		return errResult, nil
	}

	reaction, err := srv.CreateReaction(ctx, messageName, emoji)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	result := map[string]any{
		"reaction_name": reaction.Name,
	}
	if reaction.Emoji != nil {
		result["emoji"] = reaction.Emoji.Unicode
	}

	return common.MarshalToolResult(result)
}

// TestableChatDeleteReaction removes a reaction from a message.
func TestableChatDeleteReaction(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	reactionName, errResult := common.RequireStringArg(request.Params.Arguments, "reaction_name")
	if errResult != nil {
		return errResult, nil
	}

	err := srv.DeleteReaction(ctx, reactionName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{
		"deleted": true,
	})
}

// TestableChatListMembers lists members of a Chat space.
func TestableChatListMembers(ctx context.Context, request mcp.CallToolRequest, deps *ChatHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveChatServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	spaceName, errResult := common.RequireStringArg(request.Params.Arguments, "space_name")
	if errResult != nil {
		return errResult, nil
	}

	spaceName = normalizeSpaceName(spaceName)
	pageSize := common.ParseMaxResults(request.Params.Arguments, 100, 1000)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	members, nextPageToken, err := srv.ListMembers(ctx, spaceName, pageSize, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Chat API error: %v", err)), nil
	}

	formattedMembers := make([]map[string]any, 0, len(members))
	for _, member := range members {
		formattedMembers = append(formattedMembers, formatMembership(member))
	}

	result := map[string]any{
		"member_count": len(formattedMembers),
		"members":      formattedMembers,
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// normalizeSpaceName ensures a space name has the "spaces/" prefix.
func normalizeSpaceName(name string) string {
	if len(name) > 0 && name[0] != 's' {
		// If just an ID was provided, add the prefix
		return "spaces/" + name
	}
	return name
}

// formatSpace formats a Chat space for output.
func formatSpace(space *chatapi.Space) map[string]any {
	result := map[string]any{
		"name": space.Name,
	}

	if space.DisplayName != "" {
		result["display_name"] = space.DisplayName
	}
	if space.SpaceType != "" {
		result["space_type"] = space.SpaceType
	}
	if space.SpaceThreadingState != "" {
		result["threading_state"] = space.SpaceThreadingState
	}
	if space.CreateTime != "" {
		result["create_time"] = space.CreateTime
	}

	return result
}

// formatMessage formats a Chat message for output.
func formatMessage(msg *chatapi.Message) map[string]any {
	result := map[string]any{
		"name": msg.Name,
	}

	if msg.Text != "" {
		result["text"] = msg.Text
	}
	if msg.CreateTime != "" {
		result["create_time"] = msg.CreateTime
	}
	if msg.Sender != nil {
		sender := map[string]any{
			"name": msg.Sender.Name,
		}
		if msg.Sender.DisplayName != "" {
			sender["display_name"] = msg.Sender.DisplayName
		}
		result["sender"] = sender
	}
	if msg.Thread != nil && msg.Thread.Name != "" {
		result["thread_name"] = msg.Thread.Name
	}

	return result
}

// formatMembership formats a Chat membership for output.
func formatMembership(m *chatapi.Membership) map[string]any {
	result := map[string]any{
		"name": m.Name,
	}

	if m.Member != nil {
		member := map[string]any{
			"name": m.Member.Name,
		}
		if m.Member.DisplayName != "" {
			member["display_name"] = m.Member.DisplayName
		}
		result["member"] = member
	}
	if m.Role != "" {
		result["role"] = m.Role
	}
	if m.CreateTime != "" {
		result["create_time"] = m.CreateTime
	}

	return result
}
