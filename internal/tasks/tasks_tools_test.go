package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/tasks/v1"
)

func TestTasksListTasklists(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int
		wantErr   bool
		setupMock func(mock *MockTasksService)
	}{
		{
			name:      "list all task lists",
			args:      map[string]any{},
			wantCount: 2,
		},
		{
			name: "list with max_results",
			args: map[string]any{
				"max_results": float64(1),
			},
			wantCount: 2, // Mock returns same data regardless of limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksListTasklists(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			// Parse the response
			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			count := int(response["count"].(float64))
			if count != tt.wantCount {
				t.Errorf("got count %d, want %d", count, tt.wantCount)
			}

			taskLists := response["task_lists"].([]any)
			if len(taskLists) != tt.wantCount {
				t.Errorf("got %d task lists, want %d", len(taskLists), tt.wantCount)
			}
		})
	}
}

func TestTasksList(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int
		wantErr   bool
	}{
		{
			name:      "list tasks in default list",
			args:      map[string]any{},
			wantCount: 3,
		},
		{
			name: "list tasks with tasklist_id",
			args: map[string]any{
				"tasklist_id": "tasklist001",
			},
			wantCount: 3,
		},
		{
			name: "list tasks with show_completed false",
			args: map[string]any{
				"show_completed": false,
			},
			wantCount: 3, // Mock doesn't filter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksList(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			count := int(response["count"].(float64))
			if count != tt.wantCount {
				t.Errorf("got count %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestTasksGet(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantTitle string
		wantErr   bool
	}{
		{
			name: "get task by id",
			args: map[string]any{
				"task_id": "task001",
			},
			wantTitle: "Buy groceries",
		},
		{
			name:    "missing task_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksGet(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			title := response["title"].(string)
			if title != tt.wantTitle {
				t.Errorf("got title %q, want %q", title, tt.wantTitle)
			}
		})
	}
}

func TestTasksCreate(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantTitle string
		wantErr   bool
	}{
		{
			name: "create simple task",
			args: map[string]any{
				"title": "New task",
			},
			wantTitle: "New task",
		},
		{
			name: "create task with notes and due date",
			args: map[string]any{
				"title": "Task with details",
				"notes": "Some notes",
				"due":   "2024-03-01T00:00:00Z",
			},
			wantTitle: "Task with details",
		},
		{
			name:    "missing title",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksCreate(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			title := response["title"].(string)
			if title != tt.wantTitle {
				t.Errorf("got title %q, want %q", title, tt.wantTitle)
			}

			// Verify task was created with an ID
			id := response["id"].(string)
			if id == "" {
				t.Error("expected task to have an ID")
			}
		})
	}
}

func TestTasksUpdate(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantTitle string
		wantErr   bool
	}{
		{
			name: "update task title",
			args: map[string]any{
				"task_id": "task001",
				"title":   "Updated title",
			},
			wantTitle: "Updated title",
		},
		{
			name:    "missing task_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksUpdate(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			title := response["title"].(string)
			if title != tt.wantTitle {
				t.Errorf("got title %q, want %q", title, tt.wantTitle)
			}
		})
	}
}

func TestTasksComplete(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name       string
		args       map[string]any
		wantStatus string
		wantErr    bool
	}{
		{
			name: "complete task",
			args: map[string]any{
				"task_id": "task001",
			},
			wantStatus: "completed",
		},
		{
			name:    "missing task_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksComplete(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			status := response["status"].(string)
			if status != tt.wantStatus {
				t.Errorf("got status %q, want %q", status, tt.wantStatus)
			}

			message := response["message"].(string)
			if message != "Task marked as completed" {
				t.Errorf("got message %q, want 'Task marked as completed'", message)
			}
		})
	}
}

func TestTasksDelete(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "delete task",
			args: map[string]any{
				"task_id": "task001",
			},
		},
		{
			name:    "missing task_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksDelete(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			success := response["success"].(bool)
			if !success {
				t.Error("expected success to be true")
			}

			message := response["message"].(string)
			if message != "Task deleted successfully" {
				t.Errorf("got message %q, want 'Task deleted successfully'", message)
			}
		})
	}
}

func TestFormatTask(t *testing.T) {
	task := &tasks.Task{
		Id:     "test123",
		Title:  "Test Task",
		Status: "needsAction",
		Due:    "2024-02-15T00:00:00Z",
		Parent: "parent123",
		// Completed is *string, nil means not completed
	}

	result := formatTask(task)

	if result["id"] != "test123" {
		t.Errorf("got id %v, want test123", result["id"])
	}
	if result["title"] != "Test Task" {
		t.Errorf("got title %v, want Test Task", result["title"])
	}
	if result["status"] != "needsAction" {
		t.Errorf("got status %v, want needsAction", result["status"])
	}
	if result["due"] != "2024-02-15T00:00:00Z" {
		t.Errorf("got due %v, want 2024-02-15T00:00:00Z", result["due"])
	}
	if result["parent"] != "parent123" {
		t.Errorf("got parent %v, want parent123", result["parent"])
	}
}

