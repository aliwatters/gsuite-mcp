package tasks

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/tasks/v1"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleTasksListTasklists  = common.WrapHandler[TasksService](TestableTasksListTasklists)
	HandleTasksList           = common.WrapHandler[TasksService](TestableTasksList)
	HandleTasksGet            = common.WrapHandler[TasksService](TestableTasksGet)
	HandleTasksCreate         = common.WrapHandler[TasksService](TestableTasksCreate)
	HandleTasksUpdate         = common.WrapHandler[TasksService](TestableTasksUpdate)
	HandleTasksComplete       = common.WrapHandler[TasksService](TestableTasksComplete)
	HandleTasksDelete         = common.WrapHandler[TasksService](TestableTasksDelete)
	HandleTasksCreateTasklist = common.WrapHandler[TasksService](TestableTasksCreateTasklist)
	HandleTasksUpdateTasklist = common.WrapHandler[TasksService](TestableTasksUpdateTasklist)
	HandleTasksDeleteTasklist = common.WrapHandler[TasksService](TestableTasksDeleteTasklist)
	HandleTasksMove           = common.WrapHandler[TasksService](TestableTasksMove)
	HandleTasksClearCompleted = common.WrapHandler[TasksService](TestableTasksClearCompleted)
)

// formatTask formats a task for compact output
func formatTask(task *tasks.Task) map[string]any {
	result := map[string]any{
		"id":     task.Id,
		"title":  task.Title,
		"status": task.Status,
	}

	if task.Due != "" {
		result["due"] = task.Due
	}

	if task.Parent != "" {
		result["parent"] = task.Parent
	}

	if task.Completed != nil && *task.Completed != "" {
		result["completed"] = *task.Completed
	}

	return result
}

// formatTaskFull formats a task with all details
func formatTaskFull(task *tasks.Task) map[string]any {
	result := formatTask(task)

	result["updated"] = task.Updated
	result["etag"] = task.Etag
	result["position"] = task.Position
	result["self_link"] = task.SelfLink

	if task.Notes != "" {
		result["notes"] = task.Notes
	}

	if task.Hidden {
		result["hidden"] = true
	}

	if task.Deleted {
		result["deleted"] = true
	}

	if len(task.Links) > 0 {
		links := make([]map[string]any, 0, len(task.Links))
		for _, link := range task.Links {
			links = append(links, map[string]any{
				"type":        link.Type,
				"description": link.Description,
				"link":        link.Link,
			})
		}
		result["links"] = links
	}

	return result
}
