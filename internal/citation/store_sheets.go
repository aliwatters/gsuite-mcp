package citation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

// SheetsStore reads/writes citation data to a Google Sheet.
// Not a full IndexStore — used by DualStore for the write path and rebuild.
type SheetsStore struct {
	sheetsSrv *sheets.Service
	sheetID   string
}

// NewSheetsStore creates a SheetsStore for the given spreadsheet.
func NewSheetsStore(srv *sheets.Service, sheetID string) *SheetsStore {
	return &SheetsStore{sheetsSrv: srv, sheetID: sheetID}
}

// AppendChunks appends chunk rows to the chunks tab.
func (s *SheetsStore) AppendChunks(ctx context.Context, chunks []Chunk) error {
	var rows [][]any
	for _, c := range chunks {
		rows = append(rows, ChunkToSheetRow(c))
	}
	_, err := s.sheetsSrv.Spreadsheets.Values.Append(s.sheetID, "chunks!A2", &sheets.ValueRange{
		Values: rows,
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending chunks to sheet: %w", err)
	}
	return nil
}

// AppendConcepts appends concept rows to the concepts tab.
func (s *SheetsStore) AppendConcepts(ctx context.Context, mappings []ConceptMapping) error {
	var rows [][]any
	for _, m := range mappings {
		chunkIDsJSON, _ := json.Marshal(m.ChunkIDs)
		rows = append(rows, []any{m.Concept, string(chunkIDsJSON)})
	}
	_, err := s.sheetsSrv.Spreadsheets.Values.Append(s.sheetID, "concepts!A2", &sheets.ValueRange{
		Values: rows,
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending concepts to sheet: %w", err)
	}
	return nil
}

// AppendSummary appends a summary row to the summaries tab.
func (s *SheetsStore) AppendSummary(ctx context.Context, summary LevelSummary) error {
	row := [][]any{{summary.Level, summary.ParentID, summary.Summary}}
	_, err := s.sheetsSrv.Spreadsheets.Values.Append(s.sheetID, "summaries!A2", &sheets.ValueRange{
		Values: row,
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending summary to sheet: %w", err)
	}
	return nil
}

// ReadAllChunks reads all chunk rows from the Sheet.
func (s *SheetsStore) ReadAllChunks(ctx context.Context) ([]Chunk, error) {
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "chunks!A2:K").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading chunks: %w", err)
	}

	var chunks []Chunk
	for _, row := range resp.Values {
		chunks = append(chunks, SheetRowToChunk(row))
	}
	return chunks, nil
}

// ReadAllConcepts reads all concept rows from the Sheet.
func (s *SheetsStore) ReadAllConcepts(ctx context.Context) ([]ConceptMapping, error) {
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "concepts!A2:B").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading concepts: %w", err)
	}

	var mappings []ConceptMapping
	for _, row := range resp.Values {
		if len(row) < 2 {
			continue
		}
		concept, _ := row[0].(string)
		chunkIDsStr, _ := row[1].(string)
		var chunkIDs []string
		json.Unmarshal([]byte(chunkIDsStr), &chunkIDs)
		mappings = append(mappings, ConceptMapping{Concept: concept, ChunkIDs: chunkIDs})
	}
	return mappings, nil
}

// ReadAllSummaries reads all summary rows from the Sheet.
func (s *SheetsStore) ReadAllSummaries(ctx context.Context) ([]LevelSummary, error) {
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "summaries!A2:C").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading summaries: %w", err)
	}

	var summaries []LevelSummary
	for _, row := range resp.Values {
		if len(row) < 3 {
			continue
		}
		var level int
		switch v := row[0].(type) {
		case float64:
			level = int(v)
		case string:
			if v != "" {
				n, err := strconv.Atoi(v)
				if err != nil {
					log.Printf("citation: ReadAllSummaries: level %q is not an integer: %v", v, err)
				}
				level = n
			}
		}
		parentID, _ := row[1].(string)
		summary, _ := row[2].(string)

		summaries = append(summaries, LevelSummary{Level: level, ParentID: parentID, Summary: summary})
	}
	return summaries, nil
}

// ReadAllMetadata reads all metadata rows from the Sheet.
func (s *SheetsStore) ReadAllMetadata(ctx context.Context) (map[string]string, error) {
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "metadata!A2:B").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading metadata: %w", err)
	}

	meta := make(map[string]string)
	for _, row := range resp.Values {
		if len(row) < 2 {
			continue
		}
		key, _ := row[0].(string)
		value := fmt.Sprintf("%v", row[1])
		meta[key] = value
	}
	return meta, nil
}

