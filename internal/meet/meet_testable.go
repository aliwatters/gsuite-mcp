package meet

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// conferenceRecordPrefix is the expected prefix for conference record resource names.
const conferenceRecordPrefix = "conferenceRecords/"

// ensureConferenceRecordName ensures the name has the conferenceRecords/ prefix.
func ensureConferenceRecordName(name string) string {
	if strings.HasPrefix(name, conferenceRecordPrefix) {
		return name
	}
	return conferenceRecordPrefix + name
}

// TestableListConferenceRecords lists conference records visible to the authenticated user.
func TestableListConferenceRecords(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	pageSize := common.ParseMaxResults(request.Params.Arguments, defaultPageSize, 100)

	records, nextPageToken, err := srv.ListConferenceRecords(ctx, pageToken, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	formatted := make([]map[string]any, 0, len(records))
	for _, r := range records {
		formatted = append(formatted, formatConferenceRecord(r))
	}

	result := map[string]any{
		"conference_records": formatted,
		"count":              len(records),
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// TestableGetConferenceRecord retrieves a specific conference record.
func TestableGetConferenceRecord(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name, errResult := common.RequireStringArg(request.Params.Arguments, "name")
	if errResult != nil {
		return errResult, nil
	}
	name = ensureConferenceRecordName(name)

	record, err := srv.GetConferenceRecord(ctx, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatConferenceRecord(record))
}

// TestableListParticipants lists participants of a conference.
func TestableListParticipants(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	parent, errResult := common.RequireStringArg(request.Params.Arguments, "conference_record")
	if errResult != nil {
		return errResult, nil
	}
	parent = ensureConferenceRecordName(parent)

	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	pageSize := common.ParseMaxResults(request.Params.Arguments, defaultPageSize, 100)

	participants, nextPageToken, totalSize, err := srv.ListParticipants(ctx, parent, pageToken, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	formatted := make([]map[string]any, 0, len(participants))
	for _, p := range participants {
		formatted = append(formatted, formatParticipant(p))
	}

	result := map[string]any{
		"conference_record": parent,
		"participants":      formatted,
		"count":             len(participants),
	}
	if totalSize > 0 {
		result["total_size"] = totalSize
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// TestableListTranscripts lists transcripts for a conference.
func TestableListTranscripts(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	parent, errResult := common.RequireStringArg(request.Params.Arguments, "conference_record")
	if errResult != nil {
		return errResult, nil
	}
	parent = ensureConferenceRecordName(parent)

	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")

	transcripts, nextPageToken, err := srv.ListTranscripts(ctx, parent, pageToken)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	formatted := make([]map[string]any, 0, len(transcripts))
	for _, t := range transcripts {
		formatted = append(formatted, formatTranscript(t))
	}

	result := map[string]any{
		"conference_record": parent,
		"transcripts":       formatted,
		"count":             len(transcripts),
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}

// TestableGetTranscript retrieves a specific transcript.
func TestableGetTranscript(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name, errResult := common.RequireStringArg(request.Params.Arguments, "name")
	if errResult != nil {
		return errResult, nil
	}

	transcript, err := srv.GetTranscript(ctx, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	return common.MarshalToolResult(formatTranscript(transcript))
}

// TestableListTranscriptEntries lists transcript entries for summarization.
func TestableListTranscriptEntries(ctx context.Context, request mcp.CallToolRequest, deps *MeetHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveMeetServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	parent, errResult := common.RequireStringArg(request.Params.Arguments, "transcript")
	if errResult != nil {
		return errResult, nil
	}

	pageToken := common.ParseStringArg(request.Params.Arguments, "page_token", "")
	pageSize := common.ParseMaxResults(request.Params.Arguments, defaultPageSize, 250)

	entries, nextPageToken, err := srv.ListTranscriptEntries(ctx, parent, pageToken, pageSize)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Meet API error: %v", err)), nil
	}

	formatted := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		formatted = append(formatted, formatTranscriptEntry(e))
	}

	result := map[string]any{
		"transcript": parent,
		"entries":    formatted,
		"count":      len(entries),
	}
	if nextPageToken != "" {
		result["next_page_token"] = nextPageToken
	}

	return common.MarshalToolResult(result)
}
