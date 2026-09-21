package gmail

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/api/gmail/v1"
)

// === Attachment tests ===

func TestGmailGetAttachment_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add a message with an attachment
	msg := &gmail.Message{
		Id:       "msg123",
		ThreadId: "thread123",
		LabelIds: []string{"INBOX"},
		Payload: &gmail.MessagePart{
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: encodeBase64("Hello")},
				},
				{
					Filename: "document.pdf",
					MimeType: "application/pdf",
					Body: &gmail.MessagePartBody{
						AttachmentId: "att123",
						Size:         1024,
					},
				},
			},
		},
	}
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id":    "msg123",
		"attachment_id": "att123",
	})

	result, err := TestableGmailGetAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify GetAttachment was called
	if !fixtures.MockService.WasMethodCalled("GetAttachment") {
		t.Error("expected GetAttachment to be called")
	}
}

func TestGmailGetAttachment_MissingMessageID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"attachment_id": "att123",
	})

	result, err := TestableGmailGetAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing message_id")
	}
}

func TestGmailGetAttachment_MissingAttachmentID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailGetAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing attachment_id")
	}
}

func TestGmailListAttachments_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithAttachment("msg-att-1"))

	request := makeRequest(map[string]any{
		"message_id": "msg-att-1",
	})

	result, err := TestableGmailListAttachments(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["count"]; got != float64(1) {
		t.Fatalf("count: got %v, want 1", got)
	}
	attachments, ok := response["attachments"].([]any)
	if !ok {
		t.Fatalf("expected attachments array, got %T", response["attachments"])
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	attachment := attachments[0].(map[string]any)
	if got := attachment["attachment_id"]; got != "ATT-PDF-1" {
		t.Errorf("attachment_id: got %v, want ATT-PDF-1", got)
	}
	if got := attachment["filename"]; got != "spec.pdf" {
		t.Errorf("filename: got %v, want spec.pdf", got)
	}
}

func TestGmailDownloadAttachment_SingleAttachmentToOutputDir(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithAttachment("msg-att-1"))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id": "msg-att-1",
		"output_dir": outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	outputPath, ok := response["path"].(string)
	if !ok || outputPath == "" {
		t.Fatalf("expected output path, got %v", response["path"])
	}
	if outputPath != filepath.Join(outputDir, "spec.pdf") {
		t.Fatalf("path: got %q, want output dir/spec.pdf", outputPath)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != "Hello World!" {
		t.Fatalf("downloaded data: got %q, want Hello World!", data)
	}
	if got := response["bytes_written"]; got != float64(len("Hello World!")) {
		t.Errorf("bytes_written: got %v, want %d", got, len("Hello World!"))
	}
}

func TestGmailDownloadAttachment_SanitizesFilename(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	msg := newTestMessageWithAttachment("msg-att-unsafe")
	msg.Payload.Parts[1].Filename = "../../secret.txt"
	fixtures.MockService.AddMessage(msg)
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id": "msg-att-unsafe",
		"output_dir": outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	outputPath := response["path"].(string)
	if outputPath != filepath.Join(outputDir, "secret.txt") {
		t.Fatalf("path escaped output dir or kept unsafe name: %q", outputPath)
	}
}

func TestGmailDownloadAttachment_RefusesOverwriteByDefault(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithAttachment("msg-att-1"))
	outputPath := filepath.Join(t.TempDir(), "existing.pdf")
	if err := os.WriteFile(outputPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing fixture: %v", err)
	}

	request := makeRequest(map[string]any{
		"message_id":  "msg-att-1",
		"output_path": outputPath,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected overwrite protection error")
	}
}

func TestGmailDownloadAttachment_RequiresSelectionForMultipleAttachments(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(&gmail.Message{
		Id:       "msg-multi",
		ThreadId: "thread1",
		Payload: &gmail.MessagePart{
			MimeType: "multipart/mixed",
			Parts: []*gmail.MessagePart{
				{
					PartId:   "1",
					MimeType: "application/pdf",
					Filename: "first.pdf",
					Body:     &gmail.MessagePartBody{AttachmentId: "A1", Size: 100},
				},
				{
					PartId:   "2",
					MimeType: "image/jpeg",
					Filename: "second.jpg",
					Body:     &gmail.MessagePartBody{AttachmentId: "A2", Size: 200},
				},
			},
		},
	})

	request := makeRequest(map[string]any{
		"message_id": "msg-multi",
		"output_dir": t.TempDir(),
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected selection error")
	}
}

// === Filter tests ===

func TestGmailListFilters_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	// Add test filters
	fixtures.MockService.AddFilter(&gmail.Filter{
		Id: "filter1",
		Criteria: &gmail.FilterCriteria{
			From: "newsletters@example.com",
		},
		Action: &gmail.FilterAction{
			AddLabelIds: []string{"Label_1"},
		},
	})
	fixtures.MockService.AddFilter(&gmail.Filter{
		Id: "filter2",
		Criteria: &gmail.FilterCriteria{
			Subject: "urgent",
		},
		Action: &gmail.FilterAction{
			AddLabelIds: []string{"IMPORTANT"},
		},
	})

	request := makeRequest(map[string]any{})

	result, err := TestableGmailListFilters(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ListFilters was called
	if !fixtures.MockService.WasMethodCalled("ListFilters") {
		t.Error("expected ListFilters to be called")
	}

	// Verify response
	response := extractResponse(t, result)
	filters, ok := response["filters"].([]any)
	if !ok {
		t.Fatal("expected filters array in response")
	}
	if len(filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(filters))
	}
}

func TestGmailCreateFilter_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"from":          "spam@example.com",
		"add_label_ids": []any{"TRASH"},
	})

	result, err := TestableGmailCreateFilter(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify CreateFilter was called
	if !fixtures.MockService.WasMethodCalled("CreateFilter") {
		t.Error("expected CreateFilter to be called")
	}
}

