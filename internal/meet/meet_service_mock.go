package meet

import (
	"context"
	"fmt"
	"strings"

	meet "google.golang.org/api/meet/v2"
)

// MockMeetService implements MeetService for testing.
type MockMeetService struct {
	// ConferenceRecords stores mock conference record data keyed by name.
	ConferenceRecords map[string]*meet.ConferenceRecord

	// Participants stores mock participant data keyed by conference record name.
	Participants map[string][]*meet.Participant

	// Transcripts stores mock transcript data keyed by conference record name.
	Transcripts map[string][]*meet.Transcript

	// TranscriptEntries stores mock transcript entry data keyed by transcript name.
	TranscriptEntries map[string][]*meet.TranscriptEntry

	// Errors allows tests to configure specific errors for methods.
	Errors struct {
		ListConferenceRecords error
		GetConferenceRecord   error
		ListParticipants      error
		ListTranscripts       error
		GetTranscript         error
		ListTranscriptEntries error
	}

	// Calls tracks method invocations for verification.
	Calls struct {
		ListConferenceRecords []string
		GetConferenceRecord   []string
		ListParticipants      []string
		ListTranscripts       []string
		GetTranscript         []string
		ListTranscriptEntries []string
	}
}

// NewMockMeetService creates a new mock Meet service with default test data.
func NewMockMeetService() *MockMeetService {
	m := &MockMeetService{
		ConferenceRecords: make(map[string]*meet.ConferenceRecord),
		Participants:      make(map[string][]*meet.Participant),
		Transcripts:       make(map[string][]*meet.Transcript),
		TranscriptEntries: make(map[string][]*meet.TranscriptEntry),
	}

	// Add a default test conference record
	m.ConferenceRecords["conferenceRecords/abc-123"] = &meet.ConferenceRecord{
		Name:       "conferenceRecords/abc-123",
		Space:      "spaces/meeting-space-1",
		StartTime:  "2024-03-15T10:00:00Z",
		EndTime:    "2024-03-15T11:00:00Z",
		ExpireTime: "2024-04-14T11:00:00Z",
	}

	m.ConferenceRecords["conferenceRecords/def-456"] = &meet.ConferenceRecord{
		Name:      "conferenceRecords/def-456",
		Space:     "spaces/meeting-space-2",
		StartTime: "2024-03-16T14:00:00Z",
	}

	// Add test participants
	m.Participants["conferenceRecords/abc-123"] = []*meet.Participant{
		{
			Name:              "conferenceRecords/abc-123/participants/p1",
			EarliestStartTime: "2024-03-15T10:00:00Z",
			LatestEndTime:     "2024-03-15T11:00:00Z",
			SignedinUser: &meet.SignedinUser{
				DisplayName: "Alice Smith",
				User:        "users/alice123",
			},
		},
		{
			Name:              "conferenceRecords/abc-123/participants/p2",
			EarliestStartTime: "2024-03-15T10:05:00Z",
			LatestEndTime:     "2024-03-15T10:55:00Z",
			AnonymousUser: &meet.AnonymousUser{
				DisplayName: "Guest User",
			},
		},
	}

	// Add test transcripts
	m.Transcripts["conferenceRecords/abc-123"] = []*meet.Transcript{
		{
			Name:      "conferenceRecords/abc-123/transcripts/t1",
			State:     "FILE_GENERATED",
			StartTime: "2024-03-15T10:00:00Z",
			EndTime:   "2024-03-15T11:00:00Z",
			DocsDestination: &meet.DocsDestination{
				Document:  "doc-abc-123",
				ExportUri: "https://docs.google.com/document/d/doc-abc-123/view",
			},
		},
	}

	// Add test transcript entries
	m.TranscriptEntries["conferenceRecords/abc-123/transcripts/t1"] = []*meet.TranscriptEntry{
		{
			Name:         "conferenceRecords/abc-123/transcripts/t1/entries/e1",
			Participant:  "conferenceRecords/abc-123/participants/p1",
			Text:         "Hello everyone, let's get started with the meeting.",
			StartTime:    "2024-03-15T10:00:05Z",
			EndTime:      "2024-03-15T10:00:10Z",
			LanguageCode: "en-US",
		},
		{
			Name:         "conferenceRecords/abc-123/transcripts/t1/entries/e2",
			Participant:  "conferenceRecords/abc-123/participants/p2",
			Text:         "Sounds good, I have the agenda ready.",
			StartTime:    "2024-03-15T10:00:12Z",
			EndTime:      "2024-03-15T10:00:16Z",
			LanguageCode: "en-US",
		},
	}

	return m
}

