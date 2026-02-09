package contacts

import (
	"context"

	"google.golang.org/api/people/v1"
)

// Default fields to request from the People API
const (
	// DefaultPersonFields contains the fields to request when listing/getting contacts
	DefaultPersonFields = "names,emailAddresses,phoneNumbers,organizations,addresses,biographies,birthdays,urls,photos"
	// DefaultGroupFields contains the fields to request when listing/getting contact groups
	DefaultGroupFields = "name,groupType,memberCount"
)

// ContactsService defines the interface for Google People API operations.
// This interface enables dependency injection and testing with mocks.
type ContactsService interface {
	// Contacts
	ListContacts(ctx context.Context, opts *ListContactsOptions) (*people.ListConnectionsResponse, error)
	GetContact(ctx context.Context, resourceName string, opts *GetContactOptions) (*people.Person, error)
	SearchContacts(ctx context.Context, query string, opts *SearchContactsOptions) (*people.SearchResponse, error)
	CreateContact(ctx context.Context, person *people.Person) (*people.Person, error)
	UpdateContact(ctx context.Context, resourceName string, person *people.Person, updatePersonFields string) (*people.Person, error)
	DeleteContact(ctx context.Context, resourceName string) error

	// Contact Groups
	ListContactGroups(ctx context.Context, opts *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error)
	GetContactGroup(ctx context.Context, resourceName string, opts *GetContactGroupOptions) (*people.ContactGroup, error)
	CreateContactGroup(ctx context.Context, name string) (*people.ContactGroup, error)
	UpdateContactGroup(ctx context.Context, resourceName string, name string) (*people.ContactGroup, error)
	DeleteContactGroup(ctx context.Context, resourceName string) error
	ModifyContactGroupMembers(ctx context.Context, resourceName string, addMembers, removeMembers []string) (*people.ModifyContactGroupMembersResponse, error)
}

// ListContactsOptions contains optional parameters for listing contacts.
type ListContactsOptions struct {
	PageSize     int32
	PageToken    string
	PersonFields string
	SortOrder    string
}

// GetContactOptions contains optional parameters for getting a contact.
type GetContactOptions struct {
	PersonFields string
}

// SearchContactsOptions contains optional parameters for searching contacts.
type SearchContactsOptions struct {
	PageSize int32
	ReadMask string
}

// ListContactGroupsOptions contains optional parameters for listing contact groups.
type ListContactGroupsOptions struct {
	PageSize    int32
	PageToken   string
	GroupFields string
}

// GetContactGroupOptions contains optional parameters for getting a contact group.
type GetContactGroupOptions struct {
	GroupFields string
	MaxMembers  int32
}

// RealContactsService wraps the People API client and implements ContactsService.
type RealContactsService struct {
	service *people.Service
}

// NewRealContactsService creates a new RealContactsService wrapping the given People API service.
func NewRealContactsService(service *people.Service) *RealContactsService {
	return &RealContactsService{service: service}
}

