package citation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS chunks (
	id TEXT PRIMARY KEY,
	file_id TEXT NOT NULL,
	file_name TEXT,
	content TEXT NOT NULL,
	summary TEXT,
	page_number INTEGER,
	section_heading TEXT,
	paragraph_index INTEGER,
	char_start INTEGER,
	char_end INTEGER
);

CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
	content, summary, file_name,
	content=chunks, content_rowid=rowid
);

CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
	INSERT INTO chunks_fts(rowid, content, summary, file_name)
	VALUES (new.rowid, new.content, new.summary, new.file_name);
END;

CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
	INSERT INTO chunks_fts(chunks_fts, rowid, content, summary, file_name)
	VALUES ('delete', old.rowid, old.content, old.summary, old.file_name);
END;

CREATE TRIGGER IF NOT EXISTS chunks_au AFTER UPDATE ON chunks BEGIN
	INSERT INTO chunks_fts(chunks_fts, rowid, content, summary, file_name)
	VALUES ('delete', old.rowid, old.content, old.summary, old.file_name);
	INSERT INTO chunks_fts(rowid, content, summary, file_name)
	VALUES (new.rowid, new.content, new.summary, new.file_name);
END;

CREATE TABLE IF NOT EXISTS chunk_concepts (
	chunk_id TEXT NOT NULL,
	concept TEXT NOT NULL,
	PRIMARY KEY (chunk_id, concept)
);

CREATE INDEX IF NOT EXISTS idx_concept ON chunk_concepts(concept);

CREATE TABLE IF NOT EXISTS indexed_files (
	file_id TEXT PRIMARY KEY,
	file_name TEXT,
	mime_type TEXT,
	modified_time TEXT,
	chunk_count INTEGER
);

CREATE TABLE IF NOT EXISTS summaries (
	level INTEGER NOT NULL,
	parent_id TEXT NOT NULL,
	summary TEXT NOT NULL,
	PRIMARY KEY (level, parent_id)
);

