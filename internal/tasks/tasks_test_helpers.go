package tasks

import (
	"context"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/tasks/v1"
)

// Test fixture constants for task IDs, task list IDs, and dates.
const (
	testTaskID1       = "task001"
	testTaskID2       = "task002"
	testTaskID3       = "task003"
	testNewTaskID     = "newtask001"
	testTaskListID1   = "tasklist001"
	testTaskListID2   = "tasklist002"
	testNewTaskListID = "newtasklist001"
	testDate          = "2024-02-01T12:00:00Z"
	testDueDate       = "2024-02-15T00:00:00Z"
	testCompletedDate = "2024-02-01T10:00:00Z"
)

// TasksTestFixtures provides pre-configured test data for Tasks tests.
type TasksTestFixtures struct {
	DefaultEmail string
	MockService  *MockTasksService
	Deps         *TasksHandlerDeps
}

// NewTasksTestFixtures creates a new test fixtures instance with default configuration.
func NewTasksTestFixtures() *TasksTestFixtures {
	mockService := &MockTasksService{}
	setupDefaultTasksMockData(mockService)
	f := common.NewTestFixtures[TasksService](mockService)

	return &TasksTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}

// setupDefaultTasksMockData populates the mock service with standard test data.
func setupDefaultTasksMockData(mock *MockTasksService) {
	// Set up ListTaskLists to return sample task lists
	mock.ListTaskListsFunc = func(_ context.Context, _ *ListTaskListsOptions) (*tasks.TaskLists, error) {
		return &tasks.TaskLists{
			Items: []*tasks.TaskList{
				{
					Id:      testTaskListID1,
					Title:   "My Tasks",
					Updated: testDate,
				},
				{
					Id:      testTaskListID2,
					Title:   "Work Tasks",
					Updated: testDate,
				},
			},
		}, nil
	}

	// Set up ListTasks to return sample tasks
	mock.ListTasksFunc = func(_ context.Context, taskListID string, _ *ListTasksOptions) (*tasks.Tasks, error) {
		if taskListID == common.DefaultTaskListID || taskListID == testTaskListID1 {
			return &tasks.Tasks{
				Items: []*tasks.Task{
					createTestTask(testTaskID1, "Buy groceries", "Milk, eggs, bread", "needsAction", testDueDate, ""),
					createTestTask(testTaskID2, "Call doctor", "", "needsAction", "", ""),
					createTestTask(testTaskID3, "Review code", "PR #123", "completed", "", testCompletedDate),
				},
			}, nil
		}
		return &tasks.Tasks{Items: []*tasks.Task{}}, nil
	}

	// Set up GetTask
	mock.GetTaskFunc = func(_ context.Context, taskListID, taskID string) (*tasks.Task, error) {
		if taskID == testTaskID1 {
			return createTestTask(testTaskID1, "Buy groceries", "Milk, eggs, bread", "needsAction", testDueDate, ""), nil
		}
		return nil, nil
	}

	// Set up CreateTask
	mock.CreateTaskFunc = func(_ context.Context, taskListID string, task *tasks.Task, opts *CreateTaskOptions) (*tasks.Task, error) {
		task.Id = testNewTaskID
		task.Status = "needsAction"
		if opts != nil && opts.Parent != "" {
			task.Parent = opts.Parent
		}
		return task, nil
	}

	// Set up UpdateTask
	mock.UpdateTaskFunc = func(_ context.Context, taskListID, taskID string, task *tasks.Task) (*tasks.Task, error) {
		return task, nil
	}

	// Set up DeleteTask
	mock.DeleteTaskFunc = func(_ context.Context, taskListID, taskID string) error {
		return nil
	}

	// Set up GetTaskList
	mock.GetTaskListFunc = func(_ context.Context, taskListID string) (*tasks.TaskList, error) {
		return &tasks.TaskList{
			Id:      taskListID,
			Title:   "My Tasks",
			Updated: testDate,
		}, nil
	}

	// Set up CreateTaskList
	mock.CreateTaskListFunc = func(_ context.Context, taskList *tasks.TaskList) (*tasks.TaskList, error) {
		taskList.Id = testNewTaskListID
		taskList.Updated = testDate
		return taskList, nil
	}

	// Set up UpdateTaskList
	mock.UpdateTaskListFunc = func(_ context.Context, taskListID string, taskList *tasks.TaskList) (*tasks.TaskList, error) {
		taskList.Updated = testDate
		return taskList, nil
	}

	// Set up DeleteTaskList
	mock.DeleteTaskListFunc = func(_ context.Context, taskListID string) error {
		return nil
	}

	// Set up ClearCompleted
	mock.ClearCompletedFunc = func(_ context.Context, taskListID string) error {
		return nil
	}

	// Set up MoveTask
	mock.MoveTaskFunc = func(_ context.Context, taskListID, taskID string, opts *MoveTaskOptions) (*tasks.Task, error) {
		task := createTestTask(taskID, "Moved Task", "", "needsAction", "", "")
		if opts != nil && opts.Parent != "" {
			task.Parent = opts.Parent
		}
		return task, nil
	}
}

// createTestTask creates a Task with standard fields.
func createTestTask(id, title, notes, status, due, completed string) *tasks.Task {
	task := &tasks.Task{
		Id:       id,
		Title:    title,
		Notes:    notes,
		Status:   status,
		Updated:  testDate,
		Position: "00000000000000000001",
		SelfLink: "https://www.googleapis.com/tasks/v1/lists/" + testTaskListID1 + "/tasks/" + id,
	}

	if due != "" {
		task.Due = due
	}

	if completed != "" {
		task.Completed = &completed
	}

	return task
}

// createTestTaskWithParent creates a Task with a parent (subtask).
func createTestTaskWithParent(id, title, parentID string) *tasks.Task {
	task := createTestTask(id, title, "", "needsAction", "", "")
	task.Parent = parentID
	return task
}
