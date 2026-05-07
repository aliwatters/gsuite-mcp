package drive

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/drive/v3"
)

// TestableDriveShare shares a file with users.
func TestableDriveShare(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	email, errResult := common.RequireStringArg(request.GetArguments(), "email")
	if errResult != nil {
		return errResult, nil
	}

	role := common.ParseStringArg(request.GetArguments(), "role", "reader")
	permType := common.ParseStringArg(request.GetArguments(), "type", "user")

	permission := &drive.Permission{
		Type:         permType,
		Role:         role,
		EmailAddress: email,
	}

	sendNotification := common.ParseBoolArg(request.GetArguments(), "send_notification", true)

	created, err := srv.CreatePermission(ctx, fileID, permission, sendNotification)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"success":       true,
		"file_id":       fileID,
		"permission_id": created.Id,
		"type":          created.Type,
		"role":          created.Role,
		"email":         created.EmailAddress,
		"message":       fmt.Sprintf("File shared with %s as %s", email, role),
	}

	return common.MarshalToolResult(result)
}

// TestableDriveGetShareableLink gets a shareable link and sharing status for a file.
func TestableDriveGetShareableLink(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	file, err := srv.GetFile(ctx, fileID, DriveShareableLinkFields)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	result := map[string]any{
		"file_id":   file.Id,
		"name":      file.Name,
		"mime_type": file.MimeType,
		"url":       file.WebViewLink,
	}

	if file.WebContentLink != "" {
		result["download_url"] = file.WebContentLink
	}

	// Include permissions/sharing status
	permissions := make([]map[string]any, 0, len(file.Permissions))
	for _, p := range file.Permissions {
		perm := map[string]any{
			"id":   p.Id,
			"type": p.Type,
			"role": p.Role,
		}
		if p.EmailAddress != "" {
			perm["email"] = p.EmailAddress
		}
		if p.DisplayName != "" {
			perm["display_name"] = p.DisplayName
		}
		if p.Domain != "" {
			perm["domain"] = p.Domain
		}
		permissions = append(permissions, perm)
	}
	result["permissions"] = permissions
	result["permission_count"] = len(permissions)

	return common.MarshalToolResult(result)
}

// TestableDriveGetPermissions gets file permissions.
func TestableDriveGetPermissions(ctx context.Context, request mcp.CallToolRequest, deps *DriveHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDriveServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	fileID, idErrResult := extractRequiredFileID(request)
	if idErrResult != nil {
		return idErrResult, nil
	}

	resp, err := srv.ListPermissions(ctx, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error: %v", err)), nil
	}

	permissions := make([]map[string]any, 0, len(resp.Permissions))
	for _, p := range resp.Permissions {
		perm := map[string]any{
			"id":   p.Id,
			"type": p.Type,
			"role": p.Role,
		}
		if p.EmailAddress != "" {
			perm["email"] = p.EmailAddress
		}
		if p.DisplayName != "" {
			perm["display_name"] = p.DisplayName
		}
		if p.Domain != "" {
			perm["domain"] = p.Domain
		}
		permissions = append(permissions, perm)
	}

	result := map[string]any{
		"file_id":     fileID,
		"permissions": permissions,
		"count":       len(permissions),
	}

	return common.MarshalToolResult(result)
}
