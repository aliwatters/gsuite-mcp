package contacts

import (
	"context"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/people/v1"
)

// ContactsTestFixtures provides pre-configured test data for Contacts tests.
type ContactsTestFixtures struct {
	DefaultEmail string
	MockService  *MockContactsService
	Deps         *ContactsHandlerDeps
}

// NewContactsTestFixtures creates a new test fixtures instance with default configuration.
func NewContactsTestFixtures() *ContactsTestFixtures {
	mockService := &MockContactsService{}
	setupDefaultContactsMockData(mockService)
	f := common.NewTestFixtures[ContactsService](mockService)

	return &ContactsTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}

// setupDefaultContactsMockData populates the mock service with standard test data.
func setupDefaultContactsMockData(mock *MockContactsService) {
	// Set up ListContacts to return sample contacts
	mock.ListContactsFunc = func(_ context.Context, _ *ListContactsOptions) (*people.ListConnectionsResponse, error) {
		return &people.ListConnectionsResponse{
			Connections: []*people.Person{
				createTestContact("people/c001", "John", "Doe", "john@example.com", "+1-555-0001"),
				createTestContact("people/c002", "Jane", "Smith", "jane@example.com", "+1-555-0002"),
				createTestContact("people/c003", "Bob", "Wilson", "bob@example.com", "+1-555-0003"),
			},
			TotalPeople: 3,
		}, nil
	}

	// Set up GetContact
	mock.GetContactFunc = func(_ context.Context, resourceName string, _ *GetContactOptions) (*people.Person, error) {
		if resourceName == "people/c001" {
			return createTestContactFull("people/c001", "John", "Doe", "john@example.com", "+1-555-0001", "Acme Corp", "Engineer"), nil
		}
		return createTestContact("people/c001", "John", "Doe", "john@example.com", "+1-555-0001"), nil
	}

	// Set up SearchContacts
	mock.SearchContactsFunc = func(_ context.Context, query string, _ *SearchContactsOptions) (*people.SearchResponse, error) {
		return &people.SearchResponse{
			Results: []*people.SearchResult{
				{Person: createTestContact("people/c001", "John", "Doe", "john@example.com", "+1-555-0001")},
			},
		}, nil
	}

	// Set up CreateContact
	mock.CreateContactFunc = func(_ context.Context, person *people.Person) (*people.Person, error) {
		person.ResourceName = "people/c_new001"
		person.Etag = "etag_new"
		return person, nil
	}

	// Set up UpdateContact
	mock.UpdateContactFunc = func(_ context.Context, resourceName string, person *people.Person, updatePersonFields string) (*people.Person, error) {
		person.ResourceName = resourceName
		return person, nil
	}

	// Set up DeleteContact
	mock.DeleteContactFunc = func(_ context.Context, resourceName string) error {
		return nil
	}

	// Set up ListContactGroups
	mock.ListContactGroupsFunc = func(_ context.Context, _ *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error) {
		return &people.ListContactGroupsResponse{
			ContactGroups: []*people.ContactGroup{
				createTestContactGroup("contactGroups/g001", "Family", "USER_CONTACT_GROUP", 5),
				createTestContactGroup("contactGroups/g002", "Work", "USER_CONTACT_GROUP", 10),
				createTestContactGroup("contactGroups/myContacts", "My Contacts", "SYSTEM_CONTACT_GROUP", 100),
			},
			TotalItems: 3,
		}, nil
	}

	// Set up GetContactGroup
	mock.GetContactGroupFunc = func(_ context.Context, resourceName string, _ *GetContactGroupOptions) (*people.ContactGroup, error) {
		group := createTestContactGroupFull("contactGroups/g001", "Family", "USER_CONTACT_GROUP", 2)
		group.MemberResourceNames = []string{"people/c001", "people/c002"}
		return group, nil
	}

	// Set up CreateContactGroup
	mock.CreateContactGroupFunc = func(_ context.Context, name string) (*people.ContactGroup, error) {
		return &people.ContactGroup{
			ResourceName: "contactGroups/g_new001",
			Name:         name,
			GroupType:    "USER_CONTACT_GROUP",
			Etag:         "etag_new",
		}, nil
	}

	// Set up UpdateContactGroup
	mock.UpdateContactGroupFunc = func(_ context.Context, resourceName string, name string) (*people.ContactGroup, error) {
		return &people.ContactGroup{
			ResourceName: resourceName,
			Name:         name,
			GroupType:    "USER_CONTACT_GROUP",
			Etag:         "etag_updated",
		}, nil
	}

	// Set up DeleteContactGroup
	mock.DeleteContactGroupFunc = func(_ context.Context, resourceName string) error {
		return nil
	}

	// Set up ModifyContactGroupMembers
	mock.ModifyContactGroupMembersFunc = func(_ context.Context, resourceName string, addMembers, removeMembers []string) (*people.ModifyContactGroupMembersResponse, error) {
		return &people.ModifyContactGroupMembersResponse{}, nil
	}
}

// createTestContact creates a Person with standard fields for testing.
func createTestContact(resourceName, givenName, familyName, email, phone string) *people.Person {
	return &people.Person{
		ResourceName: resourceName,
		Etag:         "etag_" + resourceName,
		Names: []*people.Name{
			{
				GivenName:   givenName,
				FamilyName:  familyName,
				DisplayName: givenName + " " + familyName,
			},
		},
		EmailAddresses: []*people.EmailAddress{
			{Value: email, Type: "work"},
		},
		PhoneNumbers: []*people.PhoneNumber{
			{Value: phone, Type: "mobile"},
		},
	}
}

// createTestContactFull creates a Person with additional fields for testing.
func createTestContactFull(resourceName, givenName, familyName, email, phone, company, jobTitle string) *people.Person {
	person := createTestContact(resourceName, givenName, familyName, email, phone)
	person.Organizations = []*people.Organization{
		{Name: company, Title: jobTitle},
	}
	person.Photos = []*people.Photo{
		{Url: "https://example.com/photo.jpg"},
	}
	person.Biographies = []*people.Biography{
		{Value: "Test biography notes"},
	}
	person.Metadata = &people.PersonMetadata{
		Sources: []*people.Source{
			{Type: "CONTACT"},
		},
	}
	return person
}

// createTestContactGroup creates a ContactGroup for testing.
func createTestContactGroup(resourceName, name, groupType string, memberCount int64) *people.ContactGroup {
	return &people.ContactGroup{
		ResourceName: resourceName,
		Name:         name,
		GroupType:    groupType,
		MemberCount:  memberCount,
		Etag:         "etag_" + resourceName,
	}
}

// createTestContactGroupFull creates a ContactGroup with additional fields for testing.
func createTestContactGroupFull(resourceName, name, groupType string, memberCount int64) *people.ContactGroup {
	group := createTestContactGroup(resourceName, name, groupType, memberCount)
	group.FormattedName = name
	group.Metadata = &people.ContactGroupMetadata{
		UpdateTime: "2024-02-01T12:00:00Z",
	}
	return group
}
