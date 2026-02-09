package sheets

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// SheetsTestFixtures provides pre-configured test data for Sheets tests.
type SheetsTestFixtures struct {
	DefaultEmail string
	MockService  *MockSheetsService
	Deps         *SheetsHandlerDeps
}

// NewSheetsTestFixtures creates a new test fixtures instance with default configuration.
func NewSheetsTestFixtures() *SheetsTestFixtures {
	mockService := NewMockSheetsService()
	f := common.NewTestFixtures[SheetsService](mockService)

	return &SheetsTestFixtures{
		DefaultEmail: f.DefaultEmail,
		MockService:  mockService,
		Deps:         f.Deps,
	}
}
