package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"forsaken-mail/internal/i18n"
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
			lang := i18n.LangFromRequest(r)
			http.Error(w, i18n.T(lang, "internal_server_error"), http.StatusInternalServerError)
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
				lang := i18n.LangFromRequest(r)
				http.Error(w, i18n.T(lang, "forbidden"), http.StatusForbidden)
				return
			}
		}

		ctx := context.WithValue(r.Context(), emailKey, session.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth reads the session cookie and sets the email in context if present,
// but does NOT reject unauthenticated requests. Use for public routes that
// need to know the user identity when available (e.g. /api/config).
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.sessions.GetCookie(r)
		if err == nil && time.Now().Before(session.ExpiresAt) {
			ctx := context.WithValue(r.Context(), emailKey, session.Email)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func GetEmail(r *http.Request) string {
	email, _ := r.Context().Value(emailKey).(string)
	return email
}
