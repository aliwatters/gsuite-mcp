package contacts

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/people/v1"
)

func TestContactsList(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name:      "list all contacts",
			args:      map[string]any{},
			wantCount: 3,
		},
		{
			name: "list with max_results",
			args: map[string]any{
				"max_results": float64(10),
			},
			wantCount: 3, // Mock returns same data regardless of limit
		},
		{
			name: "list with page_token",
			args: map[string]any{
				"page_token": "next_page",
			},
			wantCount: 3,
		},
		{
			name:    "service error",
			args:    map[string]any{},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.ListContactsFunc = func(_ context.Context, _ *ListContactsOptions) (*people.ListConnectionsResponse, error) {
					return nil, errors.New("API error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock to defaults
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsList(context.Background(), request, fixtures.Deps)
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

			contacts := response["contacts"].([]any)
			if len(contacts) != tt.wantCount {
				t.Errorf("got %d contacts, want %d", len(contacts), tt.wantCount)
			}
		})
	}
}

func TestContactsGet(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantName  string
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "get contact by resource_name",
			args: map[string]any{
				"resource_name": "people/c001",
			},
			wantName: "John Doe",
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "people/invalid",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.GetContactFunc = func(_ context.Context, _ string, _ *GetContactOptions) (*people.Person, error) {
					return nil, errors.New("contact not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsGet(context.Background(), request, fixtures.Deps)
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

			name := response["name"].(string)
			if name != tt.wantName {
				t.Errorf("got name %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestContactsSearch(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "search contacts",
			args: map[string]any{
				"query": "John",
			},
			wantCount: 1,
		},
		{
			name: "search with max_results",
			args: map[string]any{
				"query":       "John",
				"max_results": float64(5),
			},
			wantCount: 1,
		},
		{
			name:    "missing query",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "empty query",
			args: map[string]any{
				"query": "",
			},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"query": "test",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.SearchContactsFunc = func(_ context.Context, _ string, _ *SearchContactsOptions) (*people.SearchResponse, error) {
					return nil, errors.New("search failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsSearch(context.Background(), request, fixtures.Deps)
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

func TestContactsCreate(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantName  string
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "create contact with name only",
			args: map[string]any{
				"given_name":  "Alice",
				"family_name": "Johnson",
			},
			wantName: "Alice Johnson",
		},
		{
			name: "create contact with all fields",
			args: map[string]any{
				"given_name":  "Bob",
				"family_name": "Williams",
				"email":       "bob@example.com",
				"phone":       "+1-555-1234",
				"company":     "Tech Corp",
				"job_title":   "Developer",
				"notes":       "Test notes",
			},
			wantName: "Bob Williams",
		},
		{
			name:    "missing given_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"given_name": "Test",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.CreateContactFunc = func(_ context.Context, _ *people.Person) (*people.Person, error) {
					return nil, errors.New("create failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsCreate(context.Background(), request, fixtures.Deps)
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

			name := response["name"].(string)
			if name != tt.wantName {
				t.Errorf("got name %q, want %q", name, tt.wantName)
			}

			// Verify resource_name is returned
			if _, ok := response["resource_name"]; !ok {
				t.Error("expected resource_name in response")
			}
		})
	}
}

func TestContactsUpdate(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "update contact name",
			args: map[string]any{
				"resource_name": "people/c001",
				"given_name":    "Johnny",
			},
		},
		{
			name: "update contact email",
			args: map[string]any{
				"resource_name": "people/c001",
				"email":         "newemail@example.com",
			},
		},
		{
			name: "update multiple fields",
			args: map[string]any{
				"resource_name": "people/c001",
				"given_name":    "Johnny",
				"family_name":   "Updated",
				"phone":         "+1-555-9999",
				"company":       "New Corp",
			},
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "people/c001",
				"given_name":    "Test",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.UpdateContactFunc = func(_ context.Context, _ string, _ *people.Person, _ string) (*people.Person, error) {
					return nil, errors.New("update failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsUpdate(context.Background(), request, fixtures.Deps)
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

			// Verify resource_name is in response
			if _, ok := response["resource_name"]; !ok {
				t.Error("expected resource_name in response")
			}
		})
	}
}

func TestContactsDelete(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "delete contact",
			args: map[string]any{
				"resource_name": "people/c001",
			},
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "people/invalid",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.DeleteContactFunc = func(_ context.Context, _ string) error {
					return errors.New("delete failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsDelete(context.Background(), request, fixtures.Deps)
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

			// Check success message
			if _, ok := response["success"]; !ok {
				t.Error("expected success field in response")
			}
		})
	}
}

func TestContactsListGroups(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantCount int
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name:      "list all groups",
			args:      map[string]any{},
			wantCount: 3,
		},
		{
			name: "list with max_results",
			args: map[string]any{
				"max_results": float64(5),
			},
			wantCount: 3,
		},
		{
			name:    "service error",
			args:    map[string]any{},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.ListContactGroupsFunc = func(_ context.Context, _ *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error) {
					return nil, errors.New("list groups failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsListGroups(context.Background(), request, fixtures.Deps)
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

			groups := response["contact_groups"].([]any)
			if len(groups) != tt.wantCount {
				t.Errorf("got %d groups, want %d", len(groups), tt.wantCount)
			}
		})
	}
}

func TestContactsGetGroup(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantName  string
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "get group by resource_name",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
			},
			wantName: "Family",
		},
		{
			name: "get group with members",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
				"max_members":   float64(10),
			},
			wantName: "Family",
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "contactGroups/invalid",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.GetContactGroupFunc = func(_ context.Context, _ string, _ *GetContactGroupOptions) (*people.ContactGroup, error) {
					return nil, errors.New("group not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsGetGroup(context.Background(), request, fixtures.Deps)
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

			name := response["name"].(string)
			if name != tt.wantName {
				t.Errorf("got name %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestContactsCreateGroup(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantName  string
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "create group",
			args: map[string]any{
				"name": "Friends",
			},
			wantName: "Friends",
		},
		{
			name:    "missing name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "empty name",
			args: map[string]any{
				"name": "",
			},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"name": "Test",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.CreateContactGroupFunc = func(_ context.Context, _ string) (*people.ContactGroup, error) {
					return nil, errors.New("create group failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsCreateGroup(context.Background(), request, fixtures.Deps)
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

			name := response["name"].(string)
			if name != tt.wantName {
				t.Errorf("got name %q, want %q", name, tt.wantName)
			}

			// Verify resource_name is returned
			if _, ok := response["resource_name"]; !ok {
				t.Error("expected resource_name in response")
			}
		})
	}
}

func TestContactsUpdateGroup(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantName  string
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "update group name",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
				"name":          "Close Friends",
			},
			wantName: "Close Friends",
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "missing name",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
			},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
				"name":          "Test",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.UpdateContactGroupFunc = func(_ context.Context, _ string, _ string) (*people.ContactGroup, error) {
					return nil, errors.New("update group failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsUpdateGroup(context.Background(), request, fixtures.Deps)
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

			name := response["name"].(string)
			if name != tt.wantName {
				t.Errorf("got name %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestContactsDeleteGroup(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "delete group",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
			},
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "contactGroups/invalid",
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.DeleteContactGroupFunc = func(_ context.Context, _ string) error {
					return errors.New("delete group failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsDeleteGroup(context.Background(), request, fixtures.Deps)
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

			// Check success message
			if _, ok := response["success"]; !ok {
				t.Error("expected success field in response")
			}
		})
	}
}

func TestContactsModifyGroupMembers(t *testing.T) {
	fixtures := NewContactsTestFixtures()

	tests := []struct {
		name      string
		args      map[string]any
		wantErr   bool
		setupMock func(mock *MockContactsService)
	}{
		{
			name: "add members",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
				"add_members":   []any{"people/c003", "people/c004"},
			},
		},
		{
			name: "remove members",
			args: map[string]any{
				"resource_name":  "contactGroups/g001",
				"remove_members": []any{"people/c001"},
			},
		},
		{
			name: "add and remove members",
			args: map[string]any{
				"resource_name":  "contactGroups/g001",
				"add_members":    []any{"people/c003"},
				"remove_members": []any{"people/c001"},
			},
		},
		{
			name:    "missing resource_name",
			args:    map[string]any{},
			wantErr: true,
		},
		{
			name: "no members to add or remove",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
			},
			wantErr: true,
		},
		{
			name: "service error",
			args: map[string]any{
				"resource_name": "contactGroups/g001",
				"add_members":   []any{"people/c003"},
			},
			wantErr: true,
			setupMock: func(mock *MockContactsService) {
				mock.ModifyContactGroupMembersFunc = func(_ context.Context, _ string, _, _ []string) (*people.ModifyContactGroupMembersResponse, error) {
					return nil, errors.New("modify members failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupDefaultContactsMockData(fixtures.MockService)
			if tt.setupMock != nil {
				tt.setupMock(fixtures.MockService)
			}

			request := mcp.CallToolRequest{}
			request.Params.Arguments = tt.args

			result, err := TestableContactsModifyGroupMembers(context.Background(), request, fixtures.Deps)
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

			// Check success message
			if _, ok := response["success"]; !ok {
				t.Error("expected success field in response")
			}
		})
	}
}
