package citation

import (
	"context"
	"testing"
)

// newTestDualStore creates a DualStore backed by an in-memory SQLite for testing.
// The sheets store is nil — the rebuild path is skipped since we pretend the DB exists.
func newTestDualStore(t *testing.T) *DualStore {
	t.Helper()
	sqliteStore, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite store: %v", err)
	}
	t.Cleanup(func() { sqliteStore.Close() })

	// Construct a DualStore without a real SheetsStore — tests exercise SQLite paths only.
	// SaveChunks and other dual-write methods require sheets; use sqlite directly for
	// pure dual-store routing tests.
	ds := &DualStore{
		indexID: "test-index",
		sheets:  nil,
		sqlite:  sqliteStore,
	}
	return ds
}

// TestDualStore_SQLiteReadPath verifies that reads delegate to SQLite.
func TestDualStore_SQLiteReadPath(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	// Populate SQLite directly
	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "deck.pptx", Content: "revenue data"},
		{ID: "c2", FileID: "f1", FileName: "deck.pptx", Content: "market share"},
	}
	if err := sqlite.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}

	// GetChunks reads from SQLite
	got, err := ds.GetChunks(ctx, []string{"c1", "c2"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(got))
	}
}

// TestDualStore_Search verifies FTS5 search through DualStore delegates to SQLite.
func TestDualStore_Search(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "report.docx", Content: "branosotine clinical trials"},
		{ID: "c2", FileID: "f2", FileName: "other.docx", Content: "market expansion Europe"},
	}
	if err := sqlite.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}

	results, err := ds.Search(ctx, "branosotine", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "c1" {
		t.Errorf("expected chunk c1, got %q", results[0].ID)
	}
}

// TestDualStore_GetConcepts verifies concept reads delegate to SQLite.
func TestDualStore_GetConcepts(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "f.pptx", Content: "test"},
	}
	sqlite.SaveChunks(ctx, chunks)

	mappings := []ConceptMapping{
		{Concept: "testing", ChunkIDs: []string{"c1"}},
	}
	if err := sqlite.SaveConcepts(ctx, mappings); err != nil {
		t.Fatalf("SaveConcepts: %v", err)
	}

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	got, err := ds.GetConcepts(ctx)
	if err != nil {
		t.Fatalf("GetConcepts: %v", err)
	}
	if len(got) != 1 || got[0].Concept != "testing" {
		t.Errorf("expected concept 'testing', got %v", got)
	}
}

// TestDualStore_GetSummaries verifies summary reads delegate to SQLite.
func TestDualStore_GetSummaries(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	if err := sqlite.SaveSummary(ctx, LevelSummary{Level: 2, ParentID: "idx", Summary: "corpus overview"}); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	got, err := ds.GetSummaries(ctx, 2)
	if err != nil {
		t.Fatalf("GetSummaries: %v", err)
	}
	if len(got) != 1 || got[0].Summary != "corpus overview" {
		t.Errorf("expected summary 'corpus overview', got %v", got)
	}
}

// TestDualStore_GetMetadata verifies metadata reads delegate to SQLite.
func TestDualStore_GetMetadata(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	sqlite.SetMetadata(ctx, "index_id", "test")
	sqlite.SetMetadata(ctx, "chunk_count", "5")

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	info, err := ds.GetMetadata(ctx)
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if info.IndexID != "test" {
		t.Errorf("expected index_id=test, got %q", info.IndexID)
	}
	if info.ChunkCount != 5 {
		t.Errorf("expected chunk_count=5, got %d", info.ChunkCount)
	}
}

// TestDualStore_GetIndexedFiles verifies file reads delegate to SQLite.
func TestDualStore_GetIndexedFiles(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	file := IndexedFile{
		FileID: "f1", FileName: "deck.pptx",
		MimeType:     "application/vnd.google-apps.presentation",
		ModifiedTime: "2026-01-01T00:00:00Z", ChunkCount: 3,
	}
	if err := sqlite.SaveIndexedFile(ctx, file); err != nil {
		t.Fatalf("SaveIndexedFile: %v", err)
	}

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	got, err := ds.GetIndexedFiles(ctx)
	if err != nil {
		t.Fatalf("GetIndexedFiles: %v", err)
	}
	if len(got) != 1 || got[0].FileID != "f1" {
		t.Errorf("expected file f1, got %v", got)
	}
}

// TestDualStore_DeleteIndexedFile verifies deletion delegates to SQLite.
func TestDualStore_DeleteIndexedFile(t *testing.T) {
	ctx := context.Background()
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	t.Cleanup(func() { sqlite.Close() })

	sqlite.SaveIndexedFile(ctx, IndexedFile{FileID: "f1", FileName: "deck.pptx"})

	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	if err := ds.DeleteIndexedFile(ctx, "f1"); err != nil {
		t.Fatalf("DeleteIndexedFile: %v", err)
	}

	got, err := ds.GetIndexedFiles(ctx)
	if err != nil {
		t.Fatalf("GetIndexedFiles: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 files after delete, got %d", len(got))
	}
}

// TestDualStore_Close verifies Close works without panic.
func TestDualStore_Close(t *testing.T) {
	sqlite, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating sqlite: %v", err)
	}
	ds := &DualStore{indexID: "idx", sheets: nil, sqlite: sqlite}
	if err := ds.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
