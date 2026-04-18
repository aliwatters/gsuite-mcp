package citation

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	gslides "google.golang.org/api/slides/v1"
)

// CitationService is the interface for all citation operations.
type CitationService interface {
	// CreateIndex creates a new index Sheet in the given folder.
	CreateIndex(ctx context.Context, name, folderID string) (*IndexInfo, error)

	// AddDocuments extracts + chunks documents from a Drive folder, writing to the index.
	AddDocuments(ctx context.Context, indexID string, fileIDs []string) (int, error)

	// SaveConcepts saves LLM-extracted concepts for chunks.
	SaveConcepts(ctx context.Context, indexID string, mappings []ConceptMapping) error

	// SaveSummary saves an LLM-generated summary.
	SaveSummary(ctx context.Context, indexID string, summary LevelSummary) error

	// ListIndexes returns known indexes.
	ListIndexes(ctx context.Context) ([]IndexInfo, error)

	// GetOverview returns top-level summary, concepts, and doc list.
	GetOverview(ctx context.Context, indexID string) (map[string]any, error)

	// Lookup searches chunks by keyword/concept.
	Lookup(ctx context.Context, indexID, query string, limit int) ([]Chunk, error)

	// GetChunks retrieves full chunk data by IDs.
	GetChunks(ctx context.Context, indexID string, chunkIDs []string) ([]Chunk, error)

	// VerifyClaim finds candidate chunks for a claim.
	VerifyClaim(ctx context.Context, indexID, claim string, limit int) ([]Chunk, error)

	// FormatCitation formats a chunk as a human-readable citation.
	FormatCitation(ctx context.Context, chunk Chunk) string

	// RefreshIndex checks for updated/removed/renamed files and re-indexes as needed.
	RefreshIndex(ctx context.Context, indexID string) (*RefreshResult, error)
}

// RealCitationService implements CitationService using Google APIs.
type RealCitationService struct {
	driveService  *drive.Service
	sheetsService *sheets.Service
	slidesService *gslides.Service
	mu            sync.RWMutex
	stores        map[string]*DualStore // indexID → store
	config        *CitationConfig
}

// CitationConfig holds citation-specific configuration.
type CitationConfig struct {
	Indexes map[string]IndexEntry `json:"indexes"`
}

// IndexEntry maps an index ID to its Sheet ID.
type IndexEntry struct {
	SheetID string `json:"sheet_id"`
}

// maxExportSize limits Drive export downloads to 50MB.
const maxExportSize = 50 * 1024 * 1024

// NewRealCitationService creates a new citation service from an HTTP client.
func NewRealCitationService(ctx context.Context, client *http.Client, cfg *CitationConfig) (*RealCitationService, error) {
	driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating drive service: %w", err)
	}
	sheetsSrv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	slidesSrv, err := gslides.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating slides service: %w", err)
	}

	if cfg == nil {
		cfg = &CitationConfig{Indexes: make(map[string]IndexEntry)}
	}

	return &RealCitationService{
		driveService:  driveSrv,
		sheetsService: sheetsSrv,
		slidesService: slidesSrv,
		stores:        make(map[string]*DualStore),
		config:        cfg,
	}, nil
}

