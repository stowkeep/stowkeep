// Package server wires the Stowkeep HTTP API and embedded frontend.
package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/http/middleware"
	"github.com/stowkeep/stowkeep/pkg/web"
)

// Server is the Stowkeep HTTP server.
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
	http   *http.Server
}

// New creates a Server with routes and middleware configured.
func New(cfg *config.Config, logger *slog.Logger, db *sql.DB) *Server {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.AccessLog(logger))

	r.Get("/healthz", healthHandler)
	r.Get("/readyz", readyHandler(db))

	r.Route("/api", func(r chi.Router) {
		r.Get("/v1/version", versionHandler(cfg))
	})

	spa := web.Handler()
	r.Handle("/*", spa)

	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		http: &http.Server{
			Addr:              cfg.HTTPAddr,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
	return s
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "not_ready",
				"error":  "database unavailable",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func versionHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"version": cfg.Version,
			"service": "stowkeep",
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Handler returns the root HTTP handler (for tests).
func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	s.logger.Info("starting HTTP server",
		slog.String("component", "http"),
		slog.String("addr", s.cfg.HTTPAddr),
	)
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
