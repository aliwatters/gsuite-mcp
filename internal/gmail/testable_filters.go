package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailListFilters lists all filters.
func TestableGmailListFilters(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resp, err := svc.ListFilters(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	filters := make([]map[string]any, 0, len(resp.Filter))
	for _, f := range resp.Filter {
		filter := map[string]any{
			"id": f.Id,
		}
		if f.Criteria != nil {
			criteria := map[string]any{}
			if f.Criteria.From != "" {
				criteria["from"] = f.Criteria.From
			}
			if f.Criteria.To != "" {
				criteria["to"] = f.Criteria.To
			}
			if f.Criteria.Subject != "" {
				criteria["subject"] = f.Criteria.Subject
			}
			if f.Criteria.Query != "" {
				criteria["query"] = f.Criteria.Query
			}
			if f.Criteria.NegatedQuery != "" {
				criteria["negated_query"] = f.Criteria.NegatedQuery
			}
			if f.Criteria.HasAttachment {
				criteria["has_attachment"] = true
			}
			filter["criteria"] = criteria
		}
		if f.Action != nil {
			action := map[string]any{}
			if len(f.Action.AddLabelIds) > 0 {
				action["add_labels"] = f.Action.AddLabelIds
			}
			if len(f.Action.RemoveLabelIds) > 0 {
				action["remove_labels"] = f.Action.RemoveLabelIds
			}
			if f.Action.Forward != "" {
				action["forward"] = f.Action.Forward
			}
			filter["action"] = action
		}
		filters = append(filters, filter)
	}

	result := map[string]any{
		"filters": filters,
		"count":   len(filters),
	}

	return common.MarshalToolResult(result)
}

// TestableGmailCreateFilter creates a new filter.
func TestableGmailCreateFilter(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	filter := &gmail.Filter{
		Criteria: &gmail.FilterCriteria{},
		Action:   &gmail.FilterAction{},
	}

	// Criteria
	if val := common.ParseStringArg(request.Params.Arguments, "from", ""); val != "" {
		filter.Criteria.From = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "to", ""); val != "" {
		filter.Criteria.To = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "subject", ""); val != "" {
		filter.Criteria.Subject = val
	}
	if val := common.ParseStringArg(request.Params.Arguments, "query", ""); val != "" {
		filter.Criteria.Query = val
	}
	if val := common.ParseBoolArg(request.Params.Arguments, "has_attachment", false); val {
		filter.Criteria.HasAttachment = true
	}

	// Action
	if add, ok := request.Params.Arguments["add_label_ids"].([]any); ok {
		for _, l := range add {
			if s, ok := l.(string); ok {
				filter.Action.AddLabelIds = append(filter.Action.AddLabelIds, s)
			}
		}
	}
	if remove, ok := request.Params.Arguments["remove_label_ids"].([]any); ok {
		for _, l := range remove {
			if s, ok := l.(string); ok {
				filter.Action.RemoveLabelIds = append(filter.Action.RemoveLabelIds, s)
			}
		}
	}
	if val := common.ParseStringArg(request.Params.Arguments, "forward", ""); val != "" {
		filter.Action.Forward = val
	}

	created, err := svc.CreateFilter(ctx, filter)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id": created.Id,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailDeleteFilter deletes a filter.
func TestableGmailDeleteFilter(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	filterID := common.ParseStringArg(request.Params.Arguments, "filter_id", "")
	if filterID == "" {
		return mcp.NewToolResultError("filter_id parameter is required"), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	err := svc.DeleteFilter(ctx, filterID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":      filterID,
		"deleted": true,
	}

	return common.MarshalToolResult(result)
}
