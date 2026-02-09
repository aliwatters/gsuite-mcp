package gmail

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/gmail/v1"
)

// MethodCall records a method invocation for verification in tests.
type MethodCall struct {
	Method string
	Args   []any
}

// MockGmailService is a mock implementation of GmailService for testing.
type MockGmailService struct {
	// Data stores
	Messages map[string]*gmail.Message
	Threads  map[string]*gmail.Thread
	Labels   map[string]*gmail.Label
	Drafts   map[string]*gmail.Draft
	Filters  map[string]*gmail.Filter
	Profile  *gmail.Profile
	Vacation *gmail.VacationSettings

	// Error to return (if set, operations return this error)
	Error error

	// MethodCalls records all method invocations for verification
	MethodCalls []MethodCall

	// Counters for generating IDs
	nextMessageID int
	nextDraftID   int
	nextFilterID  int
	nextLabelID   int
}

// NewMockGmailService creates a new MockGmailService with initialized maps.
func NewMockGmailService() *MockGmailService {
	return &MockGmailService{
		Messages: make(map[string]*gmail.Message),
		Threads:  make(map[string]*gmail.Thread),
		Labels:   make(map[string]*gmail.Label),
		Drafts:   make(map[string]*gmail.Draft),
		Filters:  make(map[string]*gmail.Filter),
		Profile: &gmail.Profile{
			EmailAddress:  common.TestEmail,
			MessagesTotal: 100,
			ThreadsTotal:  50,
			HistoryId:     12345,
		},
		Vacation: &gmail.VacationSettings{
			EnableAutoReply: false,
		},
	}
}

func (m *MockGmailService) recordCall(method string, args ...any) {
	m.MethodCalls = append(m.MethodCalls, MethodCall{Method: method, Args: args})
}

// Reset clears all data and method calls for a fresh test.
func (m *MockGmailService) Reset() {
	m.Messages = make(map[string]*gmail.Message)
	m.Threads = make(map[string]*gmail.Thread)
	m.Labels = make(map[string]*gmail.Label)
	m.Drafts = make(map[string]*gmail.Draft)
	m.Filters = make(map[string]*gmail.Filter)
	m.MethodCalls = nil
	m.Error = nil
}

// === Messages ===

func (m *MockGmailService) ListMessages(ctx context.Context, query string, maxResults int64, pageToken string) (*gmail.ListMessagesResponse, error) {
	m.recordCall("ListMessages", query, maxResults, pageToken)
	if m.Error != nil {
		return nil, m.Error
	}

	// Return all messages (simplified - doesn't filter by query)
	var messages []*gmail.Message
	for _, msg := range m.Messages {
		messages = append(messages, &gmail.Message{
			Id:       msg.Id,
			ThreadId: msg.ThreadId,
		})
		if int64(len(messages)) >= maxResults {
			break
		}
	}

	return &gmail.ListMessagesResponse{
		Messages:           messages,
		ResultSizeEstimate: int64(len(m.Messages)),
	}, nil
}

func (m *MockGmailService) GetMessage(ctx context.Context, messageID, format string) (*gmail.Message, error) {
	m.recordCall("GetMessage", messageID, format)
	if m.Error != nil {
		return nil, m.Error
	}

	msg, ok := m.Messages[messageID]
	if !ok {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}
	return msg, nil
}

func (m *MockGmailService) SendMessage(ctx context.Context, message *gmail.Message) (*gmail.Message, error) {
	m.recordCall("SendMessage", message)
	if m.Error != nil {
		return nil, m.Error
	}

	m.nextMessageID++
	message.Id = fmt.Sprintf("msg-%d", m.nextMessageID)
	if message.ThreadId == "" {
		message.ThreadId = fmt.Sprintf("thread-%d", m.nextMessageID)
	}
	message.LabelIds = []string{"SENT"}
	m.Messages[message.Id] = message
	return message, nil
}

func (m *MockGmailService) ModifyMessage(ctx context.Context, messageID string, req *gmail.ModifyMessageRequest) (*gmail.Message, error) {
	m.recordCall("ModifyMessage", messageID, req)
	if m.Error != nil {
		return nil, m.Error
	}

	msg, ok := m.Messages[messageID]
	if !ok {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}

	// Remove labels
	newLabels := []string{}
	for _, label := range msg.LabelIds {
		keep := true
		for _, remove := range req.RemoveLabelIds {
			if label == remove {
				keep = false
				break
			}
		}
		if keep {
			newLabels = append(newLabels, label)
		}
	}

	// Add labels
	for _, add := range req.AddLabelIds {
		found := false
		for _, existing := range newLabels {
			if existing == add {
				found = true
				break
			}
		}
		if !found {
			newLabels = append(newLabels, add)
		}
	}

	msg.LabelIds = newLabels
	return msg, nil
}

