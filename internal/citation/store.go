package citation

import "context"

// IndexStore defines storage operations for a citation index.
type IndexStore interface {
	// SaveChunks persists chunks (upsert).
	SaveChunks(ctx context.Context, chunks []Chunk) error

	// GetChunks retrieves chunks by IDs.
	GetChunks(ctx context.Context, ids []string) ([]Chunk, error)

	// DeleteChunksByFileID removes all chunks (and their concepts) for a file.
	DeleteChunksByFileID(ctx context.Context, fileID string) error

	// SaveConcepts persists concept-to-chunk mappings.
	SaveConcepts(ctx context.Context, mappings []ConceptMapping) error

	// GetConcepts returns all concept mappings.
	GetConcepts(ctx context.Context) ([]ConceptMapping, error)

	// SaveSummary persists a level summary (upsert by level+parent_id).
	SaveSummary(ctx context.Context, summary LevelSummary) error

	// GetSummaries returns summaries at a given level.
	GetSummaries(ctx context.Context, level int) ([]LevelSummary, error)

	// Search performs full-text search across chunk content/summaries.
	Search(ctx context.Context, query string, limit int) ([]Chunk, error)

	// SaveIndexedFile records a file that has been indexed.
	SaveIndexedFile(ctx context.Context, file IndexedFile) error

	// GetIndexedFiles returns all tracked indexed files.
	GetIndexedFiles(ctx context.Context) ([]IndexedFile, error)

	// DeleteIndexedFile removes a file tracking record.
	DeleteIndexedFile(ctx context.Context, fileID string) error

	// GetMetadata returns index metadata.
	GetMetadata(ctx context.Context) (*IndexInfo, error)

	// SetMetadata updates a single metadata key.
	SetMetadata(ctx context.Context, key, value string) error

	// Close releases resources.
	Close() error
}
