package citation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// mockCitationService is an in-memory CitationService for testing.
type mockCitationService struct {
	indexes   map[string]*IndexInfo
	chunks    map[string]map[string]Chunk // indexID → chunkID → chunk
	concepts  map[string][]ConceptMapping // indexID → mappings
	summaries map[string][]LevelSummary   // indexID → summaries
	files     map[string][]IndexedFile    // indexID → files
}

func newMockCitationService() *mockCitationService {
	return &mockCitationService{
		indexes:   make(map[string]*IndexInfo),
		chunks:    make(map[string]map[string]Chunk),
		concepts:  make(map[string][]ConceptMapping),
		summaries: make(map[string][]LevelSummary),
		files:     make(map[string][]IndexedFile),
	}
}

func (m *mockCitationService) CreateIndex(_ context.Context, name, folderID string) (*IndexInfo, error) {
	info := &IndexInfo{IndexID: name, SourceFolderID: folderID}
	m.indexes[name] = info
	return info, nil
}

func (m *mockCitationService) AddDocuments(_ context.Context, indexID string, fileIDs []string) (int, error) {
	if m.chunks[indexID] == nil {
		m.chunks[indexID] = make(map[string]Chunk)
	}
	for _, fid := range fileIDs {
		c := Chunk{ID: fid + "-chunk", FileID: fid, Content: "content for " + fid}
		m.chunks[indexID][c.ID] = c
	}
	return len(fileIDs), nil
}

func (m *mockCitationService) SaveConcepts(_ context.Context, indexID string, mappings []ConceptMapping) error {
	m.concepts[indexID] = append(m.concepts[indexID], mappings...)
	return nil
}

func (m *mockCitationService) SaveSummary(_ context.Context, indexID string, summary LevelSummary) error {
	m.summaries[indexID] = append(m.summaries[indexID], summary)
	return nil
}

func (m *mockCitationService) ListIndexes(_ context.Context) ([]IndexInfo, error) {
	var result []IndexInfo
	for _, info := range m.indexes {
		result = append(result, *info)
	}
	return result, nil
}

func (m *mockCitationService) GetOverview(_ context.Context, indexID string) (map[string]any, error) {
	return map[string]any{
		"index_id":  indexID,
		"doc_count": len(m.files[indexID]),
	}, nil
}

