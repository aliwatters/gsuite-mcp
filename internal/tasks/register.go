package tasks

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Tasks tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Tasks Core (Phase 1) ===

	// tasks_list_tasklists - List all task lists
	s.AddTool(mcp.NewTool("tasks_list_tasklists",
		mcp.WithDescription("List all Google Tasks task lists."),
		mcp.WithNumber("max_results", mcp.Description("Maximum results to return (1-100, default 100)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleTasksListTasklists)

	// tasks_list - List tasks in a task list
	s.AddTool(mcp.NewTool("tasks_list",
		mcp.WithDescription("List tasks in a Google Tasks task list."),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default' for primary list)")),
		mcp.WithNumber("max_results", mcp.Description("Maximum results to return (1-100, default 100)")),
		common.WithPageToken(),
		mcp.WithBoolean("show_completed", mcp.Description("Include completed tasks (default: true)")),
		mcp.WithBoolean("show_hidden", mcp.Description("Include hidden tasks (default: false)")),
		mcp.WithString("due_min", mcp.Description("Filter tasks due on or after this date (RFC3339)")),
		mcp.WithString("due_max", mcp.Description("Filter tasks due on or before this date (RFC3339)")),
		common.WithAccountParam(),
	), HandleTasksList)

	// tasks_get - Get task details
	s.AddTool(mcp.NewTool("tasks_get",
		mcp.WithDescription("Get full details of a Google Task."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		common.WithAccountParam(),
	), HandleTasksGet)

	// tasks_create - Create a new task
	s.AddTool(mcp.NewTool("tasks_create",
		mcp.WithDescription("Create a new Google Task."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("notes", mcp.Description("Task notes/description")),
		mcp.WithString("due", mcp.Description("Due date (RFC3339 format, e.g., 2024-02-15T00:00:00Z)")),
		mcp.WithString("parent", mcp.Description("Parent task ID (creates as subtask)")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		common.WithAccountParam(),
	), HandleTasksCreate)

	// tasks_update - Update task
	s.AddTool(mcp.NewTool("tasks_update",
		mcp.WithDescription("Update a Google Task (title, notes, due date)."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("title", mcp.Description("New task title")),
		mcp.WithString("notes", mcp.Description("New task notes/description")),
		mcp.WithString("due", mcp.Description("New due date (RFC3339 format)")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		common.WithAccountParam(),
	), HandleTasksUpdate)

	// tasks_complete - Mark task as completed
	s.AddTool(mcp.NewTool("tasks_complete",
		mcp.WithDescription("Mark a Google Task as completed."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		common.WithAccountParam(),
	), HandleTasksComplete)

	// tasks_delete - Delete a task
	s.AddTool(mcp.NewTool("tasks_delete",
		mcp.WithDescription("Delete a Google Task."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		common.WithAccountParam(),
	), HandleTasksDelete)

	// === Tasks Extended (Phase 2) ===

	// tasks_create_tasklist - Create a new task list
	s.AddTool(mcp.NewTool("tasks_create_tasklist",
		mcp.WithDescription("Create a new Google Tasks task list."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Title for the new task list")),
		common.WithAccountParam(),
	), HandleTasksCreateTasklist)

	// tasks_update_tasklist - Rename a task list
	s.AddTool(mcp.NewTool("tasks_update_tasklist",
		mcp.WithDescription("Rename a Google Tasks task list."),
		mcp.WithString("tasklist_id", mcp.Required(), mcp.Description("Task list ID to update")),
		mcp.WithString("title", mcp.Required(), mcp.Description("New title for the task list")),
		common.WithAccountParam(),
	), HandleTasksUpdateTasklist)

	// tasks_delete_tasklist - Delete a task list
	s.AddTool(mcp.NewTool("tasks_delete_tasklist",
		mcp.WithDescription("Delete a Google Tasks task list."),
		mcp.WithString("tasklist_id", mcp.Required(), mcp.Description("Task list ID to delete")),
		common.WithAccountParam(),
	), HandleTasksDeleteTasklist)

	// tasks_move - Reorder task or move to different position/parent
	s.AddTool(mcp.NewTool("tasks_move",
		mcp.WithDescription("Move a Google Task to a different position or parent. Use to reorder tasks or create subtasks."),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to move")),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default')")),
		mcp.WithString("parent", mcp.Description("Parent task ID (makes this task a subtask)")),
		mcp.WithString("previous", mcp.Description("Previous sibling task ID (positions task after this one)")),
		common.WithAccountParam(),
	), HandleTasksMove)

	// tasks_clear_completed - Clear all completed tasks from a list
	s.AddTool(mcp.NewTool("tasks_clear_completed",
		mcp.WithDescription("Clear all completed tasks from a Google Tasks task list."),
		mcp.WithString("tasklist_id", mcp.Description("Task list ID (default: '@default' for primary list)")),
		common.WithAccountParam(),
	), HandleTasksClearCompleted)
}
