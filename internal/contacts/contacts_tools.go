package contacts

import (
	"slices"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/people/v1"
)

// Resource name prefix constants.
const (
	PeoplePrefix        = "people/"
	ContactGroupsPrefix = "contactGroups/"
)

// ensurePrefix ensures a resource name has the given prefix.
func ensurePrefix(name, prefix string) string {
	if !strings.HasPrefix(name, prefix) {
		return prefix + name
	}
	return name
}

// contactFieldEntry represents a parsed value+type pair from request arguments.
type contactFieldEntry struct {
	Value string
	Type  string
}

// parseContactFields parses a list of value+type entries from request arguments.
// It supports both singular (string) and plural (array of strings or {value, type} maps) forms.
// Returns nil if no arguments are present; returns a non-nil (possibly empty) slice
// if the plural key is present, to allow callers to distinguish "not provided" from "clear all".
func parseContactFields(args map[string]any, pluralKey, singularKey string) []contactFieldEntry {
	if raw, ok := args[pluralKey].([]any); ok {
		entries := make([]contactFieldEntry, 0, len(raw))
		for _, item := range raw {
			if str, ok := item.(string); ok {
				entries = append(entries, contactFieldEntry{Value: str})
			} else if m, ok := item.(map[string]any); ok {
				entry := contactFieldEntry{}
				if v, ok := m["value"].(string); ok {
					entry.Value = v
				}
				if t, ok := m["type"].(string); ok {
					entry.Type = t
				}
				entries = append(entries, entry)
			}
		}
		return entries
	}
	if val, ok := args[singularKey].(string); ok && val != "" {
		return []contactFieldEntry{{Value: val}}
	}
	return nil
}

// parseEmailsFromRequest parses email addresses from request arguments.
func parseEmailsFromRequest(args map[string]any) []*people.EmailAddress {
	entries := parseContactFields(args, "emails", "email")
	if entries == nil {
		return nil
	}
	emails := make([]*people.EmailAddress, 0, len(entries))
	for _, e := range entries {
		emails = append(emails, &people.EmailAddress{Value: e.Value, Type: e.Type})
	}
	return emails
}

// parsePhonesFromRequest parses phone numbers from request arguments.
func parsePhonesFromRequest(args map[string]any) []*people.PhoneNumber {
	entries := parseContactFields(args, "phones", "phone")
	if entries == nil {
		return nil
	}
	phones := make([]*people.PhoneNumber, 0, len(entries))
	for _, e := range entries {
		phones = append(phones, &people.PhoneNumber{Value: e.Value, Type: e.Type})
	}
	return phones
}

// applyContactUpdates applies field updates from request arguments to an existing contact.
// Returns the list of People API update field masks for the fields that were modified.
func applyContactUpdates(existing *people.Person, args map[string]any) []string {
	var updateFields []string

	if givenName, ok := args["given_name"].(string); ok {
		ensureNames(existing)
		existing.Names[0].GivenName = givenName
		updateFields = append(updateFields, "names")
	}
	if familyName, ok := args["family_name"].(string); ok {
		ensureNames(existing)
		existing.Names[0].FamilyName = familyName
		if !slices.Contains(updateFields, "names") {
			updateFields = append(updateFields, "names")
		}
	}

	if emails := parseEmailsFromRequest(args); emails != nil {
		existing.EmailAddresses = emails
		updateFields = append(updateFields, "emailAddresses")
	}

	if phones := parsePhonesFromRequest(args); phones != nil {
		existing.PhoneNumbers = phones
		updateFields = append(updateFields, "phoneNumbers")
	}

	if company, ok := args["company"].(string); ok {
		ensureOrganizations(existing)
		existing.Organizations[0].Name = company
		updateFields = append(updateFields, "organizations")
	}
	if title, ok := args["job_title"].(string); ok {
		ensureOrganizations(existing)
		existing.Organizations[0].Title = title
		if !slices.Contains(updateFields, "organizations") {
			updateFields = append(updateFields, "organizations")
		}
	}

	if notes, ok := args["notes"].(string); ok {
		existing.Biographies = []*people.Biography{{Value: notes}}
		updateFields = append(updateFields, "biographies")
	}

	return updateFields
}

// ensureNames ensures the person has at least one Name entry.
func ensureNames(p *people.Person) {
	if len(p.Names) == 0 {
		p.Names = []*people.Name{{}}
	}
}

// ensureOrganizations ensures the person has at least one Organization entry.
func ensureOrganizations(p *people.Person) {
	if len(p.Organizations) == 0 {
		p.Organizations = []*people.Organization{{}}
	}
}

// === Handle functions - generated via WrapHandler ===

