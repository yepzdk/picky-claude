package search

import (
	"sort"

	"github.com/jesperpedersen/picky-claude/internal/db"
)

// Weights controls the blend between FTS5 and vector search scores.
type Weights struct {
	FTS    float64 // Weight for FTS5 keyword search (0-1)
	Vector float64 // Weight for vector similarity search (0-1)
}

// DefaultWeights returns the default search weights.
func DefaultWeights() Weights {
	return Weights{FTS: 0.4, Vector: 0.6}
}

// SearchQuery describes a hybrid search request.
type SearchQuery struct {
	Text    string
	Type    string
	Project string
	Limit   int
}

// HybridResult is a merged search result with a combined score.
type HybridResult struct {
	ID        int64   `json:"id"`
	Score     float64 `json:"score"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	ObsType   string  `json:"type"`
	Project   string  `json:"project"`
	SessionID string  `json:"session_id"`
}

// Orchestrator coordinates FTS5 and vector search.
type Orchestrator struct {
	db      *db.DB
	vector  *VectorStore
	weights Weights
}

// NewOrchestrator creates a hybrid search orchestrator.
func NewOrchestrator(database *db.DB) (*Orchestrator, error) {
	vs, err := NewVectorStore(database)
	if err != nil {
		return nil, err
	}
	return &Orchestrator{
		db:      database,
		vector:  vs,
		weights: DefaultWeights(),
	}, nil
}

// SetWeights configures the FTS/vector weight blend.
func (o *Orchestrator) SetWeights(w Weights) {
	o.weights = w
}

// RebuildIndex rebuilds the vector search index from all observations.
func (o *Orchestrator) RebuildIndex() error {
	return o.vector.IndexAll()
}

// Search performs a hybrid search combining FTS5 and vector similarity.
func (o *Orchestrator) Search(q SearchQuery) ([]HybridResult, error) {
	if q.Limit <= 0 {
		q.Limit = 20
	}

	// Collect results by observation ID for merging
	merged := make(map[int64]*HybridResult)

	// 1. FTS5 keyword search
	ftsResults, err := o.db.FilteredSearch(db.SearchFilter{
		Query:   q.Text,
		Type:    q.Type,
		Project: q.Project,
		Limit:   q.Limit * 2, // Fetch extra for merging
	})
	if err == nil && len(ftsResults) > 0 {
		// Normalize FTS scores: best rank = 1.0, worst = 0.0
		maxRank := 1.0
		for i, r := range ftsResults {
			// FTS5 rank is negative (more negative = better match)
			// Position-based scoring: first result = best
			score := 1.0 - float64(i)/float64(len(ftsResults))
			if score > maxRank {
				maxRank = score
			}
			merged[r.ID] = &HybridResult{
				ID:        r.ID,
				Score:     score * o.weights.FTS,
				Title:     r.Title,
				Text:      r.Text,
				ObsType:   r.Type,
				Project:   r.Project,
				SessionID: r.SessionID,
			}
		}
	}

	// 2. Vector similarity search
	vecResults, err := o.vector.Search(q.Text, q.Limit*2)
	if err == nil && len(vecResults) > 0 {
		for _, r := range vecResults {
			// Apply type/project filters
			if q.Type != "" && r.ObsType != q.Type {
				continue
			}
			if q.Project != "" && r.Project != q.Project {
				continue
			}

			vectorScore := r.Score * o.weights.Vector
			if existing, ok := merged[r.ID]; ok {
				existing.Score += vectorScore
			} else {
				merged[r.ID] = &HybridResult{
					ID:        r.ID,
					Score:     vectorScore,
					Title:     r.Title,
					Text:      r.Text,
					ObsType:   r.ObsType,
					Project:   r.Project,
					SessionID: r.SessionID,
				}
			}
		}
	}

	// 3. Sort by combined score descending
	results := make([]HybridResult, 0, len(merged))
	for _, r := range merged {
		results = append(results, *r)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 4. Truncate to limit
	if len(results) > q.Limit {
		results = results[:q.Limit]
	}

	return results, nil
}
