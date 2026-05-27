package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/stowkeep/stowkeep/pkg/docker"
)

// SwarmHandler serves read-only Swarm API endpoints.
type SwarmHandler struct {
	docker *docker.Client
}

// NewSwarmHandler returns a Swarm HTTP handler.
func NewSwarmHandler(d *docker.Client) *SwarmHandler {
	return &SwarmHandler{docker: d}
}

// Routes returns the chi router for read-only Swarm API endpoints.
func (h *SwarmHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", h.status)
	r.Get("/nodes", h.nodes)
	r.Get("/services", h.services)
	r.Get("/tasks", h.tasks)
	r.Get("/stacks", h.stacks)
	r.Get("/stacks/{name}", h.stackDetail)
	return r
}

func (h *SwarmHandler) status(w http.ResponseWriter, r *http.Request) {
	if h.docker == nil {
		writeJSON(w, http.StatusOK, docker.Status{
			DockerHost: "",
			Error:      "docker client not configured",
		})
		return
	}
	writeJSON(w, http.StatusOK, h.docker.Status(r.Context()))
}

func (h *SwarmHandler) nodes(w http.ResponseWriter, r *http.Request) {
	h.serveList(w, r, func() (any, error) {
		return h.docker.ListNodes(r.Context())
	})
}

func (h *SwarmHandler) services(w http.ResponseWriter, r *http.Request) {
	h.serveList(w, r, func() (any, error) {
		return h.docker.ListServices(r.Context())
	})
}

func (h *SwarmHandler) tasks(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	h.serveList(w, r, func() (any, error) {
		return h.docker.ListTasks(r.Context(), serviceID)
	})
}

func (h *SwarmHandler) stacks(w http.ResponseWriter, r *http.Request) {
	h.serveList(w, r, func() (any, error) {
		return h.docker.ListStacks(r.Context())
	})
}

func (h *SwarmHandler) stackDetail(w http.ResponseWriter, r *http.Request) {
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}
	name := chi.URLParam(r, "name")
	detail, err := h.docker.GetStack(r.Context(), name)
	if err != nil {
		if errors.Is(err, docker.ErrStackNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "stack not found"})
			return
		}
		slog.ErrorContext(r.Context(), "swarm stack lookup failed",
			slog.String("component", "swarm"),
			slog.String("stack", name),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "docker request failed"})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *SwarmHandler) serveList(w http.ResponseWriter, r *http.Request, fn func() (any, error)) {
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}
	items, err := fn()
	if err != nil {
		slog.ErrorContext(r.Context(), "swarm list failed",
			slog.String("component", "swarm"),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "docker request failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
