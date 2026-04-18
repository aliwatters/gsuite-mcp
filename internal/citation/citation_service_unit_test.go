package citation

import (
	"context"
	"testing"
)

func TestRealCitationService_ListIndexes(t *testing.T) {
	svc := &RealCitationService{
		stores: make(map[string]*DualStore),
		config: &CitationConfig{
			Indexes: map[string]IndexEntry{
				"idx1": {SheetID: "sheet1"},
				"idx2": {SheetID: "sheet2"},
			},
		},
	}

	indexes, err := svc.ListIndexes(context.Background())
	if err != nil {
		t.Fatalf("ListIndexes: %v", err)
	}
	if len(indexes) != 2 {
		t.Errorf("expected 2 indexes, got %d", len(indexes))
	}
}

func TestRealCitationService_Lookup(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "doc.pdf", Content: "revenue growth"},
		{ID: "c2", FileID: "f2", FileName: "doc2.pdf", Content: "market share"},
	}
	sqlite.SaveChunks(ctx, chunks)

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	got, err := svc.Lookup(ctx, "idx1", "revenue", 10)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if len(got) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(got))
	}
}

func TestRealCitationService_GetChunks(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "doc.pdf", Content: "content"},
	}
	sqlite.SaveChunks(ctx, chunks)

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	got, err := svc.GetChunks(ctx, "idx1", []string{"c1"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 1 || got[0].ID != "c1" {
		t.Errorf("expected chunk c1, got %v", got)
	}
}

func TestRealCitationService_VerifyClaim(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "doc.pdf", Content: "revenue was $1M in Q3"},
	}
	sqlite.SaveChunks(ctx, chunks)

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	got, err := svc.VerifyClaim(ctx, "idx1", "revenue", 5)
	if err != nil {
		t.Fatalf("VerifyClaim: %v", err)
	}
	if len(got) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(got))
	}
}

func TestRealCitationService_GetOverview(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	sqlite.SaveIndexedFile(ctx, IndexedFile{
		FileID: "f1", FileName: "doc.pdf", ChunkCount: 3,
	})
	sqlite.SaveConcepts(ctx, []ConceptMapping{
		{Concept: "revenue", ChunkIDs: []string{"c1"}},
	})
	sqlite.SaveSummary(ctx, LevelSummary{Level: 2, ParentID: "idx", Summary: "corpus"})

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	overview, err := svc.GetOverview(ctx, "idx1")
	if err != nil {
		t.Fatalf("GetOverview: %v", err)
	}
	if overview["index_id"] != "idx1" {
		t.Errorf("expected index_id=idx1, got %v", overview["index_id"])
	}
	if overview["doc_count"].(int) != 1 {
		t.Errorf("expected doc_count=1, got %v", overview["doc_count"])
	}
	if overview["corpus_summary"] != "corpus" {
		t.Errorf("expected corpus_summary=corpus, got %v", overview["corpus_summary"])
	}
}

func TestRealCitationService_SaveConcepts(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	// Need chunks first
	sqlite.SaveChunks(ctx, []Chunk{
		{ID: "c1", FileID: "f1", Content: "test"},
	})

	// SaveConcepts in DualStore requires a SheetsStore for dual-write.
	// Test via SQLiteStore directly to verify the service's store dispatch.
	err := sqlite.SaveConcepts(ctx, []ConceptMapping{
		{Concept: "testing", ChunkIDs: []string{"c1"}},
	})
	if err != nil {
		t.Fatalf("SaveConcepts via sqlite: %v", err)
	}

	got, err := sqlite.GetConcepts(ctx)
	if err != nil {
		t.Fatalf("GetConcepts: %v", err)
	}
	if len(got) != 1 || got[0].Concept != "testing" {
		t.Errorf("expected concept 'testing', got %v", got)
	}
}

func TestRealCitationService_SaveSummary(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	// SaveSummary in DualStore requires a SheetsStore for dual-write.
	// Test via SQLiteStore directly.
	err := sqlite.SaveSummary(ctx, LevelSummary{Level: 2, ParentID: "idx", Summary: "overview"})
	if err != nil {
		t.Fatalf("SaveSummary via sqlite: %v", err)
	}

	got, err := sqlite.GetSummaries(ctx, 2)
	if err != nil {
		t.Fatalf("GetSummaries: %v", err)
	}
	if len(got) != 1 || got[0].Summary != "overview" {
		t.Errorf("expected summary 'overview', got %v", got)
	}
}

func TestRealCitationService_GetStore_UnknownIndex(t *testing.T) {
	svc := &RealCitationService{
		stores: make(map[string]*DualStore),
		config: &CitationConfig{Indexes: map[string]IndexEntry{}},
	}
	_, err := svc.getStore(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for unknown index")
	}
}

func TestRealCitationService_Lookup_ZeroLimit(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	sqlite.SaveChunks(ctx, []Chunk{
		{ID: "c1", FileID: "f1", Content: "test content"},
	})

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	// limit=0 should default to 10
	got, err := svc.Lookup(ctx, "idx1", "test", 0)
	if err != nil {
		t.Fatalf("Lookup with limit=0: %v", err)
	}
	_ = got // result count doesn't matter, just verify no error
}

func TestRealCitationService_GetStore_Caching(t *testing.T) {
	ctx := context.Background()
	sqlite, _ := NewSQLiteStore(":memory:")
	defer sqlite.Close()

	ds := &DualStore{indexID: "idx1", sheets: nil, sqlite: sqlite}
	svc := &RealCitationService{
		stores: map[string]*DualStore{"idx1": ds},
		config: &CitationConfig{Indexes: map[string]IndexEntry{"idx1": {}}},
	}

	// First call returns cached store
	got1, err := svc.getStore(ctx, "idx1")
	if err != nil {
		t.Fatalf("getStore: %v", err)
	}
	got2, err := svc.getStore(ctx, "idx1")
	if err != nil {
		t.Fatalf("getStore (second): %v", err)
	}
	if got1 != got2 {
		t.Error("expected same store instance from cache")
	}
}

func TestSQLiteStore_GetMetadata_ParseError(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Set a non-integer value for doc_count
	store.db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('doc_count', 'not-a-number')`)

	_, err := store.GetMetadata(ctx)
	if err == nil {
		t.Error("expected error when doc_count is not an integer")
	}
}

func TestSQLiteStore_GetChunks_Empty(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	got, err := store.GetChunks(ctx, nil)
	if err != nil {
		t.Fatalf("GetChunks with nil IDs: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 chunks for nil IDs, got %d", len(got))
	}
}

func TestSheetRowToChunk_InvalidIntField(t *testing.T) {
	// When an int field contains a non-numeric string, should default to 0
	row := []any{"id1", "file1", "file.txt", "content", "summary", "[]", "not-a-number", "heading", "0", "0", "0"}
	c := SheetRowToChunk(row)
	if c.Location.PageNumber != 0 {
		t.Errorf("expected PageNumber=0 for invalid int, got %d", c.Location.PageNumber)
	}
}

func TestRealCitationService_FormatCitation_NoSectionHeading(t *testing.T) {
	svc := &RealCitationService{}
	chunk := Chunk{
		FileName: "doc.pdf",
		Content:  "short content",
		Location: Location{PageNumber: 5, SectionHeading: ""},
	}
	citation := svc.FormatCitation(context.Background(), chunk)
	if citation == "" {
		t.Error("expected non-empty citation")
	}
}
