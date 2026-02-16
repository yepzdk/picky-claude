package console

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	ctxbuilder "github.com/jesperpedersen/picky-claude/internal/console/context"
	"github.com/jesperpedersen/picky-claude/internal/db"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateObservation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Type      string `json:"type"`
		Title     string `json:"title"`
		Text      string `json:"text"`
		Project   string `json:"project"`
		Metadata  string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	id, err := s.db.InsertObservation(&db.Observation{
		SessionID: req.SessionID,
		Type:      req.Type,
		Title:     req.Title,
		Text:      req.Text,
		Project:   req.Project,
		Metadata:  req.Metadata,
	})
	if err != nil {
		s.logger.Error("insert observation", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	// Broadcast to SSE subscribers
	eventData, _ := json.Marshal(map[string]any{"id": id, "type": req.Type, "title": req.Title})
	s.sse.Send(Event{Type: "observation", Data: string(eventData)})

	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (s *Server) handleGetObservation(w http.ResponseWriter, r *http.Request) {
	id := parseID(chi.URLParam(r, "id"))
	if id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	obs, err := s.db.GetObservation(id)
	if err != nil {
		s.logger.Error("get observation", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if obs == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, obs)
}

func (s *Server) handleSearchObservations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing q parameter"})
		return
	}

	filter := db.SearchFilter{
		Query:     query,
		Type:      r.URL.Query().Get("type"),
		Project:   r.URL.Query().Get("project"),
		DateStart: r.URL.Query().Get("dateStart"),
		DateEnd:   r.URL.Query().Get("dateEnd"),
		Limit:     int(parseID(r.URL.Query().Get("limit"))),
	}

	results, err := s.db.FilteredSearch(filter)
	if err != nil {
		s.logger.Error("search observations", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	id := parseID(chi.URLParam(r, "id"))
	if id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	before := int(parseID(r.URL.Query().Get("before")))
	after := int(parseID(r.URL.Query().Get("after")))

	results, err := s.db.TimelineAround(id, before, after)
	if err != nil {
		s.logger.Error("timeline", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path      string `json:"path"`
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.Status == "" {
		req.Status = "PENDING"
	}

	id, err := s.db.InsertPlan(&db.Plan{
		Path:      req.Path,
		SessionID: req.SessionID,
		Status:    req.Status,
	})
	if err != nil {
		s.logger.Error("insert plan", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (s *Server) handleGetPlanByPath(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing path parameter"})
		return
	}

	plan, err := s.db.GetPlanByPath(path)
	if err != nil {
		s.logger.Error("get plan by path", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	if plan == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func (s *Server) handleContextInject(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	// Get session's project for filtering
	var project string
	if sessionID != "" {
		if sess, err := s.db.GetSession(sessionID); err == nil && sess != nil {
			project = sess.Project
		}
	}

	obs, err := s.db.RecentObservations(project, 50)
	if err != nil {
		s.logger.Error("recent observations", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	summaries, err := s.db.RecentSummaries(10)
	if err != nil {
		s.logger.Error("recent summaries", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	builder := ctxbuilder.NewBuilder(4000)
	ctx := builder.Build(obs, summaries)

	writeJSON(w, http.StatusOK, map[string]string{"context": ctx})
}

func (s *Server) handleUpdatePlanStatus(w http.ResponseWriter, r *http.Request) {
	id := parseID(chi.URLParam(r, "id"))
	if id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if err := s.db.UpdatePlanStatus(id, req.Status); err != nil {
		s.logger.Error("update plan status", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": req.Status})
}
