package server

import (
	"net/http"

	"github.com/stowkeep/stowkeep/pkg/config"
)

// RequireFeature gates routes behind a STOWKEEP_FEATURES flag.
func RequireFeature(cfg *config.Config, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.HasFeature(name) {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "feature_disabled"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
