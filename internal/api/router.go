package api

import (
	"log/slog"
	"net/http"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/config"
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
		var err error
		provider, err = auth.NewProvider(cfg)
		if err != nil {
			slog.Error("failed to create OAuth provider", "error", err)
		}
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
	mux.Handle("/api/config", rt.authMW.OptionalAuth(http.HandlerFunc(rt.handleConfig)))

	// Protected routes (auth middleware applied).
	mux.Handle("GET /api/mails", rt.authMW.Wrap(http.HandlerFunc(rt.handleMails)))
	mux.Handle("GET /api/mails/recent", rt.authMW.Wrap(http.HandlerFunc(rt.handleRecentMails)))
	mux.Handle("GET /api/mails/all", rt.authMW.Wrap(http.HandlerFunc(rt.handleAllMails)))
	mux.Handle("GET /api/mails/filters", rt.authMW.Wrap(http.HandlerFunc(rt.handleMailFilters)))
	mux.Handle("GET /api/mails/{id}", rt.authMW.Wrap(http.HandlerFunc(rt.handleGetMail)))
	mux.Handle("PUT /api/mails/{id}/read", rt.authMW.Wrap(http.HandlerFunc(rt.handleMailRead)))
	mux.Handle("GET /api/emails/", rt.authMW.Wrap(http.HandlerFunc(rt.handleEmails)))
	mux.Handle("POST /api/domain-test", rt.authMW.Wrap(http.HandlerFunc(rt.handleDomainTest)))
	mux.Handle("POST /api/webhook/test", rt.authMW.Wrap(http.HandlerFunc(rt.handleWebhookTest)))
	mux.Handle("POST /api/test-email", rt.authMW.Wrap(http.HandlerFunc(rt.handleTestEmail)))
	mux.Handle("GET /ws", rt.authMW.Wrap(http.HandlerFunc(rt.handleWS)))
	mux.Handle("GET /api/admin/audit-logs", rt.authMW.Wrap(http.HandlerFunc(rt.handleAuditLogs)))
	mux.Handle("GET /api/admin/settings", rt.authMW.Wrap(http.HandlerFunc(rt.handleGetSettings)))
	mux.Handle("PUT /api/admin/settings", rt.authMW.Wrap(http.HandlerFunc(rt.handleUpdateSettings)))
	mux.Handle("GET /api/admin/status", rt.authMW.Wrap(http.HandlerFunc(rt.handleStatus)))

	// Apply security headers and CSRF middleware to all routes.
	return csrfMiddleware(securityHeaders(mux))
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
