package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/api/docs/v1"
)

// extractDocumentMarkdown converts a Google Docs document structure to clean markdown.
// It handles headings, bold, italic, links, lists, and tables.
func extractDocumentMarkdown(doc *docs.Document) string {
	if doc == nil || doc.Body == nil || doc.Body.Content == nil {
		return ""
	}

	var builder strings.Builder

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			renderParagraph(&builder, elem.Paragraph, doc.Lists)
		}
		if elem.Table != nil {
			renderTable(&builder, elem.Table)
		}
	}

	return builder.String()
}

// renderParagraph converts a docs.Paragraph to markdown.
func renderParagraph(b *strings.Builder, p *docs.Paragraph, lists map[string]docs.List) {
	if p == nil {
		return
	}

	// Check for heading style
	headingPrefix := ""
	if p.ParagraphStyle != nil {
		switch p.ParagraphStyle.NamedStyleType {
		case "HEADING_1":
			headingPrefix = "# "
		case "HEADING_2":
			headingPrefix = "## "
		case "HEADING_3":
			headingPrefix = "### "
		case "HEADING_4":
			headingPrefix = "#### "
		case "HEADING_5":
			headingPrefix = "##### "
		case "HEADING_6":
			headingPrefix = "###### "
		case "TITLE":
			headingPrefix = "# "
		case "SUBTITLE":
			headingPrefix = "## "
		}
	}

	// Check for list membership
	listPrefix := ""
	if p.Bullet != nil {
		nestingLevel := p.Bullet.NestingLevel
		indent := strings.Repeat("  ", int(nestingLevel))

		// Determine bullet type
		isOrdered := false
		if p.Bullet.ListId != "" && lists != nil {
			if list, ok := lists[p.Bullet.ListId]; ok && list.ListProperties != nil {
				if int(nestingLevel) < len(list.ListProperties.NestingLevels) {
					nl := list.ListProperties.NestingLevels[nestingLevel]
					if nl.GlyphType != "" && nl.GlyphType != "GLYPH_TYPE_UNSPECIFIED" {
						isOrdered = true
					}
				}
			}
		}

		if isOrdered {
			listPrefix = indent + "1. "
		} else {
			listPrefix = indent + "- "
		}
	}

	// Build paragraph content from elements
	var content strings.Builder
	for _, elem := range p.Elements {
		if elem.TextRun != nil {
			renderTextRun(&content, elem.TextRun)
		}
	}

	text := content.String()

	// Don't emit blank heading lines or list items
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		b.WriteString("\n")
		return
	}

	if headingPrefix != "" {
		// Headings: strip trailing newline from content, then add double newline
		b.WriteString(headingPrefix)
		b.WriteString(trimmed)
		b.WriteString("\n\n")
	} else if listPrefix != "" {
		b.WriteString(listPrefix)
		b.WriteString(trimmed)
		b.WriteString("\n")
	} else {
		b.WriteString(text)
	}
}

// renderTextRun converts a docs.TextRun to markdown with inline formatting.
func renderTextRun(b *strings.Builder, tr *docs.TextRun) {
	if tr == nil || tr.Content == "" {
		return
	}

	text := tr.Content
	style := tr.TextStyle

	// If no style, write plain text
	if style == nil {
		b.WriteString(text)
		return
	}

	// Check for link
	hasLink := style.Link != nil && style.Link.Url != ""

	// Determine inline formatting
	isBold := style.Bold
	isItalic := style.Italic
	isStrikethrough := style.Strikethrough
	isCode := style.WeightedFontFamily != nil &&
		isMonospaceFont(style.WeightedFontFamily.FontFamily)

	// Trailing newlines should not be wrapped in formatting
	trailingNewline := ""
	if strings.HasSuffix(text, "\n") {
		text = strings.TrimRight(text, "\n")
		trailingNewline = "\n"
	}

	if text == "" {
		b.WriteString(trailingNewline)
		return
	}

	if isCode {
		b.WriteString("`")
		b.WriteString(text)
		b.WriteString("`")
	} else {
		// Apply formatting wrappers (order: strikethrough, bold, italic)
		prefix := ""
		suffix := ""
		if isStrikethrough {
			prefix += "~~"
			suffix = "~~" + suffix
		}
		if isBold && isItalic {
			prefix += "***"
			suffix = "***" + suffix
		} else if isBold {
			prefix += "**"
			suffix = "**" + suffix
		} else if isItalic {
			prefix += "*"
			suffix = "*" + suffix
		}

		if hasLink {
			b.WriteString("[")
			b.WriteString(prefix)
			b.WriteString(text)
			b.WriteString(suffix)
			b.WriteString("](")
			b.WriteString(style.Link.Url)
			b.WriteString(")")
		} else {
			b.WriteString(prefix)
			b.WriteString(text)
			b.WriteString(suffix)
		}
	}

	b.WriteString(trailingNewline)
}

// renderTable converts a docs.Table to a markdown table.
func renderTable(b *strings.Builder, table *docs.Table) {
	if table == nil || len(table.TableRows) == 0 {
		return
	}

	for rowIdx, row := range table.TableRows {
		b.WriteString("|")
		for _, cell := range row.TableCells {
			b.WriteString(" ")
			b.WriteString(extractCellText(cell))
			b.WriteString(" |")
		}
		b.WriteString("\n")

		// After header row, add separator
		if rowIdx == 0 {
			b.WriteString("|")
			for range row.TableCells {
				b.WriteString("---|")
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
}

// extractCellText extracts plain text from a table cell, trimming whitespace.
func extractCellText(cell *docs.TableCell) string {
	if cell == nil || cell.Content == nil {
		return ""
	}

	var sb strings.Builder
	for _, elem := range cell.Content {
		if elem.Paragraph != nil && elem.Paragraph.Elements != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(pe.TextRun.Content)
				}
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

// isMonospaceFont checks if a font name indicates a monospace/code font.
func isMonospaceFont(fontFamily string) bool {
	lower := strings.ToLower(fontFamily)
	monoFonts := []string{
		"courier", "consolas", "monaco", "menlo",
		"source code", "fira code", "jetbrains mono",
		"roboto mono", "ubuntu mono", "inconsolata",
	}
	for _, f := range monoFonts {
		if strings.Contains(lower, f) {
			return true
		}
	}
	return false
}

// TestableDocsGetAsMarkdown is the testable version of HandleDocsGetAsMarkdown.
func TestableDocsGetAsMarkdown(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	docID, errResult := extractRequiredDocID(request)
	if errResult != nil {
		return errResult, nil
	}

	doc, err := srv.GetDocument(ctx, docID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Docs API error: %v", err)), nil
	}

	markdown := extractDocumentMarkdown(doc)

	result := map[string]any{
		"document_id": doc.DocumentId,
		"title":       doc.Title,
		"markdown":    markdown,
		"url":         fmt.Sprintf(docsEditURLFormat, doc.DocumentId),
	}

	return common.MarshalToolResult(result)
}
