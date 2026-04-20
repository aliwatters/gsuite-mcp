package forms

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/api/forms/v1"
)

// FormsService defines the interface for Google Forms API operations.
// This interface enables dependency injection and testing with mocks.
type FormsService interface {
	// GetForm retrieves a form by ID.
	GetForm(ctx context.Context, formID string) (*forms.Form, error)

	// CreateForm creates a new form with the given title.
	CreateForm(ctx context.Context, title string) (*forms.Form, error)

	// BatchUpdate performs a batch update on a form.
	BatchUpdate(ctx context.Context, formID string, requests []*forms.Request) (*forms.BatchUpdateFormResponse, error)

	// ListResponses lists responses for a form.
	ListResponses(ctx context.Context, formID string) ([]*forms.FormResponse, error)

	// GetResponse retrieves a single form response by ID.
	GetResponse(ctx context.Context, formID string, responseID string) (*forms.FormResponse, error)
}

// RealFormsService wraps the Forms API client and implements FormsService.
type RealFormsService struct {
	service *forms.Service
}

// NewRealFormsService creates a new RealFormsService wrapping the given API service.
func NewRealFormsService(service *forms.Service) *RealFormsService {
	return &RealFormsService{service: service}
}

// GetForm retrieves a form by ID.
func (s *RealFormsService) GetForm(ctx context.Context, formID string) (*forms.Form, error) {
	return s.service.Forms.Get(formID).Context(ctx).Do()
}

// CreateForm creates a new form with the given title.
func (s *RealFormsService) CreateForm(ctx context.Context, title string) (*forms.Form, error) {
	form := &forms.Form{
		Info: &forms.Info{
			Title: title,
		},
	}
	return s.service.Forms.Create(form).Context(ctx).Do()
}

// BatchUpdate performs a batch update on a form.
func (s *RealFormsService) BatchUpdate(ctx context.Context, formID string, requests []*forms.Request) (*forms.BatchUpdateFormResponse, error) {
	req := &forms.BatchUpdateFormRequest{
		Requests:              requests,
		IncludeFormInResponse: true,
	}
	return s.service.Forms.BatchUpdate(formID, req).Context(ctx).Do()
}

// maxResponsePages limits pagination to prevent unbounded memory growth.
const maxResponsePages = 10

// ListResponses lists responses for a form (up to maxResponsePages pages).
func (s *RealFormsService) ListResponses(ctx context.Context, formID string) ([]*forms.FormResponse, error) {
	var allResponses []*forms.FormResponse
	pageToken := ""
	for page := 0; page < maxResponsePages; page++ {
		call := s.service.Forms.Responses.List(formID).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing form responses for %s: %w", formID, err)
		}
		allResponses = append(allResponses, resp.Responses...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return allResponses, nil
}

// GetResponse retrieves a single form response by ID.
func (s *RealFormsService) GetResponse(ctx context.Context, formID string, responseID string) (*forms.FormResponse, error) {
	return s.service.Forms.Responses.Get(formID, responseID).Context(ctx).Do()
}

// formatItem formats a form item for output.
func formatItem(item *forms.Item) map[string]any {
	result := map[string]any{
		"item_id": item.ItemId,
		"title":   item.Title,
	}

	if item.Description != "" {
		result["description"] = item.Description
	}

	if item.QuestionItem != nil {
		result["type"] = "question"
		q := item.QuestionItem.Question
		if q != nil {
			result["question_id"] = q.QuestionId
			result["required"] = q.Required
			if q.ChoiceQuestion != nil {
				result["question_type"] = "choice"
				result["choice_type"] = q.ChoiceQuestion.Type
				options := make([]string, 0, len(q.ChoiceQuestion.Options))
				for _, opt := range q.ChoiceQuestion.Options {
					options = append(options, opt.Value)
				}
				result["options"] = options
			} else if q.TextQuestion != nil {
				result["question_type"] = "text"
				result["paragraph"] = q.TextQuestion.Paragraph
			} else if q.ScaleQuestion != nil {
				result["question_type"] = "scale"
				result["low"] = q.ScaleQuestion.Low
				result["high"] = q.ScaleQuestion.High
			} else if q.DateQuestion != nil {
				result["question_type"] = "date"
			} else if q.TimeQuestion != nil {
				result["question_type"] = "time"
			} else if q.FileUploadQuestion != nil {
				result["question_type"] = "file_upload"
			} else if q.RatingQuestion != nil {
				result["question_type"] = "rating"
			}
		}
	} else if item.QuestionGroupItem != nil {
		result["type"] = "question_group"
		questions := make([]map[string]any, 0, len(item.QuestionGroupItem.Questions))
		for _, q := range item.QuestionGroupItem.Questions {
			qMap := map[string]any{
				"question_id": q.QuestionId,
				"required":    q.Required,
			}
			questions = append(questions, qMap)
		}
		result["questions"] = questions
	} else if item.PageBreakItem != nil {
		result["type"] = "page_break"
	} else if item.TextItem != nil {
		result["type"] = "text"
	} else if item.ImageItem != nil {
		result["type"] = "image"
	} else if item.VideoItem != nil {
		result["type"] = "video"
	}

	return result
}

// formatResponse formats a form response for output.
func formatResponse(resp *forms.FormResponse) map[string]any {
	result := map[string]any{
		"response_id":         resp.ResponseId,
		"create_time":         resp.CreateTime,
		"last_submitted_time": resp.LastSubmittedTime,
	}

	if resp.RespondentEmail != "" {
		result["respondent_email"] = resp.RespondentEmail
	}

	if resp.TotalScore > 0 {
		result["total_score"] = resp.TotalScore
	}

	answers := make(map[string]any, len(resp.Answers))
	for qID, answer := range resp.Answers {
		a := map[string]any{
			"question_id": answer.QuestionId,
		}
		if answer.TextAnswers != nil {
			texts := make([]string, 0, len(answer.TextAnswers.Answers))
			for _, ta := range answer.TextAnswers.Answers {
				texts = append(texts, ta.Value)
			}
			a["text_answers"] = texts
		}
		if answer.Grade != nil {
			a["score"] = answer.Grade.Score
			a["correct"] = answer.Grade.Correct
		}
		answers[qID] = a
	}
	result["answers"] = answers

	return result
}

// parseBatchUpdateRequests parses a JSON string into Forms batch update requests.
func parseBatchUpdateRequests(requestsJSON string) ([]*forms.Request, error) {
	var requests []*forms.Request
	if err := json.Unmarshal([]byte(requestsJSON), &requests); err != nil {
		return nil, fmt.Errorf("parsing forms batch update requests: %w", err)
	}
	return requests, nil
}
