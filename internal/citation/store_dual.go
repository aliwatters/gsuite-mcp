package citation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/api/sheets/v4"
)

// DualStore writes to both Sheets (source of truth) and SQLite (fast local cache).
// Reads come from SQLite. If the SQLite cache is missing, it rebuilds from Sheets.
type DualStore struct {
	indexID string
	sheets  *SheetsStore
	sqlite  *SQLiteStore
}

// cacheDir returns the SQLite cache directory.
func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".cache", "gsuite-mcp", "citation"), nil
}

// NewDualStore creates a DualStore, rebuilding the SQLite cache from Sheets if needed.
func NewDualStore(ctx context.Context, indexID, sheetID string, sheetsSrv *sheets.Service) (*DualStore, error) {
	sheetsStore := NewSheetsStore(sheetsSrv, sheetID)

	dir, err := cacheDir()
	if err != nil {
		return nil, fmt.Errorf("resolving cache dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	dbPath := filepath.Join(dir, indexID+".db")
	needsRebuild := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		needsRebuild = true
	}

	sqliteStore, err := NewSQLiteStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	ds := &DualStore{
		indexID: indexID,
		sheets:  sheetsStore,
		sqlite:  sqliteStore,
	}

	if needsRebuild {
		if err := ds.rebuild(ctx); err != nil {
			sqliteStore.Close()
			return nil, fmt.Errorf("rebuilding cache: %w", err)
		}
	}

	return ds, nil
}

// rebuild pulls all data from Sheets and populates SQLite.
func (d *DualStore) rebuild(ctx context.Context) error {
	chunks, err := d.sheets.ReadAllChunks(ctx)
	if err != nil {
		return fmt.Errorf("reading chunks from sheet: %w", err)
	}
	concepts, err := d.sheets.ReadAllConcepts(ctx)
	if err != nil {
		return fmt.Errorf("reading concepts from sheet: %w", err)
	}
	summaries, err := d.sheets.ReadAllSummaries(ctx)
	if err != nil {
		return fmt.Errorf("reading summaries from sheet: %w", err)
	}
	files, err := d.sheets.ReadAllFiles(ctx)
	if err != nil {
		return fmt.Errorf("reading files from sheet: %w", err)
	}
	meta, err := d.sheets.ReadAllMetadata(ctx)
	if err != nil {
		return fmt.Errorf("reading metadata from sheet: %w", err)
	}

	return d.sqlite.BulkInsertFromSheet(chunks, concepts, summaries, files, meta)
}

// SaveChunks writes to Sheets first, then SQLite.
func (d *DualStore) SaveChunks(ctx context.Context, chunks []Chunk) error {
	if err := d.sheets.AppendChunks(ctx, chunks); err != nil {
		return fmt.Errorf("sheets write: %w", err)
	}
	return d.sqlite.SaveChunks(ctx, chunks)
}

// GetChunks reads from SQLite.
func (d *DualStore) GetChunks(ctx context.Context, ids []string) ([]Chunk, error) {
	return d.sqlite.GetChunks(ctx, ids)
}

// DeleteChunksByFileID removes chunks for a file from both stores.
func (d *DualStore) DeleteChunksByFileID(ctx context.Context, fileID string) error {
	if err := d.sheets.RewriteChunksForFile(ctx, fileID); err != nil {
		return fmt.Errorf("sheets delete: %w", err)
	}
	return d.sqlite.DeleteChunksByFileID(ctx, fileID)
}

// SaveConcepts writes to Sheets first, then SQLite.
func (d *DualStore) SaveConcepts(ctx context.Context, mappings []ConceptMapping) error {
	if err := d.sheets.AppendConcepts(ctx, mappings); err != nil {
		return fmt.Errorf("sheets write: %w", err)
	}
	return d.sqlite.SaveConcepts(ctx, mappings)
}

// GetConcepts reads from SQLite.
func (d *DualStore) GetConcepts(ctx context.Context) ([]ConceptMapping, error) {
	return d.sqlite.GetConcepts(ctx)
}

// SaveSummary writes to Sheets first, then SQLite.
func (d *DualStore) SaveSummary(ctx context.Context, summary LevelSummary) error {
	if err := d.sheets.AppendSummary(ctx, summary); err != nil {
		return fmt.Errorf("sheets write: %w", err)
	}
	return d.sqlite.SaveSummary(ctx, summary)
}

// GetSummaries reads from SQLite.
func (d *DualStore) GetSummaries(ctx context.Context, level int) ([]LevelSummary, error) {
	return d.sqlite.GetSummaries(ctx, level)
}

// Search uses SQLite FTS5.
func (d *DualStore) Search(ctx context.Context, query string, limit int) ([]Chunk, error) {
	return d.sqlite.Search(ctx, query, limit)
}

// SaveIndexedFile writes to Sheets first, then SQLite.
func (d *DualStore) SaveIndexedFile(ctx context.Context, file IndexedFile) error {
	if err := d.sheets.AppendFiles(ctx, []IndexedFile{file}); err != nil {
		return fmt.Errorf("sheets write: %w", err)
	}
	return d.sqlite.SaveIndexedFile(ctx, file)
}

// GetIndexedFiles reads from SQLite.
func (d *DualStore) GetIndexedFiles(ctx context.Context) ([]IndexedFile, error) {
	return d.sqlite.GetIndexedFiles(ctx)
}

// DeleteIndexedFile removes a file tracking record from both stores.
func (d *DualStore) DeleteIndexedFile(ctx context.Context, fileID string) error {
	// SQLite delete is fast; Sheets rewrite happens during refresh bulk operations
	return d.sqlite.DeleteIndexedFile(ctx, fileID)
}

// GetMetadata reads from SQLite.
func (d *DualStore) GetMetadata(ctx context.Context) (*IndexInfo, error) {
	return d.sqlite.GetMetadata(ctx)
}

// SetMetadata writes to Sheets first, then SQLite.
func (d *DualStore) SetMetadata(ctx context.Context, key, value string) error {
	if err := d.sheets.UpdateMetadata(ctx, key, value); err != nil {
		return fmt.Errorf("sheets write: %w", err)
	}
	return d.sqlite.SetMetadata(ctx, key, value)
}

// Close releases SQLite resources.
func (d *DualStore) Close() error {
	return d.sqlite.Close()
}
