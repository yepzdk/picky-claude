package console

import (
	"net/http"

	"github.com/jesperpedersen/picky-claude/internal/search"
)

func (s *Server) handleHybridSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing q parameter"})
		return
	}

	if s.search == nil {
		// Fall back to FTS-only search
		s.handleSearchObservations(w, r)
		return
	}

	results, err := s.search.Search(search.SearchQuery{
		Text:    query,
		Type:    r.URL.Query().Get("type"),
		Project: r.URL.Query().Get("project"),
		Limit:   int(parseID(r.URL.Query().Get("limit"))),
	})
	if err != nil {
		s.logger.Error("hybrid search", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleReindex(w http.ResponseWriter, r *http.Request) {
	if s.search == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "search not available"})
		return
	}

	if err := s.search.RebuildIndex(); err != nil {
		s.logger.Error("reindex", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "reindex failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reindexed"})
}