var (
	HandleContactsList               = common.WrapHandler[ContactsService](TestableContactsList)
	HandleContactsGet                = common.WrapHandler[ContactsService](TestableContactsGet)
	HandleContactsSearch             = common.WrapHandler[ContactsService](TestableContactsSearch)
	HandleContactsCreate             = common.WrapHandler[ContactsService](TestableContactsCreate)
	HandleContactsUpdate             = common.WrapHandler[ContactsService](TestableContactsUpdate)
	HandleContactsDelete             = common.WrapHandler[ContactsService](TestableContactsDelete)
	HandleContactsListGroups         = common.WrapHandler[ContactsService](TestableContactsListGroups)
	HandleContactsGetGroup           = common.WrapHandler[ContactsService](TestableContactsGetGroup)
	HandleContactsCreateGroup        = common.WrapHandler[ContactsService](TestableContactsCreateGroup)
	HandleContactsUpdateGroup        = common.WrapHandler[ContactsService](TestableContactsUpdateGroup)
	HandleContactsDeleteGroup        = common.WrapHandler[ContactsService](TestableContactsDeleteGroup)
	HandleContactsModifyGroupMembers = common.WrapHandler[ContactsService](TestableContactsModifyGroupMembers)
)

// Helper functions

// formatContact formats a Person for listing
func formatContact(person *people.Person) map[string]any {
	result := map[string]any{
		"resource_name": person.ResourceName,
	}

	if len(person.Names) > 0 {
		result["name"] = person.Names[0].DisplayName
		if result["name"] == "" {
			result["name"] = person.Names[0].GivenName + " " + person.Names[0].FamilyName
		}
	}

	if len(person.EmailAddresses) > 0 {
		result["email"] = person.EmailAddresses[0].Value
	}

	if len(person.PhoneNumbers) > 0 {
		result["phone"] = person.PhoneNumbers[0].Value
	}

	if len(person.Organizations) > 0 {
		result["company"] = person.Organizations[0].Name
	}

	if len(person.Photos) > 0 {
		result["photo_url"] = person.Photos[0].Url
	}

	return result
}

// formatContactFull formats a Person with all details
func formatContactFull(person *people.Person) map[string]any {
	result := formatContact(person)
	result["etag"] = person.Etag

	if len(person.Names) > 0 {
		result["given_name"] = person.Names[0].GivenName
		result["family_name"] = person.Names[0].FamilyName
	}

	if len(person.Organizations) > 0 {
		result["job_title"] = person.Organizations[0].Title
	}

	if len(person.Addresses) > 0 {
		addresses := make([]map[string]any, 0, len(person.Addresses))
		for _, addr := range person.Addresses {
			addresses = append(addresses, map[string]any{
				"formatted_value": addr.FormattedValue,
				"type":            addr.Type,
				"city":            addr.City,
				"country":         addr.Country,
			})
		}
		result["addresses"] = addresses
	}

	if len(person.Biographies) > 0 {
		result["notes"] = person.Biographies[0].Value
	}

	if len(person.Birthdays) > 0 && person.Birthdays[0].Date != nil {
		birthday := person.Birthdays[0].Date
		result["birthday"] = map[string]any{
			"year":  birthday.Year,
			"month": birthday.Month,
			"day":   birthday.Day,
		}
	}

	if len(person.Urls) > 0 {
		urls := make([]map[string]any, 0, len(person.Urls))
		for _, url := range person.Urls {
			urls = append(urls, map[string]any{
				"value": url.Value,
				"type":  url.Type,
			})
		}
		result["urls"] = urls
	}

	// All emails
	if len(person.EmailAddresses) > 1 {
		emails := make([]map[string]any, 0, len(person.EmailAddresses))
		for _, email := range person.EmailAddresses {
			emails = append(emails, map[string]any{
				"value": email.Value,
				"type":  email.Type,
			})
		}
		result["emails"] = emails
	}

	// All phones
	if len(person.PhoneNumbers) > 1 {
		phones := make([]map[string]any, 0, len(person.PhoneNumbers))
		for _, phone := range person.PhoneNumbers {
			phones = append(phones, map[string]any{
				"value": phone.Value,
				"type":  phone.Type,
			})
		}
		result["phones"] = phones
	}

	return result
}

// formatContactGroup formats a ContactGroup for listing
func formatContactGroup(group *people.ContactGroup) map[string]any {
	return map[string]any{
		"resource_name": group.ResourceName,
		"name":          group.Name,
		"group_type":    group.GroupType,
		"member_count":  group.MemberCount,
	}
}

// formatContactGroupFull formats a ContactGroup with all details
func formatContactGroupFull(group *people.ContactGroup) map[string]any {
	result := formatContactGroup(group)
	result["etag"] = group.Etag
	result["formatted_name"] = group.FormattedName

	if group.Metadata != nil {
		result["update_time"] = group.Metadata.UpdateTime
	}

	if len(group.MemberResourceNames) > 0 {
		result["member_resource_names"] = group.MemberResourceNames
	}

	return result
}
