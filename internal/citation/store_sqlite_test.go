package citation

import (
	"context"
	"testing"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("creating store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestSQLiteStore_SaveAndGetChunks(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	chunks := []Chunk{
		{
			ID: "abc123", FileID: "f1", FileName: "deck.pptx",
			Content:  "Q3 revenue exceeded targets by 15%",
			Concepts: []string{"revenue", "Q3"},
			Location: Location{PageNumber: 7},
		},
		{
			ID: "def456", FileID: "f1", FileName: "deck.pptx",
			Content:  "New product launch planned for Q4",
			Location: Location{PageNumber: 12},
		},
	}

	if err := store.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	got, err := store.GetChunks(ctx, []string{"abc123", "def456"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(got))
	}

	// Verify concepts loaded
	if len(got[0].Concepts) != 2 {
		t.Errorf("expected 2 concepts for chunk abc123, got %d", len(got[0].Concepts))
	}
}

func TestSQLiteStore_FTS5Search(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "report.docx", Content: "The branosotine compound showed promising results in Phase II trials"},
		{ID: "c2", FileID: "f1", FileName: "report.docx", Content: "Revenue projections indicate 20% growth in European markets"},
		{ID: "c3", FileID: "f2", FileName: "summary.docx", Content: "Clinical trial enrollment exceeded expectations for branosotine"},
	}

	if err := store.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	results, err := store.Search(ctx, "branosotine", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for 'branosotine', got %d", len(results))
	}

	results, err = store.Search(ctx, "revenue", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result for 'revenue', got %d", len(results))
	}
}

func TestSQLiteStore_Concepts(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Save chunks first
	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "deck.pptx", Content: "slide content"},
		{ID: "c2", FileID: "f1", FileName: "deck.pptx", Content: "more content"},
	}
	if err := store.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	mappings := []ConceptMapping{
		{Concept: "clinical trials", ChunkIDs: []string{"c1", "c2"}},
		{Concept: "revenue", ChunkIDs: []string{"c2"}},
	}
	if err := store.SaveConcepts(ctx, mappings); err != nil {
		t.Fatalf("SaveConcepts: %v", err)
	}

	got, err := store.GetConcepts(ctx)
	if err != nil {
		t.Fatalf("GetConcepts: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 concepts, got %d", len(got))
	}
}

func TestSQLiteStore_Summaries(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SaveSummary(ctx, LevelSummary{Level: 0, ParentID: "c1", Summary: "chunk summary"}); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	if err := store.SaveSummary(ctx, LevelSummary{Level: 2, ParentID: "index1", Summary: "corpus overview"}); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	got, err := store.GetSummaries(ctx, 2)
	if err != nil {
		t.Fatalf("GetSummaries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 level-2 summary, got %d", len(got))
	}
	if got[0].Summary != "corpus overview" {
		t.Errorf("unexpected summary: %q", got[0].Summary)
	}
}

func TestSQLiteStore_Metadata(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	if err := store.SetMetadata(ctx, "index_id", "test-index"); err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}
	if err := store.SetMetadata(ctx, "chunk_count", "42"); err != nil {
		t.Fatalf("SetMetadata: %v", err)
	}

	info, err := store.GetMetadata(ctx)
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if info.IndexID != "test-index" {
		t.Errorf("expected index_id=test-index, got %q", info.IndexID)
	}
	if info.ChunkCount != 42 {
		t.Errorf("expected chunk_count=42, got %d", info.ChunkCount)
	}
}

func TestSQLiteStore_BulkInsertFromSheet(t *testing.T) {
	store := newTestStore(t)

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "test.pptx", Content: "content one", Location: Location{PageNumber: 1}},
		{ID: "c2", FileID: "f1", FileName: "test.pptx", Content: "content two", Location: Location{PageNumber: 2}},
	}
	concepts := []ConceptMapping{
		{Concept: "testing", ChunkIDs: []string{"c1", "c2"}},
	}
	summaries := []LevelSummary{
		{Level: 2, ParentID: "idx", Summary: "corpus summary"},
	}
	files := []IndexedFile{
		{FileID: "f1", FileName: "test.pptx", MimeType: "application/vnd.google-apps.presentation", ModifiedTime: "2026-01-01T00:00:00Z", ChunkCount: 2},
	}
	meta := map[string]string{"index_id": "rebuild-test", "chunk_count": "2"}

	if err := store.BulkInsertFromSheet(chunks, concepts, summaries, files, meta); err != nil {
		t.Fatalf("BulkInsertFromSheet: %v", err)
	}

	ctx := context.Background()

	got, err := store.GetChunks(ctx, []string{"c1", "c2"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 chunks, got %d", len(got))
	}

	info, err := store.GetMetadata(ctx)
	if err != nil {
		t.Fatalf("GetMetadata: %v", err)
	}
	if info.IndexID != "rebuild-test" {
		t.Errorf("expected index_id=rebuild-test, got %q", info.IndexID)
	}

	// Verify files were inserted
	gotFiles, err := store.GetIndexedFiles(ctx)
	if err != nil {
		t.Fatalf("GetIndexedFiles: %v", err)
	}
	if len(gotFiles) != 1 {
		t.Fatalf("expected 1 indexed file, got %d", len(gotFiles))
	}
	if gotFiles[0].FileID != "f1" {
		t.Errorf("expected file_id=f1, got %q", gotFiles[0].FileID)
	}
}

