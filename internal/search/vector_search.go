package search

import (
	"fmt"
	"sort"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

// VectorResult holds a search result with its similarity score.
type VectorResult struct {
	ID         int64
	Score      float64
	Title      string
	Text       string
	ObsType    string
	Project    string
	SessionID  string
}

// VectorStore manages TF-IDF embeddings stored in SQLite alongside observations.
type VectorStore struct {
	db    *db.DB
	vocab *Vocabulary
}

// NewVectorStore creates a vector store backed by the given database.
// It creates the embeddings table if it doesn't exist and rebuilds the
// vocabulary from existing observations.
func NewVectorStore(database *db.DB) (*VectorStore, error) {
	if err := createEmbeddingsTable(database); err != nil {
		return nil, err
	}
	return &VectorStore{db: database}, nil
}

func createEmbeddingsTable(database *db.DB) error {
	_, err := database.Conn().Exec(`
		CREATE TABLE IF NOT EXISTS observation_embeddings (
			observation_id INTEGER PRIMARY KEY,
			embedding BLOB NOT NULL,
			FOREIGN KEY (observation_id) REFERENCES observations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("create embeddings table: %w", err)
	}
	return nil
}

// IndexAll builds a vocabulary from all observations and indexes them.
func (vs *VectorStore) IndexAll() error {
	// Load all observation texts to build vocabulary
	rows, err := vs.db.Conn().Query(
		`SELECT id, title, text FROM observations ORDER BY id`,
	)
	if err != nil {
		return fmt.Errorf("load observations: %w", err)
	}
	defer rows.Close()

	type obsText struct {
		id    int64
		title string
		text  string
	}

	var observations []obsText
	var docs []string
	for rows.Next() {
		var o obsText
		if err := rows.Scan(&o.id, &o.title, &o.text); err != nil {
			return fmt.Errorf("scan observation: %w", err)
		}
		observations = append(observations, o)
		docs = append(docs, o.title+" "+o.text)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate observations: %w", err)
	}

	if len(docs) == 0 {
		return nil
	}

	// Build vocabulary from corpus
	vs.vocab = NewVocabulary(docs)

	// Index each observation
	tx, err := vs.db.Conn().Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.Prepare(
		`INSERT OR REPLACE INTO observation_embeddings (observation_id, embedding) VALUES (?, ?)`,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for i, o := range observations {
		vec := vs.vocab.Embed(docs[i])
		blob := EncodeVector(vec)
		if _, err := stmt.Exec(o.id, blob); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert embedding for %d: %w", o.id, err)
		}
	}

	return tx.Commit()
}

// IndexObservation indexes a single observation. If no vocabulary exists yet,
// it rebuilds from all observations.
func (vs *VectorStore) IndexObservation(id int64) error {
	obs, err := vs.db.GetObservation(id)
	if err != nil {
		return fmt.Errorf("get observation %d: %w", id, err)
	}
	if obs == nil {
		return fmt.Errorf("observation %d not found", id)
	}

	// Rebuild vocabulary if needed (ensures new terms are included)
	if err := vs.IndexAll(); err != nil {
		return err
	}

	return nil
}

// Search finds the top-K most similar observations to the query text.
func (vs *VectorStore) Search(query string, limit int) ([]VectorResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Load all embeddings
	rows, err := vs.db.Conn().Query(`
		SELECT e.observation_id, e.embedding, o.title, o.text, o.type, o.project, o.session_id
		FROM observation_embeddings e
		JOIN observations o ON o.id = e.observation_id
	`)
	if err != nil {
		return nil, fmt.Errorf("load embeddings: %w", err)
	}
	defer rows.Close()

	type stored struct {
		result VectorResult
		vec    []float64
	}

	var entries []stored
	for rows.Next() {
		var s stored
		var blob []byte
		if err := rows.Scan(
			&s.result.ID, &blob,
			&s.result.Title, &s.result.Text, &s.result.ObsType,
			&s.result.Project, &s.result.SessionID,
		); err != nil {
			return nil, fmt.Errorf("scan embedding: %w", err)
		}
		s.vec = DecodeVector(blob)
		entries = append(entries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate embeddings: %w", err)
	}

	if len(entries) == 0 || vs.vocab == nil {
		return nil, nil
	}

	// Embed the query
	queryVec := vs.vocab.Embed(query)

	// Compute similarities
	for i := range entries {
		entries[i].result.Score = CosineSimilarity(queryVec, entries[i].vec)
	}

	// Sort by score descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].result.Score > entries[j].result.Score
	})

	// Return top-K with positive scores
	var results []VectorResult
	for _, e := range entries {
		if e.result.Score <= 0 {
			break
		}
		results = append(results, e.result)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}