func TestFormatTaskFull(t *testing.T) {
	task := &tasks.Task{
		Id:       "test123",
		Title:    "Test Task",
		Status:   "needsAction",
		Notes:    "Some notes here",
		Due:      "2024-02-15T00:00:00Z",
		Updated:  "2024-02-01T12:00:00Z",
		Position: "00000000000000000001",
		SelfLink: "https://example.com/task/test123",
		Links: []*tasks.TaskLinks{
			{
				Type:        "related",
				Description: "Related doc",
				Link:        "https://example.com/doc",
			},
		},
	}

	result := formatTaskFull(task)

	if result["id"] != "test123" {
		t.Errorf("got id %v, want test123", result["id"])
	}
	if result["notes"] != "Some notes here" {
		t.Errorf("got notes %v, want 'Some notes here'", result["notes"])
	}
	if result["updated"] != "2024-02-01T12:00:00Z" {
		t.Errorf("got updated %v, want 2024-02-01T12:00:00Z", result["updated"])
	}
	if result["position"] != "00000000000000000001" {
		t.Errorf("got position %v, want 00000000000000000001", result["position"])
	}
	if result["self_link"] != "https://example.com/task/test123" {
		t.Errorf("got self_link %v, want https://example.com/task/test123", result["self_link"])
	}

	links := result["links"].([]map[string]any)
	if len(links) != 1 {
		t.Errorf("got %d links, want 1", len(links))
	}
}

// === Phase 2 Tests: Task List Management ===

func TestTasksCreateTasklist(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantTitle string
		wantErr   bool
	}{
		{
			name: "create task list",
			args: map[string]any{
				"title": "New Task List",
			},
			wantTitle: "New Task List",
		},
		{
			name:    "missing title",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksCreateTasklist(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			title := response["title"].(string)
			if title != tt.wantTitle {
				t.Errorf("got title %q, want %q", title, tt.wantTitle)
			}

			// Verify task list was created with an ID
			id := response["id"].(string)
			if id == "" {
				t.Error("expected task list to have an ID")
			}

			message := response["message"].(string)
			if message != "Task list created successfully" {
				t.Errorf("got message %q, want 'Task list created successfully'", message)
			}
		})
	}
}

func TestTasksUpdateTasklist(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantTitle string
		wantErr   bool
	}{
		{
			name: "update task list title",
			args: map[string]any{
				"tasklist_id": "tasklist001",
				"title":       "Updated Title",
			},
			wantTitle: "Updated Title",
		},
		{
			name: "missing tasklist_id",
			args: map[string]any{
				"title": "New Title",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			args: map[string]any{
				"tasklist_id": "tasklist001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksUpdateTasklist(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			title := response["title"].(string)
			if title != tt.wantTitle {
				t.Errorf("got title %q, want %q", title, tt.wantTitle)
			}

			message := response["message"].(string)
			if message != "Task list updated successfully" {
				t.Errorf("got message %q, want 'Task list updated successfully'", message)
			}
		})
	}
}

func TestTasksDeleteTasklist(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "delete task list",
			args: map[string]any{
				"tasklist_id": "tasklist001",
			},
		},
		{
			name:    "missing tasklist_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksDeleteTasklist(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			success := response["success"].(bool)
			if !success {
				t.Error("expected success to be true")
			}

			message := response["message"].(string)
			if message != "Task list deleted successfully" {
				t.Errorf("got message %q, want 'Task list deleted successfully'", message)
			}
		})
	}
}

// === Phase 2 Tests: Task Organization ===

func TestTasksMove(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "move task to different position",
			args: map[string]any{
				"task_id":  "task001",
				"previous": "task002",
			},
		},
		{
			name: "move task to become subtask",
			args: map[string]any{
				"task_id": "task001",
				"parent":  "task002",
			},
		},
		{
			name: "move task with both parent and previous",
			args: map[string]any{
				"task_id":     "task001",
				"tasklist_id": "tasklist001",
				"parent":      "task002",
				"previous":    "task003",
			},
		},
		{
			name:    "missing task_id",
			args:    map[string]any{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksMove(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			message := response["message"].(string)
			if message != "Task moved successfully" {
				t.Errorf("got message %q, want 'Task moved successfully'", message)
			}

			// If parent was specified, verify it's in the response
			if parent, ok := tt.args["parent"].(string); ok && parent != "" {
				if response["parent"] != parent {
					t.Errorf("got parent %v, want %v", response["parent"], parent)
				}
			}
		})
	}
}

func TestTasksClearCompleted(t *testing.T) {
	fixtures := NewTasksTestFixtures()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "clear completed tasks from default list",
			args: map[string]any{},
		},
		{
			name: "clear completed tasks from specific list",
			args: map[string]any{
				"tasklist_id": "tasklist001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableTasksClearCompleted(context.Background(), request, fixtures.Deps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr {
				if !result.IsError {
					t.Errorf("expected error result")
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result: %v", result)
			}

			var response map[string]any
			textContent := result.Content[0].(mcp.TextContent)
			if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			success := response["success"].(bool)
			if !success {
				t.Error("expected success to be true")
			}

			message := response["message"].(string)
			if message != "Completed tasks cleared successfully" {
				t.Errorf("got message %q, want 'Completed tasks cleared successfully'", message)
			}
		})
	}
}
