package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jesperpedersen/picky-claude/internal/db"
)

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       string `json:"id"`
		Project  string `json:"project"`
		Metadata string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Metadata == "" {
		req.Metadata = "{}"
	}

	if err := s.db.InsertSession(&db.Session{
		ID:       req.ID,
		Project:  req.Project,
		Metadata: req.Metadata,
	}); err != nil {
		s.logger.Error("insert session", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": req.ID})
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.db.ListActiveSessions()
	if err != nil {
		s.logger.Error("list sessions", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, sessions)
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sess, err := s.db.GetSession(id)
	if err != nil {
		s.logger.Error("get session", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if sess == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleEndSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.db.EndSession(id); err != nil {
		s.logger.Error("end session", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ended"})
}

func (s *Server) handleIncrementMessageCount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.db.IncrementMessageCount(id); err != nil {
		s.logger.Error("increment message count", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "incremented"})
}

func (s *Server) handleCleanupSessions(w http.ResponseWriter, r *http.Request) {
	count, err := s.db.CleanupStaleSessions(24)
	if err != nil {
		s.logger.Error("cleanup sessions", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"cleaned": count})
}

func (s *Server) handleCreateSummary(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Text      string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	id, err := s.db.InsertSummary(&db.Summary{
		SessionID: req.SessionID,
		Text:      req.Text,
	})
	if err != nil {
		s.logger.Error("insert summary", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (s *Server) handleRecentSummaries(w http.ResponseWriter, r *http.Request) {
	limit := int(parseID(r.URL.Query().Get("limit")))
	results, err := s.db.RecentSummaries(limit)
	if err != nil {
		s.logger.Error("recent summaries", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, results)
}
