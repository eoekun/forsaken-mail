package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// csrfMiddleware implements double-submit cookie CSRF protection.
// It sets a csrf_token cookie on every response and validates the
// X-CSRF-Token header on state-changing requests (POST/PUT/DELETE).
func csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CSRF cookie if not already present.
		if _, err := r.Cookie("csrf_token"); err != nil {
			token, err := generateRandomHex(32)
			if err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/",
					HttpOnly: false, // JS needs to read it
					SameSite: http.SameSiteLaxMode,
				})
			}
		}

		// Validate CSRF for state-changing methods.
		method := r.Method
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
			// Exempt WebSocket upgrade and OAuth callbacks.
			path := r.URL.Path
			if path == "/ws" || strings.HasPrefix(path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			csrfCookie, _ := r.Cookie("csrf_token")
			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfCookie == nil || csrfHeader == "" || csrfCookie.Value != csrfHeader {
				writeError(w, http.StatusForbidden, "CSRF token mismatch")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// generateRandomHex generates a random hex string of the given byte length.
func generateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
