package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"forsaken-mail/internal/settings"
)

type contextKey string

const emailKey contextKey = "email"

type Middleware struct {
	sessions *SessionManager
	settings *settings.Store
}

func NewMiddleware(sessions *SessionManager, settings *settings.Store) *Middleware {
	return &Middleware{
		sessions: sessions,
		settings: settings,
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return m.RequireAuth(next)
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.sessions.GetCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			m.sessions.ClearCookie(w)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		allowedEmails, err := m.settings.Get("allowed_emails")
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if allowedEmails != "" {
			found := false
			for _, email := range strings.Split(allowedEmails, ",") {
				if strings.TrimSpace(email) == session.Email {
					found = true
					break
				}
			}
			if !found {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		ctx := context.WithValue(r.Context(), emailKey, session.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetEmail(r *http.Request) string {
	email, _ := r.Context().Value(emailKey).(string)
	return email
}
