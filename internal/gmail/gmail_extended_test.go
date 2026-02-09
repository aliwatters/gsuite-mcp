package gmail

import (
	"context"
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
