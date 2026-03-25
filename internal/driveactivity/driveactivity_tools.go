package driveactivity

import (
	"github.com/aliwatters/gsuite-mcp/internal/common"
)

// === Handle functions - generated via WrapHandler ===

var (
	HandleDriveActivityQuery = common.WrapHandler[DriveActivityService](TestableDriveActivityQuery)
)
