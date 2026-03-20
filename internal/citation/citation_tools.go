package citation

import "github.com/aliwatters/gsuite-mcp/internal/common"

// Handle vars wrap testable functions for production use.
var (
	HandleCitationCreateIndex    = common.WrapHandler[CitationService](TestableCitationCreateIndex)
	HandleCitationAddDocuments   = common.WrapHandler[CitationService](TestableCitationAddDocuments)
	HandleCitationSaveConcepts   = common.WrapHandler[CitationService](TestableCitationSaveConcepts)
	HandleCitationSaveSummary    = common.WrapHandler[CitationService](TestableCitationSaveSummary)
	HandleCitationListIndexes    = common.WrapHandler[CitationService](TestableCitationListIndexes)
	HandleCitationGetOverview    = common.WrapHandler[CitationService](TestableCitationGetOverview)
	HandleCitationLookup         = common.WrapHandler[CitationService](TestableCitationLookup)
	HandleCitationGetChunks      = common.WrapHandler[CitationService](TestableCitationGetChunks)
	HandleCitationVerifyClaim    = common.WrapHandler[CitationService](TestableCitationVerifyClaim)
	HandleCitationFormatCitation = common.WrapHandler[CitationService](TestableCitationFormatCitation)
	HandleCitationRefresh        = common.WrapHandler[CitationService](TestableCitationRefresh)
)