func (m *MockGmailService) TrashMessage(ctx context.Context, messageID string) (*gmail.Message, error) {
	m.recordCall("TrashMessage", messageID)
	if m.Error != nil {
		return nil, m.Error
	}

	msg, ok := m.Messages[messageID]
	if !ok {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}

	// Add TRASH label, remove INBOX
	msg.LabelIds = append(msg.LabelIds, "TRASH")
	newLabels := []string{}
	for _, l := range msg.LabelIds {
		if l != "INBOX" {
			newLabels = append(newLabels, l)
		}
	}
	msg.LabelIds = newLabels
	return msg, nil
}

func (m *MockGmailService) UntrashMessage(ctx context.Context, messageID string) (*gmail.Message, error) {
	m.recordCall("UntrashMessage", messageID)
	if m.Error != nil {
		return nil, m.Error
	}

	msg, ok := m.Messages[messageID]
	if !ok {
		return nil, fmt.Errorf("message not found: %s", messageID)
	}

	// Remove TRASH label and add INBOX back
	newLabels := []string{}
	hasInbox := false
	for _, l := range msg.LabelIds {
		if l != "TRASH" {
			newLabels = append(newLabels, l)
			if l == "INBOX" {
				hasInbox = true
			}
		}
	}
	if !hasInbox {
		newLabels = append(newLabels, "INBOX")
	}
	msg.LabelIds = newLabels
	return msg, nil
}

func (m *MockGmailService) BatchModifyMessages(ctx context.Context, req *gmail.BatchModifyMessagesRequest) error {
	m.recordCall("BatchModifyMessages", req)
	if m.Error != nil {
		return m.Error
	}

	for _, id := range req.Ids {
		if msg, ok := m.Messages[id]; ok {
			// Remove labels
			newLabels := []string{}
			for _, label := range msg.LabelIds {
				keep := true
				for _, remove := range req.RemoveLabelIds {
					if label == remove {
						keep = false
						break
					}
				}
				if keep {
					newLabels = append(newLabels, label)
				}
			}
			// Add labels
			for _, add := range req.AddLabelIds {
				found := false
				for _, existing := range newLabels {
					if existing == add {
						found = true
						break
					}
				}
				if !found {
					newLabels = append(newLabels, add)
				}
			}
			msg.LabelIds = newLabels
		}
	}
	return nil
}

// === Threads ===

func (m *MockGmailService) GetThread(ctx context.Context, threadID, format string) (*gmail.Thread, error) {
	m.recordCall("GetThread", threadID, format)
	if m.Error != nil {
		return nil, m.Error
	}

	thread, ok := m.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}
	return thread, nil
}

func (m *MockGmailService) ModifyThread(ctx context.Context, threadID string, req *gmail.ModifyThreadRequest) (*gmail.Thread, error) {
	m.recordCall("ModifyThread", threadID, req)
	if m.Error != nil {
		return nil, m.Error
	}

	thread, ok := m.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}

	// Modify labels on all messages in thread
	for _, msg := range thread.Messages {
		// Remove labels
		newLabels := []string{}
		for _, label := range msg.LabelIds {
			keep := true
			for _, remove := range req.RemoveLabelIds {
				if label == remove {
					keep = false
					break
				}
			}
			if keep {
				newLabels = append(newLabels, label)
			}
		}
		// Add labels
		for _, add := range req.AddLabelIds {
			found := false
			for _, existing := range newLabels {
				if existing == add {
					found = true
					break
				}
			}
			if !found {
				newLabels = append(newLabels, add)
			}
		}
		msg.LabelIds = newLabels
	}

	return thread, nil
}

func (m *MockGmailService) TrashThread(ctx context.Context, threadID string) (*gmail.Thread, error) {
	m.recordCall("TrashThread", threadID)
	if m.Error != nil {
		return nil, m.Error
	}

	thread, ok := m.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}

	for _, msg := range thread.Messages {
		msg.LabelIds = append(msg.LabelIds, "TRASH")
	}
	return thread, nil
}

func (m *MockGmailService) UntrashThread(ctx context.Context, threadID string) (*gmail.Thread, error) {
	m.recordCall("UntrashThread", threadID)
	if m.Error != nil {
		return nil, m.Error
	}

	thread, ok := m.Threads[threadID]
	if !ok {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}

	for _, msg := range thread.Messages {
		newLabels := []string{}
		for _, l := range msg.LabelIds {
			if l != "TRASH" {
				newLabels = append(newLabels, l)
			}
		}
		msg.LabelIds = newLabels
	}
	return thread, nil
}

// === Labels ===