// ListConferenceRecords lists mock conference records.
func (m *MockMeetService) ListConferenceRecords(ctx context.Context, pageToken string, pageSize int64) ([]*meet.ConferenceRecord, string, error) {
	m.Calls.ListConferenceRecords = append(m.Calls.ListConferenceRecords, pageToken)

	if m.Errors.ListConferenceRecords != nil {
		return nil, "", m.Errors.ListConferenceRecords
	}

	records := make([]*meet.ConferenceRecord, 0, len(m.ConferenceRecords))
	for _, r := range m.ConferenceRecords {
		records = append(records, r)
	}
	return records, "", nil
}

// GetConferenceRecord retrieves a mock conference record by name.
func (m *MockMeetService) GetConferenceRecord(ctx context.Context, name string) (*meet.ConferenceRecord, error) {
	m.Calls.GetConferenceRecord = append(m.Calls.GetConferenceRecord, name)

	if m.Errors.GetConferenceRecord != nil {
		return nil, m.Errors.GetConferenceRecord
	}

	record, ok := m.ConferenceRecords[name]
	if !ok {
		return nil, fmt.Errorf("conference record not found: %s", name)
	}
	return record, nil
}

// ListParticipants lists mock participants for a conference record.
func (m *MockMeetService) ListParticipants(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.Participant, string, int64, error) {
	m.Calls.ListParticipants = append(m.Calls.ListParticipants, parent)

	if m.Errors.ListParticipants != nil {
		return nil, "", 0, m.Errors.ListParticipants
	}

	// Validate conference record exists
	if _, ok := m.ConferenceRecords[parent]; !ok {
		return nil, "", 0, fmt.Errorf("conference record not found: %s", parent)
	}

	participants := m.Participants[parent]
	if participants == nil {
		participants = []*meet.Participant{}
	}
	return participants, "", int64(len(participants)), nil
}

// ListTranscripts lists mock transcripts for a conference record.
func (m *MockMeetService) ListTranscripts(ctx context.Context, parent string, pageToken string) ([]*meet.Transcript, string, error) {
	m.Calls.ListTranscripts = append(m.Calls.ListTranscripts, parent)

	if m.Errors.ListTranscripts != nil {
		return nil, "", m.Errors.ListTranscripts
	}

	if _, ok := m.ConferenceRecords[parent]; !ok {
		return nil, "", fmt.Errorf("conference record not found: %s", parent)
	}

	transcripts := m.Transcripts[parent]
	if transcripts == nil {
		transcripts = []*meet.Transcript{}
	}
	return transcripts, "", nil
}

// GetTranscript retrieves a mock transcript by name.
func (m *MockMeetService) GetTranscript(ctx context.Context, name string) (*meet.Transcript, error) {
	m.Calls.GetTranscript = append(m.Calls.GetTranscript, name)

	if m.Errors.GetTranscript != nil {
		return nil, m.Errors.GetTranscript
	}

	// Search through all transcripts for the matching name
	for _, transcripts := range m.Transcripts {
		for _, t := range transcripts {
			if t.Name == name {
				return t, nil
			}
		}
	}
	return nil, fmt.Errorf("transcript not found: %s", name)
}

// ListTranscriptEntries lists mock transcript entries.
func (m *MockMeetService) ListTranscriptEntries(ctx context.Context, parent string, pageToken string, pageSize int64) ([]*meet.TranscriptEntry, string, error) {
	m.Calls.ListTranscriptEntries = append(m.Calls.ListTranscriptEntries, parent)

	if m.Errors.ListTranscriptEntries != nil {
		return nil, "", m.Errors.ListTranscriptEntries
	}

	// Validate transcript exists by checking if parent matches any known transcript
	found := false
	for _, transcripts := range m.Transcripts {
		for _, t := range transcripts {
			if t.Name == parent {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		// Also check by prefix pattern
		parts := strings.Split(parent, "/")
		if len(parts) < 4 {
			return nil, "", fmt.Errorf("transcript not found: %s", parent)
		}
		confName := parts[0] + "/" + parts[1]
		if _, ok := m.ConferenceRecords[confName]; !ok {
			return nil, "", fmt.Errorf("transcript not found: %s", parent)
		}
	}

	entries := m.TranscriptEntries[parent]
	if entries == nil {
		entries = []*meet.TranscriptEntry{}
	}
	return entries, "", nil
}
