package chat

import (
	"context"
	"fmt"

	chatapi "google.golang.org/api/chat/v1"
)

// MockChatService implements ChatService for testing.
type MockChatService struct {
	// Spaces stores mock space data keyed by name.
	Spaces map[string]*chatapi.Space

	// Messages stores mock messages keyed by space name.
	Messages map[string][]*chatapi.Message

	// Members stores mock members keyed by space name.
	Members map[string][]*chatapi.Membership

	// Reactions stores mock reactions keyed by message name.
	Reactions map[string][]*chatapi.Reaction

	// Errors allows tests to configure specific errors for methods.
	Errors struct {
		ListSpaces     error
		GetSpace       error
		CreateSpace    error
		ListMessages   error
		GetMessage     error
		SendMessage    error
		CreateReaction error
		DeleteReaction error
		ListMembers    error
	}

	// Calls tracks method invocations for verification.
	Calls struct {
		ListSpaces     int
		GetSpace       []string
		CreateSpace    []struct{ DisplayName, SpaceType string }
		ListMessages   []string
		GetMessage     []string
		SendMessage    []struct{ SpaceName, Text, ThreadName string }
		CreateReaction []struct{ MessageName, Emoji string }
		DeleteReaction []string
		ListMembers    []string
	}
}

// NewMockChatService creates a new mock Chat service with default test data.
func NewMockChatService() *MockChatService {
	m := &MockChatService{
		Spaces:    make(map[string]*chatapi.Space),
		Messages:  make(map[string][]*chatapi.Message),
		Members:   make(map[string][]*chatapi.Membership),
		Reactions: make(map[string][]*chatapi.Reaction),
	}

	// Add default test spaces
	m.Spaces["spaces/test-space-1"] = &chatapi.Space{
		Name:                "spaces/test-space-1",
		DisplayName:         "Test Space",
		SpaceType:           "SPACE",
		SpaceThreadingState: "THREADED_MESSAGES",
	}
	m.Spaces["spaces/test-space-2"] = &chatapi.Space{
		Name:        "spaces/test-space-2",
		DisplayName: "Another Space",
		SpaceType:   "SPACE",
	}

	// Add default test messages
	m.Messages["spaces/test-space-1"] = []*chatapi.Message{
		{
			Name:       "spaces/test-space-1/messages/msg-1",
			Text:       "Hello, world!",
			CreateTime: "2024-01-15T10:00:00.000Z",
			Sender: &chatapi.User{
				Name:        "users/12345",
				DisplayName: "Alice",
			},
			Thread: &chatapi.Thread{
				Name: "spaces/test-space-1/threads/thread-1",
			},
		},
		{
			Name:       "spaces/test-space-1/messages/msg-2",
			Text:       "How are you?",
			CreateTime: "2024-01-15T10:05:00.000Z",
			Sender: &chatapi.User{
				Name:        "users/67890",
				DisplayName: "Bob",
			},
		},
	}

	// Add default test members
	m.Members["spaces/test-space-1"] = []*chatapi.Membership{
		{
			Name: "spaces/test-space-1/members/12345",
			Member: &chatapi.User{
				Name:        "users/12345",
				DisplayName: "Alice",
			},
			Role: "ROLE_MANAGER",
		},
		{
			Name: "spaces/test-space-1/members/67890",
			Member: &chatapi.User{
				Name:        "users/67890",
				DisplayName: "Bob",
			},
			Role: "ROLE_MEMBER",
		},
	}

	return m
}

// ListSpaces lists mock Chat spaces.
func (m *MockChatService) ListSpaces(ctx context.Context, pageSize int64, pageToken string) ([]*chatapi.Space, string, error) {
	m.Calls.ListSpaces++

	if m.Errors.ListSpaces != nil {
		return nil, "", m.Errors.ListSpaces
	}

	spaces := make([]*chatapi.Space, 0, len(m.Spaces))
	for _, s := range m.Spaces {
		spaces = append(spaces, s)
	}
	return spaces, "", nil
}

// GetSpace gets a mock Chat space.
func (m *MockChatService) GetSpace(ctx context.Context, spaceName string) (*chatapi.Space, error) {
	m.Calls.GetSpace = append(m.Calls.GetSpace, spaceName)

	if m.Errors.GetSpace != nil {
		return nil, m.Errors.GetSpace
	}

	space, ok := m.Spaces[spaceName]
	if !ok {
		return nil, fmt.Errorf("space not found: %s", spaceName)
	}
	return space, nil
}

