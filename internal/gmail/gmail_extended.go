package gmail

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// === Handle functions - generated via WrapHandler ===

// Attachment Tools
var HandleGmailGetAttachment = common.WrapHandler[GmailService](TestableGmailGetAttachment)

// Filter Tools
var (
	HandleGmailListFilters  = common.WrapHandler[GmailService](TestableGmailListFilters)
	HandleGmailCreateFilter = common.WrapHandler[GmailService](TestableGmailCreateFilter)
	HandleGmailDeleteFilter = common.WrapHandler[GmailService](TestableGmailDeleteFilter)
)

// Label Management
var (
	HandleGmailCreateLabel = common.WrapHandler[GmailService](TestableGmailCreateLabel)
	HandleGmailDeleteLabel = common.WrapHandler[GmailService](TestableGmailDeleteLabel)
	HandleGmailUpdateLabel = common.WrapHandler[GmailService](TestableGmailUpdateLabel)
)

// Draft Management
var (
	HandleGmailListDrafts  = common.WrapHandler[GmailService](TestableGmailListDrafts)
	HandleGmailGetDraft    = common.WrapHandler[GmailService](TestableGmailGetDraft)
	HandleGmailUpdateDraft = common.WrapHandler[GmailService](TestableGmailUpdateDraft)
	HandleGmailDeleteDraft = common.WrapHandler[GmailService](TestableGmailDeleteDraft)
	HandleGmailSendDraft   = common.WrapHandler[GmailService](TestableGmailSendDraft)
)

// Thread Operations

// threadResult creates a standard MCP result from a Gmail thread response.
func threadResult(thread *gmail.Thread) (*mcp.CallToolResult, error) {
	result := map[string]any{
		"id":       thread.Id,
		"messages": len(thread.Messages),
	}

	return common.MarshalToolResult(result)
}

var (
	HandleGmailThreadArchive = common.WrapHandler[GmailService](TestableGmailThreadArchive)
	HandleGmailThreadTrash   = common.WrapHandler[GmailService](TestableGmailThreadTrash)
	HandleGmailThreadUntrash = common.WrapHandler[GmailService](TestableGmailThreadUntrash)
	HandleGmailModifyThread  = common.WrapHandler[GmailService](TestableGmailModifyThread)
)

// Account & Settings
var (
	HandleGmailGetProfile  = common.WrapHandler[GmailService](TestableGmailGetProfile)
	HandleGmailGetVacation = common.WrapHandler[GmailService](TestableGmailGetVacation)
	HandleGmailSetVacation = common.WrapHandler[GmailService](TestableGmailSetVacation)
)

// Spam Convenience
var (
	HandleGmailSpam    = common.WrapHandler[GmailService](TestableGmailSpam)
	HandleGmailNotSpam = common.WrapHandler[GmailService](TestableGmailNotSpam)
)