func (s *RealCitationService) getStore(ctx context.Context, indexID string) (*DualStore, error) {
	s.mu.RLock()
	store, ok := s.stores[indexID]
	s.mu.RUnlock()
	if ok {
		return store, nil
	}

	entry, ok := s.config.Indexes[indexID]
	if !ok {
		return nil, fmt.Errorf("unknown index %q — add it to config", indexID)
	}

	store, err := NewDualStore(ctx, indexID, entry.SheetID, s.sheetsService)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	// Double-check after acquiring write lock
	if existing, ok := s.stores[indexID]; ok {
		s.mu.Unlock()
		store.Close()
		return existing, nil
	}
	s.stores[indexID] = store
	s.mu.Unlock()
	return store, nil
}

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
	created, err := s.sheetsService.Spreadsheets.Create(ss).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("creating index sheet: %w", err)
	}

	// Move to folder if specified
	if folderID != "" {
		_, err = s.driveService.Files.Update(created.SpreadsheetId, nil).
			AddParents(folderID).
			SupportsAllDrives(true).
			Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("moving sheet to folder: %w", err)
		}
	}

	// Write headers
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
	_, err = s.sheetsService.Spreadsheets.Values.BatchUpdate(created.SpreadsheetId, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
		Data:             data,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("writing headers: %w", err)
	}

	// Write initial metadata
	now := time.Now().UTC().Format(time.RFC3339)
	metaRows := [][]any{
		{"index_id", name},
		{"source_folder_id", folderID},
		{"created_at", now},
		{"doc_count", "0"},
		{"chunk_count", "0"},
	}
	_, err = s.sheetsService.Spreadsheets.Values.Append(created.SpreadsheetId, "metadata!A2", &sheets.ValueRange{
		Values: metaRows,
	}).ValueInputOption("RAW").Context(ctx).Do()
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

// addDocResult holds the result of processing a single file in AddDocuments.
type addDocResult struct {
	chunks      []Chunk
	indexedFile IndexedFile
	err         error
	fileID      string
}

func (s *RealCitationService) AddDocuments(ctx context.Context, indexID string, fileIDs []string) (int, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return 0, err
	}

	// Fetch file metadata and chunk files concurrently (up to 5 at a time).
	results := make([]addDocResult, len(fileIDs))
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for i, fileID := range fileIDs {
		i, fileID := i, fileID // capture loop vars
		g.Go(func() error {
			file, err := s.driveService.Files.Get(fileID).
				Fields("id,name,mimeType,modifiedTime").
				SupportsAllDrives(true).
				Context(gCtx).Do()
			if err != nil {
				results[i] = addDocResult{fileID: fileID, err: fmt.Errorf("getting file %s: %w", fileID, err)}
				return nil // partial failure — keep going
			}

			chunks, err := s.chunkFile(gCtx, file)
			if err != nil {
				results[i] = addDocResult{fileID: fileID, err: fmt.Errorf("chunking %s: %w", file.Name, err)}
				return nil // partial failure — keep going
			}

			results[i] = addDocResult{
				fileID: fileID,
				chunks: chunks,
				indexedFile: IndexedFile{
					FileID:       file.Id,
					FileName:     file.Name,
					MimeType:     file.MimeType,
					ModifiedTime: file.ModifiedTime,
					ChunkCount:   len(chunks),
				},
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return 0, err
	}

	// Collect results and save sequentially (store is not safe for concurrent writes).
	totalChunks := 0
	var errs []string
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err.Error())
			continue
		}
		if r.chunks == nil {
			continue // empty result slot
		}

		if err := store.SaveChunks(ctx, r.chunks); err != nil {
			errs = append(errs, fmt.Sprintf("saving chunks for %s: %v", r.indexedFile.FileName, err))
			continue
		}

		if err := store.SaveIndexedFile(ctx, r.indexedFile); err != nil {
			errs = append(errs, fmt.Sprintf("tracking file %s: %v", r.indexedFile.FileName, err))
			continue
		}

		totalChunks += len(r.chunks)
	}

	if len(errs) > 0 {
		return totalChunks, fmt.Errorf("partial failures (%d/%d files succeeded): %s",
			len(fileIDs)-len(errs), len(fileIDs), strings.Join(errs, "; "))
	}

	return totalChunks, nil
}

func (s *RealCitationService) chunkFile(ctx context.Context, file *drive.File) ([]Chunk, error) {
	switch file.MimeType {
	case "application/vnd.google-apps.presentation":
		return s.chunkSlides(ctx, file)
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		// Uploaded .pptx — download raw content and extract text from XML
		return s.chunkDownloadedFile(ctx, file)
	default:
		if strings.HasPrefix(file.MimeType, "application/vnd.google-apps.") {
			// Google-native file — export as text
			return s.chunkExportedText(ctx, file)
		}
		// Non-Google file — try downloading raw content
		return s.chunkDownloadedFile(ctx, file)
	}
}

