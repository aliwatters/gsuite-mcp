package contacts

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Contacts tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Contacts Read Operations (Phase 1) ===

	// contacts_list - List contacts
	s.AddTool(mcp.NewTool("contacts_list",
		mcp.WithDescription("List contacts from Google Contacts."),
		mcp.WithNumber("page_size", mcp.Description("Maximum contacts to return (1-1000, default 100)")),
		common.WithPageToken(),
		mcp.WithString("sort_order", mcp.Description("Sort order: LAST_MODIFIED_ASCENDING, LAST_MODIFIED_DESCENDING, FIRST_NAME_ASCENDING, LAST_NAME_ASCENDING")),
		common.WithAccountParam(),
	), HandleContactsList)

	// contacts_get - Get contact details
	s.AddTool(mcp.NewTool("contacts_get",
		mcp.WithDescription("Get full details for a contact."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Contact resource name (e.g., 'people/c123456')")),
		common.WithAccountParam(),
	), HandleContactsGet)

	// contacts_search - Search contacts
	s.AddTool(mcp.NewTool("contacts_search",
		mcp.WithDescription("Search contacts by name, email, or phone number."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (matches name, email, phone)")),
		mcp.WithNumber("page_size", mcp.Description("Maximum results (1-30, default 30)")),
		common.WithAccountParam(),
	), HandleContactsSearch)

	// === Contacts Write Operations (Phase 2) ===

	// contacts_create - Create a contact
	s.AddTool(mcp.NewTool("contacts_create",
		mcp.WithDescription("Create a new contact."),
		mcp.WithString("given_name", mcp.Required(), mcp.Description("First name")),
		mcp.WithString("family_name", mcp.Description("Last name")),
		mcp.WithString("email", mcp.Description("Email address")),
		mcp.WithString("phone", mcp.Description("Phone number")),
		mcp.WithString("company", mcp.Description("Company name")),
		mcp.WithString("job_title", mcp.Description("Job title")),
		mcp.WithString("notes", mcp.Description("Notes/biography")),
		common.WithAccountParam(),
	), HandleContactsCreate)

	// contacts_update - Update a contact
	s.AddTool(mcp.NewTool("contacts_update",
		mcp.WithDescription("Update an existing contact. Only provided fields are updated."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Contact resource name (e.g., 'people/c123456')")),
		mcp.WithString("given_name", mcp.Description("First name")),
		mcp.WithString("family_name", mcp.Description("Last name")),
		mcp.WithString("email", mcp.Description("Email address")),
		mcp.WithString("phone", mcp.Description("Phone number")),
		mcp.WithString("company", mcp.Description("Company name")),
		mcp.WithString("job_title", mcp.Description("Job title")),
		mcp.WithString("notes", mcp.Description("Notes/biography")),
		common.WithAccountParam(),
	), HandleContactsUpdate)

	// contacts_delete - Delete a contact
	s.AddTool(mcp.NewTool("contacts_delete",
		mcp.WithDescription("Delete a contact."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Contact resource name (e.g., 'people/c123456')")),
		common.WithAccountParam(),
	), HandleContactsDelete)

	// === Contact Groups (Phase 3) ===

	// contacts_list_groups - List contact groups
	s.AddTool(mcp.NewTool("contacts_list_groups",
		mcp.WithDescription("List all contact groups (labels)."),
		mcp.WithNumber("page_size", mcp.Description("Maximum groups to return (1-200, default 100)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleContactsListGroups)

	// contacts_get_group - Get contact group details
	s.AddTool(mcp.NewTool("contacts_get_group",
		mcp.WithDescription("Get details for a contact group including member list."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Group resource name (e.g., 'contactGroups/123')")),
		mcp.WithNumber("max_members", mcp.Description("Maximum members to return (default: all)")),
		common.WithAccountParam(),
	), HandleContactsGetGroup)

	// contacts_create_group - Create a contact group
	s.AddTool(mcp.NewTool("contacts_create_group",
		mcp.WithDescription("Create a new contact group (label)."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Group name")),
		common.WithAccountParam(),
	), HandleContactsCreateGroup)

	// contacts_update_group - Update a contact group
	s.AddTool(mcp.NewTool("contacts_update_group",
		mcp.WithDescription("Rename a contact group."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Group resource name (e.g., 'contactGroups/123')")),
		mcp.WithString("name", mcp.Required(), mcp.Description("New group name")),
		common.WithAccountParam(),
	), HandleContactsUpdateGroup)

	// contacts_delete_group - Delete a contact group
	s.AddTool(mcp.NewTool("contacts_delete_group",
		mcp.WithDescription("Delete a contact group."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Group resource name (e.g., 'contactGroups/123')")),
		common.WithAccountParam(),
	), HandleContactsDeleteGroup)

	// contacts_modify_group_members - Add/remove members from a group
	s.AddTool(mcp.NewTool("contacts_modify_group_members",
		mcp.WithDescription("Add or remove contacts from a contact group."),
		mcp.WithString("resource_name", mcp.Required(), mcp.Description("Group resource name (e.g., 'contactGroups/123')")),
		mcp.WithArray("add_members", mcp.Description("Contact resource names to add")),
		mcp.WithArray("remove_members", mcp.Description("Contact resource names to remove")),
		common.WithAccountParam(),
	), HandleContactsModifyGroupMembers)
}
