package api

import (
	"log/slog"
	"net/http"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/config"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/service"
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
	settings   *settings.Service
	auditStore *audit.Store
	hub        *ws.Hub
	admin            *service.AdminService
	publicConfig     *service.PublicConfigService
	webhookService   *service.WebhookService
	provider         auth.Provider
	localAuth        *auth.LocalAuth // non-nil only when AUTH_MODE=local
	startTime        time.Time
	domainTestLimiter *smtp.RateLimiter
}

// NewRouter creates a new Router with the given dependencies.
func NewRouter(
	cfg *config.Config,
	sessions *auth.SessionManager,
	authMW *auth.Middleware,
	mailStore *mail.Store,
	settings *settings.Service,
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
	startTime := time.Now().UTC()
	return &Router{
		cfg:               cfg,
		sessions:          sessions,
		authMW:            authMW,
		mailStore:         mailStore,
		settings:          settings,
		auditStore:        auditStore,
		hub:               hub,
		admin:             service.NewAdminService(settings, auditStore, mailStore, hub, startTime, cfg.DBPath),
		publicConfig:      service.NewPublicConfigService(settings, cfg.AuthMode),
		webhookService:    service.NewWebhookService(webhookSender),
		provider:          provider,
		localAuth:         localAuth,
		startTime:         startTime,
		domainTestLimiter: smtp.NewRateLimiter(10.0/60.0, 10), // 10 requests per minute, burst 10
	}
}

// Handler returns the http.Handler with all routes registered.
func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	// Public routes (no auth required).
	mux.HandleFunc("GET /auth/logout", rt.handleLogout)
	mux.HandleFunc("GET /api/health", rt.handleHealth)
	mux.Handle("GET /api/config", rt.authMW.OptionalAuth(http.HandlerFunc(rt.handleConfig)))

	// Auth routes (method-aware, conditional on AUTH_MODE).
	if rt.cfg.AuthMode == "local" {
		mux.HandleFunc("POST /auth/login", rt.handleLocalLogin)
	} else {
		mux.HandleFunc("GET /auth/{provider}/login", rt.handleOAuthLogin)
		mux.HandleFunc("GET /auth/{provider}/callback", rt.handleOAuthCallback)
	}

	// Protected routes (auth middleware applied).
	mux.Handle("GET /api/mails", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleMails)))
	mux.Handle("POST /api/mails/reextract", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleReextractMails)))
	mux.Handle("GET /api/mails/recent", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleRecentMails)))
	mux.Handle("GET /api/mails/all", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleAllMails)))
	mux.Handle("GET /api/mails/filters", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleMailFilters)))
	mux.Handle("GET /api/mails/{id}", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleGetMail)))
	mux.Handle("PUT /api/mails/{id}/read", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleMailRead)))
	mux.Handle("GET /api/emails/{shortId}", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleListEmailsByShortID)))
	mux.Handle("DELETE /api/emails/{shortId}", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleDeleteAllEmailsByShortID)))
	mux.Handle("GET /api/emails/{shortId}/{mailId}", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleGetEmailByPath)))
	mux.Handle("DELETE /api/emails/{shortId}/{mailId}", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleDeleteEmailByPath)))
	mux.Handle("GET /api/domain-test", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleDomainTest)))
	mux.Handle("POST /api/webhook/test", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleWebhookTest)))
	mux.Handle("POST /api/test-email", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleTestEmail)))
	mux.Handle("GET /ws", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleWS)))
	mux.Handle("GET /api/admin/audit-logs", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleAuditLogs)))
	mux.Handle("GET /api/admin/settings", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleGetSettings)))
	mux.Handle("PUT /api/admin/settings", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleUpdateSettings)))
	mux.Handle("GET /api/admin/status", rt.authMW.RequireAPIAuth(http.HandlerFunc(rt.handleStatus)))

	// Apply security headers and CSRF middleware to all routes.
	return csrfMiddleware(securityHeaders(mux))
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
