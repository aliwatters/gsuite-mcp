package citation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestableCitationCreateIndex creates a new citation index.
func TestableCitationCreateIndex(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	name := common.ParseStringArg(request.GetArguments(), "name", "")
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}
	folderID := common.ParseStringArg(request.GetArguments(), "folder_id", "")

	info, err := srv.CreateIndex(ctx, name, folderID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(info)
}

// TestableCitationAddDocuments adds documents to an index.
func TestableCitationAddDocuments(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	fileIDsRaw := common.ParseStringArg(request.GetArguments(), "file_ids", "")
	if fileIDsRaw == "" {
		return mcp.NewToolResultError("file_ids is required (comma-separated or JSON array)"), nil
	}

	fileIDs := parseStringList(fileIDsRaw)
	if len(fileIDs) == 0 {
		return mcp.NewToolResultError("file_ids must contain at least one file ID"), nil
	}

	count, err := srv.AddDocuments(ctx, indexID, fileIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{
		"chunks_created":  count,
		"files_processed": len(fileIDs),
	})
}

// TestableCitationSaveConcepts saves LLM-extracted concepts.
func TestableCitationSaveConcepts(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	mappingsJSON := common.ParseStringArg(request.GetArguments(), "mappings", "")
	if mappingsJSON == "" {
		return mcp.NewToolResultError("mappings is required (JSON array of {concept, chunk_ids})"), nil
	}

	var mappings []ConceptMapping
	if err := json.Unmarshal([]byte(mappingsJSON), &mappings); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid mappings JSON: %v", err)), nil
	}

	if err := srv.SaveConcepts(ctx, indexID, mappings); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{"saved": len(mappings)})
}

// TestableCitationSaveSummary saves an LLM-generated summary.
func TestableCitationSaveSummary(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	args := request.GetArguments()
	level := 0
	if v, ok := args["level"].(float64); ok {
		level = int(v)
	}
	parentID := common.ParseStringArg(args, "parent_id", "")
	summary := common.ParseStringArg(args, "summary", "")
	if summary == "" {
		return mcp.NewToolResultError("summary is required"), nil
	}

	if err := srv.SaveSummary(ctx, indexID, LevelSummary{Level: level, ParentID: parentID, Summary: summary}); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(map[string]any{"saved": true})
}

// TestableCitationListIndexes lists known indexes.
func TestableCitationListIndexes(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexes, err := srv.ListIndexes(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(indexes)
}

// TestableCitationGetOverview returns index overview.
func TestableCitationGetOverview(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	overview, err := srv.GetOverview(ctx, indexID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(overview)
}

// TestableCitationLookup searches chunks by keyword/concept.
func TestableCitationLookup(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}
	query := common.ParseStringArg(request.GetArguments(), "query", "")
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	limit := 10
	if v, ok := request.GetArguments()["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}

	// Token efficiency: return compact results
	chunks, err := srv.Lookup(ctx, indexID, query, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	// Return compact representation — omit full content, include snippet
	results := make([]map[string]any, len(chunks))
	for i, c := range chunks {
		snippet := c.Content
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		results[i] = map[string]any{
			"id":       c.ID,
			"file":     c.FileName,
			"snippet":  snippet,
			"location": c.Location,
		}
	}

	return common.MarshalToolResult(map[string]any{
		"count":   len(results),
		"results": results,
	})
}

// TestableCitationGetChunks retrieves full chunk data.
func TestableCitationGetChunks(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	chunkIDsRaw := common.ParseStringArg(request.GetArguments(), "chunk_ids", "")
	if chunkIDsRaw == "" {
		return mcp.NewToolResultError("chunk_ids is required"), nil
	}

	chunkIDs := parseStringList(chunkIDsRaw)
	chunks, err := srv.GetChunks(ctx, indexID, chunkIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(chunks)
}

// TestableCitationVerifyClaim finds chunks supporting a claim.
func TestableCitationVerifyClaim(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}
	claim := common.ParseStringArg(request.GetArguments(), "claim", "")
	if claim == "" {
		return mcp.NewToolResultError("claim is required"), nil
	}

	limit := 5
	if v, ok := request.GetArguments()["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}

	chunks, err := srv.VerifyClaim(ctx, indexID, claim, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	// Token-efficient: compact results with citation formatting
	results := make([]map[string]any, len(chunks))
	for i, c := range chunks {
		results[i] = map[string]any{
			"id":       c.ID,
			"citation": srv.FormatCitation(ctx, c),
			"snippet":  truncate(c.Content, 300),
		}
	}

	return common.MarshalToolResult(map[string]any{
		"claim":      claim,
		"candidates": len(results),
		"results":    results,
	})
}

// TestableCitationFormatCitation formats a chunk as a citation string.
func TestableCitationFormatCitation(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}
	chunkID := common.ParseStringArg(request.GetArguments(), "chunk_id", "")
	if chunkID == "" {
		return mcp.NewToolResultError("chunk_id is required"), nil
	}

	chunks, err := srv.GetChunks(ctx, indexID, []string{chunkID})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}
	if len(chunks) == 0 {
		return mcp.NewToolResultError("chunk not found"), nil
	}

	citation := srv.FormatCitation(ctx, chunks[0])
	return mcp.NewToolResultText(citation), nil
}

// parseStringList parses a comma-separated string or JSON array into a string slice.
func parseStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "[") {
		var list []string
		if err := json.Unmarshal([]byte(raw), &list); err == nil {
			return list
		}
	}
	parts := strings.Split(raw, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// TestableCitationRefresh checks for updated/removed files and re-indexes.
func TestableCitationRefresh(ctx context.Context, request mcp.CallToolRequest, deps *CitationHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveCitationServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	indexID := common.ParseStringArg(request.GetArguments(), "index_id", "")
	if indexID == "" {
		return mcp.NewToolResultError("index_id is required"), nil
	}

	result, err := srv.RefreshIndex(ctx, indexID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error: %v", err)), nil
	}

	return common.MarshalToolResult(result)
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
