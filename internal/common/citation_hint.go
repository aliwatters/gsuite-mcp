package common

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// largeContentThreshold is the response size (bytes) above which we suggest indexing.
// ~20KB ≈ 5000 tokens — a practical threshold for "this content is big enough
// that citation indexing would be valuable."
const largeContentThreshold = 20_000

// citationHint is the compact hint appended to large responses.
const citationHint = "This content is large. For reliable citation tracing, use citation_add_documents to index it, then citation_lookup/citation_verify_claim to find and verify specific claims."

// WithLargeContentHint wraps an MCP handler to append a citation hint when
// the response content exceeds the threshold and citation tools are enabled.
// Zero overhead when citation is disabled — the check is a single bool read.
// Pass explicit appDeps to avoid reading from the global singleton; omit to fall
// back to GetDeps() for backward compatibility.
func WithLargeContentHint(handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), appDeps ...*Deps) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var injected *Deps
	if len(appDeps) > 0 {
		injected = appDeps[0]
	}
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := handler(ctx, request)
		if err != nil || result == nil || result.IsError {
			return result, err
		}

		d := injected
		if d == nil {
			d = GetDeps()
		}
		if d == nil || !d.CitationEnabled {
			return result, nil
		}

		// Check response size
		size := resultContentSize(result)
		if size < largeContentThreshold {
			return result, nil
		}

		// Append hint as a separate text content block
		result.Content = append(result.Content, mcp.TextContent{
			Annotated: mcp.Annotated{},
			Type:      "text",
			Text:      citationHint,
		})

		return result, nil
	}
}

// resultContentSize estimates the byte size of the tool result content.
func resultContentSize(result *mcp.CallToolResult) int {
	total := 0
	for _, c := range result.Content {
		switch v := c.(type) {
		case mcp.TextContent:
			total += len(v.Text)
		default:
			// For non-text content, marshal to estimate
			if data, err := json.Marshal(v); err == nil {
				total += len(data)
			}
		}
	}
	return total
}
