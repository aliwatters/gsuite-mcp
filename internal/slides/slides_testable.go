package slides

import (
	"context"
	"fmt"

	"github.com/aliwatters/gsuite-mcp/internal/common"
	"github.com/mark3labs/mcp-go/mcp"
)

// slidesEditURLFormat is the URL template for Google Slides edit links.
const slidesEditURLFormat = "https://docs.google.com/presentation/d/%s/edit"

// extractRequiredPresentationID extracts, validates, and normalizes the presentation_id parameter.
func extractRequiredPresentationID(request mcp.CallToolRequest) (string, *mcp.CallToolResult) {
	presID := common.ParseStringArg(request.Params.Arguments, "presentation_id", "")
	if presID == "" {
		return "", mcp.NewToolResultError("presentation_id parameter is required")
	}
	return common.ExtractGoogleResourceID(presID), nil
}

// TestableSlidesGetPresentation retrieves a presentation's metadata, slides, and structure.
func TestableSlidesGetPresentation(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSlidesServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	presID, errResult := extractRequiredPresentationID(request)
	if errResult != nil {
		return errResult, nil
	}

	pres, err := srv.GetPresentation(ctx, presID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Slides API error: %v", err)), nil
	}

	// Format slides summary
	slidesSummary := make([]map[string]any, 0, len(pres.Slides))
	for _, slide := range pres.Slides {
		summary := map[string]any{
			"page_id":        slide.ObjectId,
			"element_count":  len(slide.PageElements),
		}

		// Extract text content from the slide for a preview
		var textPreview string
		for _, elem := range slide.PageElements {
			if elem.Shape != nil && elem.Shape.Text != nil {
				text := extractTextFromTextElements(elem.Shape.Text.TextElements)
				if text != "" && textPreview == "" {
					if len(text) > 100 {
						textPreview = text[:100] + "..."
					} else {
						textPreview = text
					}
				}
			}
		}
		if textPreview != "" {
			summary["text_preview"] = textPreview
		}

		slidesSummary = append(slidesSummary, summary)
	}

	result := map[string]any{
		"presentation_id": pres.PresentationId,
		"title":           pres.Title,
		"url":             fmt.Sprintf(slidesEditURLFormat, pres.PresentationId),
		"slide_count":     len(pres.Slides),
		"master_count":    len(pres.Masters),
		"layout_count":    len(pres.Layouts),
		"slides":          slidesSummary,
	}

	if pres.PageSize != nil {
		size := map[string]any{}
		if pres.PageSize.Width != nil {
			size["width_emu"] = pres.PageSize.Width.Magnitude
		}
		if pres.PageSize.Height != nil {
			size["height_emu"] = pres.PageSize.Height.Magnitude
		}
		result["page_size"] = size
	}

	return common.MarshalToolResult(result)
}

// TestableSlidesGetPage retrieves a single page/slide with full element details.
func TestableSlidesGetPage(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSlidesServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	presID, errResult := extractRequiredPresentationID(request)
	if errResult != nil {
		return errResult, nil
	}

	pageID, errResult := common.RequireStringArg(request.Params.Arguments, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	page, err := srv.GetPage(ctx, presID, pageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Slides API error: %v", err)), nil
	}

	result := formatPage(page)
	result["presentation_id"] = presID

	return common.MarshalToolResult(result)
}

// TestableSlidesGetThumbnail retrieves a thumbnail image URL for a slide.
func TestableSlidesGetThumbnail(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSlidesServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	presID, errResult := extractRequiredPresentationID(request)
	if errResult != nil {
		return errResult, nil
	}

	pageID, errResult := common.RequireStringArg(request.Params.Arguments, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	thumbnail, err := srv.GetPageThumbnail(ctx, presID, pageID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Slides API error: %v", err)), nil
	}

	result := map[string]any{
		"presentation_id": presID,
		"page_id":         pageID,
		"content_url":     thumbnail.ContentUrl,
		"width":           thumbnail.Width,
		"height":          thumbnail.Height,
	}

	return common.MarshalToolResult(result)
}

// TestableSlidesCreate creates a new presentation.
func TestableSlidesCreate(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSlidesServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	title, errResult := common.RequireStringArg(request.Params.Arguments, "title")
	if errResult != nil {
		return errResult, nil
	}

	pres, err := srv.CreatePresentation(ctx, title)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Slides API error: %v", err)), nil
	}

	result := map[string]any{
		"presentation_id": pres.PresentationId,
		"title":           pres.Title,
		"url":             fmt.Sprintf(slidesEditURLFormat, pres.PresentationId),
		"slide_count":     len(pres.Slides),
	}

	return common.MarshalToolResult(result)
}

// TestableSlidesBatchUpdate performs a batch update on a presentation.
func TestableSlidesBatchUpdate(ctx context.Context, request mcp.CallToolRequest, deps *SlidesHandlerDeps) (*mcp.CallToolResult, error) {
	srv, errResult, ok := ResolveSlidesServiceOrError(ctx, request, deps)
	if !ok {
		return errResult, nil
	}

	presID, errResult := extractRequiredPresentationID(request)
	if errResult != nil {
		return errResult, nil
	}

	requestsJSON, errResult := common.RequireStringArg(request.Params.Arguments, "requests")
	if errResult != nil {
		return errResult, nil
	}

	requests, err := parseBatchUpdateRequests(requestsJSON)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid requests JSON: %v", err)), nil
	}

	resp, err := srv.BatchUpdate(ctx, presID, requests)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Slides API error: %v", err)), nil
	}

	result := map[string]any{
		"presentation_id": resp.PresentationId,
		"replies_count":   len(resp.Replies),
	}

	return common.MarshalToolResult(result)
}
