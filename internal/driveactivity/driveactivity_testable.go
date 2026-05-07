package driveactivity

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/driveactivity/v2"
)

// maxActivityPages limits pagination to prevent unbounded memory growth.
const maxActivityPages = 10

// TestableDriveActivityQuery queries activity for a Drive item or folder.
func TestableDriveActivityQuery(ctx context.Context, request mcp.CallToolRequest, deps *DriveActivityHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveActivityServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	itemID := common.ParseStringArg(request.GetArguments(), "item_id", "")
	folderID := common.ParseStringArg(request.GetArguments(), "folder_id", "")
	filter := common.ParseStringArg(request.GetArguments(), "filter", "")
	pageToken := common.ParseStringArg(request.GetArguments(), "page_token", "")
	pageSize := common.ParseMaxResults(request.GetArguments(), 20, 100)

	if itemID == "" && folderID == "" {
		return mcp.NewToolResultError("either item_id or folder_id parameter is required"), nil
	}

	// Normalize IDs: extract from URLs if needed
	if itemID != "" {
		itemID = common.ExtractGoogleResourceID(itemID)
	}
	if folderID != "" {
		folderID = common.ExtractGoogleResourceID(folderID)
	}

	req := &driveactivity.QueryDriveActivityRequest{
		PageSize: pageSize,
	}

	if itemID != "" {
		req.ItemName = fmt.Sprintf("items/%s", itemID)
	}
	if folderID != "" {
		req.AncestorName = fmt.Sprintf("items/%s", folderID)
	}
	if filter != "" {
		req.Filter = filter
	}
	if pageToken != "" {
		req.PageToken = pageToken
	}

	// Consolidate related actions for cleaner output
	req.ConsolidationStrategy = &driveactivity.ConsolidationStrategy{
		Legacy: &driveactivity.Legacy{},
	}

	var allActivities []*driveactivity.DriveActivity
	var nextPageToken string

	for page := 0; page < maxActivityPages; page++ {
		resp, err := srv.QueryActivity(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Drive Activity API error: %v", err)), nil
		}

		allActivities = append(allActivities, resp.Activities...)
		nextPageToken = resp.NextPageToken

		if resp.NextPageToken == "" || int64(len(allActivities)) >= pageSize {
			break
		}
		req.PageToken = resp.NextPageToken
	}

	// Format activities for output
	formattedActivities := make([]map[string]any, 0, len(allActivities))
	for _, activity := range allActivities {
		formattedActivities = append(formattedActivities, formatActivity(activity))
	}

	result := map[string]any{
		"activity_count": len(formattedActivities),
		"activities":     formattedActivities,
	}

	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// formatActivity formats a single DriveActivity for output.
func formatActivity(activity *driveactivity.DriveActivity) map[string]any {
	result := map[string]any{}

	// Timestamp
	if activity.Timestamp != "" {
		result["timestamp"] = activity.Timestamp
	}
	if activity.TimeRange != nil {
		result["time_range"] = map[string]any{
			"start_time": activity.TimeRange.StartTime,
			"end_time":   activity.TimeRange.EndTime,
		}
	}

	// Primary action
	if activity.PrimaryActionDetail != nil {
		result["action"] = formatActionDetail(activity.PrimaryActionDetail)
	}

	// Actors
	if len(activity.Actors) > 0 {
		actors := make([]map[string]any, 0, len(activity.Actors))
		for _, actor := range activity.Actors {
			actors = append(actors, formatActor(actor))
		}
		result["actors"] = actors
	}

	// Targets
	if len(activity.Targets) > 0 {
		targets := make([]map[string]any, 0, len(activity.Targets))
		for _, target := range activity.Targets {
			targets = append(targets, formatTarget(target))
		}
		result["targets"] = targets
	}

	return result
}

// formatActionDetail formats an ActionDetail into a human-readable representation.
func formatActionDetail(detail *driveactivity.ActionDetail) string {
	switch {
	case detail.Create != nil:
		return "CREATE"
	case detail.Edit != nil:
		return "EDIT"
	case detail.Move != nil:
		return "MOVE"
	case detail.Rename != nil:
		action := "RENAME"
		if detail.Rename.OldTitle != "" && detail.Rename.NewTitle != "" {
			action = fmt.Sprintf("RENAME: %s -> %s", detail.Rename.OldTitle, detail.Rename.NewTitle)
		}
		return action
	case detail.Delete != nil:
		return "DELETE"
	case detail.Restore != nil:
		return "RESTORE"
	case detail.PermissionChange != nil:
		return "PERMISSION_CHANGE"
	case detail.Comment != nil:
		return "COMMENT"
	case detail.DlpChange != nil:
		return "DLP_CHANGE"
	case detail.Reference != nil:
		return "REFERENCE"
	case detail.SettingsChange != nil:
		return "SETTINGS_CHANGE"
	case detail.AppliedLabelChange != nil:
		return "LABEL_CHANGE"
	default:
		return "UNKNOWN"
	}
}

// formatActor formats an Actor for output.
func formatActor(actor *driveactivity.Actor) map[string]any {
	result := map[string]any{}

	switch {
	case actor.User != nil:
		result["type"] = "user"
		if actor.User.KnownUser != nil {
			result["person_name"] = actor.User.KnownUser.PersonName
			result["is_current_user"] = actor.User.KnownUser.IsCurrentUser
		} else if actor.User.DeletedUser != nil {
			result["deleted"] = true
		} else if actor.User.UnknownUser != nil {
			result["unknown"] = true
		}
	case actor.Administrator != nil:
		result["type"] = "administrator"
	case actor.System != nil:
		result["type"] = "system"
	case actor.Anonymous != nil:
		result["type"] = "anonymous"
	case actor.Impersonation != nil:
		result["type"] = "impersonation"
	}

	return result
}

// formatTarget formats a Target for output.
func formatTarget(target *driveactivity.Target) map[string]any {
	result := map[string]any{}

	switch {
	case target.DriveItem != nil:
		result["type"] = "drive_item"
		// Extract item ID from "items/ITEM_ID" format
		name := target.DriveItem.Name
		if strings.HasPrefix(name, "items/") {
			result["item_id"] = strings.TrimPrefix(name, "items/")
		} else {
			result["item_id"] = name
		}
		result["title"] = target.DriveItem.Title
		if target.DriveItem.MimeType != "" {
			result["mime_type"] = target.DriveItem.MimeType
		}
	case target.Drive != nil:
		result["type"] = "shared_drive"
		result["name"] = target.Drive.Name
		result["title"] = target.Drive.Title
	case target.FileComment != nil:
		result["type"] = "file_comment"
		if target.FileComment.LegacyDiscussionId != "" {
			result["discussion_id"] = target.FileComment.LegacyDiscussionId
		}
	}

	return result
}
