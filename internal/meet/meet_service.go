package meet

import (
	"context"

	meet "google.golang.org/api/meet/v2"
)

// MeetService defines the interface for Google Meet API operations.
// This interface enables dependency injection and testing with mocks.
type MeetService interface {
	// ListConferenceRecords lists conference records visible to the authenticated user.
	ListConferenceRecords(ctx context.Context, pageToken string, pageSize int64) ([]*meet.ConferenceRecord, string, error)

	// GetConferenceRecord retrieves a specific conference record by name.
	GetConferenceRecord(ctx context.Context, name string) (*meet.ConferenceRecord, error)

	// ListParticipants lists participants of a conference record.
	ListParticipants(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.Participant, string, int64, error)

	// ListTranscripts lists transcripts for a conference record.
	ListTranscripts(ctx context.Context, parent string, pageToken string) ([]*meet.Transcript, string, error)

	// GetTranscript retrieves a specific transcript by name.
	GetTranscript(ctx context.Context, name string) (*meet.Transcript, error)

	// ListTranscriptEntries lists transcript entries for a transcript.
	ListTranscriptEntries(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.TranscriptEntry, string, error)
}

// RealMeetService wraps the Meet API client and implements MeetService.
type RealMeetService struct {
	service *meet.Service
}

// NewRealMeetService creates a new RealMeetService wrapping the given API service.
func NewRealMeetService(service *meet.Service) *RealMeetService {
	return &RealMeetService{service: service}
}

// defaultPageSize is the default number of items per page.
const defaultPageSize = 25

// ListConferenceRecords lists conference records.
func (s *RealMeetService) ListConferenceRecords(ctx context.Context, pageToken string, pageSize int64) ([]*meet.ConferenceRecord, string, error) {
	call := s.service.ConferenceRecords.List().Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.ConferenceRecords, resp.NextPageToken, nil
}

// GetConferenceRecord retrieves a specific conference record.
func (s *RealMeetService) GetConferenceRecord(ctx context.Context, name string) (*meet.ConferenceRecord, error) {
	return s.service.ConferenceRecords.Get(name).Context(ctx).Do()
}

// ListParticipants lists participants of a conference record.
func (s *RealMeetService) ListParticipants(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.Participant, string, int64, error) {
	call := s.service.ConferenceRecords.Participants.List(parent).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", 0, err
	}
	return resp.Participants, resp.NextPageToken, resp.TotalSize, nil
}

// ListTranscripts lists transcripts for a conference record.
func (s *RealMeetService) ListTranscripts(ctx context.Context, parent string, pageToken string) ([]*meet.Transcript, string, error) {
	call := s.service.ConferenceRecords.Transcripts.List(parent).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.Transcripts, resp.NextPageToken, nil
}

// GetTranscript retrieves a specific transcript.
func (s *RealMeetService) GetTranscript(ctx context.Context, name string) (*meet.Transcript, error) {
	return s.service.ConferenceRecords.Transcripts.Get(name).Context(ctx).Do()
}

// ListTranscriptEntries lists transcript entries for a transcript.
func (s *RealMeetService) ListTranscriptEntries(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.TranscriptEntry, string, error) {
	call := s.service.ConferenceRecords.Transcripts.Entries.List(parent).Context(ctx)
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.TranscriptEntries, resp.NextPageToken, nil
}

// formatConferenceRecord formats a conference record for output.
func formatConferenceRecord(record *meet.ConferenceRecord) map[string]any {
	result := map[string]any{
		"name":       record.Name,
		"space":      record.Space,
		"start_time": record.StartTime,
	}
	if record.EndTime != "" {
		result["end_time"] = record.EndTime
	}
	if record.ExpireTime != "" {
		result["expire_time"] = record.ExpireTime
	}
	return result
}

// formatParticipant formats a participant for output.
func formatParticipant(p *meet.Participant) map[string]any {
	result := map[string]any{
		"name":                p.Name,
		"earliest_start_time": p.EarliestStartTime,
	}
	if p.LatestEndTime != "" {
		result["latest_end_time"] = p.LatestEndTime
	}
	if p.SignedinUser != nil {
		result["type"] = "signed_in"
		result["display_name"] = p.SignedinUser.DisplayName
		result["user"] = p.SignedinUser.User
	} else if p.AnonymousUser != nil {
		result["type"] = "anonymous"
		result["display_name"] = p.AnonymousUser.DisplayName
	} else if p.PhoneUser != nil {
		result["type"] = "phone"
		result["display_name"] = p.PhoneUser.DisplayName
	}
	return result
}

// formatTranscript formats a transcript for output.
func formatTranscript(t *meet.Transcript) map[string]any {
	result := map[string]any{
		"name":       t.Name,
		"state":      t.State,
		"start_time": t.StartTime,
	}
	if t.EndTime != "" {
		result["end_time"] = t.EndTime
	}
	if t.DocsDestination != nil {
		result["docs_destination"] = map[string]any{
			"document_id": t.DocsDestination.Document,
			"export_uri":  t.DocsDestination.ExportUri,
		}
	}
	return result
}

// formatTranscriptEntry formats a transcript entry for output.
func formatTranscriptEntry(e *meet.TranscriptEntry) map[string]any {
	result := map[string]any{
		"name":        e.Name,
		"participant": e.Participant,
		"text":        e.Text,
		"start_time":  e.StartTime,
	}
	if e.EndTime != "" {
		result["end_time"] = e.EndTime
	}
	if e.LanguageCode != "" {
		result["language_code"] = e.LanguageCode
	}
	return result
}