func TestGmailCreateFilter_WithAllCriteria(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"from":           "sender@example.com",
		"to":             "me@example.com",
		"subject":        "important",
		"query":          "has:attachment",
		"has_attachment": true,
		"add_label_ids":  []any{"Label_1"},
	})

	result, err := TestableGmailCreateFilter(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestGmailDeleteFilter_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddFilter(&gmail.Filter{Id: "filter123"})

	request := makeRequest(map[string]any{
		"filter_id": "filter123",
	})

	result, err := TestableGmailDeleteFilter(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify DeleteFilter was called
	if !fixtures.MockService.WasMethodCalled("DeleteFilter") {
		t.Error("expected DeleteFilter to be called")
	}
}

func TestGmailDeleteFilter_MissingID(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{})

	result, err := TestableGmailDeleteFilter(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing filter_id")
	}
}

// === Label CRUD tests ===

func TestGmailCreateLabel_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"name": "My New Label",
	})

	result, err := TestableGmailCreateLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify CreateLabel was called
	if !fixtures.MockService.WasMethodCalled("CreateLabel") {
		t.Error("expected CreateLabel to be called")
	}
}

func TestGmailCreateLabel_WithVisibility(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"name":                    "Hidden Label",
		"label_list_visibility":   "labelHide",
		"message_list_visibility": "hide",
	})

	result, err := TestableGmailCreateLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

func TestGmailCreateLabel_MissingName(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{})

	result, err := TestableGmailCreateLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing name")
	}
}

func TestGmailDeleteLabel_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddLabel(newTestLabel("Label_1", "Work", "user", 10, 2))

	request := makeRequest(map[string]any{
		"label_id": "Label_1",
	})

	result, err := TestableGmailDeleteLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify DeleteLabel was called
	if !fixtures.MockService.WasMethodCalled("DeleteLabel") {
		t.Error("expected DeleteLabel to be called")
	}
}

func TestGmailUpdateLabel_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddLabel(newTestLabel("Label_1", "Old Name", "user", 10, 2))

	request := makeRequest(map[string]any{
		"label_id": "Label_1",
		"name":     "New Name",
	})

	result, err := TestableGmailUpdateLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UpdateLabel was called
	if !fixtures.MockService.WasMethodCalled("UpdateLabel") {
		t.Error("expected UpdateLabel to be called")
	}
}

// === Draft CRUD tests ===

