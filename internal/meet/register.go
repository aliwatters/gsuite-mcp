package meet

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Meet tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// meet_list_conference_records - List recent conferences
	s.AddTool(mcp.NewTool("meet_list_conference_records",
		mcp.WithDescription("List Google Meet conference records visible to the authenticated user. Returns meeting metadata including space, start/end times."),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default 25, max 100)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleListConferenceRecords)

	// meet_get_conference_record - Get a specific conference record
	s.AddTool(mcp.NewTool("meet_get_conference_record",
		mcp.WithDescription("Get details of a specific Google Meet conference record by name."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Conference record resource name (e.g., 'conferenceRecords/abc-123' or just 'abc-123')")),
		common.WithAccountParam(),
	), HandleGetConferenceRecord)

	// meet_list_participants - List conference participants
	s.AddTool(mcp.NewTool("meet_list_participants",
		mcp.WithDescription("List participants of a Google Meet conference. Shows signed-in users, anonymous users, and phone users with join/leave times."),
		mcp.WithString("conference_record", mcp.Required(), mcp.Description("Conference record resource name (e.g., 'conferenceRecords/abc-123' or just 'abc-123')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (default 25, max 100)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleListParticipants)

	// meet_list_transcripts - List transcripts for a conference
	s.AddTool(mcp.NewTool("meet_list_transcripts",
		mcp.WithDescription("List transcripts for a Google Meet conference. Shows transcript state and Google Docs destination when available."),
		mcp.WithString("conference_record", mcp.Required(), mcp.Description("Conference record resource name (e.g., 'conferenceRecords/abc-123' or just 'abc-123')")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleListTranscripts)

	// meet_get_transcript - Get transcript details
	s.AddTool(mcp.NewTool("meet_get_transcript",
		mcp.WithDescription("Get details of a specific Google Meet transcript including state and Google Docs destination."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Transcript resource name (e.g., 'conferenceRecords/abc-123/transcripts/t1')")),
		common.WithAccountParam(),
	), HandleGetTranscript)

	// meet_list_transcript_entries - Get transcript text entries
	s.AddTool(mcp.NewTool("meet_list_transcript_entries",
		mcp.WithDescription("Get transcript text entries for a Google Meet transcript. Returns speaker, text, and timestamps — useful for meeting summarization."),
		mcp.WithString("transcript", mcp.Required(), mcp.Description("Transcript resource name (e.g., 'conferenceRecords/abc-123/transcripts/t1')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of entries to return (default 25, max 250)")),
		common.WithPageToken(),
		common.WithAccountParam(),
	), HandleListTranscriptEntries)
}