func (m *mockCitationService) Lookup(_ context.Context, indexID, query string, limit int) ([]Chunk, error) {
	var result []Chunk
	for _, c := range m.chunks[indexID] {
		result = append(result, c)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (m *mockCitationService) GetChunks(_ context.Context, indexID string, chunkIDs []string) ([]Chunk, error) {
	var result []Chunk
	for _, id := range chunkIDs {
		if c, ok := m.chunks[indexID][id]; ok {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCitationService) VerifyClaim(ctx context.Context, indexID, claim string, limit int) ([]Chunk, error) {
	return m.Lookup(ctx, indexID, claim, limit)
}

func (m *mockCitationService) FormatCitation(_ context.Context, chunk Chunk) string {
	return "Source: " + chunk.FileName
}

func (m *mockCitationService) RefreshIndex(_ context.Context, _ string) (*RefreshResult, error) {
	return &RefreshResult{}, nil
}

// makeToolRequest creates a CallToolRequest with the given arguments.
func makeToolRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// extractText returns the text content from an MCP tool result.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// newCitationTestFixtures returns deps backed by the mock service.
func newCitationTestFixtures() (*mockCitationService, *CitationHandlerDeps) {
	mock := newMockCitationService()
	f := common.NewTestFixtures[CitationService](mock)
	return mock, f.Deps
}

func TestTestableCitationCreateIndex(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	// Missing name returns error
	result, err := TestableCitationCreateIndex(ctx, makeToolRequest(map[string]any{}), deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result when name is missing")
	}

	// Valid request creates index
	result, err = TestableCitationCreateIndex(ctx, makeToolRequest(map[string]any{
		"name":      "test-index",
		"folder_id": "folder1",
	}), deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", extractText(t, result))
	}

	var info IndexInfo
	if err := json.Unmarshal([]byte(extractText(t, result)), &info); err != nil {
		t.Fatalf("parsing response: %v", err)
	}
	if info.IndexID != "test-index" {
		t.Errorf("expected index_id=test-index, got %q", info.IndexID)
	}
	_ = mock
}

func TestTestableCitationAddDocuments(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	// Seed an index
	mock.indexes["idx1"] = &IndexInfo{IndexID: "idx1"}
	mock.chunks["idx1"] = make(map[string]Chunk)

	// Missing index_id
	result, err := TestableCitationAddDocuments(ctx, makeToolRequest(map[string]any{
		"file_ids": "f1,f2",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when index_id missing")
	}

	// Missing file_ids
	result, err = TestableCitationAddDocuments(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when file_ids missing")
	}

	// Valid — comma separated
	result, err = TestableCitationAddDocuments(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"file_ids": "f1,f2",
	}), deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", extractText(t, result))
	}
	var resp map[string]any
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if resp["chunks_created"].(float64) != 2 {
		t.Errorf("expected 2 chunks_created, got %v", resp["chunks_created"])
	}
}

func TestTestableCitationSaveConcepts(t *testing.T) {
	_, deps := newCitationTestFixtures()
	ctx := context.Background()

	// Valid request
	mappings, _ := json.Marshal([]ConceptMapping{
		{Concept: "revenue", ChunkIDs: []string{"c1"}},
	})
	result, err := TestableCitationSaveConcepts(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"mappings": string(mappings),
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v %v", err, result)
	}
	var resp map[string]any
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if resp["saved"].(float64) != 1 {
		t.Errorf("expected saved=1, got %v", resp["saved"])
	}

	// Invalid JSON
	result, err = TestableCitationSaveConcepts(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"mappings": "not-json",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error for invalid JSON")
	}
}

func TestTestableCitationSaveSummary(t *testing.T) {
	_, deps := newCitationTestFixtures()
	ctx := context.Background()

	// Missing summary
	result, err := TestableCitationSaveSummary(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when summary missing")
	}

	// Valid
	result, err = TestableCitationSaveSummary(ctx, makeToolRequest(map[string]any{
		"index_id":  "idx1",
		"level":     float64(2),
		"parent_id": "root",
		"summary":   "corpus overview",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTestableCitationListIndexes(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	mock.indexes["a"] = &IndexInfo{IndexID: "a"}
	mock.indexes["b"] = &IndexInfo{IndexID: "b"}

	result, err := TestableCitationListIndexes(ctx, makeToolRequest(map[string]any{}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp []IndexInfo
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 indexes, got %d", len(resp))
	}
}

func TestTestableCitationGetOverview(t *testing.T) {
	_, deps := newCitationTestFixtures()
	ctx := context.Background()

	// Missing index_id
	result, err := TestableCitationGetOverview(ctx, makeToolRequest(map[string]any{}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when index_id missing")
	}

	// Valid
	result, err = TestableCitationGetOverview(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp map[string]any
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if resp["index_id"] != "idx1" {
		t.Errorf("expected index_id=idx1, got %v", resp["index_id"])
	}
}

func TestTestableCitationLookup(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	mock.chunks["idx1"] = map[string]Chunk{
		"c1": {ID: "c1", FileID: "f1", FileName: "doc.pdf", Content: "revenue growth"},
	}

	// Missing query
	result, err := TestableCitationLookup(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when query missing")
	}

	// Valid
	result, err = TestableCitationLookup(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"query":    "revenue",
		"limit":    float64(5),
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp map[string]any
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if resp["count"].(float64) < 1 {
		t.Errorf("expected at least 1 result, got %v", resp["count"])
	}
}

func TestTestableCitationGetChunks(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	mock.chunks["idx1"] = map[string]Chunk{
		"c1": {ID: "c1", Content: "test content"},
	}

	// Missing chunk_ids
	result, err := TestableCitationGetChunks(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error when chunk_ids missing")
	}

	// Valid
	result, err = TestableCitationGetChunks(ctx, makeToolRequest(map[string]any{
		"index_id":  "idx1",
		"chunk_ids": "c1",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	var chunks []Chunk
	json.Unmarshal([]byte(extractText(t, result)), &chunks)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestTestableCitationVerifyClaim(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	mock.chunks["idx1"] = map[string]Chunk{
		"c1": {ID: "c1", FileName: "doc.pdf", Content: "revenue was $1M"},
	}

	result, err := TestableCitationVerifyClaim(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"claim":    "revenue claim",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp map[string]any
	json.Unmarshal([]byte(extractText(t, result)), &resp)
	if resp["claim"] != "revenue claim" {
		t.Errorf("expected claim=revenue claim, got %v", resp["claim"])
	}
}

func TestTestableCitationFormatCitation(t *testing.T) {
	mock, deps := newCitationTestFixtures()
	ctx := context.Background()

	mock.chunks["idx1"] = map[string]Chunk{
		"c1": {ID: "c1", FileName: "report.pdf", Content: "test"},
	}

	// chunk not found
	result, err := TestableCitationFormatCitation(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"chunk_id": "nonexistent",
	}), deps)
	if err != nil || !result.IsError {
		t.Error("expected error for nonexistent chunk")
	}

	// Valid
	result, err = TestableCitationFormatCitation(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
		"chunk_id": "c1",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if text == "" {
		t.Error("expected non-empty citation")
	}
}

func TestTestableCitationRefresh(t *testing.T) {
	_, deps := newCitationTestFixtures()
	ctx := context.Background()

	result, err := TestableCitationRefresh(ctx, makeToolRequest(map[string]any{
		"index_id": "idx1",
	}), deps)
	if err != nil || result.IsError {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseStringList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{`["x","y"]`, []string{"x", "y"}},
		{" a , b ", []string{"a", "b"}},
		{"single", []string{"single"}},
	}
	for _, tt := range tests {
		got := parseStringList(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("parseStringList(%q): got %v, want %v", tt.input, got, tt.expected)
			continue
		}
		for i, v := range got {
			if v != tt.expected[i] {
				t.Errorf("parseStringList(%q)[%d]: got %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("expected hello..., got %q", got)
	}
}
