package gmail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// emailContentType is the default Content-Type header for plain text emails.
const emailContentType = "Content-Type: text/plain; charset=\"UTF-8\""

const (
	gmailDefaultAttachmentMIMEType = "application/octet-stream"
	gmailMaxOutgoingRawBytes       = 25 * 1024 * 1024
)

// parseBodyFormat parses the "body_format" argument from a request and returns the
// corresponding BodyFormat. Defaults to BodyFormatText if not specified.
func parseBodyFormat(args map[string]any) BodyFormat {
	bf := common.ParseStringArg(args, "body_format", "")
	switch bf {
	case "html":
		return BodyFormatHTML
	case "full":
		return BodyFormatFull
	default:
		return BodyFormatText
	}
}

// extractAddRemoveLabels extracts "add_labels" and "remove_labels" string arrays
// from request arguments using extractStringArray.
func extractAddRemoveLabels(args map[string]any) (addLabels, removeLabels []string) {
	addLabels = extractStringArray(args["add_labels"])
	removeLabels = extractStringArray(args["remove_labels"])
	return
}

// buildMessageFromArgs builds a gmail.Message with a raw RFC 2822 body from common
// email arguments (to, subject, body, cc, bcc).
func buildMessageFromArgs(args map[string]any) *gmail.Message {
	raw := buildEmailMessage(EmailMessage{
		To:      common.ParseStringArg(args, "to", ""),
		Cc:      common.ParseStringArg(args, "cc", ""),
		Bcc:     common.ParseStringArg(args, "bcc", ""),
		Subject: common.ParseStringArg(args, "subject", ""),
		Body:    common.ParseStringArg(args, "body", ""),
	})
	return &gmail.Message{Raw: raw}
}

// buildMessageFromArgsWithAttachments builds a gmail.Message from common email
// arguments and optional local attachment file paths.
func buildMessageFromArgsWithAttachments(args map[string]any) (*gmail.Message, error) {
	attachments, err := loadEmailAttachments(args["attachments"])
	if err != nil {
		return nil, err
	}

	raw, err := buildEmailMessageWithAttachments(EmailMessage{
		To:      common.ParseStringArg(args, "to", ""),
		Cc:      common.ParseStringArg(args, "cc", ""),
		Bcc:     common.ParseStringArg(args, "bcc", ""),
		Subject: common.ParseStringArg(args, "subject", ""),
		Body:    common.ParseStringArg(args, "body", ""),
	}, attachments)
	if err != nil {
		return nil, err
	}

	return &gmail.Message{Raw: raw}, nil
}

// EmailMessage holds the components for building an RFC 2822 email message.
type EmailMessage struct {
	To           string
	Cc           string
	Bcc          string
	Subject      string
	Body         string
	ExtraHeaders map[string]string // Additional headers (e.g., In-Reply-To, References)
}

type emailAttachment struct {
	Path     string
	Filename string
	MIMEType string
	Data     []byte
}

