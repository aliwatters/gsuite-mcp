package gmail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/gmail/v1"
)

// TestableGmailSearch performs a Gmail search using the provided service.
func TestableGmailSearch(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	query, errResult := common.RequireStringArg(request.GetArguments(), "query")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	maxResults := common.ParseMaxResults(request.GetArguments(), common.GmailDefaultMaxResults, common.GmailMaxResultsLimit)
	pageToken := common.ParseStringArg(request.GetArguments(), "page_token", "")

	resp, err := svc.ListMessages(ctx, query, maxResults, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	type messageInfo struct {
		ID       string `json:"id"`
		ThreadID string `json:"thread_id"`
	}

	messages := make([]messageInfo, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		messages = append(messages, messageInfo{
			ID:       msg.Id,
			ThreadID: msg.ThreadId,
		})
	}

	result := map[string]any{
		"messages":        messages,
		"result_size":     resp.ResultSizeEstimate,
		"next_page_token": resp.NextPageToken,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetMessage retrieves a single message using the provided service.
func TestableGmailGetMessage(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.GetArguments(), "format", "full")

	msg, err := svc.GetMessage(ctx, messageID, format)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := FormatMessageWithOptions(msg, FormatMessageOptions{BodyFormat: parseBodyFormat(request.GetArguments())})
	return common.MarshalToolResult(result)
}

// TestableGmailGetMessages retrieves multiple messages using the provided service.
func TestableGmailGetMessages(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageIDsRaw, ok := request.GetArguments()["message_ids"].([]any)
	if !ok || len(messageIDsRaw) == 0 {
		return mcp.NewToolResultError("message_ids parameter is required (array of message IDs)"), nil
	}

	if len(messageIDsRaw) > common.GmailMaxBatchMessages {
		return mcp.NewToolResultError(fmt.Sprintf("maximum %d messages per request", common.GmailMaxBatchMessages)), nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	format := common.ParseStringArg(request.GetArguments(), "format", "full")

	messages := make([]map[string]any, 0, len(messageIDsRaw))
	var errors []string

	opts := FormatMessageOptions{BodyFormat: parseBodyFormat(request.GetArguments())}
	for _, idRaw := range messageIDsRaw {
		messageID, ok := idRaw.(string)
		if !ok {
			continue
		}

		msg, err := svc.GetMessage(ctx, messageID, format)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", messageID, err))
			continue
		}

		messages = append(messages, FormatMessageWithOptions(msg, opts))
	}

	result := map[string]any{
		"messages": messages,
		"count":    len(messages),
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSend sends a new email using the provided service.
func TestableGmailSend(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	if _, errResult := common.RequireStringArg(request.GetArguments(), "to"); errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	message, err := buildMessageFromArgsWithAttachments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sent, err := svc.SendMessage(ctx, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":        sent.Id,
		"thread_id": sent.ThreadId,
		"labels":    sent.LabelIds,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailReply replies to an email using the provided service.
func TestableGmailReply(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID := common.ParseStringArg(request.GetArguments(), "message_id", "")
	if messageID == "" {
		return mcp.NewToolResultError("message_id parameter is required (the message you're replying to)"), nil
	}

	body, bodyErr := common.RequireStringArg(request.GetArguments(), "body")
	if bodyErr != nil {
		return bodyErr, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	// Get the original message to extract headers
	origMsg, err := svc.GetMessage(ctx, messageID, "metadata")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get original message: %v", err)), nil
	}

	var origFrom, origTo, origCc, origSubject, origMessageID, origReferences string
	if origMsg.Payload != nil {
		for _, h := range origMsg.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "from":
				origFrom = h.Value
			case "to":
				origTo = h.Value
			case "cc":
				origCc = h.Value
			case "subject":
				origSubject = h.Value
			case "message-id":
				origMessageID = h.Value
			case "references":
				origReferences = h.Value
			}
		}
	}

	replyTo := origFrom
	if to := common.ParseStringArg(request.GetArguments(), "to", ""); to != "" {
		replyTo = to
	}

	replyAll := common.ParseBoolArg(request.GetArguments(), "reply_all", false)

	var cc string
	if replyAll {
		var ccList []string
		if origTo != "" {
			ccList = append(ccList, origTo)
		}
		if origCc != "" {
			ccList = append(ccList, origCc)
		}
		cc = strings.Join(ccList, ", ")
	}
	if ccOverride := common.ParseStringArg(request.GetArguments(), "cc", ""); ccOverride != "" {
		cc = ccOverride
	}

	subject := origSubject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	references := origMessageID
	if origReferences != "" {
		references = origReferences + " " + origMessageID
	}

	emailMsg := EmailMessage{
		To:      replyTo,
		Cc:      cc,
		Subject: subject,
		Body:    body,
		ExtraHeaders: map[string]string{
			"References":  references,
			"In-Reply-To": origMessageID,
		},
	}

	attachments, err := loadEmailAttachments(request.GetArguments()["attachments"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	raw, err := buildEmailMessageWithAttachments(emailMsg, attachments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	message := &gmail.Message{
		Raw:      raw,
		ThreadId: origMsg.ThreadId,
	}

	sent, err := svc.SendMessage(ctx, message)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":        sent.Id,
		"thread_id": sent.ThreadId,
		"labels":    sent.LabelIds,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailSpam marks a message as spam.
func TestableGmailSpam(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"SPAM"}, []string{"INBOX"})
}

// TestableGmailNotSpam removes a message from spam and moves to inbox.
func TestableGmailNotSpam(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"INBOX"}, []string{"SPAM"})
}

// TestableGmailTrash moves a message to trash.
func TestableGmailTrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.TrashMessage(ctx, messageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":       msg.Id,
		"labelIds": msg.LabelIds,
		"success":  true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailUntrash removes a message from trash.
func TestableGmailUntrash(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.UntrashMessage(ctx, messageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"id":       msg.Id,
		"labelIds": msg.LabelIds,
		"success":  true,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailGetAttachment downloads an attachment.
func TestableGmailGetAttachment(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	attachmentID, errResult := common.RequireStringArg(request.GetArguments(), "attachment_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	attach, err := svc.GetAttachment(ctx, messageID, attachmentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	result := map[string]any{
		"attachment_id": attachmentID,
		"message_id":    messageID,
		"size":          attach.Size,
		"data":          attach.Data,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailListAttachments lists downloadable attachments on a message.
func TestableGmailListAttachments(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	messageID, errResult := common.RequireStringArg(request.GetArguments(), "message_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.GetMessage(ctx, messageID, "full")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	attachments := ExtractAttachments(msg.Payload)
	if attachments == nil {
		attachments = []map[string]any{}
	}
	result := map[string]any{
		"message_id":   messageID,
		"count":        len(attachments),
		"attachments":  attachments,
		"has_multiple": len(attachments) > 1,
	}

	return common.MarshalToolResult(result)
}

// TestableGmailDownloadAttachment decodes a Gmail attachment and writes it to disk.
func TestableGmailDownloadAttachment(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	messageID, errResult := common.RequireStringArg(args, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	msg, err := svc.GetMessage(ctx, messageID, "full")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	attachments := ExtractAttachments(msg.Payload)
	selected, selectErr := selectAttachment(attachments, common.ParseStringArg(args, "attachment_id", ""), common.ParseStringArg(args, "part_id", ""))
	if selectErr != nil {
		return mcp.NewToolResultError(selectErr.Error()), nil
	}

	attachmentID := attachmentString(selected, "attachment_id")
	attach, err := svc.GetAttachment(ctx, messageID, attachmentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Gmail API error: %v", err)), nil
	}

	data, err := decodeGmailAttachmentData(attach.Data)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	filename := sanitizeAttachmentFilename(attachmentString(selected, "filename"), "attachment-"+attachmentID)
	outputPath, err := resolveAttachmentOutputPath(args, filename)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := writeAttachmentFile(outputPath, data, common.ParseBoolArg(args, "overwrite", false)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sum := sha256.Sum256(data)
	result := map[string]any{
		"message_id":     messageID,
		"attachment_id":  attachmentID,
		"filename":       filename,
		"mime_type":      attachmentString(selected, "mime_type"),
		"path":           outputPath,
		"size":           attach.Size,
		"bytes_written":  len(data),
		"sha256":         fmt.Sprintf("%x", sum),
		"overwrote_file": common.ParseBoolArg(args, "overwrite", false),
	}
	if partID := attachmentString(selected, "part_id"); partID != "" {
		result["part_id"] = partID
	}

	return common.MarshalToolResult(result)
}

func selectAttachment(attachments []map[string]any, attachmentID string, partID string) (map[string]any, error) {
	if len(attachments) == 0 {
		return nil, fmt.Errorf("message has no downloadable attachments")
	}

	for _, attachment := range attachments {
		if attachmentID != "" && attachmentString(attachment, "attachment_id") == attachmentID {
			if partID != "" && attachmentString(attachment, "part_id") != partID {
				return nil, fmt.Errorf("attachment_id %q is not on part_id %q", attachmentID, partID)
			}
			return attachment, nil
		}
		if attachmentID == "" && partID != "" && attachmentString(attachment, "part_id") == partID {
			return attachment, nil
		}
	}

	if attachmentID != "" {
		return nil, fmt.Errorf("attachment_id %q was not found on message", attachmentID)
	}
	if partID != "" {
		return nil, fmt.Errorf("part_id %q was not found on message", partID)
	}
	if len(attachments) > 1 {
		return nil, fmt.Errorf("message has %d attachments; provide attachment_id or part_id from gmail_list_attachments", len(attachments))
	}

	return attachments[0], nil
}

func attachmentString(attachment map[string]any, key string) string {
	value, _ := attachment[key].(string)
	return value
}

func decodeGmailAttachmentData(data string) ([]byte, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return []byte{}, nil
	}

	encodings := []*base64.Encoding{
		base64.RawURLEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.StdEncoding,
	}
	var lastErr error
	for _, encoding := range encodings {
		decoded, err := encoding.DecodeString(data)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("failed to decode attachment data: %w", lastErr)
}

func resolveAttachmentOutputPath(args map[string]any, filename string) (string, error) {
	outputPath := common.ParseStringArg(args, "output_path", "")
	outputDir := common.ParseStringArg(args, "output_dir", "")
	if outputPath != "" && outputDir != "" {
		return "", fmt.Errorf("use output_path or output_dir, not both")
	}

	var path string
	switch {
	case outputPath != "":
		path = outputPath
	case outputDir != "":
		path = filepath.Join(outputDir, filename)
	default:
		path = filepath.Join(os.TempDir(), "gsuite-mcp-attachments", filename)
	}

	expanded, err := expandUserPath(path)
	if err != nil {
		return "", err
	}
	if strings.ContainsRune(expanded, 0) {
		return "", fmt.Errorf("output path contains an invalid NUL byte")
	}
	if expanded == "" {
		return "", fmt.Errorf("output path is required")
	}
	if !filepath.IsAbs(expanded) {
		expanded, err = filepath.Abs(expanded)
		if err != nil {
			return "", fmt.Errorf("resolve output path: %w", err)
		}
	}
	return filepath.Clean(expanded), nil
}

func expandUserPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func sanitizeAttachmentFilename(filename string, fallback string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		filename = fallback
	}
	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = filepath.Base(filename)
	filename = strings.Trim(filename, ". ")

	var b strings.Builder
	for _, r := range filename {
		switch r {
		case 0, '/', '\\':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}

	cleaned := strings.TrimSpace(b.String())
	if cleaned == "" || cleaned == "." {
		return fallback
	}
	return cleaned
}

func writeAttachmentFile(path string, data []byte, overwrite bool) error {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return fmt.Errorf("output path %q is a directory", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check output path %q: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(path, flags, 0o600)
	if os.IsExist(err) {
		return fmt.Errorf("output file %q already exists; pass overwrite=true to replace it", path)
	}
	if err != nil {
		return fmt.Errorf("open output file %q: %w", path, err)
	}

	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write output file %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close output file %q: %w", path, err)
	}
	return nil
}

// TestableGmailArchive archives a message.
func TestableGmailArchive(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"INBOX"})
}

// TestableGmailMarkRead marks a message as read.
func TestableGmailMarkRead(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"UNREAD"})
}

// TestableGmailMarkUnread marks a message as unread.
func TestableGmailMarkUnread(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"UNREAD"}, nil)
}

// TestableGmailStar stars a message.
func TestableGmailStar(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, []string{"STARRED"}, nil)
}

// TestableGmailUnstar unstars a message.
func TestableGmailUnstar(ctx context.Context, request mcp.CallToolRequest, deps *GmailHandlerDeps) (*mcp.CallToolResult, error) {
	svc, errResult, ok := ResolveGmailServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}
	return modifyMessageLabels(ctx, svc, request, nil, []string{"STARRED"})
}
