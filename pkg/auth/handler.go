package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// HandlerConfig holds auth HTTP handler settings.
type HandlerConfig struct {
	SessionIdleTTL time.Duration
	CookieSecure   bool
}

// Handler serves auth and setup HTTP endpoints.
type Handler struct {
	store   *Store
	cfg     HandlerConfig
	limiter *loginLimiter
}

// NewHandler returns an auth HTTP handler.
func NewHandler(store *Store, cfg HandlerConfig) *Handler {
	if cfg.SessionIdleTTL <= 0 {
		cfg.SessionIdleTTL = 24 * time.Hour
	}
	return &Handler{
		store:   store,
		cfg:     cfg,
		limiter: newLoginLimiter(10, time.Minute),
	}
}

type setupStatusResponse struct {
	NeedsBootstrap bool `json:"needs_bootstrap"`
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SetupStatus reports whether first-run bootstrap is required.
func (h *Handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	needs, err := h.store.NeedsBootstrap(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "setup status failed",
			slog.String("component", "auth"),
			slog.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, setupStatusResponse{NeedsBootstrap: needs})
}

// SetupAdmin creates the first admin account.
func (h *Handler) SetupAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.store.CreateBootstrapAdmin(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrBootstrapComplete) {
		writeError(w, http.StatusConflict, "bootstrap already completed")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.issueSession(w, r.Context(), user); err != nil {
		slog.ErrorContext(r.Context(), "create session after bootstrap failed",
			slog.String("component", "auth"),
			slog.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

// Login authenticates a user and issues a session cookie.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.limiter.allow(requestClientKey(r.RemoteAddr)) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts")
		return
	}
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.store.Authenticate(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "login failed",
			slog.String("component", "auth"),
			slog.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := h.issueSession(w, r.Context(), user); err != nil {
		slog.ErrorContext(r.Context(), "create session after login failed",
			slog.String("component", "auth"),
			slog.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

// Logout invalidates the current session.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if token, ok := SessionTokenFromRequest(r); ok {
		if err := h.store.DeleteSession(r.Context(), HashToken(token)); err != nil {
			slog.ErrorContext(r.Context(), "logout failed",
				slog.String("component", "auth"),
				slog.String("error", err.Error()),
			)
		}
	}
	ClearSessionCookie(w, h.cfg.CookieSecure)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me returns the authenticated user.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	user, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) issueSession(w http.ResponseWriter, ctx context.Context, user *User) error {
	plain, hash, err := NewSessionToken()
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(h.cfg.SessionIdleTTL)
	if err := h.store.CreateSession(ctx, user.ID, hash, expiresAt); err != nil {
		return err
	}
	SetSessionCookie(w, plain, expiresAt, h.cfg.CookieSecure)
	return nil
}
