package common

import (
	"regexp"
	"strings"
)

// Google resource ID regex patterns, compiled once at package level.
var (
	// resourceIDFromPath matches /d/RESOURCE_ID or /document/d/RESOURCE_ID patterns.
	resourceIDFromPath = regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)

	// resourceIDFromQuery matches ?id=RESOURCE_ID or &id=RESOURCE_ID patterns.
	resourceIDFromQuery = regexp.MustCompile(`[?&]id=([a-zA-Z0-9_-]+)`)

	// spreadsheetURLPattern matches Google Sheets URLs specifically.
	spreadsheetURLPattern = regexp.MustCompile(`docs\.google\.com/spreadsheets/d/([a-zA-Z0-9_-]+)`)
)

// ExtractGoogleResourceID extracts a Google resource ID from a URL or returns the input as-is
// if it's already a plain ID. It handles common Google URL patterns:
//   - https://drive.google.com/file/d/FILE_ID/view
//   - https://drive.google.com/open?id=FILE_ID
//   - https://docs.google.com/document/d/DOC_ID/edit
//   - https://docs.google.com/spreadsheets/d/SHEET_ID/edit
//   - Plain IDs (no slashes or query params)
func ExtractGoogleResourceID(input string) string {
	// If it doesn't look like a URL, return as-is
	if !strings.Contains(input, "/") && !strings.Contains(input, "?") {
		return input
	}

	// Try to extract from /d/RESOURCE_ID pattern (covers drive, docs, sheets, slides)
	if matches := resourceIDFromPath.FindStringSubmatch(input); len(matches) >= 2 {
		return matches[1]
	}

	// Try to extract from ?id=RESOURCE_ID pattern
	if matches := resourceIDFromQuery.FindStringSubmatch(input); len(matches) >= 2 {
		return matches[1]
	}

	// Return as-is if no match
	return input
}

// ExtractSpreadsheetID extracts a spreadsheet ID from a Google Sheets URL
// or returns the input as-is if it's already a plain ID.
// This uses a Sheets-specific pattern that matches docs.google.com/spreadsheets/d/ID.
func ExtractSpreadsheetID(input string) string {
	if matches := spreadsheetURLPattern.FindStringSubmatch(input); len(matches) > 1 {
		return matches[1]
	}
	return input
}
