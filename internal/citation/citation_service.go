package citation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

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
		if r.indexedFile.FileID == "" {
			continue // uninitialized slot (goroutine did not run)
		}
		// nil chunks means the file had no extractable content — still record the file
		// with ChunkCount=0 so the index has a consistent files-row.
		if r.chunks == nil {
			r.chunks = []Chunk{}
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

// chunkFile dispatches to the appropriate chunking strategy based on MIME type.
func (s *RealCitationService) chunkFile(ctx context.Context, file *drive.File) ([]Chunk, error) {
	switch file.MimeType {
	case "application/vnd.google-apps.presentation":
		return s.chunkSlides(ctx, file)
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		// Uploaded .pptx — download raw content and extract text from XML
		return s.chunkDownloadedFile(ctx, file)
	default:
		if strings.HasPrefix(file.MimeType, "application/vnd.google-apps.") {
			// Google-native file — export as plain text
			return s.chunkExportedText(ctx, file)
		}
		// Non-Google file — try downloading raw content
		return s.chunkDownloadedFile(ctx, file)
	}
}

// chunkSlides extracts per-slide text from a Google Slides presentation via the Slides API.
func (s *RealCitationService) chunkSlides(ctx context.Context, file *drive.File) ([]Chunk, error) {
	pres, err := s.slidesService.Presentations.Get(file.Id).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return extractSlidesText(file.Id, file.Name, pres), nil
}

// chunkDownloadedFile downloads a non-Google-native file and extracts text.
// For .pptx files, reads text from XML inside the zip. For others, treats content as plain text.
func (s *RealCitationService) chunkDownloadedFile(ctx context.Context, file *drive.File) ([]Chunk, error) {
	resp, err := s.driveService.Files.Get(file.Id).SupportsAllDrives(true).Context(ctx).Download()
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", file.Name, err)
	}
	defer resp.Body.Close()

	data, err := limitedRead(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", file.Name, err)
	}

	return downloadAndChunk(data, file)
}

// chunkPptxBytes extracts text from a .pptx file (zip containing XML slides).
// Delegates to the package-level chunkPptxData function.
func (s *RealCitationService) chunkPptxBytes(file *drive.File, data []byte) ([]Chunk, error) {
	return chunkPptxData(file, data)
}

// chunkExportedText exports a Google-native file as plain text and splits into chunks.
func (s *RealCitationService) chunkExportedText(ctx context.Context, file *drive.File) ([]Chunk, error) {
	resp, err := s.driveService.Files.Export(file.Id, "text/plain").Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := limitedRead(resp.Body)
	if err != nil {
		return nil, err
	}

	return chunkText(file.Id, file.Name, string(data)), nil
}

func (s *RealCitationService) SaveConcepts(ctx context.Context, indexID string, mappings []ConceptMapping) error {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return fmt.Errorf("getting store for index %q: %w", indexID, err)
	}
	return store.SaveConcepts(ctx, mappings)
}

func (s *RealCitationService) SaveSummary(ctx context.Context, indexID string, summary LevelSummary) error {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return fmt.Errorf("getting store for index %q: %w", indexID, err)
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