// CreateSpace creates a mock Chat space.
func (m *MockChatService) CreateSpace(ctx context.Context, displayName string, spaceType string) (*chatapi.Space, error) {
	m.Calls.CreateSpace = append(m.Calls.CreateSpace, struct{ DisplayName, SpaceType string }{displayName, spaceType})

	if m.Errors.CreateSpace != nil {
		return nil, m.Errors.CreateSpace
	}

	spaceName := fmt.Sprintf("spaces/new-space-%d", len(m.Spaces)+1)
	space := &chatapi.Space{
		Name:        spaceName,
		DisplayName: displayName,
		SpaceType:   spaceType,
	}
	m.Spaces[spaceName] = space
	return space, nil
}

// ListMessages lists mock messages in a space.
func (m *MockChatService) ListMessages(ctx context.Context, spaceName string, pageSize int64, pageToken string, filter string) ([]*chatapi.Message, string, error) {
	m.Calls.ListMessages = append(m.Calls.ListMessages, spaceName)

	if m.Errors.ListMessages != nil {
		return nil, "", m.Errors.ListMessages
	}

	if _, ok := m.Spaces[spaceName]; !ok {
		return nil, "", fmt.Errorf("space not found: %s", spaceName)
	}

	messages := m.Messages[spaceName]
	if messages == nil {
		messages = []*chatapi.Message{}
	}
	return messages, "", nil
}

// GetMessage gets a mock message.
func (m *MockChatService) GetMessage(ctx context.Context, messageName string) (*chatapi.Message, error) {
	m.Calls.GetMessage = append(m.Calls.GetMessage, messageName)

	if m.Errors.GetMessage != nil {
		return nil, m.Errors.GetMessage
	}

	for _, messages := range m.Messages {
		for _, msg := range messages {
			if msg.Name == messageName {
				return msg, nil
			}
		}
	}
	return nil, fmt.Errorf("message not found: %s", messageName)
}

// SendMessage sends a mock message to a space.
func (m *MockChatService) SendMessage(ctx context.Context, spaceName string, text string, threadName string) (*chatapi.Message, error) {
	m.Calls.SendMessage = append(m.Calls.SendMessage, struct{ SpaceName, Text, ThreadName string }{spaceName, text, threadName})

	if m.Errors.SendMessage != nil {
		return nil, m.Errors.SendMessage
	}

	if _, ok := m.Spaces[spaceName]; !ok {
		return nil, fmt.Errorf("space not found: %s", spaceName)
	}

	msgName := fmt.Sprintf("%s/messages/new-msg-%d", spaceName, len(m.Messages[spaceName])+1)
	msg := &chatapi.Message{
		Name:       msgName,
		Text:       text,
		CreateTime: "2024-01-15T12:00:00.000Z",
		Sender: &chatapi.User{
			Name:        "users/me",
			DisplayName: "Me",
		},
	}
	if threadName != "" {
		msg.Thread = &chatapi.Thread{Name: threadName}
	}
	m.Messages[spaceName] = append(m.Messages[spaceName], msg)
	return msg, nil
}

// CreateReaction adds a mock reaction to a message.
func (m *MockChatService) CreateReaction(ctx context.Context, messageName string, emoji string) (*chatapi.Reaction, error) {
	m.Calls.CreateReaction = append(m.Calls.CreateReaction, struct{ MessageName, Emoji string }{messageName, emoji})

	if m.Errors.CreateReaction != nil {
		return nil, m.Errors.CreateReaction
	}

	reactionName := fmt.Sprintf("%s/reactions/new-reaction-%d", messageName, len(m.Reactions[messageName])+1)
	reaction := &chatapi.Reaction{
		Name: reactionName,
		Emoji: &chatapi.Emoji{
			Unicode: emoji,
		},
		User: &chatapi.User{
			Name:        "users/me",
			DisplayName: "Me",
		},
	}
	m.Reactions[messageName] = append(m.Reactions[messageName], reaction)
	return reaction, nil
}

// DeleteReaction removes a mock reaction.
func (m *MockChatService) DeleteReaction(ctx context.Context, reactionName string) error {
	m.Calls.DeleteReaction = append(m.Calls.DeleteReaction, reactionName)

	if m.Errors.DeleteReaction != nil {
		return m.Errors.DeleteReaction
	}

	return nil
}

// ListMembers lists mock members of a space.
func (m *MockChatService) ListMembers(ctx context.Context, spaceName string, pageSize int64, pageToken string) ([]*chatapi.Membership, string, error) {
	m.Calls.ListMembers = append(m.Calls.ListMembers, spaceName)

	if m.Errors.ListMembers != nil {
		return nil, "", m.Errors.ListMembers
	}

	if _, ok := m.Spaces[spaceName]; !ok {
		return nil, "", fmt.Errorf("space not found: %s", spaceName)
	}

	members := m.Members[spaceName]
	if members == nil {
		members = []*chatapi.Membership{}
	}
	return members, "", nil
}
