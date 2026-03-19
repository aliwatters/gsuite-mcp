package slides

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleSlidesGetPresentation = common.WrapHandler[SlidesService](TestableSlidesGetPresentation)
	HandleSlidesGetPage         = common.WrapHandler[SlidesService](TestableSlidesGetPage)
	HandleSlidesGetThumbnail    = common.WrapHandler[SlidesService](TestableSlidesGetThumbnail)
	HandleSlidesCreate          = common.WrapHandler[SlidesService](TestableSlidesCreate)
	HandleSlidesBatchUpdate     = common.WrapHandler[SlidesService](TestableSlidesBatchUpdate)
)
