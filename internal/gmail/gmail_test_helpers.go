package gmail

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// GmailTestFixtures provides pre-configured test data for Gmail tests.
type GmailTestFixtures struct {
	DefaultEmail string
	MockService  *MockGmailService
	Deps         *GmailHandlerDeps
}

// NewGmailTestFixtures creates a new test fixtures instance with default configuration.
func NewGmailTestFixtures() *GmailTestFixtures {
	mockService := NewMockGmailService()
	f := common.NewTestFixtures[GmailService](mockService)

	return &GmailTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
