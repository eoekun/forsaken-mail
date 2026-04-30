package main

import (
	"context"
	"database/sql"
	"encoding/json"
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

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/config"
	"forsaken-mail/internal/logger"
	"forsaken-mail/internal/mail"
	"forsaken-mail/internal/settings"
	fsmtp "forsaken-mail/internal/smtp"
	"forsaken-mail/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 1. Load config.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 2. Setup logger.
	logger.Setup(cfg)

	slog.Info("forsaken-mail starting",
		"port", cfg.Port,
		"mail_host", cfg.MailHost,
		"smtp_addr", fmt.Sprintf("%s:%d", cfg.MailinHost, cfg.MailinPort),
	)

	// 3. Open SQLite database.
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1) // SQLite supports a single writer.
	defer db.Close()

	// 4. Initialize stores.
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

	// 5. Seed settings from environment.
	if err := settingsStore.SeedFromEnv(cfg); err != nil {
		log.Fatalf("failed to seed settings: %v", err)
	}

	// Load blacklist from settings for the WebSocket hub.
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

	// 6. Create WebSocket hub.
	hub := ws.NewHub(blacklist, cfg.MailHost)

	// 7. Create mail router (webhook sender is nil for now; Phase 2 adds it).
	router := mail.NewRouter(mailStore, hub, settingsStore, auditStore, nil)

	// 8. Create SMTP server with rate limiter.
	smtpAddr := fmt.Sprintf("%s:%d", cfg.MailinHost, cfg.MailinPort)
	limiter := fsmtp.NewRateLimiter(10, 20) // 10 req/s, burst 20 per IP
	smtpServer := fsmtp.New(router, limiter, settingsStore, auditStore)

	// 9. Start cleanup goroutine.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go mail.StartCleanup(ctx, mailStore, auditStore, settingsStore)

	// 10. Start hub event loop.
	go hub.Run(ctx)

	// 11. Setup HTTP routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/ws", hub.HandleWS)

	// 12. Start HTTP server.
	httpAddr := fmt.Sprintf(":%d", cfg.Port)
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		slog.Info("HTTP server starting", "addr", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start SMTP server.
	go func() {
		if err := smtpServer.ListenAndServe(smtpAddr); err != nil {
			log.Fatalf("SMTP server error: %v", err)
		}
	}()

	// 13. Graceful shutdown on SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("received shutdown signal", "signal", sig)

	// Cancel context to stop background goroutines.
	cancel()

	// Shutdown HTTP server.
	if err := httpServer.Shutdown(context.Background()); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	// Close SMTP server.
	if err := smtpServer.Close(); err != nil {
		slog.Error("SMTP server shutdown error", "error", err)
	}

	// Close WebSocket hub.
	hub.Close()

	// Close database.
	if err := db.Close(); err != nil {
		slog.Error("database close error", "error", err)
	}

	slog.Info("forsaken-mail stopped")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
