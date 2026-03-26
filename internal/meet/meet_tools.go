package meet

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleListConferenceRecords = common.WrapHandler[MeetService](TestableListConferenceRecords)
	HandleGetConferenceRecord   = common.WrapHandler[MeetService](TestableGetConferenceRecord)
	HandleListParticipants      = common.WrapHandler[MeetService](TestableListParticipants)
	HandleListTranscripts       = common.WrapHandler[MeetService](TestableListTranscripts)
	HandleGetTranscript         = common.WrapHandler[MeetService](TestableGetTranscript)
	HandleListTranscriptEntries = common.WrapHandler[MeetService](TestableListTranscriptEntries)
)
