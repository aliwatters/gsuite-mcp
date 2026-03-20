package common

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestWithLargeContentHint_NoHintWhenDisabled(t *testing.T) {
	SetDeps(&Deps{CitationEnabled: false})
	defer SetDeps(nil)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(strings.Repeat("x", 30_000)), nil
	}

	wrapped := WithLargeContentHint(handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block (no hint), got %d", len(result.Content))
	}
}

func TestWithLargeContentHint_HintWhenEnabledAndLarge(t *testing.T) {
	SetDeps(&Deps{CitationEnabled: true})
	defer SetDeps(nil)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(strings.Repeat("x", 30_000)), nil
	}

	wrapped := WithLargeContentHint(handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 2 {
		t.Fatalf("expected 2 content blocks (content + hint), got %d", len(result.Content))
	}

	hintBlock, ok := result.Content[1].(mcp.TextContent)
	if !ok {
		t.Fatal("expected second block to be TextContent")
	}
	if !strings.Contains(hintBlock.Text, "citation") {
		t.Errorf("hint should mention citation, got: %q", hintBlock.Text)
	}
}

func TestWithLargeContentHint_NoHintWhenSmall(t *testing.T) {
	SetDeps(&Deps{CitationEnabled: true})
	defer SetDeps(nil)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("small response"), nil
	}

	wrapped := WithLargeContentHint(handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block (no hint for small response), got %d", len(result.Content))
	}
}

func TestWithLargeContentHint_NoHintOnError(t *testing.T) {
	SetDeps(&Deps{CitationEnabled: true})
	defer SetDeps(nil)

	handler := func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("something went wrong"), nil
	}

	wrapped := WithLargeContentHint(handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block (no hint on error), got %d", len(result.Content))
	}
}
