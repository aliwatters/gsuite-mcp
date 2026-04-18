package drive

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// formatComment formats a comment for output.
func formatComment(c *drive.Comment) map[string]any {
	result := map[string]any{
		"id":      c.Id,
		"content": c.Content,
	}
	if c.Author != nil {
		result["author"] = map[string]any{
			"display_name": c.Author.DisplayName,
			"email":        c.Author.EmailAddress,
		}
	}
	if c.CreatedTime != "" {
		result["created_time"] = c.CreatedTime
	}
	if c.ModifiedTime != "" {
		result["modified_time"] = c.ModifiedTime
	}
	if c.Resolved {
		result["resolved"] = true
	}
	if c.QuotedFileContent != nil && c.QuotedFileContent.Value != "" {
		result["quoted_content"] = c.QuotedFileContent.Value
	}
	if len(c.Replies) > 0 {
		replies := make([]map[string]any, 0, len(c.Replies))
		for _, r := range c.Replies {
			replies = append(replies, formatReply(r))
		}
		result["replies"] = replies
	}
	return result
}

// formatReply formats a reply for output.
func formatReply(r *drive.Reply) map[string]any {
	result := map[string]any{
		"id":      r.Id,
		"content": r.Content,
	}
	if r.Author != nil {
		result["author"] = map[string]any{
			"display_name": r.Author.DisplayName,
			"email":        r.Author.EmailAddress,
		}
	}
	if r.CreatedTime != "" {
		result["created_time"] = r.CreatedTime
	}
	if r.ModifiedTime != "" {
		result["modified_time"] = r.ModifiedTime
	}
	if r.Action != "" {
		result["action"] = r.Action
	}
	return result
}

// TestableDriveListComments lists comments on a file.
func TestableDriveListComments(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	resp, err := srv.ListComments(ctx, fileID, DriveCommentListFields, maxResults, pageToken, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	comments := make([]map[string]any, 0, len(resp.Comments))
	for _, c := range resp.Comments {
		comments = append(comments, formatComment(c))
	}

	result := map[string]any{
		"file_id":         fileID,
		"comments":        comments,
		"count":           len(comments),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGetComment gets a single comment.
func TestableDriveGetComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	comment, err := srv.GetComment(ctx, fileID, commentID, DriveCommentFields, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(comment)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveCreateComment creates a comment on a file.
func TestableDriveCreateComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	comment := &drive.Comment{
		Content: content,
	}

	created, err := srv.CreateComment(ctx, fileID, comment, DriveCommentFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(created)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveUpdateComment updates a comment.
func TestableDriveUpdateComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	comment := &drive.Comment{
		Content: content,
	}

	updated, err := srv.UpdateComment(ctx, fileID, commentID, comment, DriveCommentFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatComment(updated)
	result["file_id"] = fileID

	return common.MarshalToolResult(result)
}

// TestableDriveDeleteComment deletes a comment.
func TestableDriveDeleteComment(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	err := srv.DeleteComment(ctx, fileID, commentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success":    true,
		"file_id":    fileID,
		"comment_id": commentID,
		"message":    "Comment deleted",
	}

	return common.MarshalToolResult(result)
}

// TestableDriveListReplies lists replies on a comment.
func TestableDriveListReplies(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	maxResults := common.ParseMaxResults(request.Params.Arguments, common.DriveSearchDefaultMaxResults, common.DriveSearchMaxResultsLimit)
	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	includeDeleted := common.ParseBoolArg(request.Params.Arguments, "include_deleted", false)

	resp, err := srv.ListReplies(ctx, fileID, commentID, DriveReplyListFields, maxResults, pageToken, includeDeleted)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	replies := make([]map[string]any, 0, len(resp.Replies))
	for _, r := range resp.Replies {
		replies = append(replies, formatReply(r))
	}

	result := map[string]any{
		"file_id":         fileID,
		"comment_id":      commentID,
		"replies":         replies,
		"count":           len(replies),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableDriveCreateReply creates a reply on a comment.
func TestableDriveCreateReply(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	commentID, errResult := common.RequireStringArg(request.Params.Arguments, "comment_id")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := common.RequireStringArg(request.Params.Arguments, "content")
	if errResult != nil {
		return errResult, nil
	}

	reply := &drive.Reply{
		Content: content,
	}

	// Support resolve/reopen action
	action := common.ParseStringArg(request.Params.Arguments, "action", "")
	if action != "" {
		if action != "resolve" && action != "reopen" {
			return mcp.NewToolResultError("invalid action: must be 'resolve' or 'reopen'"), nil
		}
		reply.Action = action
	}

	replyFields := "id,content,author(displayName,emailAddress),createdTime,modifiedTime,action"
	created, err := srv.CreateReply(ctx, fileID, commentID, reply, replyFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := formatReply(created)
	result["file_id"] = fileID
	result["comment_id"] = commentID

	return common.MarshalToolResult(result)
}
