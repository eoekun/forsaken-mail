package mail

import (
	"context"
	"log/slog"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/settings"
)

const cleanupInterval = 5 * time.Minute

// StartCleanup runs a periodic goroutine that cleans up old mails and audit
// logs based on settings. It blocks until the context is cancelled.
func StartCleanup(ctx context.Context, mailStore *Store, auditStore *audit.Store, settingsStore *settings.Service) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	// Run once immediately on startup.
	runCleanup(mailStore, auditStore, settingsStore)

	for {
		select {
		case <-ctx.Done():
			slog.Info("cleanup goroutine stopping")
			return
		case <-ticker.C:
			runCleanup(mailStore, auditStore, settingsStore)
		}
	}
}

func runCleanup(mailStore *Store, auditStore *audit.Store, settingsStore *settings.Service) {
	values, err := settingsStore.Load()
	if err != nil {
		slog.Error("failed to load runtime settings for cleanup", "error", err)
		return
	}

	mailHours := values.MailRetentionHours
	mailMaxCount := values.MailMaxCount
	auditDays := values.AuditRetentionDays
	auditMaxCount := values.AuditMaxCount

	if mailHours > 0 {
		if err := mailStore.CleanupByAge(mailHours); err != nil {
			slog.Error("mail cleanup by age failed", "hours", mailHours, "error", err)
		}
	}
	if mailMaxCount > 0 {
		if err := mailStore.CleanupByCount(mailMaxCount); err != nil {
			slog.Error("mail cleanup by count failed", "max_count", mailMaxCount, "error", err)
		}
	}
	if auditDays > 0 {
		if err := auditStore.CleanupByAge(auditDays); err != nil {
			slog.Error("audit cleanup by age failed", "days", auditDays, "error", err)
		}
	}
	if auditMaxCount > 0 {
		if err := auditStore.CleanupByCount(auditMaxCount); err != nil {
			slog.Error("audit cleanup by count failed", "max_count", auditMaxCount, "error", err)
		}
	}

	slog.Debug("cleanup completed",
		"mail_hours", mailHours,
		"mail_max_count", mailMaxCount,
		"audit_days", auditDays,
		"audit_max_count", auditMaxCount,
	)
}