// buildEmailMessage constructs an RFC 2822 email message and returns it
// as a base64 URL-encoded string suitable for the Gmail API Raw field.
// Fields that are empty strings are omitted from the headers.
func buildEmailMessage(msg EmailMessage) string {
	var b strings.Builder

	writeEmailHeaders(&b, msg)
	b.WriteString(emailContentType + "\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.Body)

	return base64.URLEncoding.EncodeToString([]byte(b.String()))
}

func buildEmailMessageWithAttachments(msg EmailMessage, attachments []emailAttachment) (string, error) {
	if len(attachments) == 0 {
		return buildEmailMessage(msg), nil
	}

	raw, err := buildMultipartEmailBytes(msg, attachments)
	if err != nil {
		return "", err
	}
	if len(raw) > gmailMaxOutgoingRawBytes {
		return "", fmt.Errorf("email with attachments is %d bytes, exceeding the 25 MiB Gmail send limit", len(raw))
	}

	return base64.URLEncoding.EncodeToString(raw), nil
}

func buildMultipartEmailBytes(msg EmailMessage, attachments []emailAttachment) ([]byte, error) {
	var b bytes.Buffer
	writeEmailHeaders(&b, msg)
	b.WriteString("MIME-Version: 1.0\r\n")

	writer := multipart.NewWriter(&b)
	contentType := mime.FormatMediaType("multipart/mixed", map[string]string{"boundary": writer.Boundary()})
	fmt.Fprintf(&b, "Content-Type: %s\r\n", contentType)
	b.WriteString("\r\n")

	bodyHeader := textproto.MIMEHeader{}
	bodyHeader.Set("Content-Type", "text/plain; charset=\"UTF-8\"")
	bodyHeader.Set("Content-Transfer-Encoding", "8bit")
	bodyPart, err := writer.CreatePart(bodyHeader)
	if err != nil {
		return nil, fmt.Errorf("create email body part: %w", err)
	}
	if _, err := io.WriteString(bodyPart, msg.Body); err != nil {
		return nil, fmt.Errorf("write email body part: %w", err)
	}

	for _, attachment := range attachments {
		attachmentHeader := textproto.MIMEHeader{}
		attachmentHeader.Set("Content-Type", formatAttachmentContentType(attachment.MIMEType, attachment.Filename))
		attachmentHeader.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": attachment.Filename}))
		attachmentHeader.Set("Content-Transfer-Encoding", "base64")

		attachmentPart, err := writer.CreatePart(attachmentHeader)
		if err != nil {
			return nil, fmt.Errorf("create attachment part for %q: %w", attachment.Path, err)
		}
		if err := writeWrappedBase64(attachmentPart, attachment.Data); err != nil {
			return nil, fmt.Errorf("write attachment %q: %w", attachment.Path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close email multipart body: %w", err)
	}

	return b.Bytes(), nil
}

func writeEmailHeaders(w io.Writer, msg EmailMessage) {
	if msg.To != "" {
		fmt.Fprintf(w, "To: %s\r\n", msg.To)
	}
	if msg.Cc != "" {
		fmt.Fprintf(w, "Cc: %s\r\n", msg.Cc)
	}
	if msg.Bcc != "" {
		fmt.Fprintf(w, "Bcc: %s\r\n", msg.Bcc)
	}
	if msg.Subject != "" {
		fmt.Fprintf(w, "Subject: %s\r\n", msg.Subject)
	}
	for key, value := range msg.ExtraHeaders {
		fmt.Fprintf(w, "%s: %s\r\n", key, value)
	}
}

func loadEmailAttachments(value any) ([]emailAttachment, error) {
	paths, err := parseAttachmentPaths(value)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}

	attachments := make([]emailAttachment, 0, len(paths))
	var totalSize int64
	for index, path := range paths {
		if path == "" {
			return nil, fmt.Errorf("attachments[%d] must be a non-empty file path", index)
		}

		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot access attachment %q: %w", path, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("attachment %q is a directory, not a file", path)
		}

		totalSize += info.Size()
		if totalSize > gmailMaxOutgoingRawBytes {
			return nil, fmt.Errorf("attachments exceed the 25 MiB Gmail send limit after adding %q", path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read attachment %q: %w", path, err)
		}

		attachments = append(attachments, emailAttachment{
			Path:     path,
			Filename: filepath.Base(path),
			MIMEType: detectAttachmentMIMEType(path),
			Data:     data,
		})
	}

	return attachments, nil
}

func parseAttachmentPaths(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	switch typed := value.(type) {
	case []any:
		paths := make([]string, 0, len(typed))
		for index, item := range typed {
			path, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("attachments[%d] must be a file path string", index)
			}
			paths = append(paths, path)
		}
		return paths, nil
	case []string:
		return typed, nil
	default:
		return nil, fmt.Errorf("attachments must be an array of file path strings")
	}
}

func detectAttachmentMIMEType(path string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		return gmailDefaultAttachmentMIMEType
	}
	return mimeType
}

func formatAttachmentContentType(mimeType, filename string) string {
	mediaType, params, err := mime.ParseMediaType(mimeType)
	if err != nil {
		mediaType = gmailDefaultAttachmentMIMEType
		params = map[string]string{}
	}
	if params == nil {
		params = map[string]string{}
	}
	params["name"] = filename

	contentType := mime.FormatMediaType(mediaType, params)
	if contentType == "" {
		return mime.FormatMediaType(gmailDefaultAttachmentMIMEType, map[string]string{"name": filename})
	}
	return contentType
}

func writeWrappedBase64(w io.Writer, data []byte) error {
	encoded := base64.StdEncoding.EncodeToString(data)
	for len(encoded) > 76 {
		if _, err := io.WriteString(w, encoded[:76]+"\r\n"); err != nil {
			return err
		}
		encoded = encoded[76:]
	}
	_, err := io.WriteString(w, encoded+"\r\n")
	return err
}

// modifyMessageLabels is a generic helper for single-message label modification
// operations (archive, star, mark read, spam, etc.). It extracts message_id from
// the request, applies the specified label changes, and returns a standard result.
func modifyMessageLabels(ctx context.Context, svc GmailService, request mcp.CallToolRequest, addLabels, removeLabels []string) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	modifyRequest := &gmail.ModifyMessageRequest{
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	msg, err := svc.ModifyMessage(ctx, messageID, modifyRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	return messageResult(msg)
}

// batchModifyLabels is a generic helper for multi-message label modification
// operations (batch archive, batch trash). It extracts message_ids from the
// request, applies the specified label changes, and returns a standard result.
func batchModifyLabels(ctx context.Context, svc GmailService, request mcp.CallToolRequest, addLabels, removeLabels []string, countKey string) (*mcp.CallToolResult, error) {
	messageIDs := extractStringArray(request.GetArguments()["message_ids"])
	if len(messageIDs) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required"), nil
	}

	batchRequest := &gmail.BatchModifyMessagesRequest{
		Ids:            messageIDs,
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	err := svc.BatchModifyMessages(ctx, batchRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"success":     true,
		countKey:      len(messageIDs),
		"message_ids": messageIDs,
	}

	return common.MarshalToolResult(result)
}
