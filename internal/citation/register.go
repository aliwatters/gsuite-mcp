package citation

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// experimentalPrefix is prepended to all citation tool descriptions to clearly mark
// the feature as experimental. It warns users that the API may change without notice.
const experimentalPrefix = "[EXPERIMENTAL] "

// RegisterTools registers all citation tools with the MCP server.
// Tool descriptions are kept concise to minimize token usage per tool listing.
func RegisterTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("citation_create_index",
		mcp.WithDescription(experimentalPrefix+"Create a citation index Sheet in a Drive folder for document tracing."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("folder_id", mcp.Description("Drive folder ID for source docs")),
		common.WithAccountParam(),
	), HandleCitationCreateIndex)

	s.AddTool(mcp.NewTool("citation_add_documents",
		mcp.WithDescription(experimentalPrefix+"Extract and chunk documents into a citation index."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("file_ids", mcp.Required(), mcp.Description("Comma-separated Drive file IDs")),
		common.WithAccountParam(),
	), HandleCitationAddDocuments)

	s.AddTool(mcp.NewTool("citation_save_concepts",
		mcp.WithDescription(experimentalPrefix+"Save extracted concepts for chunks (from LLM analysis)."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("mappings", mcp.Required(), mcp.Description("JSON: [{concept, chunk_ids}]")),
		common.WithAccountParam(),
	), HandleCitationSaveConcepts)

	s.AddTool(mcp.NewTool("citation_save_summary",
		mcp.WithDescription(experimentalPrefix+"Save a hierarchical summary (level 0=chunk, 1=doc, 2=corpus)."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithNumber("level", mcp.Required(), mcp.Description("0=chunk, 1=doc, 2=corpus")),
		mcp.WithString("parent_id", mcp.Required(), mcp.Description("Chunk/file/index ID being summarized")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Summary text")),
		common.WithAccountParam(),
	), HandleCitationSaveSummary)

	s.AddTool(mcp.NewTool("citation_list_indexes",
		mcp.WithDescription(experimentalPrefix+"List configured citation indexes."),
		common.WithAccountParam(),
	), HandleCitationListIndexes)

	s.AddTool(mcp.NewTool("citation_get_overview",
		mcp.WithDescription(experimentalPrefix+"Get index summary: doc count, concepts, corpus summary."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		common.WithAccountParam(),
	), HandleCitationGetOverview)

	s.AddTool(mcp.NewTool("citation_lookup",
		mcp.WithDescription(experimentalPrefix+"Search chunks by keyword (FTS5). Returns snippets, not full text."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 10)")),
		common.WithAccountParam(),
	), HandleCitationLookup)

	s.AddTool(mcp.NewTool("citation_get_chunks",
		mcp.WithDescription(experimentalPrefix+"Get full chunk text + location by IDs."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("chunk_ids", mcp.Required(), mcp.Description("Comma-separated chunk IDs")),
		common.WithAccountParam(),
	), HandleCitationGetChunks)

	s.AddTool(mcp.NewTool("citation_verify_claim",
		mcp.WithDescription(experimentalPrefix+"Find chunks supporting a claim. Returns candidates with citations."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("claim", mcp.Required(), mcp.Description("Claim text to verify")),
		mcp.WithNumber("limit", mcp.Description("Max candidates (default 5)")),
		common.WithAccountParam(),
	), HandleCitationVerifyClaim)

	s.AddTool(mcp.NewTool("citation_format_citation",
		mcp.WithDescription(experimentalPrefix+"Format a chunk as a human-readable citation string."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		mcp.WithString("chunk_id", mcp.Required(), mcp.Description("Chunk ID")),
		common.WithAccountParam(),
	), HandleCitationFormatCitation)

	s.AddTool(mcp.NewTool("citation_refresh",
		mcp.WithDescription(experimentalPrefix+"Refresh index: detect updated/removed/renamed files and re-index."),
		mcp.WithString("index_id", mcp.Required(), mcp.Description("Index name")),
		common.WithAccountParam(),
	), HandleCitationRefresh)
}