func TestGmailListDrafts_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddDraft(&gmail.Draft{
		Id: "draft1",
		Message: &gmail.Message{
			Id: "msg1",
			Payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "Subject", Value: "Draft 1"},
				},
			},
		},
	})
	fixtures.MockService.AddDraft(&gmail.Draft{
		Id: "draft2",
		Message: &gmail.Message{
			Id: "msg2",
			Payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "Subject", Value: "Draft 2"},
				},
			},
		},
	})

	request := makeRequest(map[string]any{})

	result, err := TestableGmailListDrafts(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ListDrafts was called
	if !fixtures.MockService.WasMethodCalled("ListDrafts") {
		t.Error("expected ListDrafts to be called")
	}
}

func TestGmailGetDraft_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddDraft(&gmail.Draft{
		Id: "draft123",
		Message: &gmail.Message{
			Id: "msg123",
			Payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "Subject", Value: "My Draft"},
					{Name: "To", Value: "recipient@example.com"},
				},
				Body: &gmail.MessagePartBody{
					Data: encodeBase64("Draft content"),
				},
			},
		},
	})

	request := makeRequest(map[string]any{
		"draft_id": "draft123",
	})

	result, err := TestableGmailGetDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify GetDraft was called
	if !fixtures.MockService.WasMethodCalled("GetDraft") {
		t.Error("expected GetDraft to be called")
	}
}

func TestGmailUpdateDraft_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddDraft(&gmail.Draft{
		Id: "draft123",
		Message: &gmail.Message{
			Id: "msg123",
		},
	})

	request := makeRequest(map[string]any{
		"draft_id": "draft123",
		"to":       "newrecipient@example.com",
		"subject":  "Updated Subject",
		"body":     "Updated body content",
	})

	result, err := TestableGmailUpdateDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UpdateDraft was called
	if !fixtures.MockService.WasMethodCalled("UpdateDraft") {
		t.Error("expected UpdateDraft to be called")
	}
}

func TestGmailDeleteDraft_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddDraft(&gmail.Draft{Id: "draft123"})

	request := makeRequest(map[string]any{
		"draft_id": "draft123",
	})

	result, err := TestableGmailDeleteDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify DeleteDraft was called
	if !fixtures.MockService.WasMethodCalled("DeleteDraft") {
		t.Error("expected DeleteDraft to be called")
	}
}

func TestGmailSendDraft_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	fixtures.MockService.AddDraft(&gmail.Draft{
		Id: "draft123",
		Message: &gmail.Message{
			Id: "msg123",
		},
	})

	request := makeRequest(map[string]any{
		"draft_id": "draft123",
	})

	result, err := TestableGmailSendDraft(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify SendDraft was called
	if !fixtures.MockService.WasMethodCalled("SendDraft") {
		t.Error("expected SendDraft to be called")
	}
}

// === Profile tests ===

func TestGmailGetProfile_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.Profile = &gmail.Profile{
		EmailAddress:  "me@example.com",
		MessagesTotal: 1000,
		ThreadsTotal:  500,
	}

	request := makeRequest(map[string]any{})

	result, err := TestableGmailGetProfile(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify GetProfile was called
	if !fixtures.MockService.WasMethodCalled("GetProfile") {
		t.Error("expected GetProfile to be called")
	}

	// Verify response
	response := extractResponse(t, result)
	if response["email_address"] != "me@example.com" {
		t.Errorf("expected email_address=me@example.com, got %v", response["email_address"])
	}
}

// === Vacation settings tests ===

func TestGmailGetVacation_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.Vacation = &gmail.VacationSettings{
		EnableAutoReply:    true,
		ResponseSubject:    "Out of Office",
		ResponseBodyHtml:   "<p>I am on vacation.</p>",
		RestrictToContacts: true,
	}

	request := makeRequest(map[string]any{})

	result, err := TestableGmailGetVacation(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify GetVacationSettings was called
	if !fixtures.MockService.WasMethodCalled("GetVacationSettings") {
		t.Error("expected GetVacationSettings to be called")
	}
}

