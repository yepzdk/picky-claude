// Package console provides the HTTP server for memory, sessions, MCP, and the
// web viewer. It is the central coordination point — hooks and CLI commands
// communicate with it via HTTP.
package console

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jesperpedersen/picky-claude/internal/assets"
	"github.com/jesperpedersen/picky-claude/internal/config"
	"github.com/jesperpedersen/picky-claude/internal/db"
	"github.com/jesperpedersen/picky-claude/internal/search"
	"github.com/mark3labs/mcp-go/server"
)

// Server is the console HTTP server.
type Server struct {
	port          int
	logger        *slog.Logger
	db            *db.DB
	search        *search.Orchestrator
	http          *http.Server
	router        chi.Router
	sse           *Broadcaster
	stopRetention func() // stops background retention scheduler
}

// New creates a console server on the given port. It opens (or creates) the
// SQLite database and registers all routes.
func New(port int, logger *slog.Logger) (*Server, error) {
	database, err := db.Open(config.DBPath(), logger)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	s := &Server{
		port:   port,
		logger: logger,
		db:     database,
		router: r,
		sse:    NewBroadcaster(),
	}

	// Initialize hybrid search (optional — falls back to FTS-only)
	if orch, err := search.NewOrchestrator(database); err == nil {
		s.search = orch
	} else {
		logger.Warn("hybrid search unavailable, using FTS only", "error", err)
	}

	// Start background retention scheduler
	ret := search.NewRetention(database)
	s.stopRetention = ret.StartScheduler(search.DefaultRetentionConfig())

	s.registerRoutes()

	s.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s, nil
}

// NewWithDB creates a console server with an externally provided database.
// Useful for testing.
func NewWithDB(port int, logger *slog.Logger, database *db.DB) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	s := &Server{
		port:   port,
		logger: logger,
		db:     database,
		router: r,
		sse:    NewBroadcaster(),
	}

	if orch, err := search.NewOrchestrator(database); err == nil {
		s.search = orch
	}

	s.registerRoutes()

	s.http = &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

func (s *Server) registerRoutes() {
	s.router.Get("/health", s.handleHealth)

	s.router.Route("/api", func(r chi.Router) {
		r.Get("/events", s.handleSSE)

		r.Post("/observations", s.handleCreateObservation)
		r.Get("/observations/{id}", s.handleGetObservation)
		r.Get("/observations/search", s.handleSearchObservations)
		r.Get("/observations/hybrid-search", s.handleHybridSearch)
		r.Post("/search/reindex", s.handleReindex)
		r.Get("/observations/timeline/{id}", s.handleTimeline)

		r.Post("/sessions", s.handleCreateSession)
		r.Get("/sessions", s.handleListSessions)
		r.Post("/sessions/cleanup", s.handleCleanupSessions)
		r.Get("/sessions/{id}", s.handleGetSession)
		r.Post("/sessions/{id}/end", s.handleEndSession)
		r.Post("/sessions/{id}/message-count", s.handleIncrementMessageCount)

		r.Post("/summaries", s.handleCreateSummary)
		r.Get("/summaries/recent", s.handleRecentSummaries)

		r.Post("/plans", s.handleCreatePlan)
		r.Get("/plans/by-path", s.handleGetPlanByPath)
		r.Patch("/plans/{id}/status", s.handleUpdatePlanStatus)

		r.Get("/context/inject", s.handleContextInject)
	})

	// Mount MCP server at /mcp
	mcpSrv := s.newMCPServer()
	streamable := server.NewStreamableHTTPServer(mcpSrv)
	s.router.Handle("/mcp", streamable)
	s.router.Handle("/mcp/*", streamable)

	// Serve embedded web viewer at root
	viewerFS, err := assets.ViewerFS()
	if err == nil {
		s.router.Handle("/*", http.FileServer(http.FS(viewerFS)))
	}
}

// Handler returns the http.Handler for testing.
func (s *Server) Handler() http.Handler {
	return s.router
}

// Port returns the actual port the server is listening on.
// Only valid after Start() has been called and the server is running.
func (s *Server) Port() int {
	return s.port
}

// StartWithReady begins listening and signals readyCh once bound.
// If the configured port is busy, it tries up to 10 consecutive ports.
// readyCh can be nil if no notification is needed.
func (s *Server) StartWithReady(readyCh chan<- struct{}) error {
	const maxAttempts = 10
	for i := range maxAttempts {
		tryPort := s.port + i
		addr := fmt.Sprintf(":%d", tryPort)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Debug("port busy, trying next", "port", tryPort, "error", err)
			continue
		}
		if tryPort != s.port {
			s.logger.Info("default port busy, using alternative", "wanted", s.port, "actual", tryPort)
			s.port = tryPort
			s.http.Addr = addr
		}
		s.logger.Debug("console server started", "port", tryPort)
		if readyCh != nil {
			close(readyCh)
		}
		return s.http.Serve(ln)
	}
	return fmt.Errorf("no available port in range %d–%d", s.port, s.port+maxAttempts-1)
}

// Start begins listening. If the configured port is busy, it tries up to
// 10 consecutive ports before giving up.
func (s *Server) Start() error {
	return s.StartWithReady(nil)
}

// Stop gracefully shuts down the server with a 5-second deadline.
func (s *Server) Stop() error {
	if s.stopRetention != nil {
		s.stopRetention()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.logger.Debug("console server stopping")
	if err := s.http.Shutdown(ctx); err != nil {
		s.db.Close()
		return err
	}
	return s.db.Close()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func parseID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
