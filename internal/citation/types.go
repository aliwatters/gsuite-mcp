package citation

// Chunk represents a text chunk extracted from a source document.
type Chunk struct {
	ID       string   `json:"id"`
	FileID   string   `json:"file_id"`
	FileName string   `json:"file_name"`
	Content  string   `json:"content"`
	Summary  string   `json:"summary,omitempty"`
	Concepts []string `json:"concepts,omitempty"`
	Location Location `json:"location"`
}

// Location identifies where a chunk lives in the source document.
type Location struct {
	PageNumber     int    `json:"page_number,omitempty"`
	SectionHeading string `json:"section_heading,omitempty"`
	ParagraphIndex int    `json:"paragraph_index,omitempty"`
	CharStart      int    `json:"char_start,omitempty"`
	CharEnd        int    `json:"char_end,omitempty"`
}

// ConceptMapping links a concept to the chunks that contain it.
type ConceptMapping struct {
	Concept  string   `json:"concept"`
	ChunkIDs []string `json:"chunk_ids"`
}

// LevelSummary stores a hierarchical summary at a given level.
type LevelSummary struct {
	Level    int    `json:"level"`
	ParentID string `json:"parent_id"`
	Summary  string `json:"summary"`
}

// IndexedFile tracks a file that has been indexed, with its Drive modification time.
type IndexedFile struct {
	FileID       string `json:"file_id"`
	FileName     string `json:"file_name"`
	MimeType     string `json:"mime_type"`
	ModifiedTime string `json:"modified_time"`
	ChunkCount   int    `json:"chunk_count"`
}

// IndexInfo holds metadata about a citation index.
type IndexInfo struct {
	IndexID        string `json:"index_id"`
	SheetID        string `json:"sheet_id"`
	SourceFolderID string `json:"source_folder_id,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	DocCount       int    `json:"doc_count,omitempty"`
	ChunkCount     int    `json:"chunk_count,omitempty"`
}

// RefreshResult summarizes what changed during a refresh.
type RefreshResult struct {
	Updated []string `json:"updated,omitempty"`
	Removed []string `json:"removed,omitempty"`
	Renamed []string `json:"renamed,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}