func TestGmailSetVacation_Enable(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"enabled":    true,
		"subject":    "Out of Office",
		"body":       "I am currently on vacation.",
		"start_time": float64(1704067200000),
		"end_time":   float64(1704672000000),
	})

	result, err := TestableGmailSetVacation(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify UpdateVacationSettings was called
	if !fixtures.MockService.WasMethodCalled("UpdateVacationSettings") {
		t.Error("expected UpdateVacationSettings to be called")
	}
}

func TestGmailSetVacation_Disable(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	request := makeRequest(map[string]any{
		"enabled": false,
	})

	result, err := TestableGmailSetVacation(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}
}

// === Spam tests ===

func TestGmailSpam_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "spammer@example.com", "me@example.com", "Spam content", []string{"INBOX"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailSpam(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyMessage was called to add SPAM label
	if !fixtures.MockService.WasMethodCalled("ModifyMessage") {
		t.Error("expected ModifyMessage to be called")
	}
}

func TestGmailNotSpam_Success(t *testing.T) {
	fixtures := NewGmailTestFixtures()

	msg := newTestMessage("msg123", "thread123", "Subject", "sender@example.com", "me@example.com", "Not spam", []string{"SPAM"})
	fixtures.MockService.AddMessage(msg)

	request := makeRequest(map[string]any{
		"message_id": "msg123",
	})

	result, err := TestableGmailNotSpam(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	// Verify ModifyMessage was called to remove SPAM and add INBOX
	if !fixtures.MockService.WasMethodCalled("ModifyMessage") {
		t.Error("expected ModifyMessage to be called")
	}
}

// === Error handling tests ===

func TestGmailGetAttachment_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"message_id":    "msg123",
		"attachment_id": "att123",
	})

	result, err := TestableGmailGetAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