CREATE TABLE IF NOT EXISTS metadata (
	key TEXT PRIMARY KEY,
	value TEXT
);
`

// SQLiteStore implements IndexStore backed by a local SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given path.
// Use ":memory:" for in-memory testing.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}
	if _, err := db.Exec(sqliteSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating schema: %w", err)
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) SaveChunks(_ context.Context, chunks []Chunk) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO chunks
		(id, file_id, file_name, content, summary, page_number, section_heading, paragraph_index, char_start, char_end)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing chunk statement: %w", err)
	}
	defer stmt.Close()

	conceptStmt, err := tx.Prepare(`INSERT OR REPLACE INTO chunk_concepts (chunk_id, concept) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing concept statement: %w", err)
	}
	defer conceptStmt.Close()

	for _, c := range chunks {
		_, err := stmt.Exec(c.ID, c.FileID, c.FileName, c.Content, c.Summary,
			c.Location.PageNumber, c.Location.SectionHeading, c.Location.ParagraphIndex,
			c.Location.CharStart, c.Location.CharEnd)
		if err != nil {
			return fmt.Errorf("inserting chunk %s: %w", c.ID, err)
		}
		for _, concept := range c.Concepts {
			if _, err := conceptStmt.Exec(c.ID, concept); err != nil {
				return fmt.Errorf("inserting concept %q for chunk %s: %w", concept, c.ID, err)
			}
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetChunks(_ context.Context, ids []string) ([]Chunk, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, file_id, file_name, content, summary,
		page_number, section_heading, paragraph_index, char_start, char_end
		FROM chunks WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying chunks: %w", err)
	}
	defer rows.Close()

	chunks, err := scanChunks(rows)
	if err != nil {
		return nil, fmt.Errorf("scanning chunks: %w", err)
	}

	for i := range chunks {
		concepts, err := s.getChunkConcepts(chunks[i].ID)
		if err != nil {
			return nil, fmt.Errorf("getting concepts for chunk %s: %w", chunks[i].ID, err)
		}
		chunks[i].Concepts = concepts
	}

	return chunks, nil
}

func (s *SQLiteStore) DeleteChunksByFileID(_ context.Context, fileID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete concepts for chunks belonging to this file
	_, err = tx.Exec(`DELETE FROM chunk_concepts WHERE chunk_id IN (SELECT id FROM chunks WHERE file_id = ?)`, fileID)
	if err != nil {
		return fmt.Errorf("deleting concepts: %w", err)
	}

	// Delete the chunks (triggers handle FTS5 cleanup)
	_, err = tx.Exec(`DELETE FROM chunks WHERE file_id = ?`, fileID)
	if err != nil {
		return fmt.Errorf("deleting chunks: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) getChunkConcepts(chunkID string) ([]string, error) {
	rows, err := s.db.Query(`SELECT concept FROM chunk_concepts WHERE chunk_id = ?`, chunkID)
	if err != nil {
		return nil, fmt.Errorf("querying concepts for chunk %s: %w", chunkID, err)
	}
	defer rows.Close()

	var concepts []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("scanning concept for chunk %s: %w", chunkID, err)
		}
		concepts = append(concepts, c)
	}
	return concepts, rows.Err()
}

func (s *SQLiteStore) SaveConcepts(_ context.Context, mappings []ConceptMapping) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO chunk_concepts (chunk_id, concept) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing concept statement: %w", err)
	}
	defer stmt.Close()

	for _, m := range mappings {
		for _, chunkID := range m.ChunkIDs {
			if _, err := stmt.Exec(chunkID, m.Concept); err != nil {
				return fmt.Errorf("inserting concept %q for chunk %s: %w", m.Concept, chunkID, err)
			}
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetConcepts(_ context.Context) ([]ConceptMapping, error) {
	rows, err := s.db.Query(`SELECT concept, GROUP_CONCAT(chunk_id) FROM chunk_concepts GROUP BY concept ORDER BY concept`)
	if err != nil {
		return nil, fmt.Errorf("querying concepts: %w", err)
	}
	defer rows.Close()

	var mappings []ConceptMapping
	for rows.Next() {
		var concept, chunkIDs string
		if err := rows.Scan(&concept, &chunkIDs); err != nil {
			return nil, fmt.Errorf("scanning concept: %w", err)
		}
		mappings = append(mappings, ConceptMapping{
			Concept:  concept,
			ChunkIDs: strings.Split(chunkIDs, ","),
		})
	}
	return mappings, rows.Err()
}

func (s *SQLiteStore) SaveSummary(_ context.Context, summary LevelSummary) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO summaries (level, parent_id, summary) VALUES (?, ?, ?)`,
		summary.Level, summary.ParentID, summary.Summary)
	if err != nil {
		return fmt.Errorf("saving summary: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSummaries(_ context.Context, level int) ([]LevelSummary, error) {
	rows, err := s.db.Query(`SELECT level, parent_id, summary FROM summaries WHERE level = ?`, level)
	if err != nil {
		return nil, fmt.Errorf("querying summaries: %w", err)
	}
	defer rows.Close()

	var summaries []LevelSummary
	for rows.Next() {
		var ls LevelSummary
		if err := rows.Scan(&ls.Level, &ls.ParentID, &ls.Summary); err != nil {
			return nil, fmt.Errorf("scanning summary: %w", err)
		}
		summaries = append(summaries, ls)
	}
	return summaries, rows.Err()
}

// sanitizeFTS5Query wraps user input in double quotes to prevent FTS5 syntax injection.
// FTS5 metacharacters (AND, OR, NOT, NEAR, *, column filters) are neutralized.
func sanitizeFTS5Query(query string) string {
	escaped := strings.ReplaceAll(query, `"`, `""`)
	return `"` + escaped + `"`
}

func (s *SQLiteStore) Search(_ context.Context, query string, limit int) ([]Chunk, error) {
	sanitized := sanitizeFTS5Query(query)
	rows, err := s.db.Query(`SELECT c.id, c.file_id, c.file_name, c.content, c.summary,
		c.page_number, c.section_heading, c.paragraph_index, c.char_start, c.char_end
		FROM chunks_fts f
		JOIN chunks c ON c.rowid = f.rowid
		WHERE chunks_fts MATCH ?
		ORDER BY rank
		LIMIT ?`, sanitized, limit)
	if err != nil {
		return nil, fmt.Errorf("FTS5 search: %w", err)
	}
	defer rows.Close()

	return scanChunks(rows)
}

func (s *SQLiteStore) SaveIndexedFile(_ context.Context, file IndexedFile) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO indexed_files (file_id, file_name, mime_type, modified_time, chunk_count)
		VALUES (?, ?, ?, ?, ?)`, file.FileID, file.FileName, file.MimeType, file.ModifiedTime, file.ChunkCount)
	if err != nil {
		return fmt.Errorf("saving indexed file %s: %w", file.FileID, err)
	}
	return nil
}

func (s *SQLiteStore) GetIndexedFiles(_ context.Context) ([]IndexedFile, error) {
	rows, err := s.db.Query(`SELECT file_id, file_name, mime_type, modified_time, chunk_count FROM indexed_files`)
	if err != nil {
		return nil, fmt.Errorf("querying indexed files: %w", err)
	}
	defer rows.Close()

	var files []IndexedFile
	for rows.Next() {
		var f IndexedFile
		if err := rows.Scan(&f.FileID, &f.FileName, &f.MimeType, &f.ModifiedTime, &f.ChunkCount); err != nil {
			return nil, fmt.Errorf("scanning indexed file: %w", err)
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (s *SQLiteStore) DeleteIndexedFile(_ context.Context, fileID string) error {
	_, err := s.db.Exec(`DELETE FROM indexed_files WHERE file_id = ?`, fileID)
	if err != nil {
		return fmt.Errorf("deleting indexed file %s: %w", fileID, err)
	}
	return nil
}

func (s *SQLiteStore) GetMetadata(_ context.Context) (*IndexInfo, error) {
	rows, err := s.db.Query(`SELECT key, value FROM metadata`)
	if err != nil {
		return nil, fmt.Errorf("querying metadata: %w", err)
	}
	defer rows.Close()

	info := &IndexInfo{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scanning metadata: %w", err)
		}
		switch key {
		case "index_id":
			info.IndexID = value
		case "sheet_id":
			info.SheetID = value
		case "source_folder_id":
			info.SourceFolderID = value
		case "created_at":
			info.CreatedAt = value
		case "doc_count":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("parsing doc_count %q: %w", value, err)
			}
			info.DocCount = n
		case "chunk_count":
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("parsing chunk_count %q: %w", value, err)
			}
			info.ChunkCount = n
		}
	}
	return info, rows.Err()
}

func (s *SQLiteStore) SetMetadata(_ context.Context, key, value string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?)`, key, value)
	if err != nil {
		return fmt.Errorf("setting metadata %q: %w", key, err)
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// BulkInsertFromSheet populates the SQLite database from Sheet data.
func (s *SQLiteStore) BulkInsertFromSheet(chunks []Chunk, concepts []ConceptMapping, summaries []LevelSummary, files []IndexedFile, meta map[string]string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning bulk transaction: %w", err)
	}
	defer tx.Rollback()

	chunkStmt, err := tx.Prepare(`INSERT OR REPLACE INTO chunks
		(id, file_id, file_name, content, summary, page_number, section_heading, paragraph_index, char_start, char_end)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing bulk chunk statement: %w", err)
	}
	defer chunkStmt.Close()

	for _, c := range chunks {
		if _, err := chunkStmt.Exec(c.ID, c.FileID, c.FileName, c.Content, c.Summary,
			c.Location.PageNumber, c.Location.SectionHeading, c.Location.ParagraphIndex,
			c.Location.CharStart, c.Location.CharEnd); err != nil {
			return fmt.Errorf("bulk inserting chunk %s: %w", c.ID, err)
		}
	}

	conceptStmt, err := tx.Prepare(`INSERT OR REPLACE INTO chunk_concepts (chunk_id, concept) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing bulk concept statement: %w", err)
	}
	defer conceptStmt.Close()

	for _, m := range concepts {
		for _, chunkID := range m.ChunkIDs {
			if _, err := conceptStmt.Exec(chunkID, m.Concept); err != nil {
				return fmt.Errorf("bulk inserting concept %q for chunk %s: %w", m.Concept, chunkID, err)
			}
		}
	}

	sumStmt, err := tx.Prepare(`INSERT OR REPLACE INTO summaries (level, parent_id, summary) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing bulk summary statement: %w", err)
	}
	defer sumStmt.Close()

	for _, ls := range summaries {
		if _, err := sumStmt.Exec(ls.Level, ls.ParentID, ls.Summary); err != nil {
			return fmt.Errorf("bulk inserting summary level %d: %w", ls.Level, err)
		}
	}

	fileStmt, err := tx.Prepare(`INSERT OR REPLACE INTO indexed_files (file_id, file_name, mime_type, modified_time, chunk_count)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing bulk file statement: %w", err)
	}
	defer fileStmt.Close()

	for _, f := range files {
		if _, err := fileStmt.Exec(f.FileID, f.FileName, f.MimeType, f.ModifiedTime, f.ChunkCount); err != nil {
			return fmt.Errorf("bulk inserting file %s: %w", f.FileID, err)
		}
	}

	metaStmt, err := tx.Prepare(`INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing bulk metadata statement: %w", err)
	}
	defer metaStmt.Close()

	for k, v := range meta {
		if _, err := metaStmt.Exec(k, v); err != nil {
			return fmt.Errorf("bulk inserting metadata %q: %w", k, err)
		}
	}

	return tx.Commit()
}

// scanChunks scans rows into Chunk slices.
func scanChunks(rows *sql.Rows) ([]Chunk, error) {
	var chunks []Chunk
	for rows.Next() {
		var c Chunk
		var summary sql.NullString
		if err := rows.Scan(&c.ID, &c.FileID, &c.FileName, &c.Content, &summary,
			&c.Location.PageNumber, &c.Location.SectionHeading, &c.Location.ParagraphIndex,
			&c.Location.CharStart, &c.Location.CharEnd); err != nil {
			return nil, err
		}
		if summary.Valid {
			c.Summary = summary.String
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

// ChunkToSheetRow converts a chunk to a Sheet row.
func ChunkToSheetRow(c Chunk) []any {
	conceptsJSON, _ := json.Marshal(c.Concepts)
	return []any{
		c.ID, c.FileID, c.FileName, c.Content, c.Summary, string(conceptsJSON),
		c.Location.PageNumber, c.Location.SectionHeading, c.Location.ParagraphIndex,
		c.Location.CharStart, c.Location.CharEnd,
	}
}

// SheetRowToChunk parses a Sheet row back into a Chunk.
func SheetRowToChunk(row []any) Chunk {
	getString := func(i int) string {
		if i >= len(row) {
			return ""
		}
		s, _ := row[i].(string)
		return s
	}
	getInt := func(i int) int {
		if i >= len(row) {
			return 0
		}
		switch v := row[i].(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			n, err := strconv.Atoi(v)
			if err != nil {
				log.Printf("citation: SheetRowToChunk: field[%d] %q is not an integer: %v", i, v, err)
			}
			return n
		}
		return 0
	}

	c := Chunk{
		ID:       getString(0),
		FileID:   getString(1),
		FileName: getString(2),
		Content:  getString(3),
		Summary:  getString(4),
		Location: Location{
			PageNumber:     getInt(6),
			SectionHeading: getString(7),
			ParagraphIndex: getInt(8),
			CharStart:      getInt(9),
			CharEnd:        getInt(10),
		},
	}

	if conceptsStr := getString(5); conceptsStr != "" {
		if err := json.Unmarshal([]byte(conceptsStr), &c.Concepts); err != nil {
			log.Printf("citation: warning: malformed concepts JSON for chunk %s: %v", c.ID, err)
		}
	}

	return c
}
