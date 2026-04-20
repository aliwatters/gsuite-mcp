package slides

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/api/slides/v1"
)

// SlidesService defines the interface for Google Slides API operations.
// This interface enables dependency injection and testing with mocks.
type SlidesService interface {
	// GetPresentation retrieves a presentation by ID.
	GetPresentation(ctx context.Context, presentationID string) (*slides.Presentation, error)

	// GetPage retrieves a single page (slide) by ID.
	GetPage(ctx context.Context, presentationID string, pageID string) (*slides.Page, error)

	// GetPageThumbnail retrieves a thumbnail image URL for a page.
	GetPageThumbnail(ctx context.Context, presentationID string, pageID string) (*slides.Thumbnail, error)

	// CreatePresentation creates a new presentation with the given title.
	CreatePresentation(ctx context.Context, title string) (*slides.Presentation, error)

	// BatchUpdate performs a batch update on a presentation.
	BatchUpdate(ctx context.Context, presentationID string, requests []*slides.Request) (*slides.BatchUpdatePresentationResponse, error)
}

// RealSlidesService wraps the Slides API client and implements SlidesService.
type RealSlidesService struct {
	service *slides.Service
}

// NewRealSlidesService creates a new RealSlidesService wrapping the given API service.
func NewRealSlidesService(service *slides.Service) *RealSlidesService {
	return &RealSlidesService{service: service}
}

// GetPresentation retrieves a presentation by ID.
func (s *RealSlidesService) GetPresentation(ctx context.Context, presentationID string) (*slides.Presentation, error) {
	return s.service.Presentations.Get(presentationID).Context(ctx).Do()
}

// GetPage retrieves a single page by ID.
func (s *RealSlidesService) GetPage(ctx context.Context, presentationID string, pageID string) (*slides.Page, error) {
	return s.service.Presentations.Pages.Get(presentationID, pageID).Context(ctx).Do()
}

// GetPageThumbnail retrieves a thumbnail URL for a page.
func (s *RealSlidesService) GetPageThumbnail(ctx context.Context, presentationID string, pageID string) (*slides.Thumbnail, error) {
	return s.service.Presentations.Pages.GetThumbnail(presentationID, pageID).
		ThumbnailPropertiesThumbnailSize("LARGE").
		Context(ctx).Do()
}

// CreatePresentation creates a new presentation.
func (s *RealSlidesService) CreatePresentation(ctx context.Context, title string) (*slides.Presentation, error) {
	pres := &slides.Presentation{
		Title: title,
	}
	return s.service.Presentations.Create(pres).Context(ctx).Do()
}

// BatchUpdate performs a batch update on a presentation.
func (s *RealSlidesService) BatchUpdate(ctx context.Context, presentationID string, requests []*slides.Request) (*slides.BatchUpdatePresentationResponse, error) {
	req := &slides.BatchUpdatePresentationRequest{
		Requests: requests,
	}
	return s.service.Presentations.BatchUpdate(presentationID, req).Context(ctx).Do()
}

// formatPageElement formats a page element for output.
func formatPageElement(elem *slides.PageElement) map[string]any {
	result := map[string]any{
		"object_id": elem.ObjectId,
	}

	if elem.Size != nil {
		result["width"] = elem.Size.Width
		result["height"] = elem.Size.Height
	}

	if elem.Transform != nil {
		result["translate_x"] = elem.Transform.TranslateX
		result["translate_y"] = elem.Transform.TranslateY
	}

	if elem.Shape != nil {
		result["type"] = "shape"
		result["shape_type"] = elem.Shape.ShapeType
		if elem.Shape.Text != nil {
			result["text"] = extractTextFromTextElements(elem.Shape.Text.TextElements)
		}
	} else if elem.Image != nil {
		result["type"] = "image"
		result["source_url"] = elem.Image.SourceUrl
		result["content_url"] = elem.Image.ContentUrl
	} else if elem.Table != nil {
		result["type"] = "table"
		result["rows"] = elem.Table.Rows
		result["columns"] = elem.Table.Columns
	} else if elem.Video != nil {
		result["type"] = "video"
		result["video_id"] = elem.Video.Id
		result["source"] = elem.Video.Source
	} else if elem.SheetsChart != nil {
		result["type"] = "sheets_chart"
		result["spreadsheet_id"] = elem.SheetsChart.SpreadsheetId
		result["chart_id"] = elem.SheetsChart.ChartId
	} else if elem.ElementGroup != nil {
		result["type"] = "group"
		children := make([]map[string]any, 0, len(elem.ElementGroup.Children))
		for _, child := range elem.ElementGroup.Children {
			children = append(children, formatPageElement(child))
		}
		result["children"] = children
	}

	return result
}

// extractTextFromTextElements extracts plain text from Slides text elements.
func extractTextFromTextElements(elements []*slides.TextElement) string {
	var text string
	for _, elem := range elements {
		if elem.TextRun != nil {
			text += elem.TextRun.Content
		}
	}
	return text
}

// formatPage formats a page/slide for output.
func formatPage(page *slides.Page) map[string]any {
	result := map[string]any{
		"page_id":   page.ObjectId,
		"page_type": page.PageType,
	}

	if page.SlideProperties != nil {
		result["layout_id"] = page.SlideProperties.LayoutObjectId
		result["master_id"] = page.SlideProperties.MasterObjectId
	}

	elements := make([]map[string]any, 0, len(page.PageElements))
	for _, elem := range page.PageElements {
		elements = append(elements, formatPageElement(elem))
	}
	result["elements"] = elements

	return result
}

// parseBatchUpdateRequests parses a JSON string into Slides batch update requests.
func parseBatchUpdateRequests(requestsJSON string) ([]*slides.Request, error) {
	var requests []*slides.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("parsing slides batch update requests: %w", err)
	}
	return requests, nil
}