func (s *RealCitationService) chunkSlides(ctx context.Context, file *drive.File) ([]Chunk, error) {
	pres, err := s.slidesService.Presentations.Get(file.Id).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var chunks []Chunk
	for i, slide := range pres.Slides {
		var textParts []string
		for _, elem := range slide.PageElements {
			if elem.Shape != nil && elem.Shape.Text != nil {
				for _, te := range elem.Shape.Text.TextElements {
					if te.TextRun != nil {
						textParts = append(textParts, te.TextRun.Content)
					}
				}
			}
			if elem.Table != nil {
				for _, row := range elem.Table.TableRows {
					for _, cell := range row.TableCells {
						if cell.Text != nil {
							for _, te := range cell.Text.TextElements {
								if te.TextRun != nil {
									textParts = append(textParts, te.TextRun.Content)
								}
							}
						}
					}
				}
			}
		}

		content := strings.TrimSpace(strings.Join(textParts, ""))
		if content == "" {
			continue
		}

		chunks = append(chunks, Chunk{
			ID:       chunkID(file.Id, i),
			FileID:   file.Id,
			FileName: file.Name,
			Content:  content,
			Location: Location{PageNumber: i + 1},
		})
	}
	return chunks, nil
}

// chunkDownloadedFile downloads a non-Google-native file and extracts text.
// For .pptx files, it reads text from the XML entries inside the zip.
// For other binary formats, it returns a single chunk with available text.
func (s *RealCitationService) chunkDownloadedFile(ctx context.Context, file *drive.File) ([]Chunk, error) {
	resp, err := s.driveService.Files.Get(file.Id).SupportsAllDrives(true).Context(ctx).Download()
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", file.Name, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxExportSize))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", file.Name, err)
	}

	if file.MimeType == "application/vnd.openxmlformats-officedocument.presentationml.presentation" {
		return s.chunkPptxBytes(file, data)
	}

	// Fallback: treat as plain text
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, nil
	}
	return chunkText(file.Id, file.Name, text), nil
}

// chunkPptxBytes extracts text from a pptx file (zip containing XML slides).
// Each slide/N.xml becomes a separate chunk.
func (s *RealCitationService) chunkPptxBytes(file *drive.File, data []byte) ([]Chunk, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading pptx zip: %w", err)
	}

	// Collect slide XML files (ppt/slides/slide1.xml, slide2.xml, ...)
	type slideEntry struct {
		num  int
		file *zip.File
	}
	var slides []slideEntry
	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, "ppt/slides/slide") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		// Extract slide number from "ppt/slides/slide12.xml"
		numStr := strings.TrimPrefix(f.Name, "ppt/slides/slide")
		numStr = strings.TrimSuffix(numStr, ".xml")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		slides = append(slides, slideEntry{num: num, file: f})
	}

	// Sort by slide number
	sort.Slice(slides, func(i, j int) bool { return slides[i].num < slides[j].num })

	var chunks []Chunk
	for _, s := range slides {
		text, err := extractTextFromXML(s.file)
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		chunks = append(chunks, Chunk{
			ID:       chunkID(file.Id, s.num-1),
			FileID:   file.Id,
			FileName: file.Name,
			Content:  text,
			Location: Location{PageNumber: s.num},
		})
	}
	return chunks, nil
}

// extractTextFromXML reads a zip entry and extracts all text content from XML <a:t> tags.
func extractTextFromXML(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	decoder := xml.NewDecoder(rc)
	var textParts []string
	var inText bool
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// <a:t> contains text in OOXML presentations
			if t.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				textParts = append(textParts, string(t))
			}
		}
	}
	return strings.Join(textParts, " "), nil
}

// chunkExportedText exports a file as plain text via Drive API and chunks by paragraphs.
// Handles Google Docs, Sheets, and any other exportable format.
func (s *RealCitationService) chunkExportedText(ctx context.Context, file *drive.File) ([]Chunk, error) {
	resp, err := s.driveService.Files.Export(file.Id, "text/plain").Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxExportSize))
	if err != nil {
		return nil, err
	}

	return chunkText(file.Id, file.Name, string(data)), nil
}