func TestGmailCreateLabel_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{
		"name": "New Label",
	})

	result, err := TestableGmailCreateLabel(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

func TestGmailGetProfile_ServiceError(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.SetError("simulated API error")

	request := makeRequest(map[string]any{})

	result, err := TestableGmailGetProfile(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error when service fails")
	}
}

// === Attachment selection under Gmail's regenerated attachment_id (#204) ===

// newTestMessageWithTwoAttachments builds a two-attachment message whose
// attachment ids are the ones a *current* listing would report. Tests pass a
// different id to stand in for one minted by an earlier read.
func newTestMessageWithTwoAttachments(id string, firstSize, secondSize int64) *gmail.Message {
	return &gmail.Message{
		Id:       id,
		ThreadId: "thread1",
		Payload: &gmail.MessagePart{
			MimeType: "multipart/mixed",
			Parts: []*gmail.MessagePart{
				{
					PartId:   "1",
					MimeType: "image/jpeg",
					Filename: "first.jpg",
					Body:     &gmail.MessagePartBody{AttachmentId: "FRESH-1", Size: firstSize},
				},
				{
					PartId:   "2",
					MimeType: "image/png",
					Filename: "second.png",
					Body:     &gmail.MessagePartBody{AttachmentId: "FRESH-2", Size: secondSize},
				},
			},
		},
	}
}

// helloWorldBody is what the mock serves by default: "Hello World!" base64'd.
func helloWorldBody(size int64) *gmail.MessagePartBody {
	return &gmail.MessagePartBody{Size: size, Data: "SGVsbG8gV29ybGQh"}
}

// TestGmailDownloadAttachment_StaleAttachmentIDIsPassedThrough is the #204
// regression. Gmail re-mints body.attachmentId on every messages.get, so an id
// copied from an earlier gmail_list_attachments call matches nothing in the
// current listing — yet it still resolves at the API. The download must
// succeed rather than reject the token locally.
func TestGmailDownloadAttachment_StaleAttachmentIDIsPassedThrough(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-stale", 100, 200))
	fixtures.MockService.AddAttachmentBody("STALE-FROM-EARLIER-READ", helloWorldBody(100))

	request := makeRequest(map[string]any{
		"message_id":    "msg-stale",
		"attachment_id": "STALE-FROM-EARLIER-READ",
		"output_dir":    t.TempDir(),
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("stale attachment_id must be passed through, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["attachment_id"]; got != "STALE-FROM-EARLIER-READ" {
		t.Errorf("attachment_id: got %v, want the caller's token verbatim", got)
	}
	if got := response["attachment_id_rematched"]; got != true {
		t.Errorf("attachment_id_rematched: got %v, want true", got)
	}
}

// TestGmailDownloadAttachment_StaleIDWithPartIDUsesPartMetadata checks that an
// explicit part_id supplies the filename when the token itself cannot.
func TestGmailDownloadAttachment_StaleIDWithPartIDUsesPartMetadata(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-stale-part", 100, 200))
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(200))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-stale-part",
		"attachment_id": "STALE",
		"part_id":       "2",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["metadata_source"]; got != "part_id" {
		t.Errorf("metadata_source: got %v, want part_id", got)
	}
	if got := response["path"]; got != filepath.Join(outputDir, "second.png") {
		t.Errorf("path: got %v, want second.png from part 2", got)
	}
	// The caller's token must reach the API, not part 2's freshly listed one.
	if got := response["attachment_id"]; got != "STALE" {
		t.Errorf("attachment_id: got %v, want STALE", got)
	}
}

// TestGmailDownloadAttachment_StaleIDRecoversMetadataBySize covers the case
// with no part_id: a unique size match identifies which part was fetched.
func TestGmailDownloadAttachment_StaleIDRecoversMetadataBySize(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-size", 100, 200))
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(200))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-size",
		"attachment_id": "STALE",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["metadata_source"]; got != "size_match" {
		t.Errorf("metadata_source: got %v, want size_match", got)
	}
	if got := response["path"]; got != filepath.Join(outputDir, "second.png") {
		t.Errorf("path: got %v, want second.png (the 200-byte part)", got)
	}
	if got := response["part_id"]; got != "2" {
		t.Errorf("part_id: got %v, want 2", got)
	}
}

// TestGmailDownloadAttachment_StaleIDAmbiguousSizeFallsBack checks that two
// equally sized parts are treated as unidentifiable rather than guessed at:
// the bytes are still written, under a generated name, with no part_id claimed.
func TestGmailDownloadAttachment_StaleIDAmbiguousSizeFallsBack(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-ambiguous", 150, 150))
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(150))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-ambiguous",
		"attachment_id": "STALE",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["metadata_source"]; got != "none" {
		t.Errorf("metadata_source: got %v, want none", got)
	}
	if _, claimed := response["part_id"]; claimed {
		t.Errorf("part_id must be omitted when the part is unidentifiable, got %v", response["part_id"])
	}
	if got := response["path"]; got != filepath.Join(outputDir, "attachment-STALE") {
		t.Errorf("path: got %v, want the generated attachment-STALE name", got)
	}
}

// TestGmailDownloadAttachment_PartIDUsesFreshlyListedToken confirms part_id
// selection still resolves through the current listing's token.
func TestGmailDownloadAttachment_PartIDUsesFreshlyListedToken(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-part", 100, 200))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id": "msg-part",
		"part_id":    "1",
		"output_dir": outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["attachment_id"]; got != "FRESH-1" {
		t.Errorf("attachment_id: got %v, want FRESH-1", got)
	}
	if got := response["metadata_source"]; got != "part_id" {
		t.Errorf("metadata_source: got %v, want part_id", got)
	}
	if got := response["attachment_id_rematched"]; got != false {
		t.Errorf("attachment_id_rematched: got %v, want false", got)
	}
	if got := response["path"]; got != filepath.Join(outputDir, "first.jpg") {
		t.Errorf("path: got %v, want first.jpg", got)
	}
}

