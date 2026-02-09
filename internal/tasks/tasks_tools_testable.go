package tasks

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/tasks/v1"
)

// TestableTasksListTasklists is the testable version of handleTasksListTasklists.
func TestableTasksListTasklists(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	opts := &ListTaskListsOptions{}
	opts.MaxResults = common.ParseMaxResults(request.Params.Arguments, common.TasksDefaultMaxResults, common.TasksMaxResultsLimit)
	opts.PageToken = common.ParseStringArg(request.Params.Arguments, "page_token", "")

	resp, err := srv.ListTaskLists(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	taskLists := make([]map[string]any, 0, len(resp.Items))
	for _, tl := range resp.Items {
		taskLists = append(taskLists, map[string]any{
			"id":      tl.Id,
			"title":   tl.Title,
			"updated": tl.Updated,
		})
	}

	result := map[string]any{
		"task_lists":      taskLists,
		"count":           len(taskLists),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableTasksList is the testable version of handleTasksList.
func TestableTasksList(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	opts := &ListTasksOptions{}
	opts.MaxResults = common.ParseMaxResults(request.Params.Arguments, common.TasksDefaultMaxResults, common.TasksMaxResultsLimit)
	opts.PageToken = common.ParseStringArg(request.Params.Arguments, "page_token", "")
	opts.ShowCompleted = common.ParseBoolArg(request.Params.Arguments, "show_completed", true)
	if common.ParseBoolArg(request.Params.Arguments, "show_hidden", false) {
		opts.ShowHidden = true
	}
	opts.DueMin = common.ParseStringArg(request.Params.Arguments, "due_min", "")
	opts.DueMax = common.ParseStringArg(request.Params.Arguments, "due_max", "")

	resp, err := srv.ListTasks(ctx, taskListID, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	taskItems := make([]map[string]any, 0, len(resp.Items))
	for _, t := range resp.Items {
		taskItems = append(taskItems, formatTask(t))
	}

	result := map[string]any{
		"tasks":           taskItems,
		"count":           len(taskItems),
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableTasksGet is the testable version of handleTasksGet.
func TestableTasksGet(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskID := common.ParseStringArg(request.Params.Arguments, "task_id", "")
	if taskID == "" {
		return mcp.NewToolResultError("task_id parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	task, err := srv.GetTask(ctx, taskListID, taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := formatTaskFull(task)

	return common.MarshalToolResult(result)
}

// TestableTasksCreate is the testable version of handleTasksCreate.
func TestableTasksCreate(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title := common.ParseStringArg(request.Params.Arguments, "title", "")
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	task := &tasks.Task{
		Title: title,
	}

	if notes := common.ParseStringArg(request.Params.Arguments, "notes", ""); notes != "" {
		task.Notes = notes
	}

	if due := common.ParseStringArg(request.Params.Arguments, "due", ""); due != "" {
		task.Due = due
	}

	var createOpts *CreateTaskOptions
	if parent := common.ParseStringArg(request.Params.Arguments, "parent", ""); parent != "" {
		createOpts = &CreateTaskOptions{Parent: parent}
	}

	created, err := srv.CreateTask(ctx, taskListID, task, createOpts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := formatTask(created)

	return common.MarshalToolResult(result)
}

// TestableTasksUpdate is the testable version of handleTasksUpdate.
func TestableTasksUpdate(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskID := common.ParseStringArg(request.Params.Arguments, "task_id", "")
	if taskID == "" {
		return mcp.NewToolResultError("task_id parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	// First, get the existing task
	task, err := srv.GetTask(ctx, taskListID, taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task: %v", err)), nil
	}

	// Update fields that are provided
	if title := common.ParseStringArg(request.Params.Arguments, "title", ""); title != "" {
		task.Title = title
	}
	// For notes and due, we need to handle empty string as "no change" if the key is missing,
	// but common.ParseStringArg handles that by checking for existence.
	// However, if the user explicitly passes empty string to clear it, we might want to allow it.
	// But the current implementation logic was:
	// if notes, ok := request.Params.Arguments["notes"].(string); ok { task.Notes = notes }
	// This means if "notes" is present (even empty), it updates.
	// ParseStringArg returns default if missing OR empty. This might be a slight behavior change if empty string was allowed.
	// But usually empty string means clear.
	// Let's use direct map access for update to be safe or update ParseStringArg to support "present check".
	// Or just use ParseStringArg but be aware.
	// Actually, `ParseStringArg` returns defaultVal if `val == ""`.
	// If I want to support clearing, I need to know if it was passed.
	// Let's stick to the original logic for updates where empty string matters.

	if val, ok := request.Params.Arguments["notes"].(string); ok {
		task.Notes = val
	}
	if val, ok := request.Params.Arguments["due"].(string); ok {
		task.Due = val
	}

	updated, err := srv.UpdateTask(ctx, taskListID, taskID, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := formatTask(updated)

	return common.MarshalToolResult(result)
}

// TestableTasksComplete is the testable version of handleTasksComplete.
func TestableTasksComplete(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskID := common.ParseStringArg(request.Params.Arguments, "task_id", "")
	if taskID == "" {
		return mcp.NewToolResultError("task_id parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	// Get the existing task
	task, err := srv.GetTask(ctx, taskListID, taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task: %v", err)), nil
	}

	// Mark as completed
	task.Status = "completed"

	updated, err := srv.UpdateTask(ctx, taskListID, taskID, task)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := formatTask(updated)
	result["message"] = "Task marked as completed"

	return common.MarshalToolResult(result)
}

// TestableTasksDelete is the testable version of handleTasksDelete.
func TestableTasksDelete(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskID := common.ParseStringArg(request.Params.Arguments, "task_id", "")
	if taskID == "" {
		return mcp.NewToolResultError("task_id parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	err := srv.DeleteTask(ctx, taskListID, taskID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"task_id":     taskID,
		"tasklist_id": taskListID,
		"message":     "Task deleted successfully",
	}

	return common.MarshalToolResult(result)
}

// === Phase 2 Testable Functions: Task List Management ===

// TestableTasksCreateTasklist is the testable version of handleTasksCreateTasklist.
func TestableTasksCreateTasklist(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title := common.ParseStringArg(request.Params.Arguments, "title", "")
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	taskList := &tasks.TaskList{
		Title: title,
	}

	created, err := srv.CreateTaskList(ctx, taskList)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := map[string]any{
		"id":      created.Id,
		"title":   created.Title,
		"updated": created.Updated,
		"message": "Task list created successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableTasksUpdateTasklist is the testable version of handleTasksUpdateTasklist.
func TestableTasksUpdateTasklist(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", "")
	if taskListID == "" {
		return mcp.NewToolResultError("tasklist_id parameter is required"), nil
	}

	title := common.ParseStringArg(request.Params.Arguments, "title", "")
	if title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	// Get the existing task list
	taskList, err := srv.GetTaskList(ctx, taskListID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get task list: %v", err)), nil
	}

	// Update the title
	taskList.Title = title

	updated, err := srv.UpdateTaskList(ctx, taskListID, taskList)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := map[string]any{
		"id":      updated.Id,
		"title":   updated.Title,
		"updated": updated.Updated,
		"message": "Task list updated successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableTasksDeleteTasklist is the testable version of handleTasksDeleteTasklist.
func TestableTasksDeleteTasklist(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", "")
	if taskListID == "" {
		return mcp.NewToolResultError("tasklist_id parameter is required"), nil
	}

	err := srv.DeleteTaskList(ctx, taskListID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"tasklist_id": taskListID,
		"message":     "Task list deleted successfully",
	}

	return common.MarshalToolResult(result)
}

// === Phase 2 Testable Functions: Task Organization ===

// TestableTasksMove is the testable version of handleTasksMove.
func TestableTasksMove(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskID := common.ParseStringArg(request.Params.Arguments, "task_id", "")
	if taskID == "" {
		return mcp.NewToolResultError("task_id parameter is required"), nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	opts := &MoveTaskOptions{}

	// Optional: set parent (for making subtasks)
	if parent := common.ParseStringArg(request.Params.Arguments, "parent", ""); parent != "" {
		opts.Parent = parent
	}

	// Optional: set previous sibling (for ordering)
	if previous := common.ParseStringArg(request.Params.Arguments, "previous", ""); previous != "" {
		opts.Previous = previous
	}

	moved, err := srv.MoveTask(ctx, taskListID, taskID, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := formatTask(moved)
	result["message"] = "Task moved successfully"

	return common.MarshalToolResult(result)
}

// TestableTasksClearCompleted is the testable version of handleTasksClearCompleted.
func TestableTasksClearCompleted(ctx context.Context, request mcp.CallToolRequest, deps *TasksHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveTasksServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	taskListID := common.ParseStringArg(request.Params.Arguments, "tasklist_id", common.DefaultTaskListID)

	err := srv.ClearCompleted(ctx, taskListID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Tasks API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		"tasklist_id": taskListID,
		"message":     "Completed tasks cleared successfully",
	}

	return common.MarshalToolResult(result)
}
