package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// === Send-As Testable Functions ===

// TestableGmailListSendAs lists all send-as aliases for the account.
func TestableGmailListSendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	aliases, err := svc.ListSendAs(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	results := make([]map[string]any, 0, len(aliases))
	for _, sa := range aliases {
		results = append(results, formatSendAs(sa))
	}

	return common.MarshalToolResult(map[string]any{
		"send_as_count": len(aliases),
		"send_as":       results,
	})
}

// TestableGmailGetSendAs gets a specific send-as alias.
func TestableGmailGetSendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "send_as_email")
	if errResult != nil {
		return errResult, nil
	}

	sa, err := svc.GetSendAs(ctx, email)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatSendAs(sa))
}

// TestableGmailCreateSendAs creates a new send-as alias.
func TestableGmailCreateSendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "send_as_email")
	if errResult != nil {
		return errResult, nil
	}

	sendAs := &gmail.SendAs{
		SendAsEmail: email,
	}

	if val := common.ParseStringArg(request.Params.Arguments, "display_name", ""); val != "" {
		sendAs.DisplayName = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "reply_to_address", ""); val != "" {
		sendAs.ReplyToAddress = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "signature", ""); val != "" {
		sendAs.Signature = val
	}
	if val, ok := request.Params.Arguments["is_default"].(bool); ok && val {
		sendAs.IsDefault = val
	}

	// SMTP configuration for external send-as
	smtpHost := common.ParseStringArg(request.Params.Arguments, "smtp_host", "")
	if smtpHost != "" {
		sendAs.SmtpMsa = &gmail.SmtpMsa{
			Host: smtpHost,
		}
		if port, ok := request.Params.Arguments["smtp_port"].(float64); ok {
			sendAs.SmtpMsa.Port = int64(port)
		}
		if val := common.ParseStringArg(request.Params.Arguments, "smtp_username", ""); val != "" {
			sendAs.SmtpMsa.Username = val
		}
		if val := common.ParseStringArg(request.Params.Arguments, "smtp_password", ""); val != "" {
			sendAs.SmtpMsa.Password = val
		}
		if val := common.ParseStringArg(request.Params.Arguments, "smtp_security_mode", ""); val != "" {
			sendAs.SmtpMsa.SecurityMode = val
		}
	}

	created, err := svc.CreateSendAs(ctx, sendAs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := formatSendAs(created)
	result["success"] = true

	return common.MarshalToolResult(result)
}

// TestableGmailUpdateSendAs updates an existing send-as alias.
func TestableGmailUpdateSendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "send_as_email")
	if errResult != nil {
		return errResult, nil
	}

	sendAs := &gmail.SendAs{}

	if val := common.ParseStringArg(request.Params.Arguments, "display_name", ""); val != "" {
		sendAs.DisplayName = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "reply_to_address", ""); val != "" {
		sendAs.ReplyToAddress = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "signature", ""); val != "" {
		sendAs.Signature = val
	}
	if val, ok := request.Params.Arguments["is_default"].(bool); ok {
		sendAs.IsDefault = val
	}

	updated, err := svc.UpdateSendAs(ctx, email, sendAs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := formatSendAs(updated)
	result["success"] = true

	return common.MarshalToolResult(result)
}

// TestableGmailDeleteSendAs deletes a send-as alias.
func TestableGmailDeleteSendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "send_as_email")
	if errResult != nil {
		return errResult, nil
	}

	err := svc.DeleteSendAs(ctx, email)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{
		"success":       true,
		"send_as_email": email,
		"action":        "deleted",
	})
}

// TestableGmailVerifySendAs initiates verification for a send-as alias.
func TestableGmailVerifySendAs(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "send_as_email")
	if errResult != nil {
		return errResult, nil
	}

	err := svc.VerifySendAs(ctx, email)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{
		"success":       true,
		"send_as_email": email,
		"action":        "verification_sent",
	})
}

// === Delegate Testable Functions ===

// TestableGmailListDelegates lists all delegates for the account.
func TestableGmailListDelegates(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	delegates, err := svc.ListDelegates(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	results := make([]map[string]any, 0, len(delegates))
	for _, d := range delegates {
		results = append(results, formatDelegate(d))
	}

	return common.MarshalToolResult(map[string]any{
		"delegate_count": len(delegates),
		"delegates":      results,
	})
}

// TestableGmailCreateDelegate adds a new delegate to the account.
func TestableGmailCreateDelegate(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "delegate_email")
	if errResult != nil {
		return errResult, nil
	}

	delegate := &gmail.Delegate{
		DelegateEmail: email,
	}

	created, err := svc.CreateDelegate(ctx, delegate)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := formatDelegate(created)
	result["success"] = true

	return common.MarshalToolResult(result)
}

// TestableGmailDeleteDelegate removes a delegate from the account.
func TestableGmailDeleteDelegate(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	email, errResult := common.RequireStringArg(request.Params.Arguments, "delegate_email")
	if errResult != nil {
		return errResult, nil
	}

	err := svc.DeleteDelegate(ctx, email)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{
		"success":        true,
		"delegate_email": email,
		"action":         "deleted",
	})
}

// === Format Helpers ===

// formatSendAs formats a SendAs alias for MCP output.
func formatSendAs(sa *gmail.SendAs) map[string]any {
	result := map[string]any{
		"send_as_email": sa.SendAsEmail,
		"display_name":  sa.DisplayName,
		"is_default":    sa.IsDefault,
		"is_primary":    sa.IsPrimary,
	}

	if sa.ReplyToAddress != "" {
		result["reply_to_address"] = sa.ReplyToAddress
	}
	if sa.Signature != "" {
		result["signature"] = sa.Signature
	}
	if sa.VerificationStatus != "" {
		result["verification_status"] = sa.VerificationStatus
	}
	if sa.SmtpMsa != nil {
		result["smtp_msa"] = map[string]any{
			"host":          sa.SmtpMsa.Host,
			"port":          sa.SmtpMsa.Port,
			"security_mode": sa.SmtpMsa.SecurityMode,
		}
	}

	return result
}

// formatDelegate formats a Delegate for MCP output.
func formatDelegate(d *gmail.Delegate) map[string]any {
	return map[string]any{
		"delegate_email":      d.DelegateEmail,
		"verification_status": d.VerificationStatus,
	}
}
