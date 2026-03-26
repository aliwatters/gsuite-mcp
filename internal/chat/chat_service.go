package chat

import (
	"context"

	chatapi "google.golang.org/api/chat/v1"
)

// ChatService defines the interface for Google Chat API operations.
// This interface enables dependency injection and testing with mocks.
type ChatService interface {
	// ListSpaces lists Chat spaces the user is a member of.
	ListSpaces(ctx context.Context, pageSize int64, pageToken string) ([]*chatapi.Space, string, error)

	// GetSpace gets details about a Chat space.
	GetSpace(ctx context.Context, spaceName string) (*chatapi.Space, error)

	// CreateSpace creates a new Chat space.
	CreateSpace(ctx context.Context, displayName string, spaceType string) (*chatapi.Space, error)

	// ListMessages lists messages in a Chat space.
	ListMessages(ctx context.Context, spaceName string, pageSize int64, pageToken string, filter string) ([]*chatapi.Message, string, error)

	// GetMessage gets a specific message.
	GetMessage(ctx context.Context, messageName string) (*chatapi.Message, error)

	// SendMessage sends a message to a Chat space.
	SendMessage(ctx context.Context, spaceName string, text string, threadName string) (*chatapi.Message, error)

	// CreateReaction adds a reaction to a message.
	CreateReaction(ctx context.Context, messageName string, emoji string) (*chatapi.Reaction, error)

	// DeleteReaction removes a reaction from a message.
	DeleteReaction(ctx context.Context, reactionName string) error

	// ListMembers lists members of a Chat space.
	ListMembers(ctx context.Context, spaceName string, pageSize int64, pageToken string) ([]*chatapi.Membership, string, error)
}

// maxPages limits pagination to prevent unbounded memory growth.
const maxPages = 10

// RealChatService wraps the Chat API client and implements ChatService.
type RealChatService struct {
	service *chatapi.Service
}

// NewRealChatService creates a new RealChatService wrapping the given API service.
func NewRealChatService(service *chatapi.Service) *RealChatService {
	return &RealChatService{service: service}
}

// ListSpaces lists Chat spaces the user is a member of.
func (s *RealChatService) ListSpaces(ctx context.Context, pageSize int64, pageToken string) ([]*chatapi.Space, string, error) {
	call := s.service.Spaces.List().Context(ctx)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.Spaces, resp.NextPageToken, nil
}

// GetSpace gets details about a Chat space.
func (s *RealChatService) GetSpace(ctx context.Context, spaceName string) (*chatapi.Space, error) {
	return s.service.Spaces.Get(spaceName).Context(ctx).Do()
}

// CreateSpace creates a new Chat space.
func (s *RealChatService) CreateSpace(ctx context.Context, displayName string, spaceType string) (*chatapi.Space, error) {
	space := &chatapi.Space{
		DisplayName: displayName,
		SpaceType:   spaceType,
	}
	return s.service.Spaces.Create(space).Context(ctx).Do()
}

// ListMessages lists messages in a Chat space.
func (s *RealChatService) ListMessages(ctx context.Context, spaceName string, pageSize int64, pageToken string, filter string) ([]*chatapi.Message, string, error) {
	call := s.service.Spaces.Messages.List(spaceName).Context(ctx)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if filter != "" {
		call = call.Filter(filter)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.Messages, resp.NextPageToken, nil
}

// GetMessage gets a specific message.
func (s *RealChatService) GetMessage(ctx context.Context, messageName string) (*chatapi.Message, error) {
	return s.service.Spaces.Messages.Get(messageName).Context(ctx).Do()
}

// SendMessage sends a message to a Chat space.
func (s *RealChatService) SendMessage(ctx context.Context, spaceName string, text string, threadName string) (*chatapi.Message, error) {
	msg := &chatapi.Message{
		Text: text,
	}
	if threadName != "" {
		msg.Thread = &chatapi.Thread{
			Name: threadName,
		}
	}
	return s.service.Spaces.Messages.Create(spaceName, msg).Context(ctx).Do()
}

// CreateReaction adds a reaction to a message.
func (s *RealChatService) CreateReaction(ctx context.Context, messageName string, emoji string) (*chatapi.Reaction, error) {
	reaction := &chatapi.Reaction{
		Emoji: &chatapi.Emoji{
			Unicode: emoji,
		},
	}
	return s.service.Spaces.Messages.Reactions.Create(messageName, reaction).Context(ctx).Do()
}

// DeleteReaction removes a reaction from a message.
func (s *RealChatService) DeleteReaction(ctx context.Context, reactionName string) error {
	_, err := s.service.Spaces.Messages.Reactions.Delete(reactionName).Context(ctx).Do()
	return err
}

// ListMembers lists members of a Chat space.
func (s *RealChatService) ListMembers(ctx context.Context, spaceName string, pageSize int64, pageToken string) ([]*chatapi.Membership, string, error) {
	call := s.service.Spaces.Members.List(spaceName).Context(ctx)
	if pageSize > 0 {
		call = call.PageSize(pageSize)
	}
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	resp, err := call.Do()
	if err != nil {
		return nil, "", err
	}
	return resp.Memberships, resp.NextPageToken, nil
}
