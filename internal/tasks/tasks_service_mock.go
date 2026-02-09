package tasks

import (
	"context"

	"google.golang.org/api/tasks/v1"
)

// MockTasksService implements TasksService for testing.
type MockTasksService struct {
	// Task Lists
	ListTaskListsFunc  func(ctx context.Context, opts *ListTaskListsOptions) (*tasks.TaskLists, error)
	GetTaskListFunc    func(ctx context.Context, taskListID string) (*tasks.TaskList, error)
	CreateTaskListFunc func(ctx context.Context, taskList *tasks.TaskList) (*tasks.TaskList, error)
	UpdateTaskListFunc func(ctx context.Context, taskListID string, taskList *tasks.TaskList) (*tasks.TaskList, error)
	DeleteTaskListFunc func(ctx context.Context, taskListID string) error

	// Tasks
	ListTasksFunc      func(ctx context.Context, taskListID string, opts *ListTasksOptions) (*tasks.Tasks, error)
	GetTaskFunc        func(ctx context.Context, taskListID, taskID string) (*tasks.Task, error)
	CreateTaskFunc     func(ctx context.Context, taskListID string, task *tasks.Task, opts *CreateTaskOptions) (*tasks.Task, error)
	UpdateTaskFunc     func(ctx context.Context, taskListID, taskID string, task *tasks.Task) (*tasks.Task, error)
	DeleteTaskFunc     func(ctx context.Context, taskListID, taskID string) error
	ClearCompletedFunc func(ctx context.Context, taskListID string) error
	MoveTaskFunc       func(ctx context.Context, taskListID, taskID string, opts *MoveTaskOptions) (*tasks.Task, error)
}

// Task List methods

func (m *MockTasksService) ListTaskLists(ctx context.Context, opts *ListTaskListsOptions) (*tasks.TaskLists, error) {
	if m.ListTaskListsFunc != nil {
		return m.ListTaskListsFunc(ctx, opts)
	}
	return &tasks.TaskLists{}, nil
}

func (m *MockTasksService) GetTaskList(ctx context.Context, taskListID string) (*tasks.TaskList, error) {
	if m.GetTaskListFunc != nil {
		return m.GetTaskListFunc(ctx, taskListID)
	}
	return &tasks.TaskList{}, nil
}

func (m *MockTasksService) CreateTaskList(ctx context.Context, taskList *tasks.TaskList) (*tasks.TaskList, error) {
	if m.CreateTaskListFunc != nil {
		return m.CreateTaskListFunc(ctx, taskList)
	}
	return taskList, nil
}

func (m *MockTasksService) UpdateTaskList(ctx context.Context, taskListID string, taskList *tasks.TaskList) (*tasks.TaskList, error) {
	if m.UpdateTaskListFunc != nil {
		return m.UpdateTaskListFunc(ctx, taskListID, taskList)
	}
	return taskList, nil
}

func (m *MockTasksService) DeleteTaskList(ctx context.Context, taskListID string) error {
	if m.DeleteTaskListFunc != nil {
		return m.DeleteTaskListFunc(ctx, taskListID)
	}
	return nil
}

// Task methods

func (m *MockTasksService) ListTasks(ctx context.Context, taskListID string, opts *ListTasksOptions) (*tasks.Tasks, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, taskListID, opts)
	}
	return &tasks.Tasks{}, nil
}

func (m *MockTasksService) GetTask(ctx context.Context, taskListID, taskID string) (*tasks.Task, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, taskListID, taskID)
	}
	return &tasks.Task{}, nil
}

func (m *MockTasksService) CreateTask(ctx context.Context, taskListID string, task *tasks.Task, opts *CreateTaskOptions) (*tasks.Task, error) {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, taskListID, task, opts)
	}
	return task, nil
}

func (m *MockTasksService) UpdateTask(ctx context.Context, taskListID, taskID string, task *tasks.Task) (*tasks.Task, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, taskListID, taskID, task)
	}
	return task, nil
}

func (m *MockTasksService) DeleteTask(ctx context.Context, taskListID, taskID string) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, taskListID, taskID)
	}
	return nil
}

func (m *MockTasksService) ClearCompleted(ctx context.Context, taskListID string) error {
	if m.ClearCompletedFunc != nil {
		return m.ClearCompletedFunc(ctx, taskListID)
	}
	return nil
}

func (m *MockTasksService) MoveTask(ctx context.Context, taskListID, taskID string, opts *MoveTaskOptions) (*tasks.Task, error) {
	if m.MoveTaskFunc != nil {
		return m.MoveTaskFunc(ctx, taskListID, taskID, opts)
	}
	return &tasks.Task{}, nil
}
