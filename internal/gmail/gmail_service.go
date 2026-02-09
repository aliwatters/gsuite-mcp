package gmail

import (
	"context"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/gmail/v1"
)

// GmailService defines the interface for Gmail operations.
// This abstraction enables dependency injection and testing.
type GmailService interface {
	// Messages
	ListMessages(ctx context.Context, query string, maxResults int64, pageToken string) (*gmail.ListMessagesResponse, error)
	GetMessage(ctx context.Context, messageID, format string) (*gmail.Message, error)
	SendMessage(ctx context.Context, message *gmail.Message) (*gmail.Message, error)
	ModifyMessage(ctx context.Context, messageID string, req *gmail.ModifyMessageRequest) (*gmail.Message, error)
	TrashMessage(ctx context.Context, messageID string) (*gmail.Message, error)
	UntrashMessage(ctx context.Context, messageID string) (*gmail.Message, error)
	BatchModifyMessages(ctx context.Context, req *gmail.BatchModifyMessagesRequest) error

	// Threads
	GetThread(ctx context.Context, threadID, format string) (*gmail.Thread, error)
	ModifyThread(ctx context.Context, threadID string, req *gmail.ModifyThreadRequest) (*gmail.Thread, error)
	TrashThread(ctx context.Context, threadID string) (*gmail.Thread, error)
	UntrashThread(ctx context.Context, threadID string) (*gmail.Thread, error)

	// Labels
	ListLabels(ctx context.Context) (*gmail.ListLabelsResponse, error)
	GetLabel(ctx context.Context, labelID string) (*gmail.Label, error)
	CreateLabel(ctx context.Context, label *gmail.Label) (*gmail.Label, error)
	UpdateLabel(ctx context.Context, labelID string, label *gmail.Label) (*gmail.Label, error)
	DeleteLabel(ctx context.Context, labelID string) error

	// Drafts
	ListDrafts(ctx context.Context, maxResults int64, pageToken string) (*gmail.ListDraftsResponse, error)
	GetDraft(ctx context.Context, draftID, format string) (*gmail.Draft, error)
	CreateDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Draft, error)
	UpdateDraft(ctx context.Context, draftID string, draft *gmail.Draft) (*gmail.Draft, error)
	DeleteDraft(ctx context.Context, draftID string) error
	SendDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Message, error)

	// Attachments
	GetAttachment(ctx context.Context, messageID, attachmentID string) (*gmail.MessagePartBody, error)

	// Filters
	ListFilters(ctx context.Context) (*gmail.ListFiltersResponse, error)
	CreateFilter(ctx context.Context, filter *gmail.Filter) (*gmail.Filter, error)
	DeleteFilter(ctx context.Context, filterID string) error

	// Profile & Settings
	GetProfile(ctx context.Context) (*gmail.Profile, error)
	GetVacationSettings(ctx context.Context) (*gmail.VacationSettings, error)
	UpdateVacationSettings(ctx context.Context, settings *gmail.VacationSettings) (*gmail.VacationSettings, error)
}

// RealGmailService wraps the actual Gmail API service.
type RealGmailService struct {
	service *gmail.Service
}

// NewRealGmailService creates a new RealGmailService.
func NewRealGmailService(service *gmail.Service) *RealGmailService {
	return &RealGmailService{service: service}
}

// === Messages ===

func (s *RealGmailService) ListMessages(ctx context.Context, query string, maxResults int64, pageToken string) (*gmail.ListMessagesResponse, error) {
	call := s.service.Users.Messages.List(common.GmailUserMe).Q(query).MaxResults(maxResults)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Context(ctx).Do()
}

func (s *RealGmailService) GetMessage(ctx context.Context, messageID, format string) (*gmail.Message, error) {
	return s.service.Users.Messages.Get(common.GmailUserMe, messageID).Format(format).Context(ctx).Do()
}

func (s *RealGmailService) SendMessage(ctx context.Context, message *gmail.Message) (*gmail.Message, error) {
	return s.service.Users.Messages.Send(common.GmailUserMe, message).Context(ctx).Do()
}

func (s *RealGmailService) ModifyMessage(ctx context.Context, messageID string, req *gmail.ModifyMessageRequest) (*gmail.Message, error) {
	return s.service.Users.Messages.Modify(common.GmailUserMe, messageID, req).Context(ctx).Do()
}

func (s *RealGmailService) TrashMessage(ctx context.Context, messageID string) (*gmail.Message, error) {
	return s.service.Users.Messages.Trash(common.GmailUserMe, messageID).Context(ctx).Do()
}

func (s *RealGmailService) UntrashMessage(ctx context.Context, messageID string) (*gmail.Message, error) {
	return s.service.Users.Messages.Untrash(common.GmailUserMe, messageID).Context(ctx).Do()
}

func (s *RealGmailService) BatchModifyMessages(ctx context.Context, req *gmail.BatchModifyMessagesRequest) error {
	return s.service.Users.Messages.BatchModify(common.GmailUserMe, req).Context(ctx).Do()
}

// === Threads ===