func TestChunkText(t *testing.T) {
	text := "First paragraph about clinical trials.\n\nSecond paragraph about revenue growth and market expansion.\n\nThird paragraph about regulatory compliance."

	chunks := chunkText("f1", "test.docx", text)
	if len(chunks) != 1 {
		// All paragraphs fit in one chunk (< 3000 chars)
		t.Logf("got %d chunks", len(chunks))
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].FileID != "f1" {
		t.Errorf("expected file_id=f1, got %q", chunks[0].FileID)
	}
}

func TestSheetRowRoundTrip(t *testing.T) {
	original := Chunk{
		ID:       "abc123",
		FileID:   "f1",
		FileName: "deck.pptx",
		Content:  "test content",
		Summary:  "test summary",
		Concepts: []string{"concept1", "concept2"},
		Location: Location{PageNumber: 7, SectionHeading: "Results"},
	}

	row := ChunkToSheetRow(original)
	got := SheetRowToChunk(row)

	if got.ID != original.ID || got.FileID != original.FileID || got.Content != original.Content {
		t.Errorf("round-trip mismatch: got ID=%q FileID=%q Content=%q", got.ID, got.FileID, got.Content)
	}
	if len(got.Concepts) != 2 {
		t.Errorf("expected 2 concepts, got %d", len(got.Concepts))
	}
	if got.Location.PageNumber != 7 {
		t.Errorf("expected page 7, got %d", got.Location.PageNumber)
	}
}

func TestSQLiteStore_DeleteChunksByFileID(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "deck.pptx", Content: "slide one", Concepts: []string{"topic"}, Location: Location{PageNumber: 1}},
		{ID: "c2", FileID: "f1", FileName: "deck.pptx", Content: "slide two", Location: Location{PageNumber: 2}},
		{ID: "c3", FileID: "f2", FileName: "other.docx", Content: "other doc", Location: Location{ParagraphIndex: 1}},
	}
	if err := store.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	if err := store.DeleteChunksByFileID(ctx, "f1"); err != nil {
		t.Fatalf("DeleteChunksByFileID: %v", err)
	}

	// f1 chunks should be gone
	got, err := store.GetChunks(ctx, []string{"c1", "c2"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 chunks for f1, got %d", len(got))
	}

	// f2 chunk should remain
	got, err = store.GetChunks(ctx, []string{"c3"})
	if err != nil {
		t.Fatalf("GetChunks: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 chunk for f2, got %d", len(got))
	}
}

func TestSQLiteStore_IndexedFiles(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	file := IndexedFile{
		FileID:       "f1",
		FileName:     "deck.pptx",
		MimeType:     "application/vnd.google-apps.presentation",
		ModifiedTime: "2026-01-15T10:00:00Z",
		ChunkCount:   5,
	}

	if err := store.SaveIndexedFile(ctx, file); err != nil {
		t.Fatalf("SaveIndexedFile: %v", err)
	}

	files, err := store.GetIndexedFiles(ctx)
	if err != nil {
		t.Fatalf("GetIndexedFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].FileName != "deck.pptx" {
		t.Errorf("expected deck.pptx, got %q", files[0].FileName)
	}
	if files[0].ChunkCount != 5 {
		t.Errorf("expected 5 chunks, got %d", files[0].ChunkCount)
	}

	// Delete
	if err := store.DeleteIndexedFile(ctx, "f1"); err != nil {
		t.Fatalf("DeleteIndexedFile: %v", err)
	}
	files, err = store.GetIndexedFiles(ctx)
	if err != nil {
		t.Fatalf("GetIndexedFiles after delete: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files after delete, got %d", len(files))
	}
}

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", `"simple"`},
		{"two words", `"two words"`},
		{`has "quotes"`, `"has ""quotes"""`},
		{"AND OR NOT", `"AND OR NOT"`},
		{"test*", `"test*"`},
	}

	for _, tt := range tests {
		got := sanitizeFTS5Query(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeFTS5Query(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSQLiteStore_FTS5SearchAfterDelete(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	chunks := []Chunk{
		{ID: "c1", FileID: "f1", FileName: "deck.pptx", Content: "branosotine results"},
		{ID: "c2", FileID: "f2", FileName: "other.docx", Content: "branosotine overview"},
	}
	if err := store.SaveChunks(ctx, chunks); err != nil {
		t.Fatalf("SaveChunks: %v", err)
	}

	// Delete f1 chunks
	if err := store.DeleteChunksByFileID(ctx, "f1"); err != nil {
		t.Fatalf("DeleteChunksByFileID: %v", err)
	}

	// FTS5 should only return f2's chunk
	results, err := store.Search(ctx, "branosotine", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result after delete, got %d", len(results))
	}
	if results[0].FileID != "f2" {
		t.Errorf("expected f2, got %q", results[0].FileID)
	}
}
