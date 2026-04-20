package citation

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/sheets/v4"
)

// CreateIndex creates a new citation index Sheet in the given Drive folder.
// It sets up the required tabs (chunks, concepts, summaries, files, metadata),
// writes column headers, and populates initial metadata.
func (s *RealCitationService) CreateIndex(ctx context.Context, name, folderID string) (*IndexInfo, error) {
	// Create spreadsheet with required tabs
	ss := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: "Citation Index: " + name},
		Sheets: []*sheets.Sheet{
			{Properties: &sheets.SheetProperties{Title: "chunks"}},
			{Properties: &sheets.SheetProperties{Title: "concepts"}},
			{Properties: &sheets.SheetProperties{Title: "summaries"}},
			{Properties: &sheets.SheetProperties{Title: "files"}},
			{Properties: &sheets.SheetProperties{Title: "metadata"}},
		},
	}
	created, err := s.sheets.CreateSpreadsheet(ctx, ss)
	if err != nil {
		return nil, fmt.Errorf("creating index sheet: %w", err)
	}

	// Move to folder if specified
	if folderID != "" {
		_, err = s.drive.MoveFile(ctx, created.SpreadsheetId, folderID)
		if err != nil {
			return nil, fmt.Errorf("moving sheet to folder: %w", err)
		}
	}

	// Write column headers for each tab
	headers := map[string][][]any{
		"chunks!A1":    {{"id", "file_id", "file_name", "content", "summary", "concepts", "page_number", "section_heading", "paragraph_index", "char_start", "char_end"}},
		"concepts!A1":  {{"concept", "chunk_ids"}},
		"summaries!A1": {{"level", "parent_id", "summary"}},
		"files!A1":     {{"file_id", "file_name", "mime_type", "modified_time", "chunk_count"}},
		"metadata!A1":  {{"key", "value"}},
	}
	var data []*sheets.ValueRange
	for r, v := range headers {
		data = append(data, &sheets.ValueRange{Range: r, Values: v})
	}
	err = s.sheets.BatchUpdateValues(ctx, created.SpreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
		Data:             data,
	})
	if err != nil {
		return nil, fmt.Errorf("writing headers: %w", err)
	}

	// Write initial metadata rows
	now := time.Now().UTC().Format(time.RFC3339)
	metaRows := [][]any{
		{"index_id", name},
		{"source_folder_id", folderID},
		{"created_at", now},
		{"doc_count", "0"},
		{"chunk_count", "0"},
	}
	err = s.sheets.AppendValues(ctx, created.SpreadsheetId, "metadata!A2", &sheets.ValueRange{
		Values: metaRows,
	})
	if err != nil {
		return nil, fmt.Errorf("writing metadata: %w", err)
	}

	return &IndexInfo{
		IndexID:        name,
		SheetID:        created.SpreadsheetId,
		SourceFolderID: folderID,
		CreatedAt:      now,
	}, nil
}
