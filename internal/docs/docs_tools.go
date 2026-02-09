package docs

import (
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"google.golang.org/api/docs/v1"
)

// extractDocumentText extracts plain text from a Google Docs document structure.
func extractDocumentText(doc *docs.Document) string {
	if doc == nil || doc.Body == nil || doc.Body.Content == nil {
		return ""
	}

	var builder strings.Builder

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil && elem.Paragraph.Elements != nil {
			for _, paraElem := range elem.Paragraph.Elements {
				if paraElem.TextRun != nil {
					builder.WriteString(paraElem.TextRun.Content)
				}
			}
		}
		if elem.Table != nil {
			// Extract text from table cells
			for _, row := range elem.Table.TableRows {
				for _, cell := range row.TableCells {
					for _, cellContent := range cell.Content {
						if cellContent.Paragraph != nil && cellContent.Paragraph.Elements != nil {
							for _, paraElem := range cellContent.Paragraph.Elements {
								if paraElem.TextRun != nil {
									builder.WriteString(paraElem.TextRun.Content)
								}
							}
						}
					}
				}
			}
		}
	}

	return builder.String()
}

// parseColor parses a hex color string (e.g., "#FF0000" or "FF0000") into RGB values (0-1 range).
func parseColor(colorStr string) (float64, float64, float64, error) {
	// Remove leading # if present
	colorStr = strings.TrimPrefix(colorStr, "#")
	if len(colorStr) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid color format: expected 6 hex characters, got %d", len(colorStr))
	}

	r, err := parseHexByte(colorStr[0:2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red component: %w", err)
	}
	g, err := parseHexByte(colorStr[2:4])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green component: %w", err)
	}
	b, err := parseHexByte(colorStr[4:6])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue component: %w", err)
	}

	return float64(r) / 255.0, float64(g) / 255.0, float64(b) / 255.0, nil
}

// parseHexByte parses a 2-character hex string into a byte value.
func parseHexByte(s string) (byte, error) {
	var val byte
	for i := 0; i < 2; i++ {
		val <<= 4
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			val |= c - '0'
		case c >= 'A' && c <= 'F':
			val |= c - 'A' + 10
		case c >= 'a' && c <= 'f':
			val |= c - 'a' + 10
		default:
			return 0, fmt.Errorf("invalid hex character: %c", c)
		}
	}
	return val, nil
}

// === Handle functions - generated via WrapHandler ===

var (
	HandleDocsCreate            = common.WrapHandler[DocsService](TestableDocsCreate)
	HandleDocsGet               = common.WrapHandler[DocsService](TestableDocsGet)
	HandleDocsGetMetadata       = common.WrapHandler[DocsService](TestableDocsGetMetadata)
	HandleDocsAppendText        = common.WrapHandler[DocsService](TestableDocsAppendText)
	HandleDocsInsertText        = common.WrapHandler[DocsService](TestableDocsInsertText)
	HandleDocsReplaceText       = common.WrapHandler[DocsService](TestableDocsReplaceText)
	HandleDocsDeleteText        = common.WrapHandler[DocsService](TestableDocsDeleteText)
	HandleDocsInsertTable       = common.WrapHandler[DocsService](TestableDocsInsertTable)
	HandleDocsInsertLink        = common.WrapHandler[DocsService](TestableDocsInsertLink)
	HandleDocsBatchUpdate       = common.WrapHandler[DocsService](TestableDocsBatchUpdate)
	HandleDocsFormatText        = common.WrapHandler[DocsService](TestableDocsFormatText)
	HandleDocsClearFormatting   = common.WrapHandler[DocsService](TestableDocsClearFormatting)
	HandleDocsSetParagraphStyle = common.WrapHandler[DocsService](TestableDocsSetParagraphStyle)
	HandleDocsCreateList        = common.WrapHandler[DocsService](TestableDocsCreateList)
	HandleDocsRemoveList        = common.WrapHandler[DocsService](TestableDocsRemoveList)
	HandleDocsInsertPageBreak   = common.WrapHandler[DocsService](TestableDocsInsertPageBreak)
	HandleDocsInsertImage       = common.WrapHandler[DocsService](TestableDocsInsertImage)
	HandleDocsCreateHeader      = common.WrapHandler[DocsService](TestableDocsCreateHeader)
	HandleDocsCreateFooter      = common.WrapHandler[DocsService](TestableDocsCreateFooter)
)
