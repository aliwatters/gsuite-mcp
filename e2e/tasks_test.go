//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aliwatters/gsuite-mcp/internal/tasks"
)

// TestTasksListTasklists verifies listing task lists.
func TestTasksListTasklists(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := tasksDeps()

	result, err := tasks.TestableTasksListTasklists(ctx, makeRequest(nil), deps)
	cr := requireSuccess(t, result, err)

	tasklists := requireArrayField(t, cr, "task_lists")
	if len(tasklists) == 0 {
		t.Fatal("expected at least one task list")
	}

	t.Logf("found %d task lists", len(tasklists))
}

// TestTasksCreateListAndTasks tests creating a task list, adding tasks,
// completing a task, and cleaning up.
func TestTasksCreateListAndTasks(t *testing.T) {
	setupAuth(t)
	ctx := context.Background()
	deps := tasksDeps()

	prefix := e2ePrefix()
	listTitle := fmt.Sprintf("%s tasks %d", prefix, time.Now().UnixMilli())

	// Create task list
	t.Logf("creating task list %q", listTitle)
	result, err := tasks.TestableTasksCreateTasklist(ctx, makeRequest(map[string]any{
		"title": listTitle,
	}), deps)
	createListResult := requireSuccess(t, result, err)

	taskListID := requireStringField(t, createListResult, "id")
	t.Logf("created task list id=%s", taskListID)

	defer func() {
		t.Log("cleanup: deleting test task list")
		_, _ = tasks.TestableTasksDeleteTasklist(ctx, makeRequest(map[string]any{
			"tasklist_id": taskListID,
		}), deps)
	}()

	// Create a task in the list
	taskTitle := fmt.Sprintf("%s task item", prefix)
	t.Logf("creating task %q", taskTitle)
	result, err = tasks.TestableTasksCreate(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
		"title":       taskTitle,
		"notes":       "E2E test task. Safe to delete.",
	}), deps)
	createTaskResult := requireSuccess(t, result, err)

	taskID := requireStringField(t, createTaskResult, "id")
	t.Logf("created task id=%s", taskID)

	// List tasks in the list
	t.Log("listing tasks")
	result, err = tasks.TestableTasksList(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
	}), deps)
	listResult := requireSuccess(t, result, err)

	taskItems := requireArrayField(t, listResult, "tasks")
	if len(taskItems) == 0 {
		t.Fatal("expected at least one task in the list")
	}

	// Get the task
	t.Log("getting task")
	result, err = tasks.TestableTasksGet(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
		"task_id":     taskID,
	}), deps)
	getResult := requireSuccess(t, result, err)

	gotTitle := requireStringField(t, getResult, "title")
	if gotTitle != taskTitle {
		t.Errorf("expected task title %q, got %q", taskTitle, gotTitle)
	}

	// Update the task
	updatedTitle := taskTitle + " (updated)"
	t.Log("updating task")
	result, err = tasks.TestableTasksUpdate(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
		"task_id":     taskID,
		"title":       updatedTitle,
	}), deps)
	requireSuccess(t, result, err)

	// Complete the task
	t.Log("completing task")
	result, err = tasks.TestableTasksComplete(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
		"task_id":     taskID,
	}), deps)
	requireSuccess(t, result, err)

	// Delete the task
	t.Log("deleting task")
	result, err = tasks.TestableTasksDelete(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
		"task_id":     taskID,
	}), deps)
	requireSuccess(t, result, err)

	// Delete the task list
	t.Log("deleting task list")
	result, err = tasks.TestableTasksDeleteTasklist(ctx, makeRequest(map[string]any{
		"tasklist_id": taskListID,
	}), deps)
	requireSuccess(t, result, err)
}
