package gmail

import (
	"encoding/base64"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/gmail/v1"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleGmailSearch      = common.WrapHandler[GmailService](TestableGmailSearch)
	HandleGmailGetMessage  = common.WrapHandler[GmailService](TestableGmailGetMessage)
	HandleGmailGetMessages = common.WrapHandler[GmailService](TestableGmailGetMessages)
	HandleGmailGetThread   = common.WrapHandler[GmailService](TestableGmailGetThread)
	HandleGmailSend        = common.WrapHandler[GmailService](TestableGmailSend)
	HandleGmailReply       = common.WrapHandler[GmailService](TestableGmailReply)
	HandleGmailDraft       = common.WrapHandler[GmailService](TestableGmailDraft)
	HandleGmailListLabels  = common.WrapHandler[GmailService](TestableGmailListLabels)
)

// === Message formatting utilities (used by Testable functions) ===

// BodyFormat specifies the preferred format for email body extraction.
type BodyFormat string

const (
	// BodyFormatText prefers text/plain content (default, reduces token usage)
	BodyFormatText BodyFormat = "text"
	// BodyFormatHTML prefers text/html content
	BodyFormatHTML BodyFormat = "html"
	// BodyFormatFull returns both text and html if available
	BodyFormatFull BodyFormat = "full"
)

// FormatMessageOptions configures how messages are formatted.
type FormatMessageOptions struct {
	BodyFormat BodyFormat
}

// formatMessage extracts useful fields from a Gmail message
func FormatMessage(msg *gmail.Message) map[string]any {
	return FormatMessageWithOptions(msg, FormatMessageOptions{BodyFormat: BodyFormatText})
}

// FormatMessageWithOptions extracts useful fields from a Gmail message with configurable options.
func FormatMessageWithOptions(msg *gmail.Message, opts FormatMessageOptions) map[string]any {
	result := map[string]any{
		"id":        msg.Id,
		"thread_id": msg.ThreadId,
		"label_ids": msg.LabelIds,
		"snippet":   msg.Snippet,
	}

	if msg.Payload != nil {
		headers := make(map[string]string)
		for _, h := range msg.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "from", "to", "cc", "bcc", "subject", "date", "message-id":
				headers[strings.ToLower(h.Name)] = h.Value
			}
		}
		result["headers"] = headers

		// Extract body based on format preference
		switch opts.BodyFormat {
		case BodyFormatFull:
			text, html := ExtractBodyParts(msg.Payload)
			if text != "" {
				result["body_text"] = text
			}
			if html != "" {
				result["body_html"] = html
			}
			// Also include "body" for backward compatibility (prefer text)
			if text != "" {
				result["body"] = text
			} else if html != "" {
				result["body"] = html
			}
		case BodyFormatHTML:
			body := ExtractBodyPreferHTML(msg.Payload)
			if body != "" {
				result["body"] = body
			}
		default: // BodyFormatText
			body := ExtractBody(msg.Payload)
			if body != "" {
				result["body"] = body
			}
		}
	}

	if msg.InternalDate > 0 {
		result["internal_date"] = msg.InternalDate
	}

	// Surface attachment metadata so callers can pair gmail_get* with
	// gmail_get_attachment. The Gmail API exposes attachment_id only via
	// payload.parts[].body — historically dropped by this wrapper, which left
	// gmail_get_attachment unreachable in practice (#156).
	if attachments := ExtractAttachments(msg.Payload); len(attachments) > 0 {
		result["attachments"] = attachments
	}

	return result
}

// ExtractBody gets the message body from parts, preferring text/plain.
func ExtractBody(payload *gmail.MessagePart) string {
	text, html := ExtractBodyParts(payload)
	if text != "" {
		return text
	}
	return html
}

// ExtractBodyPreferHTML gets the message body from parts, preferring text/html.
func ExtractBodyPreferHTML(payload *gmail.MessagePart) string {
	text, html := ExtractBodyParts(payload)
	if html != "" {
		return html
	}
	return text
}

// ExtractAttachments walks the payload tree and returns one entry per part
// that carries a Body.AttachmentId, with the four fields a caller needs to
// invoke gmail_get_attachment: attachment_id, filename, mime_type, size.
// part_id is included for traceability (Gmail uses dotted paths like "0.1").
//
// Returns an empty slice if there are no attachments — callers should treat
// nil and empty identically.
func ExtractAttachments(payload *gmail.MessagePart) []map[string]any {
	if payload == nil {
		return nil
	}
	var out []map[string]any
	walkParts(payload, &out)
	return out
}

// walkParts is the recursion helper for ExtractAttachments. It appends one
// entry per attachment-bearing part encountered anywhere in the part tree.
func walkParts(part *gmail.MessagePart, out *[]map[string]any) {
	if part == nil {
		return
	}
	if part.Body != nil && part.Body.AttachmentId != "" {
		entry := map[string]any{
			"attachment_id": part.Body.AttachmentId,
			"filename":      part.Filename,
			"mime_type":     part.MimeType,
			"size":          part.Body.Size,
		}
		if part.PartId != "" {
			entry["part_id"] = part.PartId
		}
		*out = append(*out, entry)
	}
	for _, p := range part.Parts {
		walkParts(p, out)
	}
}

// ExtractBodyParts extracts both text/plain and text/html bodies from a message.
// Returns (textBody, htmlBody). Either may be empty if not present.
func ExtractBodyParts(payload *gmail.MessagePart) (textBody, htmlBody string) {
	if payload == nil {
		return "", ""
	}

	// Check if body is directly in payload (single-part message)
	if payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			decoded := string(data)
			// Determine type based on MIME type
			if strings.HasPrefix(payload.MimeType, "text/html") {
				return "", decoded
			}
			// Default to text for text/plain or unknown
			return decoded, ""
		}
	}

	// Check parts for text/plain and text/html
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil && textBody == "" {
				textBody = string(data)
			}
		} else if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			data, err := base64.URLEncoding.DecodeString(part.Body.Data)
			if err == nil && htmlBody == "" {
				htmlBody = string(data)
			}
		} else if len(part.Parts) > 0 {
			// Recursively check nested parts
			nestedText, nestedHTML := ExtractBodyParts(part)
			if textBody == "" && nestedText != "" {
				textBody = nestedText
			}
			if htmlBody == "" && nestedHTML != "" {
				htmlBody = nestedHTML
			}
		}
	}

	return textBody, htmlBody
}
