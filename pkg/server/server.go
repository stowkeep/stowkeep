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

	"github.com/stowkeep/stowkeep/pkg/audit"
	"github.com/stowkeep/stowkeep/pkg/auth"
	"github.com/stowkeep/stowkeep/pkg/config"
	"github.com/stowkeep/stowkeep/pkg/docker"
	"github.com/stowkeep/stowkeep/pkg/http/middleware"
	"github.com/stowkeep/stowkeep/pkg/rbac"
	"github.com/stowkeep/stowkeep/pkg/web"
)

// Server is the Stowkeep HTTP server.
type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	db     *sql.DB
	auth   *auth.Store
	audit  *audit.Store
	docker *docker.Client
	router chi.Router
	http   *http.Server
}

// New creates a Server with routes and middleware configured.
func New(cfg *config.Config, logger *slog.Logger, db *sql.DB) *Server {
	authStore := auth.NewStore(db, cfg.ResolvedDriver())
	authHandler := auth.NewHandler(authStore, auth.HandlerConfig{
		SessionIdleTTL: cfg.SessionIdleTTL,
		CookieSecure:   cfg.CookieSecure,
	})

	var dockerClient *docker.Client
	if cli, err := docker.New(cfg.DockerHost, cfg.DockerTimeout); err != nil {
		logger.Warn("docker client disabled",
			slog.String("component", "swarm"),
			slog.String("error", err.Error()),
		)
	} else {
		dockerClient = cli
	}

	swarmHandler := NewSwarmHandler(dockerClient)
	auditStore := audit.NewStore(db, cfg.ResolvedDriver())
	stackHandler := NewStackHandler(dockerClient, auditStore, rbac.AdminOnly{})

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.AccessLog(logger))

	r.Get("/healthz", healthHandler)
	r.Get("/readyz", readyHandler(db))

	r.Route("/api", func(r chi.Router) {
		r.Get("/v1/version", versionHandler(cfg))

		r.Route("/v1/setup", func(r chi.Router) {
			r.Get("/status", authHandler.SetupStatus)
			r.Post("/admin", authHandler.SetupAdmin)
		})

		r.Post("/v1/auth/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(authStore))
			r.Post("/v1/auth/logout", authHandler.Logout)
			r.Get("/v1/auth/me", authHandler.Me)

			r.Route("/v1/swarm", func(r chi.Router) {
				r.Use(RequireFeature(cfg, "swarm_readonly"))
				r.Mount("/", swarmHandler.Routes())
			})

			r.Route("/v1/stacks", func(r chi.Router) {
				r.Use(RequireFeature(cfg, "stack_deploy"))
				r.Mount("/", stackHandler.Routes())
			})
		})
	})

	spa := web.Handler()
	r.Handle("/*", spa)

	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		auth:   authStore,
		audit:  auditStore,
		docker: dockerClient,
		router: r,
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
		writeJSON(w, http.StatusOK, map[string]any{
			"version":  cfg.Version,
			"service":  "stowkeep",
			"features": cfg.EnabledFeatures(),
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
	return s.router
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	if s.audit != nil {
		audit.StartVerifier(context.Background(), s.audit, func(res audit.VerifyResult) {
			s.logger.Error("audit chain integrity failure",
				slog.String("component", "audit"),
				slog.Int64("break_at_event_id", res.BreakAtEventID),
				slog.String("detail", res.Detail),
			)
			_ = audit.RecordIntegrityBreak(context.Background(), s.db, s.cfg.ResolvedDriver(), res.BreakAtEventID, res.Detail)
		})
	}
	s.logger.Info("feature flags enabled",
		slog.String("component", "config"),
		slog.Any("features", s.cfg.EnabledFeatures()),
	)
	s.logger.Info("starting HTTP server",
		slog.String("component", "http"),
		slog.String("addr", s.cfg.HTTPAddr),
	)
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.docker != nil {
		if err := s.docker.Close(); err != nil {
			s.logger.Warn("docker client close failed", slog.String("error", err.Error()))
		}
	}
	return s.http.Shutdown(ctx)
}