// ListContacts lists all contacts for the authenticated user.
func (s *RealContactsService) ListContacts(ctx context.Context, opts *ListContactsOptions) (*people.ListConnectionsResponse, error) {
	call := s.service.People.Connections.List("people/me").Context(ctx)

	personFields := DefaultPersonFields
	if opts != nil {
		if opts.PageSize > 0 {
			call = call.PageSize(int64(opts.PageSize))
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.PersonFields != "" {
			personFields = opts.PersonFields
		}
		if opts.SortOrder != "" {
			call = call.SortOrder(opts.SortOrder)
		}
	}
	call = call.PersonFields(personFields)

	return call.Do()
}

// GetContact gets a single contact by resource name.
func (s *RealContactsService) GetContact(ctx context.Context, resourceName string, opts *GetContactOptions) (*people.Person, error) {
	personFields := DefaultPersonFields
	if opts != nil && opts.PersonFields != "" {
		personFields = opts.PersonFields
	}

	return s.service.People.Get(resourceName).
		Context(ctx).
		PersonFields(personFields).
		Do()
}

// SearchContacts searches contacts by query.
func (s *RealContactsService) SearchContacts(ctx context.Context, query string, opts *SearchContactsOptions) (*people.SearchResponse, error) {
	call := s.service.People.SearchContacts().
		Context(ctx).
		Query(query)

	readMask := DefaultPersonFields
	if opts != nil {
		if opts.PageSize > 0 {
			call = call.PageSize(int64(opts.PageSize))
		}
		if opts.ReadMask != "" {
			readMask = opts.ReadMask
		}
	}
	call = call.ReadMask(readMask)

	return call.Do()
}

// CreateContact creates a new contact.
func (s *RealContactsService) CreateContact(ctx context.Context, person *people.Person) (*people.Person, error) {
	return s.service.People.CreateContact(person).
		Context(ctx).
		Do()
}

// UpdateContact updates an existing contact.
func (s *RealContactsService) UpdateContact(ctx context.Context, resourceName string, person *people.Person, updatePersonFields string) (*people.Person, error) {
	return s.service.People.UpdateContact(resourceName, person).
		Context(ctx).
		UpdatePersonFields(updatePersonFields).
		Do()
}

// DeleteContact deletes a contact.
func (s *RealContactsService) DeleteContact(ctx context.Context, resourceName string) error {
	_, err := s.service.People.DeleteContact(resourceName).
		Context(ctx).
		Do()
	return err
}

// ListContactGroups lists all contact groups.
func (s *RealContactsService) ListContactGroups(ctx context.Context, opts *ListContactGroupsOptions) (*people.ListContactGroupsResponse, error) {
	call := s.service.ContactGroups.List().Context(ctx)

	groupFields := DefaultGroupFields
	if opts != nil {
		if opts.PageSize > 0 {
			call = call.PageSize(int64(opts.PageSize))
		}
		if opts.PageToken != "" {
			call = call.PageToken(opts.PageToken)
		}
		if opts.GroupFields != "" {
			groupFields = opts.GroupFields
		}
	}
	call = call.GroupFields(groupFields)

	return call.Do()
}

// GetContactGroup gets a single contact group by resource name.
func (s *RealContactsService) GetContactGroup(ctx context.Context, resourceName string, opts *GetContactGroupOptions) (*people.ContactGroup, error) {
	call := s.service.ContactGroups.Get(resourceName).Context(ctx)

	groupFields := DefaultGroupFields
	if opts != nil {
		if opts.GroupFields != "" {
			groupFields = opts.GroupFields
		}
		if opts.MaxMembers > 0 {
			call = call.MaxMembers(int64(opts.MaxMembers))
		}
	}
	call = call.GroupFields(groupFields)

	return call.Do()
}

// CreateContactGroup creates a new contact group.
func (s *RealContactsService) CreateContactGroup(ctx context.Context, name string) (*people.ContactGroup, error) {
	req := &people.CreateContactGroupRequest{
		ContactGroup: &people.ContactGroup{
			Name: name,
		},
	}
	return s.service.ContactGroups.Create(req).Context(ctx).Do()
}

// UpdateContactGroup updates a contact group's name.
func (s *RealContactsService) UpdateContactGroup(ctx context.Context, resourceName string, name string) (*people.ContactGroup, error) {
	req := &people.UpdateContactGroupRequest{
		ContactGroup: &people.ContactGroup{
			Name: name,
		},
	}
	return s.service.ContactGroups.Update(resourceName, req).Context(ctx).Do()
}

// DeleteContactGroup deletes a contact group.
func (s *RealContactsService) DeleteContactGroup(ctx context.Context, resourceName string) error {
	_, err := s.service.ContactGroups.Delete(resourceName).
		Context(ctx).
		Do()
	return err
}

// ModifyContactGroupMembers adds or removes members from a contact group.
func (s *RealContactsService) ModifyContactGroupMembers(ctx context.Context, resourceName string, addMembers, removeMembers []string) (*people.ModifyContactGroupMembersResponse, error) {
	req := &people.ModifyContactGroupMembersRequest{
		ResourceNamesToAdd:    addMembers,
		ResourceNamesToRemove: removeMembers,
	}
	return s.service.ContactGroups.Members.Modify(resourceName, req).
		Context(ctx).
		Do()
}
