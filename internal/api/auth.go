package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/i18n"
)

// handleOAuthLogin handles GET /auth/{provider}/login.
// It generates a random state, stores it in a cookie, and redirects to the OAuth provider.
func (rt *Router) handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	provider := extractProvider(r.URL.Path)
	if provider == "" {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "provider_required"))
		return
	}

	// Verify the requested provider matches the configured one.
	if strings.ToLower(provider) != strings.ToLower(rt.cfg.OAuthProvider) {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "unsupported_provider"))
		return
	}

	state, err := auth.GenerateState()
	if err != nil {
		slog.Error("failed to generate OAuth state", "error", err)
		writeError(w, http.StatusInternalServerError, i18n.T(i18n.LangFromRequest(r), "internal_server_error"))
		return
	}

	// Store state in cookie for verification on callback.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600, // 10 minutes
	})

	redirectURI := buildRedirectURI(r, provider)
	authURL := rt.provider.AuthCodeURL(state, redirectURI)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleOAuthCallback handles GET /auth/{provider}/callback.
// It verifies the state, exchanges the code for a token, gets the user email,
// creates a session, and redirects to /.
func (rt *Router) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	provider := extractProvider(r.URL.Path)
	if provider == "" {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "provider_required"))
		return
	}

	lang := i18n.LangFromRequest(r)

	// Verify state matches cookie.
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "missing_state_cookie"))
		return
	}
	queryState := r.URL.Query().Get("state")
	if queryState != stateCookie.Value {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "state_mismatch"))
		return
	}

	// Clear the state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, i18n.T(lang, "code_required"))
		return
	}

	redirectURI := buildRedirectURI(r, provider)
	token, err := rt.provider.Exchange(r.Context(), code, redirectURI)
	if err != nil {
		slog.Error("OAuth exchange failed", "provider", provider, "error", err)
		writeError(w, http.StatusUnauthorized, i18n.T(lang, "oauth_auth_failed"))
		return
	}

	email, err := rt.provider.GetEmail(r.Context(), token)
	if err != nil {
		slog.Error("failed to get user email", "provider", provider, "error", err)
		writeError(w, http.StatusUnauthorized, i18n.T(lang, "get_email_failed"))
		return
	}

	// Check email whitelist before creating session.
	allowedEmails, err := rt.settings.Get("allowed_emails")
	if err != nil {
		slog.Error("failed to get allowed_emails setting", "error", err)
		writeError(w, http.StatusInternalServerError, i18n.T(lang, "internal_server_error"))
		return
	}
	if allowedEmails != "" {
		allowed := false
		for _, e := range strings.Split(allowedEmails, ",") {
			if strings.TrimSpace(e) == email {
				allowed = true
				break
			}
		}
		if !allowed {
			slog.Warn("OAuth login rejected by email whitelist", "email", email, "provider", provider)
			ip := r.RemoteAddr
			if auditErr := rt.auditStore.Record("LOGIN_REJECTED", email, `{"reason":"email not in whitelist"}`, ip); auditErr != nil {
				slog.Error("failed to record rejected login audit", "error", auditErr)
			}
			http.Redirect(w, r, "/login?error=unauthorized_email", http.StatusFound)
			return
		}
	}

	// Create session.
	rt.sessions.SetCookie(w, &auth.SessionData{
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	// Record audit event.
	ip := r.RemoteAddr
	if auditErr := rt.auditStore.Record(auth.EventLogin, email, "{}", ip); auditErr != nil {
		slog.Error("failed to record login audit", "error", auditErr)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// handleLogout handles GET /auth/logout.
// It clears the session cookie and redirects to /login.
func (rt *Router) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	email := auth.GetEmail(r)
	rt.sessions.ClearCookie(w)

	// Record audit event.
	if email != "" {
		ip := r.RemoteAddr
		if err := rt.auditStore.Record(auth.EventLogout, email, "{}", ip); err != nil {
			slog.Error("failed to record logout audit", "error", err)
		}
	}

	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

// handleLocalLogin handles POST /auth/login for local authentication mode.
func (rt *Router) handleLocalLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, i18n.T(i18n.LangFromRequest(r), "invalid_request"))
		return
	}

	ip := r.RemoteAddr

	if rt.localAuth == nil || !rt.localAuth.Verify(req.Username, req.Password) {
		if auditErr := rt.auditStore.Record(auth.EventLoginFailed, req.Username, "{}", ip); auditErr != nil {
			slog.Error("failed to record login failed audit", "error", auditErr)
		}
		writeError(w, http.StatusUnauthorized, i18n.T(i18n.LangFromRequest(r), "login_failed"))
		return
	}

	rt.sessions.SetCookie(w, &auth.SessionData{
		Email:     req.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	if auditErr := rt.auditStore.Record(auth.EventLogin, req.Username, "{}", ip); auditErr != nil {
		slog.Error("failed to record login audit", "error", auditErr)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "email": req.Username})
}

// extractProvider extracts the provider name from a path like /auth/{provider}/login.
func extractProvider(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/auth/"), "/")
	if len(parts) >= 1 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// buildRedirectURI constructs the OAuth callback redirect URI from the request.
func buildRedirectURI(r *http.Request, provider string) string {
	scheme := "https"
	if r.TLS == nil {
		if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
			scheme = fwd
		} else {
			scheme = "http"
		}
	}
	host := r.Host
	return fmt.Sprintf("%s://%s/auth/%s/callback", scheme, host, provider)
}
