package tasks

import (
	"context"

	"google.golang.org/api/tasks/v1"
)

// TasksService defines the interface for Google Tasks API operations.
// This interface enables dependency injection and testing with mocks.
type TasksService interface {
	// Task Lists
	ListTaskLists(ctx context.Context, opts *ListTaskListsOptions) (*tasks.TaskLists, error)
	GetTaskList(ctx context.Context, taskListID string) (*tasks.TaskList, error)
	CreateTaskList(ctx context.Context, taskList *tasks.TaskList) (*tasks.TaskList, error)
	UpdateTaskList(ctx context.Context, taskListID string, taskList *tasks.TaskList) (*tasks.TaskList, error)
	DeleteTaskList(ctx context.Context, taskListID string) error

	// Tasks
	ListTasks(ctx context.Context, taskListID string, opts *ListTasksOptions) (*tasks.Tasks, error)
	GetTask(ctx context.Context, taskListID, taskID string) (*tasks.Task, error)
	CreateTask(ctx context.Context, taskListID string, task *tasks.Task, opts *CreateTaskOptions) (*tasks.Task, error)
	UpdateTask(ctx context.Context, taskListID, taskID string, task *tasks.Task) (*tasks.Task, error)
	DeleteTask(ctx context.Context, taskListID, taskID string) error
	ClearCompleted(ctx context.Context, taskListID string) error
	MoveTask(ctx context.Context, taskListID, taskID string, opts *MoveTaskOptions) (*tasks.Task, error)
}

// ListTaskListsOptions contains optional parameters for listing task lists.
type ListTaskListsOptions struct {
	MaxResults int64
	PageToken  string
}

// ListTasksOptions contains optional parameters for listing tasks.
type ListTasksOptions struct {
	MaxResults         int64
	PageToken          string
	ShowCompleted      bool
	ShowHidden         bool
	ShowDeleted        bool
	DueMin             string // RFC 3339 timestamp
	DueMax             string // RFC 3339 timestamp
	CompletedMin       string // RFC 3339 timestamp
	CompletedMax       string // RFC 3339 timestamp
	UpdatedMin         string // RFC 3339 timestamp
	ShowAssigned       bool
	ShowOnlyAssignedMe bool
}

// CreateTaskOptions contains optional parameters for creating a task.
type CreateTaskOptions struct {
	Parent string // Parent task ID (for creating subtasks)
}

// MoveTaskOptions contains parameters for moving a task.
type MoveTaskOptions struct {
	Parent   string // Parent task ID (for subtasks)
	Previous string // Previous sibling task ID (for ordering)
}

// RealTasksService wraps the Tasks API client and implements TasksService.
type RealTasksService struct {
	service *tasks.Service
}

// NewRealTasksService creates a new RealTasksService wrapping the given Tasks API service.
func NewRealTasksService(service *tasks.Service) *RealTasksService {
	return &RealTasksService{service: service}
}

// ListTaskLists lists all task lists for the user.
func (s *RealTasksService) ListTaskLists(ctx context.Context, opts *ListTaskListsOptions) (*tasks.TaskLists, error) {
	call := s.service.Tasklists.List().Context(ctx)

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
	}

	return call.Do()
}

// GetTaskList gets a single task list by ID.
func (s *RealTasksService) GetTaskList(ctx context.Context, taskListID string) (*tasks.TaskList, error) {
	return s.service.Tasklists.Get(taskListID).Context(ctx).Do()
}

// CreateTaskList creates a new task list.
func (s *RealTasksService) CreateTaskList(ctx context.Context, taskList *tasks.TaskList) (*tasks.TaskList, error) {
	return s.service.Tasklists.Insert(taskList).Context(ctx).Do()
}

// UpdateTaskList updates a task list.
func (s *RealTasksService) UpdateTaskList(ctx context.Context, taskListID string, taskList *tasks.TaskList) (*tasks.TaskList, error) {
	return s.service.Tasklists.Update(taskListID, taskList).Context(ctx).Do()
}

// DeleteTaskList deletes a task list.
func (s *RealTasksService) DeleteTaskList(ctx context.Context, taskListID string) error {
	return s.service.Tasklists.Delete(taskListID).Context(ctx).Do()
}

// ListTasks lists tasks in a task list.
func (s *RealTasksService) ListTasks(ctx context.Context, taskListID string, opts *ListTasksOptions) (*tasks.Tasks, error) {
	call := s.service.Tasks.List(taskListID).Context(ctx)

	if opts != nil {
		if opts.MaxResults > 0 {
			call = call.MaxResults(opts.MaxResults)
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.ShowCompleted {
			call = call.ShowCompleted(true)
		}
		if opts.ShowHidden {
			call = call.ShowHidden(true)
		}
		if opts.ShowDeleted {
			call = call.ShowDeleted(true)
		}
		if opts.DueMin != "" {
			call = call.DueMin(opts.DueMin)
		}
		if opts.DueMax != "" {
			call = call.DueMax(opts.DueMax)
		}
		if opts.CompletedMin != "" {
			call = call.CompletedMin(opts.CompletedMin)
		}
		if opts.CompletedMax != "" {
			call = call.CompletedMax(opts.CompletedMax)
		}
		if opts.UpdatedMin != "" {
			call = call.UpdatedMin(opts.UpdatedMin)
		}
	}

	return call.Do()
}

// GetTask gets a single task by ID.
func (s *RealTasksService) GetTask(ctx context.Context, taskListID, taskID string) (*tasks.Task, error) {
	return s.service.Tasks.Get(taskListID, taskID).Context(ctx).Do()
}

// CreateTask creates a new task in a task list.
func (s *RealTasksService) CreateTask(ctx context.Context, taskListID string, task *tasks.Task, opts *CreateTaskOptions) (*tasks.Task, error) {
	call := s.service.Tasks.Insert(taskListID, task).Context(ctx)
	if opts != nil && opts.Parent != "" {
		call = call.Parent(opts.Parent)
	}
	return call.Do()
}

// UpdateTask updates an existing task.
func (s *RealTasksService) UpdateTask(ctx context.Context, taskListID, taskID string, task *tasks.Task) (*tasks.Task, error) {
	return s.service.Tasks.Update(taskListID, taskID, task).Context(ctx).Do()
}

// DeleteTask deletes a task.
func (s *RealTasksService) DeleteTask(ctx context.Context, taskListID, taskID string) error {
	return s.service.Tasks.Delete(taskListID, taskID).Context(ctx).Do()
}

// ClearCompleted removes all completed tasks from a task list.
func (s *RealTasksService) ClearCompleted(ctx context.Context, taskListID string) error {
	return s.service.Tasks.Clear(taskListID).Context(ctx).Do()
}

// MoveTask moves a task to a different position.
func (s *RealTasksService) MoveTask(ctx context.Context, taskListID, taskID string, opts *MoveTaskOptions) (*tasks.Task, error) {
	call := s.service.Tasks.Move(taskListID, taskID).Context(ctx)

	if opts != nil {
		if opts.Parent != "" {
			call = call.Parent(opts.Parent)
		}
		if opts.Previous != "" {
			call = call.Previous(opts.Previous)
		}
	}

	return call.Do()
}
