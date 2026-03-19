package forms

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleFormsGet           = common.WrapHandler[FormsService](TestableFormsGet)
	HandleFormsCreate        = common.WrapHandler[FormsService](TestableFormsCreate)
	HandleFormsBatchUpdate   = common.WrapHandler[FormsService](TestableFormsBatchUpdate)
	HandleFormsListResponses = common.WrapHandler[FormsService](TestableFormsListResponses)
	HandleFormsGetResponse   = common.WrapHandler[FormsService](TestableFormsGetResponse)
)