// TestGmailDownloadAttachment_UnknownPartIDStillErrors keeps part_id strict:
// it is the durable handle, so a miss is a real caller error, not a stale token.
func TestGmailDownloadAttachment_UnknownPartIDStillErrors(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-badpart", 100, 200))

	request := makeRequest(map[string]any{
		"message_id": "msg-badpart",
		"part_id":    "99",
		"output_dir": t.TempDir(),
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected an error for a part_id that is not on the message")
	}
}

// TestGmailDownloadAttachment_ExactMatchOnConflictingPartIDErrors preserves the
// pre-existing contradiction check.
func TestGmailDownloadAttachment_ExactMatchOnConflictingPartIDErrors(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-conflict", 100, 200))

	request := makeRequest(map[string]any{
		"message_id":    "msg-conflict",
		"attachment_id": "FRESH-1",
		"part_id":       "2",
		"output_dir":    t.TempDir(),
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected an error when attachment_id and part_id disagree")
	}
}

// TestGmailDownloadAttachment_SingleAttachmentAcceptsAnyID: with one
// attachment there is nothing to disambiguate, so an unrecognised token still
// downloads and reports the sole attachment's metadata.
func TestGmailDownloadAttachment_SingleAttachmentAcceptsAnyID(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithAttachment("msg-single-stale"))
	// Same byte count as the message's only attachment, so the token is
	// consistent with it.
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(123456))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-single-stale",
		"attachment_id": "STALE",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["metadata_source"]; got != "sole_attachment" {
		t.Errorf("metadata_source: got %v, want sole_attachment", got)
	}
	if got := response["path"]; got != filepath.Join(outputDir, "spec.pdf") {
		t.Errorf("path: got %v, want spec.pdf", got)
	}
}

// TestGmailDownloadAttachment_StaleIDContradictingPartIDErrors: a token the
// listing does not know, paired with a part_id whose size it does not match,
// means the caller named two different attachments. Writing the fetched bytes
// under the named part's filename would put real data behind a wrong name, so
// this must fail rather than mislabel. The exact-match branch already rejects
// the same contradiction; this keeps the pass-through branch consistent.
func TestGmailDownloadAttachment_StaleIDContradictingPartIDErrors(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-mislabel", 100, 200))
	// The stale token serves part 1's 100 bytes while the caller names part 2.
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(100))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-mislabel",
		"attachment_id": "STALE",
		"part_id":       "2",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected an error when the token's bytes do not match the named part")
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("nothing may be written on a mislabel refusal, found %d file(s)", len(entries))
	}
}

// TestGmailDownloadAttachment_StaleIDContradictingSoleAttachmentErrors covers
// the same guard on the single-attachment path, where the token may well have
// come from an entirely different message.
func TestGmailDownloadAttachment_StaleIDContradictingSoleAttachmentErrors(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithAttachment("msg-single-mismatch"))
	// The message's only attachment is 123456 bytes; the token serves 100.
	fixtures.MockService.AddAttachmentBody("STALE", helloWorldBody(100))

	request := makeRequest(map[string]any{
		"message_id":    "msg-single-mismatch",
		"attachment_id": "STALE",
		"output_dir":    t.TempDir(),
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected an error when the token's bytes do not match the sole attachment")
	}
}

// TestGmailDownloadAttachment_ExactMatchOnMultiAttachmentMessage pins the most
// common path: an attachment_id still present in the current listing, on a
// message where the choice actually matters. The other exact-match test only
// reaches this branch's error return, which left the success return free to
// regress silently.
func TestGmailDownloadAttachment_ExactMatchOnMultiAttachmentMessage(t *testing.T) {
	fixtures := NewGmailTestFixtures()
	fixtures.MockService.AddMessage(newTestMessageWithTwoAttachments("msg-exact", 100, 200))
	fixtures.MockService.AddAttachmentBody("FRESH-2", helloWorldBody(200))
	outputDir := t.TempDir()

	request := makeRequest(map[string]any{
		"message_id":    "msg-exact",
		"attachment_id": "FRESH-2",
		"output_dir":    outputDir,
	})

	result, err := TestableGmailDownloadAttachment(context.Background(), request, fixtures.Deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	response := extractResponse(t, result)
	if got := response["metadata_source"]; got != "exact" {
		t.Errorf("metadata_source: got %v, want exact", got)
	}
	if got := response["attachment_id_rematched"]; got != false {
		t.Errorf("attachment_id_rematched: got %v, want false", got)
	}
	if got := response["part_id"]; got != "2" {
		t.Errorf("part_id: got %v, want 2", got)
	}
	if got := response["path"]; got != filepath.Join(outputDir, "second.png") {
		t.Errorf("path: got %v, want second.png", got)
	}
	if got := response["mime_type"]; got != "image/png" {
		t.Errorf("mime_type: got %v, want image/png", got)
	}
}
