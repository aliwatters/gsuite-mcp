package citation

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
)

// RefreshIndex checks all indexed files against Drive for modifications, renames, or deletions,
// and re-indexes any files that have changed.
func (s *RealCitationService) RefreshIndex(ctx context.Context, indexID string) (*RefreshResult, error) {
	store, err := s.getStore(ctx, indexID)
	if err != nil {
		return nil, err
	}

	// Get currently indexed files
	indexed, err := store.GetIndexedFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting indexed files: %w", err)
	}

	result := &RefreshResult{}

	// Check each indexed file against Drive
	for _, prev := range indexed {
		current, err := s.driveService.Files.Get(prev.FileID).
			Fields("id,name,mimeType,modifiedTime,trashed").
			SupportsAllDrives(true).
			Context(ctx).Do()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", prev.FileName, err))
			continue
		}

		// File was trashed → remove from index
		if current.Trashed {
			if delErr := s.removeFileFromIndex(ctx, store, prev.FileID); delErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("removing %s: %v", prev.FileName, delErr))
			} else {
				result.Removed = append(result.Removed, prev.FileName)
			}
			continue
		}

		// File was renamed
		if current.Name != prev.FileName {
			result.Renamed = append(result.Renamed, fmt.Sprintf("%s → %s", prev.FileName, current.Name))
		}

		// File was modified → re-chunk
		if current.ModifiedTime != prev.ModifiedTime {
			if reErr := s.reindexFile(ctx, store, current); reErr != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("re-indexing %s: %v", current.Name, reErr))
			} else {
				result.Updated = append(result.Updated, current.Name)
			}
		}
	}

	return result, nil
}

// removeFileFromIndex deletes a file's chunks and its tracking record.
func (s *RealCitationService) removeFileFromIndex(ctx context.Context, store *DualStore, fileID string) error {
	if err := store.DeleteChunksByFileID(ctx, fileID); err != nil {
		return fmt.Errorf("deleting chunks for file %s: %w", fileID, err)
	}
	return store.DeleteIndexedFile(ctx, fileID)
}

// reindexFile removes a file's old chunks and re-processes it with the current content.
func (s *RealCitationService) reindexFile(ctx context.Context, store *DualStore, file *drive.File) error {
	// Remove old chunks
	if err := store.DeleteChunksByFileID(ctx, file.Id); err != nil {
		return fmt.Errorf("deleting old chunks: %w", err)
	}

	// Re-chunk
	chunks, err := s.chunkFile(ctx, file)
	if err != nil {
		return fmt.Errorf("chunking: %w", err)
	}

	if err := store.SaveChunks(ctx, chunks); err != nil {
		return fmt.Errorf("saving chunks: %w", err)
	}

	// Update tracking record
	return store.SaveIndexedFile(ctx, IndexedFile{
		FileID:       file.Id,
		FileName:     file.Name,
		MimeType:     file.MimeType,
		ModifiedTime: file.ModifiedTime,
		ChunkCount:   len(chunks),
	})
}
