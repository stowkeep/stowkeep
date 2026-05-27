package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type contextKey int

const userContextKey contextKey = iota

// UserFromContext returns the authenticated user attached to ctx.
func UserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(userContextKey).(*User)
	return u, ok && u != nil
}

// RequireAuth rejects unauthenticated requests with 401.
func RequireAuth(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := SessionTokenFromRequest(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			user, err := store.GetUserBySessionToken(r.Context(), HashToken(token))
			if err != nil {
				if errors.Is(err, ErrSessionNotFound) {
					writeError(w, http.StatusUnauthorized, "authentication required")
					return
				}
				slog.ErrorContext(r.Context(), "session lookup failed",
					slog.String("component", "auth"),
					slog.String("error", err.Error()),
				)
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth attaches the user to context when a valid session exists.
func OptionalAuth(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := SessionTokenFromRequest(r); ok {
				if user, err := store.GetUserBySessionToken(r.Context(), HashToken(token)); err == nil {
					r = r.WithContext(context.WithValue(r.Context(), userContextKey, user))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