func (s *RealCitationService) SaveConcepts(ctx context.Context, indexID string, mappings []ConceptMapping) error {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return err
	}
	return store.SaveConcepts(ctx, mappings)
}

func (s *RealCitationService) SaveSummary(ctx context.Context, indexID string, summary LevelSummary) error {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return err
	}
	return store.SaveSummary(ctx, summary)
}

func (s *RealCitationService) ListIndexes(_ context.Context) ([]IndexInfo, error) {
	var indexes []IndexInfo
	for id, entry := range s.config.Indexes {
		indexes = append(indexes, IndexInfo{
			IndexID: id,
			SheetID: entry.SheetID,
		})
	}
	return indexes, nil
}

func (s *RealCitationService) GetOverview(ctx context.Context, indexID string) (map[string]any, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return nil, err
	}

	// Compute counts from actual data, not metadata (which may be stale)
	files, err := store.GetIndexedFiles(ctx)
	if err != nil {
		return nil, err
	}

	totalChunks := 0
	for _, f := range files {
		totalChunks += f.ChunkCount
	}

	concepts, err := store.GetConcepts(ctx)
	if err != nil {
		return nil, err
	}

	// Get corpus summary (level 2)
	summaries, err := store.GetSummaries(ctx, 2)
	if err != nil {
		return nil, err
	}

	conceptNames := make([]string, len(concepts))
	for i, c := range concepts {
		conceptNames[i] = c.Concept
	}

	fileNames := make([]string, len(files))
	for i, f := range files {
		fileNames[i] = f.FileName
	}

	result := map[string]any{
		"index_id":      indexID,
		"doc_count":     len(files),
		"chunk_count":   totalChunks,
		"concept_count": len(concepts),
		"concepts":      conceptNames,
		"files":         fileNames,
	}
	if len(summaries) > 0 {
		result["corpus_summary"] = summaries[0].Summary
	}
	return result, nil
}

func (s *RealCitationService) Lookup(ctx context.Context, indexID, query string, limit int) ([]Chunk, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	return store.Search(ctx, query, limit)
}

func (s *RealCitationService) GetChunks(ctx context.Context, indexID string, chunkIDs []string) ([]Chunk, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return nil, err
	}
	return store.GetChunks(ctx, chunkIDs)
}

func (s *RealCitationService) VerifyClaim(ctx context.Context, indexID, claim string, limit int) ([]Chunk, error) {
	return s.Lookup(ctx, indexID, claim, limit)
}

func (s *RealCitationService) FormatCitation(_ context.Context, chunk Chunk) string {
	var parts []string
	parts = append(parts, chunk.FileName)
	if chunk.Location.PageNumber > 0 {
		parts = append(parts, fmt.Sprintf("Slide %d", chunk.Location.PageNumber))
	}
	if chunk.Location.SectionHeading != "" {
		parts = append(parts, fmt.Sprintf("Section: %s", chunk.Location.SectionHeading))
	}
	// Include a snippet (first 80 chars)
	snippet := chunk.Content
	if len(snippet) > 80 {
		snippet = snippet[:80] + "..."
	}
	parts = append(parts, fmt.Sprintf("'%s'", snippet))
	return "Source: " + strings.Join(parts, ", ")
}