func (s *RealGmailService) GetThread(ctx context.Context, threadID, format string) (*gmail.Thread, error) {
	return s.service.Users.Threads.Get(common.GmailUserMe, threadID).Format(format).Context(ctx).Do()
}

func (s *RealGmailService) ModifyThread(ctx context.Context, threadID string, req *gmail.ModifyThreadRequest) (*gmail.Thread, error) {
	return s.service.Users.Threads.Modify(common.GmailUserMe, threadID, req).Context(ctx).Do()
}

func (s *RealGmailService) TrashThread(ctx context.Context, threadID string) (*gmail.Thread, error) {
	return s.service.Users.Threads.Trash(common.GmailUserMe, threadID).Context(ctx).Do()
}

func (s *RealGmailService) UntrashThread(ctx context.Context, threadID string) (*gmail.Thread, error) {
	return s.service.Users.Threads.Untrash(common.GmailUserMe, threadID).Context(ctx).Do()
}

// === Labels ===

func (s *RealGmailService) ListLabels(ctx context.Context) (*gmail.ListLabelsResponse, error) {
	return s.service.Users.Labels.List(common.GmailUserMe).Context(ctx).Do()
}

func (s *RealGmailService) GetLabel(ctx context.Context, labelID string) (*gmail.Label, error) {
	return s.service.Users.Labels.Get(common.GmailUserMe, labelID).Context(ctx).Do()
}

func (s *RealGmailService) CreateLabel(ctx context.Context, label *gmail.Label) (*gmail.Label, error) {
	return s.service.Users.Labels.Create(common.GmailUserMe, label).Context(ctx).Do()
}

func (s *RealGmailService) UpdateLabel(ctx context.Context, labelID string, label *gmail.Label) (*gmail.Label, error) {
	return s.service.Users.Labels.Update(common.GmailUserMe, labelID, label).Context(ctx).Do()
}

func (s *RealGmailService) DeleteLabel(ctx context.Context, labelID string) error {
	return s.service.Users.Labels.Delete(common.GmailUserMe, labelID).Context(ctx).Do()
}

// === Drafts ===

func (s *RealGmailService) ListDrafts(ctx context.Context, maxResults int64, pageToken string) (*gmail.ListDraftsResponse, error) {
	call := s.service.Users.Drafts.List(common.GmailUserMe).MaxResults(maxResults)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	return call.Context(ctx).Do()
}

func (s *RealGmailService) GetDraft(ctx context.Context, draftID, format string) (*gmail.Draft, error) {
	return s.service.Users.Drafts.Get(common.GmailUserMe, draftID).Format(format).Context(ctx).Do()
}

func (s *RealGmailService) CreateDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Draft, error) {
	return s.service.Users.Drafts.Create(common.GmailUserMe, draft).Context(ctx).Do()
}

func (s *RealGmailService) UpdateDraft(ctx context.Context, draftID string, draft *gmail.Draft) (*gmail.Draft, error) {
	return s.service.Users.Drafts.Update(common.GmailUserMe, draftID, draft).Context(ctx).Do()
}

func (s *RealGmailService) DeleteDraft(ctx context.Context, draftID string) error {
	return s.service.Users.Drafts.Delete(common.GmailUserMe, draftID).Context(ctx).Do()
}

func (s *RealGmailService) SendDraft(ctx context.Context, draft *gmail.Draft) (*gmail.Message, error) {
	return s.service.Users.Drafts.Send(common.GmailUserMe, draft).Context(ctx).Do()
}

// === Attachments ===

func (s *RealGmailService) GetAttachment(ctx context.Context, messageID, attachmentID string) (*gmail.MessagePartBody, error) {
	return s.service.Users.Messages.Attachments.Get(common.GmailUserMe, messageID, attachmentID).Context(ctx).Do()
}

// === Filters ===

func (s *RealGmailService) ListFilters(ctx context.Context) (*gmail.ListFiltersResponse, error) {
	return s.service.Users.Settings.Filters.List(common.GmailUserMe).Context(ctx).Do()
}

func (s *RealGmailService) CreateFilter(ctx context.Context, filter *gmail.Filter) (*gmail.Filter, error) {
	return s.service.Users.Settings.Filters.Create(common.GmailUserMe, filter).Context(ctx).Do()
}

func (s *RealGmailService) DeleteFilter(ctx context.Context, filterID string) error {
	return s.service.Users.Settings.Filters.Delete(common.GmailUserMe, filterID).Context(ctx).Do()
}

// === Profile & Settings ===

func (s *RealGmailService) GetProfile(ctx context.Context) (*gmail.Profile, error) {
	return s.service.Users.GetProfile(common.GmailUserMe).Context(ctx).Do()
}

func (s *RealGmailService) GetVacationSettings(ctx context.Context) (*gmail.VacationSettings, error) {
	return s.service.Users.Settings.GetVacation(common.GmailUserMe).Context(ctx).Do()
}

func (s *RealGmailService) UpdateVacationSettings(ctx context.Context, settings *gmail.VacationSettings) (*gmail.VacationSettings, error) {
	return s.service.Users.Settings.UpdateVacation(common.GmailUserMe, settings).Context(ctx).Do()
}