func (m *MockGmailService) ListLabels(ctx context.Context) (*gmail.ListLabelsResponse, error) {
	m.recordCall("ListLabels")
	if m.Error != nil {
		return nil, m.Error
	}

	var labels []*gmail.Label
	for _, label := range m.Labels {
		labels = append(labels, label)
	}
	return &gmail.ListLabelsResponse{Labels: labels}, nil
}

func (m *MockGmailService) GetLabel(ctx context.Context, labelID string) (*gmail.Label, error) {
	m.recordCall("GetLabel", labelID)
	if m.Error != nil {
		return nil, m.Error
	}

	label, ok := m.Labels[labelID]
	if !ok {
		return nil, fmt.Errorf("label not found: %s", labelID)
	}
	return label, nil
}

func (m *MockGmailService) CreateLabel(ctx context.Context, label *gmail.Label) (*gmail.Label, error) {
	m.recordCall("CreateLabel", label)
	if m.Error != nil {
		return nil, m.Error
	}

	m.nextLabelID++
	label.Id = fmt.Sprintf("Label_%d", m.nextLabelID)
	label.Type = "user"
	m.Labels[label.Id] = label
	return label, nil
}

func (m *MockGmailService) UpdateLabel(ctx context.Context, labelID string, label *gmail.Label) (*gmail.Label, error) {
	m.recordCall("UpdateLabel", labelID, label)
	if m.Error != nil {
		return nil, m.Error
	}

	if _, ok := m.Labels[labelID]; !ok {
		return nil, fmt.Errorf("label not found: %s", labelID)
	}

	label.Id = labelID
	m.Labels[labelID] = label
	return label, nil
}

func (m *MockGmailService) DeleteLabel(ctx context.Context, labelID string) error {
	m.recordCall("DeleteLabel", labelID)
	if m.Error != nil {
		return m.Error
	}

	if _, ok := m.Labels[labelID]; !ok {
		return fmt.Errorf("label not found: %s", labelID)
	}

	delete(m.Labels, labelID)
	return nil
}

// === Drafts ===

func (m *MockGmailService) ListDrafts(ctx context.Context, maxResults int64, pageToken string) (*gmail.ListDraftsResponse, error) {
	m.recordCall("ListDrafts", maxResults, pageToken)
	if m.Error != nil {
		return nil, m.Error
	}

	var drafts []*gmail.Draft
	for _, draft := range m.Drafts {
		drafts = append(drafts, draft)
		if int64(len(drafts)) >= maxResults {
			break
		}
	}
	return &gmail.ListDraftsResponse{Drafts: drafts}, nil
}

func (m *MockGmailService) GetDraft(ctx context.Context, draftID, format string) (*gmail.Draft, error) {
	m.recordCall("GetDraft", draftID, format)
	if m.Error != nil {
		return nil, m.Error
	}

	draft, ok := m.Drafts[draftID]
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", draftID)
	}
	return draft, nil
}

func (m *MockGmailService) CreateDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Draft, error) {
	m.recordCall("CreateDraft", draft)
	if m.Error != nil {
		return nil, m.Error
	}

	m.nextDraftID++
	draft.Id = fmt.Sprintf("draft-%d", m.nextDraftID)
	if draft.Message != nil {
		m.nextMessageID++
		draft.Message.Id = fmt.Sprintf("msg-%d", m.nextMessageID)
		if draft.Message.ThreadId == "" {
			draft.Message.ThreadId = fmt.Sprintf("thread-%d", m.nextMessageID)
		}
	}
	m.Drafts[draft.Id] = draft
	return draft, nil
}

func (m *MockGmailService) UpdateDraft(ctx context.Context, draftID string, draft *gmail.Draft) (*gmail.Draft, error) {
	m.recordCall("UpdateDraft", draftID, draft)
	if m.Error != nil {
		return nil, m.Error
	}

	if _, ok := m.Drafts[draftID]; !ok {
		return nil, fmt.Errorf("draft not found: %s", draftID)
	}

	draft.Id = draftID
	if draft.Message != nil && draft.Message.Id == "" {
		m.nextMessageID++
		draft.Message.Id = fmt.Sprintf("msg-%d", m.nextMessageID)
	}
	m.Drafts[draftID] = draft
	return draft, nil
}

func (m *MockGmailService) DeleteDraft(ctx context.Context, draftID string) error {
	m.recordCall("DeleteDraft", draftID)
	if m.Error != nil {
		return m.Error
	}

	if _, ok := m.Drafts[draftID]; !ok {
		return fmt.Errorf("draft not found: %s", draftID)
	}

	delete(m.Drafts, draftID)
	return nil
}

