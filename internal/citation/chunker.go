package citation

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/drive/v3"
	gslides "google.golang.org/api/slides/v1"
)

// DocumentChunker splits a document into searchable Chunk values.
// Each format (Google Slides, PPTX, plain text) has its own strategy.
type DocumentChunker interface {
	// Chunk splits the document into chunks.
	Chunk(file *drive.File, data []byte) ([]Chunk, error)
}

// PptxChunker extracts chunks from an uploaded .pptx file (zip containing XML slides).
type PptxChunker struct{}

// Chunk extracts text from each slide XML in a .pptx zip.
// Empty slides are skipped. Slides are returned in slide-number order.
func (c *PptxChunker) Chunk(file *drive.File, data []byte) ([]Chunk, error) {
	return chunkPptxData(file, data)
}

// PlainTextChunker splits plain text into paragraph-based chunks.
type PlainTextChunker struct{}

// Chunk splits text by paragraph boundaries (~3000 chars per chunk).
func (c *PlainTextChunker) Chunk(file *drive.File, data []byte) ([]Chunk, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, nil
	}
	return chunkText(file.Id, file.Name, text), nil
}

// chunkPptxData extracts text from a .pptx zip and returns per-slide chunks.
// Each slide/N.xml becomes a separate chunk. Slides are sorted by number.
func chunkPptxData(file *drive.File, data []byte) ([]Chunk, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading pptx zip: %w", err)
	}

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
// These are the standard OOXML text run elements used in PowerPoint presentations.
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

// extractSlidesText extracts text from a Google Slides presentation object.
// Called by RealCitationService.chunkSlides after fetching the presentation via API.
func extractSlidesText(fileID, fileName string, presentation *gslides.Presentation) []Chunk {
	var chunks []Chunk
	for i, slide := range presentation.Slides {
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
			ID:       chunkID(fileID, i),
			FileID:   fileID,
			FileName: fileName,
			Content:  content,
			Location: Location{PageNumber: i + 1},
		})
	}
	return chunks
}

// downloadAndChunk downloads a Drive file and dispatches to the appropriate chunker.
// For .pptx files, uses chunkPptxData. For others, uses chunkText.
func downloadAndChunk(data []byte, file *drive.File) ([]Chunk, error) {
	if file.MimeType == "application/vnd.openxmlformats-officedocument.presentationml.presentation" {
		return chunkPptxData(file, data)
	}
	// Fallback: treat as plain text
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, nil
	}
	return chunkText(file.Id, file.Name, text), nil
}

// exportSizeLimit limits Drive export downloads.
// Uses the shared constant from common to prevent divergence.
const exportSizeLimit = common.DriveMaxExportSize

// limitedRead reads at most exportSizeLimit bytes from r.
func limitedRead(r io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(r, exportSizeLimit))
}
