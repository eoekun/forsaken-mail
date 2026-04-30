package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"forsaken-mail/internal/api"
	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/auth"
	"forsaken-mail/internal/config"
	"forsaken-mail/internal/logger"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
	fsmtp "forsaken-mail/internal/smtp"
	"forsaken-mail/internal/webhook"
	"forsaken-mail/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger.Setup(cfg)

	slog.Info("forsaken-mail starting",
		"port", cfg.Port,
		"mail_host", cfg.MailHost,
		"smtp_addr", fmt.Sprintf("%s:%d", cfg.MailinHost, cfg.MailinPort),
	)

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	settingsStore := settings.NewStore(db)
	if err := settingsStore.Init(); err != nil {
		log.Fatalf("failed to init settings store: %v", err)
	}

	auditStore := audit.NewStore(db)
	if err := auditStore.Init(); err != nil {
		log.Fatalf("failed to init audit store: %v", err)
	}

	mailStore := mail.NewStore(db)
	if err := mailStore.Init(); err != nil {
		log.Fatalf("failed to init mail store: %v", err)
	}

	if err := settingsStore.SeedFromEnv(cfg); err != nil {
		log.Fatalf("failed to seed settings: %v", err)
	}

	blacklistStr, err := settingsStore.Get("keyword_blacklist")
	if err != nil {
		log.Fatalf("failed to read keyword_blacklist: %v", err)
	}
	var blacklist []string
	if blacklistStr != "" {
		for _, kw := range strings.Split(blacklistStr, ",") {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				blacklist = append(blacklist, kw)
			}
		}
	}

	hub := ws.NewHub(blacklist, cfg.MailHost)

	webhookSender := webhook.NewSender(settingsStore)

	router := mail.NewRouter(mailStore, hub, settingsStore, auditStore, webhookSender)

	smtpAddr := fmt.Sprintf("%s:%d", cfg.MailinHost, cfg.MailinPort)
	limiter := fsmtp.NewRateLimiter(10, 20)
	smtpServer := fsmtp.New(router, limiter, settingsStore, auditStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mail.StartCleanup(ctx, mailStore, auditStore, settingsStore)
	go hub.Run(ctx)

	// Auth
	sessions := auth.NewSessionManager(cfg.SessionSecret)
	authMW := auth.NewMiddleware(sessions, settingsStore)

	// API router
	apiRouter := api.NewRouter(cfg, sessions, authMW, mailStore, settingsStore, auditStore, hub, webhookSender)

	// Combine API routes with static file serving for the SPA.
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	httpServer.Handler = spaHandler(apiRouter.Handler(), "embed")

	go func() {
		slog.Info("HTTP server starting", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := smtpServer.ListenAndServe(smtpAddr); err != nil {
			log.Fatalf("SMTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("received shutdown signal", "signal", sig)

	cancel()

	if err := httpServer.Shutdown(context.Background()); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	if err := smtpServer.Close(); err != nil {
		slog.Error("SMTP server shutdown error", "error", err)
	}

	hub.Close()

	if err := db.Close(); err != nil {
		slog.Error("database close error", "error", err)
	}

	slog.Info("forsaken-mail stopped")
}

// spaHandler wraps an API handler with static file serving for the SPA.
// Requests for existing files in the staticDir are served directly.
// All other requests fall through to the API handler, and if that returns
// 404, the SPA's index.html is served instead.
func spaHandler(apiHandler http.Handler, staticDir string) http.Handler {
	staticFS := os.DirFS(staticDir)
	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API and auth routes go directly to the API handler.
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") || r.URL.Path == "/ws" {
			apiHandler.ServeHTTP(w, r)
			return
		}

		// Try to serve a static file.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := staticFS.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