func (m *MockGmailService) SendDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Message, error) {
	m.recordCall("SendDraft", draft)
	if m.Error != nil {
		return nil, m.Error
	}

	existing, ok := m.Drafts[draft.Id]
	if !ok {
		return nil, fmt.Errorf("draft not found: %s", draft.Id)
	}

	// Convert draft to sent message
	msg := existing.Message
	if msg == nil {
		msg = &gmail.Message{}
	}
	m.nextMessageID++
	msg.Id = fmt.Sprintf("msg-%d", m.nextMessageID)
	msg.LabelIds = []string{"SENT"}

	delete(m.Drafts, draft.Id)
	m.Messages[msg.Id] = msg

	return msg, nil
}

// === Attachments ===

func (m *MockGmailService) GetAttachment(ctx context.Context, messageID, attachmentID string) (*gmail.MessagePartBody, error) {
	m.recordCall("GetAttachment", messageID, attachmentID)
	if m.Error != nil {
		return nil, m.Error
	}

	// Return mock attachment data
	return &gmail.MessagePartBody{
		AttachmentId: attachmentID,
		Size:         1024,
		Data:         "SGVsbG8gV29ybGQh", // base64 encoded "Hello World!"
	}, nil
}

// === Filters ===

func (m *MockGmailService) ListFilters(ctx context.Context) (*gmail.ListFiltersResponse, error) {
	m.recordCall("ListFilters")
	if m.Error != nil {
		return nil, m.Error
	}

	var filters []*gmail.Filter
	for _, filter := range m.Filters {
		filters = append(filters, filter)
	}
	return &gmail.ListFiltersResponse{Filter: filters}, nil
}

func (m *MockGmailService) CreateFilter(ctx context.Context, filter *gmail.Filter) (*gmail.Filter, error) {
	m.recordCall("CreateFilter", filter)
	if m.Error != nil {
		return nil, m.Error
	}

	m.nextFilterID++
	filter.Id = fmt.Sprintf("filter-%d", m.nextFilterID)
	m.Filters[filter.Id] = filter
	return filter, nil
}

func (m *MockGmailService) DeleteFilter(ctx context.Context, filterID string) error {
	m.recordCall("DeleteFilter", filterID)
	if m.Error != nil {
		return m.Error
	}

	if _, ok := m.Filters[filterID]; !ok {
		return fmt.Errorf("filter not found: %s", filterID)
	}

	delete(m.Filters, filterID)
	return nil
}

// === Profile & Settings ===

func (m *MockGmailService) GetProfile(ctx context.Context) (*gmail.Profile, error) {
	m.recordCall("GetProfile")
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Profile, nil
}

func (m *MockGmailService) GetVacationSettings(ctx context.Context) (*gmail.VacationSettings, error) {
	m.recordCall("GetVacationSettings")
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Vacation, nil
}

func (m *MockGmailService) UpdateVacationSettings(ctx context.Context, settings *gmail.VacationSettings) (*gmail.VacationSettings, error) {
	m.recordCall("UpdateVacationSettings", settings)
	if m.Error != nil {
		return nil, m.Error
	}
	m.Vacation = settings
	return settings, nil
}

// === Test Helpers ===

// AddMessage adds a message to the mock store.
func (m *MockGmailService) AddMessage(msg *gmail.Message) {
	m.Messages[msg.Id] = msg
}

// AddThread adds a thread to the mock store.
func (m *MockGmailService) AddThread(thread *gmail.Thread) {
	m.Threads[thread.Id] = thread
}

// AddLabel adds a label to the mock store.
func (m *MockGmailService) AddLabel(label *gmail.Label) {
	m.Labels[label.Id] = label
}

// AddDraft adds a draft to the mock store.
func (m *MockGmailService) AddDraft(draft *gmail.Draft) {
	m.Drafts[draft.Id] = draft
}

// AddFilter adds a filter to the mock store.
func (m *MockGmailService) AddFilter(filter *gmail.Filter) {
	m.Filters[filter.Id] = filter
}

// GetLastCall returns the last method call recorded.
func (m *MockGmailService) GetLastCall() *MethodCall {
	if len(m.MethodCalls) == 0 {
		return nil
	}
	return &m.MethodCalls[len(m.MethodCalls)-1]
}

// WasMethodCalled checks if a method was called.
func (m *MockGmailService) WasMethodCalled(method string) bool {
	for _, call := range m.MethodCalls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// SetError sets the error that will be returned by all subsequent operations.
func (m *MockGmailService) SetError(errMsg string) {
	m.Error = fmt.Errorf("%s", errMsg)
}

// VacationSettings is an alias for Vacation field for convenience in tests.
func (m *MockGmailService) SetVacationSettings(settings *gmail.VacationSettings) {
	m.Vacation = settings
}
