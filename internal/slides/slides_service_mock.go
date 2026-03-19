package slides

import (
	"context"
	"fmt"

	"google.golang.org/api/slides/v1"
)

// MockSlidesService implements SlidesService for testing.
type MockSlidesService struct {
	// Presentations stores mock presentation data keyed by ID.
	Presentations map[string]*slides.Presentation

	// Errors allows tests to configure specific errors for methods.
	Errors struct {
		GetPresentation  error
		GetPage          error
		GetPageThumbnail error
		Create           error
		BatchUpdate      error
	}

	// Calls tracks method invocations for verification.
	Calls struct {
		GetPresentation  []string
		GetPage          []struct{ PresentationID, PageID string }
		GetPageThumbnail []struct{ PresentationID, PageID string }
		Create           []string
		BatchUpdate      []struct {
			PresentationID string
			Requests       []*slides.Request
		}
	}
}

// NewMockSlidesService creates a new mock Slides service with default test data.
func NewMockSlidesService() *MockSlidesService {
	m := &MockSlidesService{
		Presentations: make(map[string]*slides.Presentation),
	}

	// Add a default test presentation
	m.Presentations["test-pres-1"] = &slides.Presentation{
		PresentationId: "test-pres-1",
		Title:          "Test Presentation",
		PageSize: &slides.Size{
			Width:  &slides.Dimension{Magnitude: 9144000, Unit: "EMU"},
			Height: &slides.Dimension{Magnitude: 6858000, Unit: "EMU"},
		},
		Slides: []*slides.Page{
			{
				ObjectId: "slide-1",
				PageType: "SLIDE",
				SlideProperties: &slides.SlideProperties{
					LayoutObjectId: "layout-1",
					MasterObjectId: "master-1",
				},
				PageElements: []*slides.PageElement{
					{
						ObjectId: "title-1",
						Shape: &slides.Shape{
							ShapeType: "TEXT_BOX",
							Text: &slides.TextContent{
								TextElements: []*slides.TextElement{
									{TextRun: &slides.TextRun{Content: "Slide 1 Title\n"}},
								},
							},
						},
					},
				},
			},
			{
				ObjectId: "slide-2",
				PageType: "SLIDE",
				SlideProperties: &slides.SlideProperties{
					LayoutObjectId: "layout-1",
					MasterObjectId: "master-1",
				},
				PageElements: []*slides.PageElement{
					{
						ObjectId: "body-1",
						Shape: &slides.Shape{
							ShapeType: "TEXT_BOX",
							Text: &slides.TextContent{
								TextElements: []*slides.TextElement{
									{TextRun: &slides.TextRun{Content: "Slide 2 Body\n"}},
								},
							},
						},
					},
				},
			},
		},
		Masters: []*slides.Page{
			{ObjectId: "master-1", PageType: "MASTER"},
		},
		Layouts: []*slides.Page{
			{ObjectId: "layout-1", PageType: "LAYOUT"},
		},
	}

	return m
}

// GetPresentation retrieves a mock presentation by ID.
func (m *MockSlidesService) GetPresentation(ctx context.Context, presentationID string) (*slides.Presentation, error) {
	m.Calls.GetPresentation = append(m.Calls.GetPresentation, presentationID)

	if m.Errors.GetPresentation != nil {
		return nil, m.Errors.GetPresentation
	}

	pres, ok := m.Presentations[presentationID]
	if !ok {
		return nil, fmt.Errorf("presentation not found: %s", presentationID)
	}

	return pres, nil
}

// GetPage retrieves a mock page by ID.
func (m *MockSlidesService) GetPage(ctx context.Context, presentationID string, pageID string) (*slides.Page, error) {
	m.Calls.GetPage = append(m.Calls.GetPage, struct{ PresentationID, PageID string }{presentationID, pageID})

	if m.Errors.GetPage != nil {
		return nil, m.Errors.GetPage
	}

	pres, ok := m.Presentations[presentationID]
	if !ok {
		return nil, fmt.Errorf("presentation not found: %s", presentationID)
	}

	// Search all page types
	for _, page := range pres.Slides {
		if page.ObjectId == pageID {
			return page, nil
		}
	}
	for _, page := range pres.Masters {
		if page.ObjectId == pageID {
			return page, nil
		}
	}
	for _, page := range pres.Layouts {
		if page.ObjectId == pageID {
			return page, nil
		}
	}

	return nil, fmt.Errorf("page not found: %s", pageID)
}

// GetPageThumbnail retrieves a mock thumbnail for a page.
func (m *MockSlidesService) GetPageThumbnail(ctx context.Context, presentationID string, pageID string) (*slides.Thumbnail, error) {
	m.Calls.GetPageThumbnail = append(m.Calls.GetPageThumbnail, struct{ PresentationID, PageID string }{presentationID, pageID})

	if m.Errors.GetPageThumbnail != nil {
		return nil, m.Errors.GetPageThumbnail
	}

	pres, ok := m.Presentations[presentationID]
	if !ok {
		return nil, fmt.Errorf("presentation not found: %s", presentationID)
	}

	// Verify the page exists
	found := false
	for _, page := range pres.Slides {
		if page.ObjectId == pageID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("page not found: %s", pageID)
	}

	return &slides.Thumbnail{
		ContentUrl: fmt.Sprintf("https://slides.googleapis.com/thumbnail/%s/%s", presentationID, pageID),
		Width:      1600,
		Height:     900,
	}, nil
}

// CreatePresentation creates a mock presentation.
func (m *MockSlidesService) CreatePresentation(ctx context.Context, title string) (*slides.Presentation, error) {
	m.Calls.Create = append(m.Calls.Create, title)

	if m.Errors.Create != nil {
		return nil, m.Errors.Create
	}

	presID := fmt.Sprintf("new-pres-%d", len(m.Presentations)+1)
	pres := &slides.Presentation{
		PresentationId: presID,
		Title:          title,
		PageSize: &slides.Size{
			Width:  &slides.Dimension{Magnitude: 9144000, Unit: "EMU"},
			Height: &slides.Dimension{Magnitude: 6858000, Unit: "EMU"},
		},
		Slides: []*slides.Page{
			{
				ObjectId: fmt.Sprintf("%s-slide-1", presID),
				PageType: "SLIDE",
			},
		},
	}

	m.Presentations[presID] = pres
	return pres, nil
}

// BatchUpdate performs a mock batch update.
func (m *MockSlidesService) BatchUpdate(ctx context.Context, presentationID string, requests []*slides.Request) (*slides.BatchUpdatePresentationResponse, error) {
	m.Calls.BatchUpdate = append(m.Calls.BatchUpdate, struct {
		PresentationID string
		Requests       []*slides.Request
	}{presentationID, requests})

	if m.Errors.BatchUpdate != nil {
		return nil, m.Errors.BatchUpdate
	}

	_, ok := m.Presentations[presentationID]
	if !ok {
		return nil, fmt.Errorf("presentation not found: %s", presentationID)
	}

	return &slides.BatchUpdatePresentationResponse{
		PresentationId: presentationID,
	}, nil
}
