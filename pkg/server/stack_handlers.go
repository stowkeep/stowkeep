package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/stowkeep/stowkeep/pkg/audit"
	"github.com/stowkeep/stowkeep/pkg/auth"
	"github.com/stowkeep/stowkeep/pkg/compose"
	"github.com/stowkeep/stowkeep/pkg/docker"
	applog "github.com/stowkeep/stowkeep/pkg/observability/log"
	"github.com/stowkeep/stowkeep/pkg/rbac"
)

// StackHandler serves stack deploy and lifecycle endpoints.
type StackHandler struct {
	docker docker.SwarmWriter
	audit  *audit.Store
	authz  rbac.Authorizer
}

// NewStackHandler returns a stack HTTP handler.
func NewStackHandler(d docker.SwarmWriter, auditStore *audit.Store, authz rbac.Authorizer) *StackHandler {
	if authz == nil {
		authz = rbac.AdminOnly{}
	}
	return &StackHandler{docker: d, audit: auditStore, authz: authz}
}

// Routes returns chi routes for stack deploy APIs.
func (h *StackHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/validate", h.validate)
	r.Post("/", h.deploy)
	r.Delete("/{name}", h.remove)
	r.Patch("/services/{id}/scale", h.scaleService)
	r.Get("/services/{id}/logs", h.serviceLogs)
	return r
}

type validateRequest struct {
	Name    string `json:"name"`
	Compose string `json:"compose"`
}

type deployRequest struct {
	Name    string            `json:"name"`
	Compose string            `json:"compose"`
	Env     map[string]string `json:"env,omitempty"`
}

type scaleRequest struct {
	Replicas uint64 `json:"replicas"`
}

const maxStackRequestBytes = compose.MaxFileSize + 1024

func decodeStackJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxStackRequestBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return err
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return err
	}
	return nil
}

func (h *StackHandler) validate(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || !h.authz.Allow(r.Context(), user, "swarm.stacks.deploy", "stack", "") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	var req validateRequest
	if err := decodeStackJSON(w, r, &req); err != nil {
		return
	}
	result := compose.Validate(r.Context(), []byte(req.Compose), req.Name)
	writeJSON(w, http.StatusOK, result)
}

func (h *StackHandler) deploy(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok || !h.authz.Allow(r.Context(), user, "swarm.stacks.deploy", "stack", "") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}

	var req deployRequest
	if err := decodeStackJSON(w, r, &req); err != nil {
		return
	}
	content := []byte(req.Compose)
	result := compose.Validate(r.Context(), content, req.Name)
	if !result.Valid {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "compose validation failed",
			"errors": result.Errors,
		})
		return
	}

	if err := h.docker.DeployStack(r.Context(), req.Name, content, req.Env); err != nil {
		slog.ErrorContext(r.Context(), "stack deploy failed",
			slog.String("component", "stacks"),
			slog.String("stack", req.Name),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "docker deploy failed"})
		return
	}

	h.writeAudit(r, audit.Event{
		ActorID: user.ID, Action: "stack.deploy", ResourceType: "stack", ResourceID: req.Name,
		RequestID: applog.RequestIDFromContext(r.Context()), AfterHash: result.Hash,
	})

	detail, err := h.docker.GetStack(r.Context(), req.Name)
	if err != nil {
		if errors.Is(err, docker.ErrStackNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"name": req.Name, "services": []any{}})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "docker request failed"})
		return
	}
	writeJSON(w, http.StatusCreated, detail)
}

func (h *StackHandler) remove(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	name := chi.URLParam(r, "name")
	if !ok || !h.authz.Allow(r.Context(), user, "swarm.stacks.remove", "stack", name) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}
	if err := compose.ValidateStackName(name); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.docker.RemoveStack(r.Context(), name); err != nil {
		slog.ErrorContext(r.Context(), "stack remove failed",
			slog.String("component", "stacks"),
			slog.String("stack", name),
			slog.String("error", err.Error()),
		)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "docker remove failed"})
		return
	}
	h.writeAudit(r, audit.Event{
		ActorID: user.ID, Action: "stack.remove", ResourceType: "stack", ResourceID: name,
		RequestID: applog.RequestIDFromContext(r.Context()),
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func (h *StackHandler) scaleService(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	serviceID := chi.URLParam(r, "id")
	if !ok || !h.authz.Allow(r.Context(), user, "swarm.services.scale", "service", serviceID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}
	var req scaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if err := h.docker.ScaleService(r.Context(), serviceID, req.Replicas); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "scale failed"})
		return
	}
	payload := compose.ContentHash([]byte(strconv.FormatUint(req.Replicas, 10)))
	h.writeAudit(r, audit.Event{
		ActorID: user.ID, Action: "service.scale", ResourceType: "service", ResourceID: serviceID,
		RequestID: applog.RequestIDFromContext(r.Context()), AfterHash: payload,
	})
	writeJSON(w, http.StatusOK, map[string]any{"service_id": serviceID, "replicas": req.Replicas})
}

func (h *StackHandler) serviceLogs(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	serviceID := chi.URLParam(r, "id")
	if !ok || !h.authz.Allow(r.Context(), user, "swarm.services.logs", "service", serviceID) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}
	if h.docker == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "docker unavailable"})
		return
	}
	follow := r.URL.Query().Get("follow") == "true"
	tail := r.URL.Query().Get("tail")
	stream, err := h.docker.ServiceLogs(r.Context(), serviceID, follow, tail)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "logs unavailable"})
		return
	}
	defer func() { _ = stream.Close() }()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, stream); err != nil && !errors.Is(err, context.Canceled) {
		slog.WarnContext(r.Context(), "log stream ended",
			slog.String("component", "stacks"),
			slog.String("error", err.Error()),
		)
	}
}

func (h *StackHandler) writeAudit(r *http.Request, e audit.Event) {
	if h.audit == nil {
		return
	}
	if _, err := h.audit.Append(r.Context(), e); err != nil {
		slog.ErrorContext(r.Context(), "audit append failed",
			slog.String("component", "audit"),
			slog.String("action", e.Action),
			slog.String("error", err.Error()),
		)
	}
}
