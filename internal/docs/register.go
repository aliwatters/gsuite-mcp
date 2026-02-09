package docs

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all Docs tools with the MCP server.
func RegisterTools(s *server.MCPServer) {
	// === Docs Core (Phase 1) ===

	// docs_create - Create a new Google Doc
	s.AddTool(mcp.NewTool("docs_create",
		mcp.WithDescription("Create a new Google Doc with the given title."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Document title")),
		common.WithAccountParam(),
	), HandleDocsCreate)

	// docs_get - Get document content
	s.AddTool(mcp.NewTool("docs_get",
		mcp.WithDescription("Get the full content of a Google Doc as plain text."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		common.WithAccountParam(),
	), HandleDocsGet)

	// docs_get_metadata - Get document metadata
	s.AddTool(mcp.NewTool("docs_get_metadata",
		mcp.WithDescription("Get document metadata (title, revision, word count) without full content."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		common.WithAccountParam(),
	), HandleDocsGetMetadata)

	// docs_append_text - Append text to document
	s.AddTool(mcp.NewTool("docs_append_text",
		mcp.WithDescription("Append text to the end of a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to append to the document")),
		common.WithAccountParam(),
	), HandleDocsAppendText)

	// docs_insert_text - Insert text at specific location
	s.AddTool(mcp.NewTool("docs_insert_text",
		mcp.WithDescription("Insert text at a specific position in a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Text to insert")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("1-based position in document where text will be inserted")),
		common.WithAccountParam(),
	), HandleDocsInsertText)

	// === Docs Extended (Phase 2) ===

	// docs_replace_text - Find and replace text in document
	s.AddTool(mcp.NewTool("docs_replace_text",
		mcp.WithDescription("Find and replace all occurrences of text in a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("find_text", mcp.Required(), mcp.Description("Text to find")),
		mcp.WithString("replace_text", mcp.Description("Text to replace with (empty to delete matches)")),
		mcp.WithBoolean("match_case", mcp.Description("Case-sensitive matching (default: false)")),
		common.WithAccountParam(),
	), HandleDocsReplaceText)

	// docs_delete_text - Delete text at specified range
	s.AddTool(mcp.NewTool("docs_delete_text",
		mcp.WithDescription("Delete text at a specified range in a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		common.WithAccountParam(),
	), HandleDocsDeleteText)

	// docs_insert_table - Insert a table at specified location
	s.AddTool(mcp.NewTool("docs_insert_table",
		mcp.WithDescription("Insert a table at a specified position in a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("rows", mcp.Required(), mcp.Description("Number of rows")),
		mcp.WithNumber("columns", mcp.Required(), mcp.Description("Number of columns")),
		mcp.WithNumber("index", mcp.Description("1-based position for table insertion (default: 1, beginning of document)")),
		common.WithAccountParam(),
	), HandleDocsInsertTable)

	// docs_insert_link - Insert a hyperlink
	s.AddTool(mcp.NewTool("docs_insert_link",
		mcp.WithDescription("Insert a hyperlink at a specified position in a Google Doc."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("text", mcp.Required(), mcp.Description("Display text for the link")),
		mcp.WithString("url", mcp.Required(), mcp.Description("URL to link to")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("1-based position where link will be inserted")),
		common.WithAccountParam(),
	), HandleDocsInsertLink)

	// docs_batch_update - Raw batchUpdate for power users
	s.AddTool(mcp.NewTool("docs_batch_update",
		mcp.WithDescription("Execute raw batchUpdate requests on a Google Doc. Power user escape hatch for advanced operations."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("requests", mcp.Required(), mcp.Description("JSON array of batch update requests (see Google Docs API docs)")),
		common.WithAccountParam(),
	), HandleDocsBatchUpdate)

	// === Docs Extended (Phase 3) - Advanced Formatting ===

	// docs_format_text - Apply text formatting
	s.AddTool(mcp.NewTool("docs_format_text",
		mcp.WithDescription("Apply formatting (bold, italic, underline, font, size, color) to text range."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		mcp.WithBoolean("bold", mcp.Description("Apply bold formatting")),
		mcp.WithBoolean("italic", mcp.Description("Apply italic formatting")),
		mcp.WithBoolean("underline", mcp.Description("Apply underline formatting")),
		mcp.WithBoolean("strikethrough", mcp.Description("Apply strikethrough formatting")),
		mcp.WithBoolean("small_caps", mcp.Description("Apply small caps formatting")),
		mcp.WithString("font_family", mcp.Description("Font family name (e.g., 'Arial', 'Times New Roman')")),
		mcp.WithNumber("font_weight", mcp.Description("Font weight (100-900, e.g., 400=normal, 700=bold)")),
		mcp.WithNumber("font_size", mcp.Description("Font size in points (e.g., 12, 14, 16)")),
		mcp.WithString("foreground_color", mcp.Description("Text color as hex (e.g., '#FF0000' or 'FF0000')")),
		mcp.WithString("background_color", mcp.Description("Background/highlight color as hex")),
		mcp.WithString("baseline_offset", mcp.Description("Vertical offset: NONE, SUPERSCRIPT, or SUBSCRIPT")),
		common.WithAccountParam(),
	), HandleDocsFormatText)

	// docs_clear_formatting - Remove text formatting
	s.AddTool(mcp.NewTool("docs_clear_formatting",
		mcp.WithDescription("Remove formatting from text range."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		common.WithAccountParam(),
	), HandleDocsClearFormatting)

	// docs_set_paragraph_style - Set paragraph formatting
	s.AddTool(mcp.NewTool("docs_set_paragraph_style",
		mcp.WithDescription("Set alignment, spacing, indentation, and heading styles for paragraphs."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		mcp.WithString("alignment", mcp.Description("Text alignment: START, CENTER, END, JUSTIFIED")),
		mcp.WithString("named_style_type", mcp.Description("Style type: NORMAL_TEXT, TITLE, SUBTITLE, HEADING_1 through HEADING_6")),
		mcp.WithNumber("line_spacing", mcp.Description("Line spacing percentage (100=single, 150=1.5, 200=double)")),
		mcp.WithNumber("indent_start", mcp.Description("Start indentation in points")),
		mcp.WithNumber("indent_end", mcp.Description("End indentation in points")),
		mcp.WithNumber("indent_first_line", mcp.Description("First line indentation in points")),
		mcp.WithNumber("space_above", mcp.Description("Space above paragraph in points")),
		mcp.WithNumber("space_below", mcp.Description("Space below paragraph in points")),
		common.WithAccountParam(),
	), HandleDocsSetParagraphStyle)

	// docs_create_list - Create bulleted or numbered list
	s.AddTool(mcp.NewTool("docs_create_list",
		mcp.WithDescription("Convert text to bulleted/numbered list."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		mcp.WithString("bullet_preset", mcp.Description("Bullet preset (default: BULLET_DISC_CIRCLE_SQUARE). Use NUMBERED_DECIMAL_ALPHA_ROMAN for numbered lists.")),
		common.WithAccountParam(),
	), HandleDocsCreateList)

	// docs_remove_list - Remove list formatting
	s.AddTool(mcp.NewTool("docs_remove_list",
		mcp.WithDescription("Remove bullets/numbering from paragraphs."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("start_index", mcp.Required(), mcp.Description("1-based start position (inclusive)")),
		mcp.WithNumber("end_index", mcp.Required(), mcp.Description("1-based end position (exclusive)")),
		common.WithAccountParam(),
	), HandleDocsRemoveList)

	// docs_insert_page_break - Insert page break
	s.AddTool(mcp.NewTool("docs_insert_page_break",
		mcp.WithDescription("Insert page break at position."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("1-based position in document")),
		common.WithAccountParam(),
	), HandleDocsInsertPageBreak)

	// docs_insert_image - Insert inline image from URL
	s.AddTool(mcp.NewTool("docs_insert_image",
		mcp.WithDescription("Insert image from URL."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("uri", mcp.Required(), mcp.Description("Public URL of the image (PNG, JPEG, or GIF, max 50MB)")),
		mcp.WithNumber("index", mcp.Required(), mcp.Description("1-based position in document")),
		mcp.WithNumber("width", mcp.Description("Image width in points (aspect ratio preserved if only one dimension set)")),
		mcp.WithNumber("height", mcp.Description("Image height in points")),
		common.WithAccountParam(),
	), HandleDocsInsertImage)

	// docs_create_header - Create document header
	s.AddTool(mcp.NewTool("docs_create_header",
		mcp.WithDescription("Add/update document header."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("content", mcp.Description("Optional text content for the header")),
		common.WithAccountParam(),
	), HandleDocsCreateHeader)

	// docs_create_footer - Create document footer
	s.AddTool(mcp.NewTool("docs_create_footer",
		mcp.WithDescription("Add/update document footer."),
		mcp.WithString("document_id", mcp.Required(), mcp.Description("Document ID or full Google Docs URL")),
		mcp.WithString("content", mcp.Description("Optional text content for the footer")),
		common.WithAccountParam(),
	), HandleDocsCreateFooter)
}
