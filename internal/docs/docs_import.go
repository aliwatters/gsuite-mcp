package docs

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestableDocsImportToGoogleDoc creates a new Google Doc from provided content.
// Uses the Drive API to upload content with conversion to Google Docs format.
func TestableDocsImportToGoogleDoc(ctx context.Context, request mcp.CallToolRequest, deps *DocsHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveDocsServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title, errResult := common.RequireStringArg(request.GetArguments(), "title")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := common.RequireStringArg(request.GetArguments(), "content")
	if errResult != nil {
		return errResult, nil
	}

	contentType := common.ParseStringArg(request.GetArguments(), "content_type", "text/plain")

	// Validate content type
	validTypes := map[string]bool{
		"text/plain":    true,
		"text/html":     true,
		"text/markdown": true,
	}
	if !validTypes[contentType] {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Invalid content_type '%s'. Supported: text/plain, text/html, text/markdown", contentType)), nil
	}

	// For markdown, convert to HTML first (Google Drive doesn't natively support markdown import)
	uploadContent := content
	if contentType == "text/markdown" {
		uploadContent = markdownToBasicHTML(content)
	}

	parentID := ""
	if pid := common.ParseStringArg(request.GetArguments(), "parent_id", ""); pid != "" {
		parentID = common.ExtractGoogleResourceID(pid)
	}

	// Determine the upload content type for Drive API
	uploadMimeType := contentType
	if contentType == "text/markdown" {
		uploadMimeType = "text/html"
	}

	created, err := srv.ImportDocument(ctx, title, []byte(uploadContent), uploadMimeType, parentID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Drive API error creating document: %v", err)), nil
	}

	docURL := fmt.Sprintf(docsEditURLFormat, created.Id)
	if created.WebViewLink != "" {
		docURL = created.WebViewLink
	}

	result := map[string]any{
		"document_id":         created.Id,
		"title":               title,
		"source_content_type": contentType,
		"url":                 docURL,
		"message":             fmt.Sprintf("Created Google Doc '%s' from %s content", title, contentType),
	}

	return common.MarshalToolResult(result)
}

// markdownToBasicHTML performs a basic conversion of markdown to HTML.
// Handles headings, bold, italic, and paragraphs.
func markdownToBasicHTML(md string) string {
	lines := strings.Split(md, "\n")
	var b strings.Builder
	b.WriteString("<html><body>")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			b.WriteString("<br>")
			continue
		}

		// Headings — escape user content to prevent HTML injection
		if strings.HasPrefix(trimmed, "######") {
			b.WriteString("<h6>" + html.EscapeString(strings.TrimSpace(trimmed[6:])) + "</h6>")
		} else if strings.HasPrefix(trimmed, "#####") {
			b.WriteString("<h5>" + html.EscapeString(strings.TrimSpace(trimmed[5:])) + "</h5>")
		} else if strings.HasPrefix(trimmed, "####") {
			b.WriteString("<h4>" + html.EscapeString(strings.TrimSpace(trimmed[4:])) + "</h4>")
		} else if strings.HasPrefix(trimmed, "###") {
			b.WriteString("<h3>" + html.EscapeString(strings.TrimSpace(trimmed[3:])) + "</h3>")
		} else if strings.HasPrefix(trimmed, "##") {
			b.WriteString("<h2>" + html.EscapeString(strings.TrimSpace(trimmed[2:])) + "</h2>")
		} else if strings.HasPrefix(trimmed, "#") {
			b.WriteString("<h1>" + html.EscapeString(strings.TrimSpace(trimmed[1:])) + "</h1>")
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			b.WriteString("<ul><li>" + html.EscapeString(strings.TrimSpace(trimmed[2:])) + "</li></ul>")
		} else {
			text := applyInlineFormatting(trimmed)
			b.WriteString("<p>" + text + "</p>")
		}
	}

	b.WriteString("</body></html>")
	return b.String()
}

// applyInlineFormatting converts markdown inline formatting to HTML.
func applyInlineFormatting(text string) string {
	// Bold+italic (***text***)
	for {
		start := strings.Index(text, "***")
		if start < 0 {
			break
		}
		end := strings.Index(text[start+3:], "***")
		if end < 0 {
			break
		}
		end += start + 3
		inner := text[start+3 : end]
		text = text[:start] + "<b><i>" + inner + "</i></b>" + text[end+3:]
	}
	// Bold (**text**)
	for {
		start := strings.Index(text, "**")
		if start < 0 {
			break
		}
		end := strings.Index(text[start+2:], "**")
		if end < 0 {
			break
		}
		end += start + 2
		inner := text[start+2 : end]
		text = text[:start] + "<b>" + inner + "</b>" + text[end+2:]
	}
	// Italic (*text*)
	for {
		start := strings.Index(text, "*")
		if start < 0 {
			break
		}
		end := strings.Index(text[start+1:], "*")
		if end < 0 {
			break
		}
		end += start + 1
		inner := text[start+1 : end]
		text = text[:start] + "<i>" + inner + "</i>" + text[end+1:]
	}
	return text
}
