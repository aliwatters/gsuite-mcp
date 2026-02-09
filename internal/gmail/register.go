package gmail

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Gmail tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Phase 1: Gmail Core ===

	// gmail_search - Search messages with query
	s.AddTool(mcp.NewTool("gmail_search",
		mcp.WithDescription("Search Gmail messages with query. Returns message IDs for use with gmail_get_message/gmail_get_messages."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Gmail search query (e.g., 'is:unread', 'from:amazon newer_than:7d')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum results to return (1-100, default 20)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleGmailSearch)

	// gmail_get_message - Read single message
	s.AddTool(mcp.NewTool("gmail_get_message",
		mcp.WithDescription("Get a single Gmail message by ID. Returns full message with headers and body."),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		mcp.WithString("format", mcp.Description("Response format: full (default), metadata, minimal, raw")),
		mcp.WithString("body_format", mcp.Description("Body content format: text (default, plain text for reduced tokens), html (full HTML), full (both text and html)")),
		common.WithAccountParam(),
	), HandleGmailGetMessage)

	// gmail_get_messages - Read batch of messages
	s.AddTool(mcp.NewTool("gmail_get_messages",
		mcp.WithDescription("Get multiple Gmail messages by ID (max 25). More efficient than multiple gmail_get_message calls."),
		mcp.WithArray("message_ids", mcp.Required(), mcp.Description("Array of Gmail message IDs (max 25)")),
		mcp.WithString("format", mcp.Description("Response format: full (default), metadata, minimal")),
		mcp.WithString("body_format", mcp.Description("Body content format: text (default, plain text for reduced tokens), html (full HTML), full (both text and html)")),
		common.WithAccountParam(),
	), HandleGmailGetMessages)

	// gmail_get_thread - Read full conversation
	s.AddTool(mcp.NewTool("gmail_get_thread",
		mcp.WithDescription("Get all messages in a Gmail thread/conversation."),
		mcp.WithString("thread_id", mcp.Required(), mcp.Description("Gmail thread ID")),
		mcp.WithString("format", mcp.Description("Response format: full (default), metadata, minimal")),
		mcp.WithString("body_format", mcp.Description("Body content format: text (default, plain text for reduced tokens), html (full HTML), full (both text and html)")),
		common.WithAccountParam(),
	), HandleGmailGetThread)

	// gmail_send - Send new email
	s.AddTool(mcp.NewTool("gmail_send",
		mcp.WithDescription("Send a new email message. For replies, use gmail_reply instead."),
		mcp.WithString("to", mcp.Required(), mcp.Description("Recipient email address(es), comma-separated")),
		mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Email body (plain text)")),
		mcp.WithString("cc", mcp.Description("CC recipients, comma-separated")),
		mcp.WithString("bcc", mcp.Description("BCC recipients, comma-separated")),
		common.WithAccountParam(),
	), HandleGmailSend)

	// gmail_reply - Reply to an email
	s.AddTool(mcp.NewTool("gmail_reply",
		mcp.WithDescription("Reply to an existing email. Keeps the conversation in the same thread with proper headers."),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("ID of the message to reply to")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Reply body (plain text)")),
		mcp.WithBoolean("reply_all", mcp.Description("Reply to all recipients (default: false)")),
		mcp.WithString("to", mcp.Description("Override recipient (default: reply to sender)")),
		mcp.WithString("cc", mcp.Description("Additional CC recipients")),
		common.WithAccountParam(),
	), HandleGmailReply)

	// gmail_draft - Create draft
	s.AddTool(mcp.NewTool("gmail_draft",
		mcp.WithDescription("Create a draft email."),
		mcp.WithString("to", mcp.Description("Recipient email address(es)")),
		mcp.WithString("subject", mcp.Description("Email subject")),
		mcp.WithString("body", mcp.Description("Email body (plain text)")),
		mcp.WithString("cc", mcp.Description("CC recipients")),
		mcp.WithString("bcc", mcp.Description("BCC recipients")),
		mcp.WithString("thread_id", mcp.Description("Thread ID for reply drafts")),
		common.WithAccountParam(),
	), HandleGmailDraft)

	// gmail_list_labels - List all labels
	s.AddTool(mcp.NewTool("gmail_list_labels",
		mcp.WithDescription("List all Gmail labels for an account with message/thread counts."),
		common.WithAccountParam(),
	), HandleGmailListLabels)

	// === Phase 2: Gmail Management ===

	// gmail_modify_message - Add/remove labels from a message
	s.AddTool(mcp.NewTool("gmail_modify_message",
		mcp.WithDescription("Modify Gmail message labels (archive, mark read/unread, star, trash)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
		mcp.WithArray("add_labels", mcp.Description("Labels to add (e.g., STARRED, TRASH)")),
		mcp.WithArray("remove_labels", mcp.Description("Labels to remove (e.g., UNREAD, INBOX)")),
	), HandleGmailModifyMessage)

	// gmail_batch_modify - Batch modify multiple messages
	s.AddTool(mcp.NewTool("gmail_batch_modify",
		mcp.WithDescription("Batch modify labels on multiple Gmail messages"),
		common.WithAccountParam(),
		mcp.WithArray("message_ids", mcp.Required(), mcp.Description("List of Gmail message IDs")),
		mcp.WithArray("add_labels", mcp.Description("Labels to add to all messages")),
		mcp.WithArray("remove_labels", mcp.Description("Labels to remove from all messages")),
	), HandleGmailBatchModify)

	// gmail_trash - Move message to trash
	s.AddTool(mcp.NewTool("gmail_trash",
		mcp.WithDescription("Move Gmail message to trash"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailTrash)

	// gmail_archive - Archive message (remove from inbox)
	s.AddTool(mcp.NewTool("gmail_archive",
		mcp.WithDescription("Archive Gmail message (remove INBOX label)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailArchive)

	// gmail_mark_read - Mark message as read
	s.AddTool(mcp.NewTool("gmail_mark_read",
		mcp.WithDescription("Mark Gmail message as read (remove UNREAD label)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailMarkRead)

	// gmail_batch_archive - Archive multiple messages
	s.AddTool(mcp.NewTool("gmail_batch_archive",
		mcp.WithDescription("Archive multiple Gmail messages"),
		common.WithAccountParam(),
		mcp.WithArray("message_ids", mcp.Required(), mcp.Description("List of Gmail message IDs")),
	), HandleGmailBatchArchive)

	// gmail_batch_trash - Trash multiple messages
	s.AddTool(mcp.NewTool("gmail_batch_trash",
		mcp.WithDescription("Move multiple Gmail messages to trash"),
		common.WithAccountParam(),
		mcp.WithArray("message_ids", mcp.Required(), mcp.Description("List of Gmail message IDs")),
	), HandleGmailBatchTrash)

	// gmail_untrash - Restore message from trash
	s.AddTool(mcp.NewTool("gmail_untrash",
		mcp.WithDescription("Restore Gmail message from trash"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailUntrash)

	// gmail_mark_unread - Mark message as unread
	s.AddTool(mcp.NewTool("gmail_mark_unread",
		mcp.WithDescription("Mark Gmail message as unread (add UNREAD label)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailMarkUnread)

	// gmail_star - Star a message
	s.AddTool(mcp.NewTool("gmail_star",
		mcp.WithDescription("Star a Gmail message (add STARRED label)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailStar)

	// gmail_unstar - Unstar a message
	s.AddTool(mcp.NewTool("gmail_unstar",
		mcp.WithDescription("Unstar a Gmail message (remove STARRED label)"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailUnstar)

	// === Phase 3: Gmail Extended ===

	// gmail_get_attachment - Download attachment
	s.AddTool(mcp.NewTool("gmail_get_attachment",
		mcp.WithDescription("Download attachment content by ID. Returns base64-encoded data."),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID containing the attachment")),
		mcp.WithString("attachment_id", mcp.Required(), mcp.Description("Attachment ID from message payload")),
		common.WithAccountParam(),
	), HandleGmailGetAttachment)

	// gmail_list_filters - List all filters
	s.AddTool(mcp.NewTool("gmail_list_filters",
		mcp.WithDescription("List all Gmail filters for an account"),
		common.WithAccountParam(),
	), HandleGmailListFilters)

	// gmail_create_filter - Create new filter
	s.AddTool(mcp.NewTool("gmail_create_filter",
		mcp.WithDescription("Create a new Gmail filter with criteria and actions"),
		mcp.WithString("from", mcp.Description("Match messages from this sender")),
		mcp.WithString("to", mcp.Description("Match messages to this recipient")),
		mcp.WithString("subject", mcp.Description("Match messages with this subject")),
		mcp.WithString("query", mcp.Description("Gmail search query to match")),
		mcp.WithBoolean("has_attachment", mcp.Description("Match messages with attachments")),
		mcp.WithArray("add_label_ids", mcp.Description("Labels to add to matching messages")),
		mcp.WithArray("remove_label_ids", mcp.Description("Labels to remove from matching messages")),
		mcp.WithString("forward", mcp.Description("Email address to forward matching messages to")),
		common.WithAccountParam(),
	), HandleGmailCreateFilter)

	// gmail_delete_filter - Delete filter
	s.AddTool(mcp.NewTool("gmail_delete_filter",
		mcp.WithDescription("Delete a Gmail filter by ID"),
		mcp.WithString("filter_id", mcp.Required(), mcp.Description("Filter ID to delete")),
		common.WithAccountParam(),
	), HandleGmailDeleteFilter)

	// gmail_create_label - Create new label
	s.AddTool(mcp.NewTool("gmail_create_label",
		mcp.WithDescription("Create a new Gmail label"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Label name")),
		mcp.WithString("label_list_visibility", mcp.Description("Visibility: labelShow, labelShowIfUnread, labelHide")),
		mcp.WithString("message_list_visibility", mcp.Description("Visibility: show, hide")),
		common.WithAccountParam(),
	), HandleGmailCreateLabel)

	// gmail_delete_label - Delete label
	s.AddTool(mcp.NewTool("gmail_delete_label",
		mcp.WithDescription("Delete a Gmail label by ID"),
		mcp.WithString("label_id", mcp.Required(), mcp.Description("Label ID to delete")),
		common.WithAccountParam(),
	), HandleGmailDeleteLabel)

	// gmail_update_label - Update label
	s.AddTool(mcp.NewTool("gmail_update_label",
		mcp.WithDescription("Update a Gmail label (rename or change visibility)"),
		mcp.WithString("label_id", mcp.Required(), mcp.Description("Label ID to update")),
		mcp.WithString("name", mcp.Description("New label name")),
		mcp.WithString("label_list_visibility", mcp.Description("Visibility: labelShow, labelShowIfUnread, labelHide")),
		mcp.WithString("message_list_visibility", mcp.Description("Visibility: show, hide")),
		common.WithAccountParam(),
	), HandleGmailUpdateLabel)

	// gmail_list_drafts - List all drafts
	s.AddTool(mcp.NewTool("gmail_list_drafts",
		mcp.WithDescription("List all draft messages"),
		mcp.WithNumber("max_results", mcp.Description("Maximum results to return (1-100, default 20)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleGmailListDrafts)

	// gmail_get_draft - Get draft content
	s.AddTool(mcp.NewTool("gmail_get_draft",
		mcp.WithDescription("Get draft content by ID"),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("Draft ID")),
		mcp.WithString("format", mcp.Description("Response format: full (default), metadata, minimal")),
		common.WithAccountParam(),
	), HandleGmailGetDraft)

	// gmail_update_draft - Update draft
	s.AddTool(mcp.NewTool("gmail_update_draft",
		mcp.WithDescription("Update an existing draft"),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("Draft ID to update")),
		mcp.WithString("to", mcp.Description("Recipient email address(es)")),
		mcp.WithString("subject", mcp.Description("Email subject")),
		mcp.WithString("body", mcp.Description("Email body (plain text)")),
		mcp.WithString("cc", mcp.Description("CC recipients")),
		mcp.WithString("bcc", mcp.Description("BCC recipients")),
		common.WithAccountParam(),
	), HandleGmailUpdateDraft)

	// gmail_delete_draft - Delete draft
	s.AddTool(mcp.NewTool("gmail_delete_draft",
		mcp.WithDescription("Delete a draft"),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("Draft ID to delete")),
		common.WithAccountParam(),
	), HandleGmailDeleteDraft)

	// gmail_send_draft - Send draft
	s.AddTool(mcp.NewTool("gmail_send_draft",
		mcp.WithDescription("Send an existing draft"),
		mcp.WithString("draft_id", mcp.Required(), mcp.Description("Draft ID to send")),
		common.WithAccountParam(),
	), HandleGmailSendDraft)

	// gmail_thread_archive - Archive thread
	s.AddTool(mcp.NewTool("gmail_thread_archive",
		mcp.WithDescription("Archive entire conversation (remove INBOX label from all messages)"),
		mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID to archive")),
		common.WithAccountParam(),
	), HandleGmailThreadArchive)

	// gmail_thread_trash - Trash thread
	s.AddTool(mcp.NewTool("gmail_thread_trash",
		mcp.WithDescription("Move entire conversation to trash"),
		mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID to trash")),
		common.WithAccountParam(),
	), HandleGmailThreadTrash)

	// gmail_thread_untrash - Restore thread from trash
	s.AddTool(mcp.NewTool("gmail_thread_untrash",
		mcp.WithDescription("Restore entire conversation from trash"),
		mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID to restore")),
		common.WithAccountParam(),
	), HandleGmailThreadUntrash)

	// gmail_modify_thread - Modify thread labels
	s.AddTool(mcp.NewTool("gmail_modify_thread",
		mcp.WithDescription("Modify labels on entire conversation"),
		mcp.WithString("thread_id", mcp.Required(), mcp.Description("Thread ID to modify")),
		mcp.WithArray("add_labels", mcp.Description("Labels to add to all messages")),
		mcp.WithArray("remove_labels", mcp.Description("Labels to remove from all messages")),
		common.WithAccountParam(),
	), HandleGmailModifyThread)

	// gmail_get_profile - Get account profile
	s.AddTool(mcp.NewTool("gmail_get_profile",
		mcp.WithDescription("Get email address, message count, and thread count for account"),
		common.WithAccountParam(),
	), HandleGmailGetProfile)

	// gmail_get_vacation - Get vacation settings
	s.AddTool(mcp.NewTool("gmail_get_vacation",
		mcp.WithDescription("Get vacation auto-reply settings"),
		common.WithAccountParam(),
	), HandleGmailGetVacation)

	// gmail_set_vacation - Set vacation settings
	s.AddTool(mcp.NewTool("gmail_set_vacation",
		mcp.WithDescription("Set vacation auto-reply settings"),
		mcp.WithBoolean("enabled", mcp.Description("Enable or disable auto-reply")),
		mcp.WithString("subject", mcp.Description("Auto-reply subject")),
		mcp.WithString("body", mcp.Description("Auto-reply body (plain text)")),
		mcp.WithString("body_html", mcp.Description("Auto-reply body (HTML)")),
		mcp.WithBoolean("restrict_to_contacts", mcp.Description("Only reply to contacts")),
		mcp.WithBoolean("restrict_to_domain", mcp.Description("Only reply to same domain")),
		mcp.WithNumber("start_time", mcp.Description("Start time (Unix timestamp in ms)")),
		mcp.WithNumber("end_time", mcp.Description("End time (Unix timestamp in ms)")),
		common.WithAccountParam(),
	), HandleGmailSetVacation)

	// gmail_spam - Mark as spam
	s.AddTool(mcp.NewTool("gmail_spam",
		mcp.WithDescription("Move message to spam"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailSpam)

	// gmail_not_spam - Remove from spam
	s.AddTool(mcp.NewTool("gmail_not_spam",
		mcp.WithDescription("Remove message from spam and move to inbox"),
		mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
		common.WithAccountParam(),
	), HandleGmailNotSpam)
}