func (s *RealCitationService) RefreshIndex(ctx context.Context, indexID string) (*RefreshResult, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return nil, err
	}

	// Get currently indexed files
	indexed, err := store.GetIndexedFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting indexed files: %w", err)
	}

	result := &RefreshResult{}
	indexedMap := make(map[string]IndexedFile, len(indexed))
	for _, f := range indexed {
		indexedMap[f.FileID] = f
	}

	// refreshFileResult holds the outcome of checking a single indexed file.
	type refreshFileResult struct {
		prev    IndexedFile
		current *drive.File
		err     error
	}

	// Fetch current Drive metadata for all indexed files concurrently (up to 5 at a time).
	fileResults := make([]refreshFileResult, len(indexed))
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for i, prev := range indexed {
		i, prev := i, prev
		g.Go(func() error {
			current, err := s.driveService.Files.Get(prev.FileID).
				Fields("id,name,mimeType,modifiedTime,trashed").
				SupportsAllDrives(true).
				Context(gCtx).Do()
			fileResults[i] = refreshFileResult{prev: prev, current: current, err: err}
			return nil // partial failure — keep going
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Process results sequentially (store writes and result mutation are not concurrent-safe).
	for _, fr := range fileResults {
		if fr.err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", fr.prev.FileName, fr.err))
			continue
		}
		current := fr.current
		prev := fr.prev

		// File was trashed → remove
		if current.Trashed {
			if delErr := s.removeFileFromIndex(ctx, store, prev.FileID); delErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("removing %s: %v", prev.FileName, delErr))
			} else {
				result.Removed = append(result.Removed, prev.FileName)
			}
			continue
		}

		// File was renamed
		if current.Name != prev.FileName {
			result.Renamed = append(result.Renamed, fmt.Sprintf("%s → %s", prev.FileName, current.Name))
		}

		// File was modified → re-chunk
		if current.ModifiedTime != prev.ModifiedTime {
			if reErr := s.reindexFile(ctx, store, current); reErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("re-indexing %s: %v", current.Name, reErr))
			} else {
				result.Updated = append(result.Updated, current.Name)
			}
		}
	}

	return result, nil
}

// removeFileFromIndex deletes a file's chunks and tracking record.
func (s *RealCitationService) removeFileFromIndex(ctx context.Context, store *DualStore, fileID string) error {
	if err := store.DeleteChunksByFileID(ctx, fileID); err != nil {
		return err
	}
	return store.DeleteIndexedFile(ctx, fileID)
}

// reindexFile removes old chunks and re-chunks the file.
func (s *RealCitationService) reindexFile(ctx context.Context, store *DualStore, file *drive.File) error {
	// Remove old chunks
	if err := store.DeleteChunksByFileID(ctx, file.Id); err != nil {
		return fmt.Errorf("deleting old chunks: %w", err)
	}

	// Re-chunk
	chunks, err := s.chunkFile(ctx, file)
	if err != nil {
		return fmt.Errorf("chunking: %w", err)
	}

	if err := store.SaveChunks(ctx, chunks); err != nil {
		return fmt.Errorf("saving chunks: %w", err)
	}

	// Update tracking
	return store.SaveIndexedFile(ctx, IndexedFile{
		FileID:       file.Id,
		FileName:     file.Name,
		MimeType:     file.MimeType,
		ModifiedTime: file.ModifiedTime,
		ChunkCount:   len(chunks),
	})
}

// chunkID generates a deterministic chunk ID from file ID and offset.
func chunkID(fileID string, offset int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", fileID, offset)))
	return fmt.Sprintf("%x", h[:8])
}

// chunkText splits text into chunks by paragraph boundaries.
// Targets ~500-1000 tokens (~2000-4000 chars) per chunk.
func chunkText(fileID, fileName, text string) []Chunk {
	const maxChunkSize = 3000

	paragraphs := strings.Split(text, "\n\n")
	var chunks []Chunk
	var current strings.Builder
	paraIdx := 0
	charPos := 0
	chunkStart := 0

	flush := func() {
		content := strings.TrimSpace(current.String())
		if content == "" {
			return
		}
		chunks = append(chunks, Chunk{
			ID:       chunkID(fileID, len(chunks)),
			FileID:   fileID,
			FileName: fileName,
			Content:  content,
			Location: Location{
				ParagraphIndex: paraIdx,
				CharStart:      chunkStart,
				CharEnd:        charPos,
			},
		})
		current.Reset()
		chunkStart = charPos
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			charPos += 2
			continue
		}

		if current.Len()+len(para) > maxChunkSize && current.Len() > 0 {
			flush()
		}

		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(para)
		charPos += len(para) + 2
		paraIdx++
	}
	flush()

	return chunks
}
