package drive

import (
	"context"
	"strings"
)

const maxPathDepth = 20

// PathResolver resolves the folder path for Drive files.
// It caches folder lookups to avoid redundant API calls when
// multiple files share the same parent chain.
type PathResolver struct {
	srv        DriveService
	folderPath map[string]string // folderID → full path
	driveName  map[string]string // driveID → drive name
}

// NewPathResolver creates a PathResolver backed by the given DriveService.
func NewPathResolver(srv DriveService) *PathResolver {
	return &PathResolver{
		srv:        srv,
		folderPath: make(map[string]string),
		driveName:  make(map[string]string),
	}
}

// ResolvePath returns the folder path for a file (e.g. "My Drive/Projects/2025").
// Returns "" if the file has no parents or resolution fails.
func (r *PathResolver) ResolvePath(ctx context.Context, parents []string) string {
	if len(parents) == 0 {
		return ""
	}
	return r.resolveParentPath(ctx, parents[0], 0)
}

// resolveParentPath recursively walks up the parent chain to build the full path.
func (r *PathResolver) resolveParentPath(ctx context.Context, folderID string, depth int) string {
	if depth >= maxPathDepth {
		return ""
	}

	if cached, ok := r.folderPath[folderID]; ok {
		return cached
	}

	parent, err := r.srv.GetFile(ctx, folderID, "id,name,parents,driveId")
	if err != nil {
		return ""
	}

	// Root folder: no parents
	if len(parent.Parents) == 0 {
		var name string
		if parent.DriveId != "" {
			name = r.resolveDriveName(ctx, parent.DriveId)
		} else {
			name = parent.Name
		}
		r.folderPath[folderID] = name
		return name
	}

	// Recurse up the chain
	parentPath := r.resolveParentPath(ctx, parent.Parents[0], depth+1)
	var path string
	if parentPath != "" {
		path = parentPath + "/" + parent.Name
	} else {
		path = parent.Name
	}

	r.folderPath[folderID] = path
	return path
}

// resolveDriveName returns the display name for a shared drive, falling back to the raw ID.
func (r *PathResolver) resolveDriveName(ctx context.Context, driveID string) string {
	if driveID == "" {
		return ""
	}

	if cached, ok := r.driveName[driveID]; ok {
		return cached
	}

	d, err := r.srv.GetDrive(ctx, driveID)
	if err != nil {
		r.driveName[driveID] = driveID
		return driveID
	}

	name := strings.TrimSpace(d.Name)
	if name == "" {
		name = driveID
	}

	r.driveName[driveID] = name
	return name
}
