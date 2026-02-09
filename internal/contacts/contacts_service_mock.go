package contacts

import (
	"context"

	"google.golang.org/api/people/v1"
)

// MockContactsService implements ContactsService for testing.
type MockContactsService struct {
	// Contacts
	ListContactsFunc   func(ctx context.Context, opts *ListContactsOptions) (*people.ListConnectionsResponse, error)
	GetContactFunc     func(ctx context.Context, resourceName string, opts *GetContactOptions) (*people.Person, error)
	SearchContactsFunc func(ctx context.Context, query string, opts *SearchContactsOptions) (*people.SearchResponse, error)
	CreateContactFunc  func(ctx context.Context, person *people.Person) (*people.Person, error)
	UpdateContactFunc  func(ctx context.Context, resourceName string, person *people.Person, updatePersonFields string) (*people.Person, error)
	DeleteContactFunc  func(ctx context.Context, resourceName string) error

	// Contact Groups
	ListContactGroupsFunc         func(ctx context.Context, opts *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error)
	GetContactGroupFunc           func(ctx context.Context, resourceName string, opts *GetContactGroupOptions) (*people.ContactGroup, error)
	CreateContactGroupFunc        func(ctx context.Context, name string) (*people.ContactGroup, error)
	UpdateContactGroupFunc        func(ctx context.Context, resourceName string, name string) (*people.ContactGroup, error)
	DeleteContactGroupFunc        func(ctx context.Context, resourceName string) error
	ModifyContactGroupMembersFunc func(ctx context.Context, resourceName string, addMembers, removeMembers []string) (*people.ModifyContactGroupMembersResponse, error)
}

// Contact methods

func (m *MockContactsService) ListContacts(ctx context.Context, opts *ListContactsOptions) (*people.ListConnectionsResponse, error) {
	if m.ListContactsFunc != nil {
		return m.ListContactsFunc(ctx, opts)
	}
	return &people.ListConnectionsResponse{}, nil
}

func (m *MockContactsService) GetContact(ctx context.Context, resourceName string, opts *GetContactOptions) (*people.Person, error) {
	if m.GetContactFunc != nil {
		return m.GetContactFunc(ctx, resourceName, opts)
	}
	return &people.Person{}, nil
}

func (m *MockContactsService) SearchContacts(ctx context.Context, query string, opts *SearchContactsOptions) (*people.SearchResponse, error) {
	if m.SearchContactsFunc != nil {
		return m.SearchContactsFunc(ctx, query, opts)
	}
	return &people.SearchResponse{}, nil
}

func (m *MockContactsService) CreateContact(ctx context.Context, person *people.Person) (*people.Person, error) {
	if m.CreateContactFunc != nil {
		return m.CreateContactFunc(ctx, person)
	}
	return person, nil
}

func (m *MockContactsService) UpdateContact(ctx context.Context, resourceName string, person *people.Person, updatePersonFields string) (*people.Person, error) {
	if m.UpdateContactFunc != nil {
		return m.UpdateContactFunc(ctx, resourceName, person, updatePersonFields)
	}
	return person, nil
}

func (m *MockContactsService) DeleteContact(ctx context.Context, resourceName string) error {
	if m.DeleteContactFunc != nil {
		return m.DeleteContactFunc(ctx, resourceName)
	}
	return nil
}

// Contact Group methods

func (m *MockContactsService) ListContactGroups(ctx context.Context, opts *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error) {
	if m.ListContactGroupsFunc != nil {
		return m.ListContactGroupsFunc(ctx, opts)
	}
	return &people.ListContactGroupsResponse{}, nil
}

func (m *MockContactsService) GetContactGroup(ctx context.Context, resourceName string, opts *GetContactGroupOptions) (*people.ContactGroup, error) {
	if m.GetContactGroupFunc != nil {
		return m.GetContactGroupFunc(ctx, resourceName, opts)
	}
	return &people.ContactGroup{}, nil
}

func (m *MockContactsService) CreateContactGroup(ctx context.Context, name string) (*people.ContactGroup, error) {
	if m.CreateContactGroupFunc != nil {
		return m.CreateContactGroupFunc(ctx, name)
	}
	return &people.ContactGroup{Name: name}, nil
}

func (m *MockContactsService) UpdateContactGroup(ctx context.Context, resourceName string, name string) (*people.ContactGroup, error) {
	if m.UpdateContactGroupFunc != nil {
		return m.UpdateContactGroupFunc(ctx, resourceName, name)
	}
	return &people.ContactGroup{ResourceName: resourceName, Name: name}, nil
}

func (m *MockContactsService) DeleteContactGroup(ctx context.Context, resourceName string) error {
	if m.DeleteContactGroupFunc != nil {
		return m.DeleteContactGroupFunc(ctx, resourceName)
	}
	return nil
}

func (m *MockContactsService) ModifyContactGroupMembers(ctx context.Context, resourceName string, addMembers, removeMembers []string) (*people.ModifyContactGroupMembersResponse, error) {
	if m.ModifyContactGroupMembersFunc != nil {
		return m.ModifyContactGroupMembersFunc(ctx, resourceName, addMembers, removeMembers)
	}
	return &people.ModifyContactGroupMembersResponse{}, nil
}
