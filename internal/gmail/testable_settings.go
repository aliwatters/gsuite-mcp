package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailGetProfile gets the user's profile.
func TestableGmailGetProfile(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	profile, err := svc.GetProfile(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"email_address":  profile.EmailAddress,
		"messages_total": profile.MessagesTotal,
		"threads_total":  profile.ThreadsTotal,
		"history_id":     profile.HistoryId,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetVacation gets vacation settings.
func TestableGmailGetVacation(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	settings, err := svc.GetVacationSettings(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"enabled":              settings.EnableAutoReply,
		"subject":              settings.ResponseSubject,
		"body_plain_text":      settings.ResponseBodyPlainText,
		"body_html":            settings.ResponseBodyHtml,
		"restrict_to_contacts": settings.RestrictToContacts,
		"restrict_to_domain":   settings.RestrictToDomain,
	}

	if settings.StartTime > 0 {
		result["start_time"] = settings.StartTime
	}
	if settings.EndTime > 0 {
		result["end_time"] = settings.EndTime
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSetVacation sets vacation settings.
func TestableGmailSetVacation(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	settings := &gmail.VacationSettings{}

	if val, ok := request.Params.Arguments["enabled"].(bool); ok {
		settings.EnableAutoReply = val
	}
	if val, ok := request.Params.Arguments["subject"].(string); ok {
		settings.ResponseSubject = val
	}
	if val, ok := request.Params.Arguments["body"].(string); ok {
		settings.ResponseBodyPlainText = val
	}
	if val, ok := request.Params.Arguments["body_html"].(string); ok {
		settings.ResponseBodyHtml = val
	}
	if val, ok := request.Params.Arguments["restrict_to_contacts"].(bool); ok {
		settings.RestrictToContacts = val
	}
	if val, ok := request.Params.Arguments["restrict_to_domain"].(bool); ok {
		settings.RestrictToDomain = val
	}
	if val, ok := request.Params.Arguments["start_time"].(float64); ok {
		settings.StartTime = int64(val)
	}
	if val, ok := request.Params.Arguments["end_time"].(float64); ok {
		settings.EndTime = int64(val)
	}

	updated, err := svc.UpdateVacationSettings(ctx, settings)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"enabled": updated.EnableAutoReply,
		"subject": updated.ResponseSubject,
		"success": true,
	}

	return common.MarshalToolResult(result)
}
