package api

import (
	"net/http"
	"strings"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/config"
	"forsaken-mail/internal/i18n"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
	"forsaken-mail/internal/smtp"
	"forsaken-mail/internal/webhook"
	"forsaken-mail/internal/ws"
)

// Router holds all dependencies and registers HTTP routes.
type Router struct {
	cfg        *config.Config
	sessions   *auth.SessionManager
	authMW     *auth.Middleware
	mailStore  *mail.Store
	settings   *settings.Store
	auditStore *audit.Store
	hub        *ws.Hub
	webhook    *webhook.Sender
	provider          auth.Provider
	localAuth         *auth.LocalAuth // non-nil only when AUTH_MODE=local
	startTime         time.Time
	domainTestLimiter *smtp.RateLimiter
}

// NewRouter creates a new Router with the given dependencies.
func NewRouter(
	cfg *config.Config,
	sessions *auth.SessionManager,
	authMW *auth.Middleware,
	mailStore *mail.Store,
	settings *settings.Store,
	auditStore *audit.Store,
	hub *ws.Hub,
	webhookSender *webhook.Sender,
	localAuth *auth.LocalAuth,
) *Router {
	var provider auth.Provider
	if cfg.AuthMode == "oauth" {
		provider, _ = auth.NewProvider(cfg)
	}
	return &Router{
		cfg:               cfg,
		sessions:          sessions,
		authMW:            authMW,
		mailStore:         mailStore,
		settings:          settings,
		auditStore:        auditStore,
		hub:               hub,
		webhook:           webhookSender,
		provider:          provider,
		localAuth:         localAuth,
		startTime:         time.Now().UTC(),
		domainTestLimiter: smtp.NewRateLimiter(10.0/60.0, 10), // 10 requests per minute, burst 10
	}
}

// Handler returns the http.Handler with all routes registered.
func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public routes (no auth required).
	mux.HandleFunc("/auth/", rt.routeAuth)
	mux.HandleFunc("/api/health", rt.handleHealth)

	// Protected routes (auth middleware applied).
	mux.Handle("/api/config", rt.authMW.Wrap(http.HandlerFunc(rt.handleConfig)))
	mux.Handle("/api/mails", rt.authMW.Wrap(http.HandlerFunc(rt.handleMails)))
	mux.Handle("/api/mails/", rt.authMW.Wrap(http.HandlerFunc(rt.routeMailsSubpath)))
	mux.Handle("/api/emails/", rt.authMW.Wrap(http.HandlerFunc(rt.handleEmails)))
	mux.Handle("/api/domain-test", rt.authMW.Wrap(http.HandlerFunc(rt.handleDomainTest)))
	mux.Handle("/api/webhook/test", rt.authMW.Wrap(http.HandlerFunc(rt.handleWebhookTest)))
	mux.Handle("/api/test-email", rt.authMW.Wrap(http.HandlerFunc(rt.handleTestEmail)))
	mux.Handle("/ws", rt.authMW.Wrap(http.HandlerFunc(rt.handleWS)))
	mux.Handle("/api/admin/audit-logs", rt.authMW.Wrap(http.HandlerFunc(rt.handleAuditLogs)))
	mux.Handle("/api/admin/settings", rt.authMW.Wrap(http.HandlerFunc(rt.routeAdminSettings)))
	mux.Handle("/api/admin/status", rt.authMW.Wrap(http.HandlerFunc(rt.handleStatus)))

	// Apply security headers middleware to all routes.
	return securityHeaders(mux)
}

// routeAuth dispatches /auth/* routes based on the path suffix.
func (rt *Router) routeAuth(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /auth/logout
	if path == "/auth/logout" {
		rt.handleLogout(w, r)
		return
	}

	// AUTH_MODE=local: /auth/login
	if rt.cfg.AuthMode == "local" && path == "/auth/login" {
		rt.handleLocalLogin(w, r)
		return
	}

	// AUTH_MODE=oauth: /auth/{provider}/login or /auth/{provider}/callback
	if rt.cfg.AuthMode == "oauth" && len(path) > len("/auth/") {
		remainder := path[len("/auth/"):]
		if len(remainder) > 0 {
			// Find the action part after the provider.
			slashIdx := -1
			for i, c := range remainder {
				if c == '/' {
					slashIdx = i
					break
				}
			}
			if slashIdx >= 0 {
				action := remainder[slashIdx+1:]
				switch action {
				case "login":
					rt.handleOAuthLogin(w, r)
					return
				case "callback":
					rt.handleOAuthCallback(w, r)
					return
				}
			}
		}
	}

	http.NotFound(w, r)
}

// routeMailsSubpath dispatches /api/mails/{id}/read.
func (rt *Router) routeMailsSubpath(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/read") {
		rt.handleMailRead(w, r)
		return
	}
	http.NotFound(w, r)
}

// routeAdminSettings dispatches /api/admin/settings based on HTTP method.
func (rt *Router) routeAdminSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rt.handleGetSettings(w, r)
	case http.MethodPut:
		rt.handleUpdateSettings(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, i18n.T(i18n.LangFromRequest(r), "method_not_allowed"))
	}
}

// securityHeaders wraps an http.Handler with security headers.
func securityHeaders(next http.Handler) http.Handler {
	csp := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' ws: wss:"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r)
	})
}
