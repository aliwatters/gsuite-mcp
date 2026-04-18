package citation

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
)

// buildPptxZip creates a minimal .pptx zip in memory with the given slide texts.
// Each slide is written as ppt/slides/slideN.xml with the text in <a:t> elements.
func buildPptxZip(t *testing.T, slides []string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for i, text := range slides {
		name := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("creating zip entry %s: %v", name, err)
		}
		xml := fmt.Sprintf(`<?xml version="1.0"?>
<root xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
<a:t>%s</a:t>
</root>`, text)
		f.Write([]byte(xml))
	}
	w.Close()
	return buf.Bytes()
}

func newDriveFile(id, name, mimeType string) *drive.File {
	return &drive.File{Id: id, Name: name, MimeType: mimeType}
}

func TestChunkPptxBytes_Basic(t *testing.T) {
	slides := []string{"Revenue data", "Market share analysis", ""}
	data := buildPptxZip(t, slides)

	svc := &RealCitationService{}
	file := newDriveFile("f1", "deck.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	chunks, err := svc.chunkPptxBytes(file, data)
	if err != nil {
		t.Fatalf("chunkPptxBytes: %v", err)
	}
	// Slide 3 is empty so only 2 chunks
	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks (non-empty slides), got %d", len(chunks))
	}
	if chunks[0].Location.PageNumber != 1 {
		t.Errorf("expected slide 1 page number, got %d", chunks[0].Location.PageNumber)
	}
	if chunks[1].Location.PageNumber != 2 {
		t.Errorf("expected slide 2 page number, got %d", chunks[1].Location.PageNumber)
	}
}

func TestChunkPptxBytes_SlideOrdering(t *testing.T) {
	// Deliberately add slides in reverse order in the zip — should be sorted
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, n := range []int{3, 1, 2} {
		f, _ := w.Create(fmt.Sprintf("ppt/slides/slide%d.xml", n))
		f.Write([]byte(fmt.Sprintf(`<root xmlns:a="x"><a:t>slide %d content</a:t></root>`, n)))
	}
	w.Close()

	svc := &RealCitationService{}
	file := newDriveFile("f1", "deck.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	chunks, err := svc.chunkPptxBytes(file, buf.Bytes())
	if err != nil {
		t.Fatalf("chunkPptxBytes: %v", err)
	}
	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}
	// Should be in slide order 1, 2, 3
	for i, c := range chunks {
		if c.Location.PageNumber != i+1 {
			t.Errorf("chunk[%d]: expected page %d, got %d", i, i+1, c.Location.PageNumber)
		}
	}
}

func TestChunkPptxBytes_InvalidZip(t *testing.T) {
	svc := &RealCitationService{}
	file := newDriveFile("f1", "bad.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	_, err := svc.chunkPptxBytes(file, []byte("not a zip"))
	if err == nil {
		t.Error("expected error for invalid zip data")
	}
}

func TestChunkPptxBytes_NonSlideEntries(t *testing.T) {
	// Zip with non-slide entries should be ignored
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	// Valid slide
	f, _ := w.Create("ppt/slides/slide1.xml")
	f.Write([]byte(`<root xmlns:a="x"><a:t>real slide</a:t></root>`))
	// Non-slide entries
	g, _ := w.Create("ppt/slides/layout1.xml")
	g.Write([]byte(`<root xmlns:a="x"><a:t>layout text</a:t></root>`))
	h, _ := w.Create("docProps/app.xml")
	h.Write([]byte(`<root><text>app</text></root>`))
	w.Close()

	svc := &RealCitationService{}
	file := newDriveFile("f1", "deck.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	chunks, err := svc.chunkPptxBytes(file, buf.Bytes())
	if err != nil {
		t.Fatalf("chunkPptxBytes: %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk (only real slide), got %d", len(chunks))
	}
	if !strings.Contains(chunks[0].Content, "real slide") {
		t.Errorf("expected 'real slide' in chunk content, got %q", chunks[0].Content)
	}
}

func TestExtractTextFromXML_MultipleElements(t *testing.T) {
	// Build a zip with multiple <a:t> elements
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("ppt/slides/slide1.xml")
	f.Write([]byte(`<root xmlns:a="x"><a:t>Hello</a:t><a:t> world</a:t><other>ignored</other></root>`))
	w.Close()

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("creating zip reader: %v", err)
	}
	text, err := extractTextFromXML(r.File[0])
	if err != nil {
		t.Fatalf("extractTextFromXML: %v", err)
	}
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "world") {
		t.Errorf("expected text to contain 'Hello' and 'world', got %q", text)
	}
}

func TestExtractTextFromXML_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("ppt/slides/slide1.xml")
	f.Write([]byte(`<root xmlns:a="x"></root>`))
	w.Close()

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("creating zip reader: %v", err)
	}
	text, err := extractTextFromXML(r.File[0])
	if err != nil {
		t.Fatalf("extractTextFromXML: %v", err)
	}
	if text != "" {
		t.Errorf("expected empty text, got %q", text)
	}
}

func TestExtractTextFromXML_NestedText(t *testing.T) {
	// Test with deeply nested <a:t> elements typical in real OOXML
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, _ := w.Create("ppt/slides/slide1.xml")
	f.Write([]byte(`<spTree xmlns:a="x">
		<sp><txBody><a:bodyPr/><a:lstStyle/>
			<a:p><a:r><a:t>Title Text</a:t></a:r></a:p>
			<a:p><a:r><a:t>Body Text</a:t></a:r></a:p>
		</txBody></sp>
	</spTree>`))
	w.Close()

	r, _ := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	text, err := extractTextFromXML(r.File[0])
	if err != nil {
		t.Fatalf("extractTextFromXML: %v", err)
	}
	if !strings.Contains(text, "Title Text") {
		t.Errorf("expected 'Title Text' in %q", text)
	}
	if !strings.Contains(text, "Body Text") {
		t.Errorf("expected 'Body Text' in %q", text)
	}
}
