package common

// Google API default identifiers.
const (
	// GmailUserMe is the special Gmail user ID that refers to the authenticated user.
	GmailUserMe = "me"

	// DefaultCalendarID is the default calendar identifier for the user's primary calendar.
	DefaultCalendarID = "primary"

	// DefaultTaskListID is the default task list identifier.
	DefaultTaskListID = "@default"
)

// Gmail API limits.
const (
	GmailDefaultMaxResults = 20
	GmailMaxResultsLimit   = 100
	GmailMaxBatchMessages  = 25
)

// Calendar API limits.
const (
	CalendarDefaultMaxResults = 25
	CalendarMaxResultsLimit   = 250
)

// Contacts API limits.
const (
	ContactsDefaultPageSize       = 100
	ContactsMaxPageSize           = 1000
	ContactsSearchDefaultPageSize = 30
	ContactsSearchMaxPageSize     = 30
)

// Tasks API limits.
const (
	TasksDefaultMaxResults = 100
	TasksMaxResultsLimit   = 100
)

// Drive API limits.
const (
	DriveSearchDefaultMaxResults = 20
	DriveSearchMaxResultsLimit   = 100
	DriveListDefaultMaxResults   = 100
	DriveListMaxResultsLimit     = 1000
	DriveMaxFileSize             = 10 * 1024 * 1024 // 10MB
)

// Sheets value input options.
const (
	// ValueInputUserEntered parses values as if typed into the UI (formulas, dates, etc.).
	ValueInputUserEntered = "USER_ENTERED"

	// ValueInputRaw stores values exactly as provided without parsing.
	ValueInputRaw = "RAW"
)