// AppendFiles appends indexed file tracking rows to the files tab.
func (s *SheetsStore) AppendFiles(ctx context.Context, files []IndexedFile) error {
	var rows [][]any
	for _, f := range files {
		rows = append(rows, []any{f.FileID, f.FileName, f.MimeType, f.ModifiedTime, f.ChunkCount})
	}
	_, err := s.sheetsSrv.Spreadsheets.Values.Append(s.sheetID, "files!A2", &sheets.ValueRange{
		Values: rows,
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending files to sheet: %w", err)
	}
	return nil
}

// ReadAllFiles reads all indexed file tracking rows from the Sheet.
func (s *SheetsStore) ReadAllFiles(ctx context.Context) ([]IndexedFile, error) {
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "files!A2:E").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("reading files: %w", err)
	}

	var files []IndexedFile
	for _, row := range resp.Values {
		if len(row) < 5 {
			continue
		}
		chunkCount := 0
		switch v := row[4].(type) {
		case float64:
			chunkCount = int(v)
		case string:
			n, err := strconv.Atoi(v)
			if err != nil {
				log.Printf("citation: ReadAllFiles: chunk_count %q is not an integer: %v", v, err)
			}
			chunkCount = n
		}
		fileID, _ := row[0].(string)
		fileName, _ := row[1].(string)
		mimeType, _ := row[2].(string)
		modifiedTime, _ := row[3].(string)
		files = append(files, IndexedFile{
			FileID:       fileID,
			FileName:     fileName,
			MimeType:     mimeType,
			ModifiedTime: modifiedTime,
			ChunkCount:   chunkCount,
		})
	}
	return files, nil
}

// RewriteFilesTab clears and rewrites the files tab (used after refresh).
func (s *SheetsStore) RewriteFilesTab(ctx context.Context, files []IndexedFile) error {
	// Clear existing data
	_, err := s.sheetsSrv.Spreadsheets.Values.Clear(s.sheetID, "files!A2:E", &sheets.ClearValuesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("clearing files tab: %w", err)
	}
	if len(files) == 0 {
		return nil
	}
	return s.AppendFiles(ctx, files)
}

// RewriteChunksForFile clears chunks for a specific file from the chunks tab.
// This reads all chunks, filters out the file, and rewrites.
func (s *SheetsStore) RewriteChunksForFile(ctx context.Context, fileID string) error {
	// Read all chunks
	allChunks, err := s.ReadAllChunks(ctx)
	if err != nil {
		return fmt.Errorf("reading all chunks: %w", err)
	}

	// Filter out chunks for this file
	var remaining []Chunk
	for _, c := range allChunks {
		if c.FileID != fileID {
			remaining = append(remaining, c)
		}
	}

	// Clear and rewrite
	_, err = s.sheetsSrv.Spreadsheets.Values.Clear(s.sheetID, "chunks!A2:K", &sheets.ClearValuesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("clearing chunks tab: %w", err)
	}
	if len(remaining) > 0 {
		return s.AppendChunks(ctx, remaining)
	}
	return nil
}

// UpdateMetadata updates a metadata key. Finds the row and updates in place,
// or appends if not found.
func (s *SheetsStore) UpdateMetadata(ctx context.Context, key, value string) error {
	// Read existing metadata to find the row
	resp, err := s.sheetsSrv.Spreadsheets.Values.Get(s.sheetID, "metadata!A2:B").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reading metadata: %w", err)
	}

	for i, row := range resp.Values {
		if len(row) > 0 {
			k, _ := row[0].(string)
			if strings.EqualFold(k, key) {
				rowNum := i + 2 // 1-indexed, skip header
				updateRange := fmt.Sprintf("metadata!B%d", rowNum)
				_, err := s.sheetsSrv.Spreadsheets.Values.Update(s.sheetID, updateRange, &sheets.ValueRange{
					Values: [][]any{{value}},
				}).ValueInputOption("RAW").Context(ctx).Do()
				if err != nil {
					return fmt.Errorf("updating metadata %q: %w", key, err)
				}
				return nil
			}
		}
	}

	// Not found — append
	_, err = s.sheetsSrv.Spreadsheets.Values.Append(s.sheetID, "metadata!A2", &sheets.ValueRange{
		Values: [][]any{{key, value}},
	}).ValueInputOption("RAW").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("appending metadata %q: %w", key, err)
	}
	return nil
}
