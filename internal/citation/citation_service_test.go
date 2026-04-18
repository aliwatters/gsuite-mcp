package citation

import (
	"context"
	"strings"
	"testing"
)

func TestChunkText_SingleChunk(t *testing.T) {
	text := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	chunks := chunkText("f1", "doc.txt", text)
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	for _, c := range chunks {
		if c.FileID != "f1" {
			t.Errorf("expected file_id=f1, got %q", c.FileID)
		}
		if c.FileName != "doc.txt" {
			t.Errorf("expected file_name=doc.txt, got %q", c.FileName)
		}
		if c.ID == "" {
			t.Error("chunk ID should not be empty")
		}
	}
}

func TestChunkText_MultipleChunks(t *testing.T) {
	// Build text that exceeds maxChunkSize (3000 chars)
	var sb strings.Builder
	for i := range 10 {
		_ = i
		sb.WriteString(strings.Repeat("word ", 400))
		sb.WriteString("\n\n")
	}
	chunks := chunkText("f2", "large.txt", sb.String())
	if len(chunks) < 2 {
		t.Errorf("expected multiple chunks for large text, got %d", len(chunks))
	}
}

func TestChunkText_EmptyInput(t *testing.T) {
	chunks := chunkText("f1", "empty.txt", "")
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkText_WhitespaceOnly(t *testing.T) {
	chunks := chunkText("f1", "ws.txt", "   \n\n   \n\n   ")
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for whitespace-only text, got %d", len(chunks))
	}
}

func TestChunkText_LocationTracking(t *testing.T) {
	text := "Alpha paragraph.\n\nBeta paragraph.\n\nGamma paragraph."
	chunks := chunkText("f1", "doc.txt", text)
	if len(chunks) == 0 {
		t.Fatal("expected chunks")
	}
	// CharStart should be 0 for first chunk
	if chunks[0].Location.CharStart != 0 {
		t.Errorf("expected CharStart=0 for first chunk, got %d", chunks[0].Location.CharStart)
	}
}

func TestChunkID_Deterministic(t *testing.T) {
	id1 := chunkID("file1", 0)
	id2 := chunkID("file1", 0)
	if id1 != id2 {
		t.Errorf("chunkID is not deterministic: %q != %q", id1, id2)
	}

	// Different inputs produce different IDs
	id3 := chunkID("file1", 1)
	if id1 == id3 {
		t.Error("different offsets should produce different chunk IDs")
	}
	id4 := chunkID("file2", 0)
	if id1 == id4 {
		t.Error("different file IDs should produce different chunk IDs")
	}
}

func TestChunkID_Format(t *testing.T) {
	id := chunkID("file1", 0)
	// Should be a hex string of length 16 (8 bytes)
	if len(id) != 16 {
		t.Errorf("expected chunk ID length 16, got %d: %q", len(id), id)
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("chunk ID contains non-hex character %q", c)
		}
	}
}

func TestFormatCitation_AllFields(t *testing.T) {
	svc := &RealCitationService{}
	chunk := Chunk{
		FileName: "report.pdf",
		Content:  "This is a test content string.",
		Location: Location{PageNumber: 3, SectionHeading: "Results"},
	}
	citation := svc.FormatCitation(context.Background(), chunk)
	if !strings.HasPrefix(citation, "Source:") {
		t.Errorf("expected citation to start with Source:, got %q", citation)
	}
	if !strings.Contains(citation, "report.pdf") {
		t.Errorf("expected citation to contain file name, got %q", citation)
	}
	if !strings.Contains(citation, "Slide 3") {
		t.Errorf("expected citation to contain Slide 3, got %q", citation)
	}
	if !strings.Contains(citation, "Results") {
		t.Errorf("expected citation to contain section heading, got %q", citation)
	}
}

func TestFormatCitation_Truncated(t *testing.T) {
	svc := &RealCitationService{}
	chunk := Chunk{
		FileName: "doc.txt",
		Content:  strings.Repeat("word ", 30), // > 80 chars
	}
	citation := svc.FormatCitation(context.Background(), chunk)
	if !strings.Contains(citation, "...") {
		t.Errorf("expected truncated content to have ellipsis, got %q", citation)
	}
}

func TestFormatCitation_NoPageNumber(t *testing.T) {
	svc := &RealCitationService{}
	chunk := Chunk{
		FileName: "doc.txt",
		Content:  "short",
		Location: Location{PageNumber: 0},
	}
	citation := svc.FormatCitation(context.Background(), chunk)
	if strings.Contains(citation, "Slide") {
		t.Errorf("expected no slide reference when PageNumber is 0, got %q", citation)
	}
}
