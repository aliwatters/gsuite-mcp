package forms

import (
	"context"
	"fmt"

	"google.golang.org/api/forms/v1"
)

// MockFormsService implements FormsService for testing.
type MockFormsService struct {
	// Forms stores mock form data keyed by ID.
	Forms map[string]*forms.Form

	// Responses stores mock response data keyed by form ID.
	Responses map[string][]*forms.FormResponse

	// Errors allows tests to configure specific errors for methods.
	Errors struct {
		GetForm       error
		Create        error
		BatchUpdate   error
		ListResponses error
		GetResponse   error
	}

	// Calls tracks method invocations for verification.
	Calls struct {
		GetForm     []string
		Create      []string
		BatchUpdate []struct {
			FormID   string
			Requests []*forms.Request
		}
		ListResponses []string
		GetResponse   []struct{ FormID, ResponseID string }
	}
}

// NewMockFormsService creates a new mock Forms service with default test data.
func NewMockFormsService() *MockFormsService {
	m := &MockFormsService{
		Forms:     make(map[string]*forms.Form),
		Responses: make(map[string][]*forms.FormResponse),
	}

	// Add a default test form
	m.Forms["test-form-1"] = &forms.Form{
		FormId: "test-form-1",
		Info: &forms.Info{
			Title:         "Test Form",
			Description:   "A test form",
			DocumentTitle: "Test Form Doc",
		},
		ResponderUri: "https://docs.google.com/forms/d/test-form-1/viewform",
		Items: []*forms.Item{
			{
				ItemId: "item-1",
				Title:  "What is your name?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						QuestionId: "q-1",
						Required:   true,
						TextQuestion: &forms.TextQuestion{
							Paragraph: false,
						},
					},
				},
			},
			{
				ItemId: "item-2",
				Title:  "Favorite color?",
				QuestionItem: &forms.QuestionItem{
					Question: &forms.Question{
						QuestionId: "q-2",
						ChoiceQuestion: &forms.ChoiceQuestion{
							Type: "RADIO",
							Options: []*forms.Option{
								{Value: "Red"},
								{Value: "Blue"},
								{Value: "Green"},
							},
						},
					},
				},
			},
			{
				ItemId:        "item-3",
				Title:         "Section 2",
				PageBreakItem: &forms.PageBreakItem{},
			},
		},
	}

	// Add test responses
	m.Responses["test-form-1"] = []*forms.FormResponse{
		{
			ResponseId:        "resp-1",
			FormId:            "test-form-1",
			CreateTime:        "2024-01-15T10:00:00Z",
			LastSubmittedTime: "2024-01-15T10:00:00Z",
			RespondentEmail:   "user@example.com",
			Answers: map[string]forms.Answer{
				"q-1": {
					QuestionId: "q-1",
					TextAnswers: &forms.TextAnswers{
						Answers: []*forms.TextAnswer{
							{Value: "Alice"},
						},
					},
				},
				"q-2": {
					QuestionId: "q-2",
					TextAnswers: &forms.TextAnswers{
						Answers: []*forms.TextAnswer{
							{Value: "Blue"},
						},
					},
				},
			},
		},
	}

	return m
}

// GetForm retrieves a mock form by ID.
func (m *MockFormsService) GetForm(ctx context.Context, formID string) (*forms.Form, error) {
	m.Calls.GetForm = append(m.Calls.GetForm, formID)

	if m.Errors.GetForm != nil {
		return nil, m.Errors.GetForm
	}

	form, ok := m.Forms[formID]
	if !ok {
		return nil, fmt.Errorf("form not found: %s", formID)
	}

	return form, nil
}

// CreateForm creates a mock form.
func (m *MockFormsService) CreateForm(ctx context.Context, title string) (*forms.Form, error) {
	m.Calls.Create = append(m.Calls.Create, title)

	if m.Errors.Create != nil {
		return nil, m.Errors.Create
	}

	formID := fmt.Sprintf("new-form-%d", len(m.Forms)+1)
	form := &forms.Form{
		FormId: formID,
		Info: &forms.Info{
			Title: title,
		},
		ResponderUri: fmt.Sprintf("https://docs.google.com/forms/d/%s/viewform", formID),
	}

	m.Forms[formID] = form
	return form, nil
}

// BatchUpdate performs a mock batch update.
func (m *MockFormsService) BatchUpdate(ctx context.Context, formID string, requests []*forms.Request) (*forms.BatchUpdateFormResponse, error) {
	m.Calls.BatchUpdate = append(m.Calls.BatchUpdate, struct {
		FormID   string
		Requests []*forms.Request
	}{formID, requests})

	if m.Errors.BatchUpdate != nil {
		return nil, m.Errors.BatchUpdate
	}

	_, ok := m.Forms[formID]
	if !ok {
		return nil, fmt.Errorf("form not found: %s", formID)
	}

	return &forms.BatchUpdateFormResponse{
		Replies: make([]*forms.Response, len(requests)),
	}, nil
}

// ListResponses lists mock responses for a form.
func (m *MockFormsService) ListResponses(ctx context.Context, formID string) ([]*forms.FormResponse, error) {
	m.Calls.ListResponses = append(m.Calls.ListResponses, formID)

	if m.Errors.ListResponses != nil {
		return nil, m.Errors.ListResponses
	}

	_, ok := m.Forms[formID]
	if !ok {
		return nil, fmt.Errorf("form not found: %s", formID)
	}

	responses := m.Responses[formID]
	if responses == nil {
		responses = []*forms.FormResponse{}
	}
	return responses, nil
}

// GetResponse retrieves a mock form response by ID.
func (m *MockFormsService) GetResponse(ctx context.Context, formID string, responseID string) (*forms.FormResponse, error) {
	m.Calls.GetResponse = append(m.Calls.GetResponse, struct{ FormID, ResponseID string }{formID, responseID})

	if m.Errors.GetResponse != nil {
		return nil, m.Errors.GetResponse
	}

	responses := m.Responses[formID]
	for _, resp := range responses {
		if resp.ResponseId == responseID {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("response not found: %s", responseID)
}
