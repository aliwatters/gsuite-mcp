package contacts

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/people/v1"
)

// === Phase 1: Read Operations ===

// TestableContactsList is the testable version of handleContactsList.
func TestableContactsList(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	opts := &ListContactsOptions{}

	// Person fields
	personFields := DefaultPersonFields
	if pf, ok := request.Params.Arguments["person_fields"].(string); ok && pf != "" {
		personFields = pf
	}
	opts.PersonFields = personFields

	// Page size (default 100, max 1000)
	pageSize := int32(common.ContactsDefaultPageSize)
	if ps, ok := request.Params.Arguments["page_size"].(float64); ok && ps > 0 {
		pageSize = int32(ps)
		if pageSize > common.ContactsMaxPageSize {
			pageSize = common.ContactsMaxPageSize
		}
	}
	opts.PageSize = pageSize

	// Page token for pagination
	if pageToken, ok := request.Params.Arguments["page_token"].(string); ok && pageToken != "" {
		opts.PageToken = pageToken
	}

	// Sort order
	if sortOrder, ok := request.Params.Arguments["sort_order"].(string); ok && sortOrder != "" {
		opts.SortOrder = sortOrder
	}

	resp, err := srv.ListContacts(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	contacts := make([]map[string]any, 0, len(resp.Connections))
	for _, p := range resp.Connections {
		contacts = append(contacts, formatContact(p))
	}

	result := map[string]any{
		"contacts":        contacts,
		"count":           len(contacts),
		"total_people":    resp.TotalPeople,
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableContactsGet is the testable version of handleContactsGet.
func TestableContactsGet(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, PeoplePrefix)

	// Person fields
	personFields := DefaultPersonFields
	if pf, ok := request.Params.Arguments["person_fields"].(string); ok && pf != "" {
		personFields = pf
	}

	opts := &GetContactOptions{
		PersonFields: personFields,
	}

	person, err := srv.GetContact(ctx, resourceName, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactFull(person)

	return common.MarshalToolResult(result)
}

// TestableContactsSearch is the testable version of handleContactsSearch.
func TestableContactsSearch(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	query, _ := request.Params.Arguments["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	opts := &SearchContactsOptions{}

	// Read mask (default person fields)
	readMask := DefaultPersonFields
	if rm, ok := request.Params.Arguments["read_mask"].(string); ok && rm != "" {
		readMask = rm
	}
	opts.ReadMask = readMask

	// Page size (default 30, max 30 for search)
	pageSize := int32(common.ContactsSearchDefaultPageSize)
	if ps, ok := request.Params.Arguments["page_size"].(float64); ok && ps > 0 {
		pageSize = int32(ps)
		if pageSize > common.ContactsSearchMaxPageSize {
			pageSize = common.ContactsSearchMaxPageSize
		}
	}
	opts.PageSize = pageSize

	resp, err := srv.SearchContacts(ctx, query, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	contacts := make([]map[string]any, 0)
	if resp.Results != nil {
		for _, r := range resp.Results {
			if r.Person != nil {
				contacts = append(contacts, formatContact(r.Person))
			}
		}
	}

	result := map[string]any{
		"contacts": contacts,
		"count":    len(contacts),
	}

	return common.MarshalToolResult(result)
}

// === Phase 2: Write Operations ===

// TestableContactsCreate is the testable version of handleContactsCreate.
func TestableContactsCreate(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	person := &people.Person{}

	// Names (given_name is required)
	givenName, _ := request.Params.Arguments["given_name"].(string)
	if givenName == "" {
		return mcp.NewToolResultError("given_name parameter is required"), nil
	}
	person.Names = []*people.Name{{GivenName: givenName}}
	if familyName, ok := request.Params.Arguments["family_name"].(string); ok && familyName != "" {
		person.Names[0].FamilyName = familyName
	}

	// Email addresses
	if emails := parseEmailsFromRequest(request.Params.Arguments); emails != nil {
		person.EmailAddresses = emails
	}

	// Phone numbers
	if phones := parsePhonesFromRequest(request.Params.Arguments); phones != nil {
		person.PhoneNumbers = phones
	}

	// Organization
	if company, ok := request.Params.Arguments["company"].(string); ok && company != "" {
		org := &people.Organization{Name: company}
		if title, ok := request.Params.Arguments["job_title"].(string); ok {
			org.Title = title
		}
		person.Organizations = []*people.Organization{org}
	}

	// Notes/biography
	if notes, ok := request.Params.Arguments["notes"].(string); ok && notes != "" {
		person.Biographies = []*people.Biography{{Value: notes}}
	}

	created, err := srv.CreateContact(ctx, person)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactFull(created)
	result["message"] = "Contact created successfully"

	return common.MarshalToolResult(result)
}

// TestableContactsUpdate is the testable version of handleContactsUpdate.
func TestableContactsUpdate(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, PeoplePrefix)

	existing, err := srv.GetContact(ctx, resourceName, &GetContactOptions{PersonFields: DefaultPersonFields})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get contact: %v", err)), nil
	}

	updateFields := applyContactUpdates(existing, request.Params.Arguments)
	if len(updateFields) == 0 {
		return mcp.NewToolResultError("No fields to update - provide at least one field to modify"), nil
	}

	updated, err := srv.UpdateContact(ctx, resourceName, existing, strings.Join(updateFields, ","))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactFull(updated)
	result["message"] = "Contact updated successfully"

	return common.MarshalToolResult(result)
}

// TestableContactsDelete is the testable version of handleContactsDelete.
func TestableContactsDelete(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, PeoplePrefix)

	err := srv.DeleteContact(ctx, resourceName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := map[string]any{
		"success":       true,
		"resource_name": resourceName,
		"message":       "Contact deleted successfully",
	}

	return common.MarshalToolResult(result)
}

// === Phase 3: Contact Groups ===

// TestableContactsListGroups is the testable version of handleContactsListGroups.
func TestableContactsListGroups(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	opts := &ListContactGroupsOptions{}

	// Group fields
	groupFields := DefaultGroupFields
	if gf, ok := request.Params.Arguments["group_fields"].(string); ok && gf != "" {
		groupFields = gf
	}
	opts.GroupFields = groupFields

	// Page size
	if pageSize, ok := request.Params.Arguments["page_size"].(float64); ok && pageSize > 0 {
		opts.PageSize = int32(pageSize)
	}

	// Page token
	if pageToken, ok := request.Params.Arguments["page_token"].(string); ok && pageToken != "" {
		opts.PageToken = pageToken
	}

	resp, err := srv.ListContactGroups(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	groups := make([]map[string]any, 0)
	if resp.ContactGroups != nil {
		for _, g := range resp.ContactGroups {
			groups = append(groups, formatContactGroup(g))
		}
	}

	result := map[string]any{
		"contact_groups":  groups,
		"count":           len(groups),
		"total_items":     resp.TotalItems,
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableContactsGetGroup is the testable version of handleContactsGetGroup.
func TestableContactsGetGroup(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, ContactGroupsPrefix)

	opts := &GetContactGroupOptions{}

	// Group fields
	groupFields := DefaultGroupFields + ",memberResourceNames"
	if gf, ok := request.Params.Arguments["group_fields"].(string); ok && gf != "" {
		groupFields = gf
	}
	opts.GroupFields = groupFields

	// Max members
	if maxMembers, ok := request.Params.Arguments["max_members"].(float64); ok && maxMembers > 0 {
		opts.MaxMembers = int32(maxMembers)
	}

	group, err := srv.GetContactGroup(ctx, resourceName, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactGroupFull(group)

	return common.MarshalToolResult(result)
}

// TestableContactsCreateGroup is the testable version of handleContactsCreateGroup.
func TestableContactsCreateGroup(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name, _ := request.Params.Arguments["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	group, err := srv.CreateContactGroup(ctx, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactGroup(group)
	result["message"] = "Contact group created successfully"

	return common.MarshalToolResult(result)
}

// TestableContactsUpdateGroup is the testable version of handleContactsUpdateGroup.
func TestableContactsUpdateGroup(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	name, _ := request.Params.Arguments["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, ContactGroupsPrefix)

	group, err := srv.UpdateContactGroup(ctx, resourceName, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := formatContactGroup(group)
	result["message"] = "Contact group updated successfully"

	return common.MarshalToolResult(result)
}

// TestableContactsDeleteGroup is the testable version of handleContactsDeleteGroup.
func TestableContactsDeleteGroup(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, ContactGroupsPrefix)

	err := srv.DeleteContactGroup(ctx, resourceName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := map[string]any{
		"success":       true,
		"resource_name": resourceName,
		"message":       "Contact group deleted successfully",
	}

	return common.MarshalToolResult(result)
}

// TestableContactsModifyGroupMembers is the testable version of handleContactsModifyGroupMembers.
func TestableContactsModifyGroupMembers(ctx context.Context, request mcp.CallToolRequest, deps *ContactsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveContactsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	resourceName, _ := request.Params.Arguments["resource_name"].(string)
	if resourceName == "" {
		return mcp.NewToolResultError("resource_name parameter is required"), nil
	}

	resourceName = ensurePrefix(resourceName, ContactGroupsPrefix)

	var addMembers, removeMembers []string

	// Add members
	if addRaw, ok := request.Params.Arguments["add_members"].([]any); ok && len(addRaw) > 0 {
		addMembers = make([]string, 0, len(addRaw))
		for _, m := range addRaw {
			if memberStr, ok := m.(string); ok {
				addMembers = append(addMembers, ensurePrefix(memberStr, PeoplePrefix))
			}
		}
	}

	// Remove members
	if removeRaw, ok := request.Params.Arguments["remove_members"].([]any); ok && len(removeRaw) > 0 {
		removeMembers = make([]string, 0, len(removeRaw))
		for _, m := range removeRaw {
			if memberStr, ok := m.(string); ok {
				removeMembers = append(removeMembers, ensurePrefix(memberStr, PeoplePrefix))
			}
		}
	}

	if len(addMembers) == 0 && len(removeMembers) == 0 {
		return mcp.NewToolResultError("Either add_members or remove_members must be provided"), nil
	}

	resp, err := srv.ModifyContactGroupMembers(ctx, resourceName, addMembers, removeMembers)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("People API error: %v", err)), nil
	}

	result := map[string]any{
		"success":                  true,
		"resource_name":            resourceName,
		"members_added":            len(addMembers),
		"members_removed":          len(removeMembers),
		"not_found_resource_names": resp.NotFoundResourceNames,
		"cannot_remove_last_contact_group_resource_names": resp.CanNotRemoveLastContactGroupResourceNames,
		"message": "Contact group members modified successfully",
	}

	return common.MarshalToolResult(result)
}
