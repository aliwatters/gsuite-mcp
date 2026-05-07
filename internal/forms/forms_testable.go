package forms

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// formsEditURLFormat is the URL template for Google Forms edit links.
const formsEditURLFormat = "https://docs.google.com/forms/d/%s/edit"

// extractRequiredFormID extracts, validates, and normalizes the form_id parameter.
func extractRequiredFormID(request mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	formID := common.ParseStringArg(request.GetArguments(), "form_id", "")
	if formID == "" {
		return "", mcp.NewToolResultError("form_id parameter is required")
	}
	return common.ExtractGoogleResourceID(formID), nil
}

// TestableFormsGet retrieves a form's metadata, items, and structure.
func TestableFormsGet(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveFormsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	formID, errResult := extractRequiredFormID(request)
	if errResult != nil {
		return errResult, nil
	}

	form, err := srv.GetForm(ctx, formID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Forms API error: %v", err)), nil
	}

	// Format items summary
	items := make([]map[string]any, 0, len(form.Items))
	for _, item := range form.Items {
		items = append(items, formatItem(item))
	}

	result := map[string]any{
		"form_id":    form.FormId,
		"title":      form.Info.Title,
		"edit_url":   fmt.Sprintf(formsEditURLFormat, form.FormId),
		"item_count": len(form.Items),
		"items":      items,
	}

	if form.Info.Description != "" {
		result["description"] = form.Info.Description
	}
	if form.Info.DocumentTitle != "" {
		result["document_title"] = form.Info.DocumentTitle
	}
	if form.ResponderUri != "" {
		result["responder_url"] = form.ResponderUri
	}
	if form.LinkedSheetId != "" {
		result["linked_sheet_id"] = form.LinkedSheetId
	}

	return common.MarshalToolResult(result)
}

// TestableFormsCreate creates a new form.
func TestableFormsCreate(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveFormsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title, errResult := common.RequireStringArg(request.GetArguments(), "title")
	if errResult != nil {
		return errResult, nil
	}

	form, err := srv.CreateForm(ctx, title)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Forms API error: %v", err)), nil
	}

	result := map[string]any{
		"form_id":  form.FormId,
		"title":    form.Info.Title,
		"edit_url": fmt.Sprintf(formsEditURLFormat, form.FormId),
	}

	if form.ResponderUri != "" {
		result["responder_url"] = form.ResponderUri
	}

	return common.MarshalToolResult(result)
}

// TestableFormsBatchUpdate performs a batch update on a form.
func TestableFormsBatchUpdate(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveFormsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	formID, errResult := extractRequiredFormID(request)
	if errResult != nil {
		return errResult, nil
	}

	requestsJSON, errResult := common.RequireStringArg(request.GetArguments(), "requests")
	if errResult != nil {
		return errResult, nil
	}

	requests, err := parseBatchUpdateRequests(requestsJSON)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid requests JSON: %v", err)), nil
	}

	resp, err := srv.BatchUpdate(ctx, formID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Forms API error: %v", err)), nil
	}

	result := map[string]any{
		"form_id":       formID,
		"replies_count": len(resp.Replies),
	}

	return common.MarshalToolResult(result)
}

// TestableFormsListResponses lists all responses for a form.
func TestableFormsListResponses(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveFormsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	formID, errResult := extractRequiredFormID(request)
	if errResult != nil {
		return errResult, nil
	}

	responses, err := srv.ListResponses(ctx, formID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Forms API error: %v", err)), nil
	}

	formattedResponses := make([]map[string]any, 0, len(responses))
	for _, resp := range responses {
		formattedResponses = append(formattedResponses, formatResponse(resp))
	}

	result := map[string]any{
		"form_id":        formID,
		"response_count": len(responses),
		"responses":      formattedResponses,
	}

	return common.MarshalToolResult(result)
}

// TestableFormsGetResponse retrieves a single form response by ID.
func TestableFormsGetResponse(ctx context.Context, request mcp.CallToolRequest, deps *FormsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveFormsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	formID, errResult := extractRequiredFormID(request)
	if errResult != nil {
		return errResult, nil
	}

	responseID, errResult := common.RequireStringArg(request.GetArguments(), "response_id")
	if errResult != nil {
		return errResult, nil
	}

	resp, err := srv.GetResponse(ctx, formID, responseID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Forms API error: %v", err)), nil
	}

	result := formatResponse(resp)
	result["form_id"] = formID

	return common.MarshalToolResult(result)
}
