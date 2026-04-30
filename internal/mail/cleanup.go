package mail

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/settings"
)

const cleanupInterval = 5 * time.Minute

// StartCleanup runs a periodic goroutine that cleans up old mails and audit
// logs based on settings. It blocks until the context is cancelled.
func StartCleanup(ctx context.Context, mailStore *Store, auditStore *audit.Store, settingsStore *settings.Store) {
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

func runCleanup(mailStore *Store, auditStore *audit.Store, settingsStore *settings.Store) {
	mailHours := getSettingInt(settingsStore, "mail_retention_hours", 1)
	mailMaxCount := getSettingInt(settingsStore, "mail_max_count", 100)
	auditDays := getSettingInt(settingsStore, "audit_retention_days", 7)
	auditMaxCount := getSettingInt(settingsStore, "audit_max_count", 5000)

	if err := mailStore.CleanupByAge(mailHours); err != nil {
		slog.Error("mail cleanup by age failed", "hours", mailHours, "error", err)
	}
	if err := mailStore.CleanupByCount(mailMaxCount); err != nil {
		slog.Error("mail cleanup by count failed", "max_count", mailMaxCount, "error", err)
	}
	if err := auditStore.CleanupByAge(auditDays); err != nil {
		slog.Error("audit cleanup by age failed", "days", auditDays, "error", err)
	}
	if err := auditStore.CleanupByCount(auditMaxCount); err != nil {
		slog.Error("audit cleanup by count failed", "max_count", auditMaxCount, "error", err)
	}

	slog.Debug("cleanup completed",
		"mail_hours", mailHours,
		"mail_max_count", mailMaxCount,
		"audit_days", auditDays,
		"audit_max_count", auditMaxCount,
	)
}

// getSettingInt reads a setting by key and converts it to int. Returns the
// default value if the setting is empty or not a valid integer.
func getSettingInt(s *settings.Store, key string, defaultVal int) int {
	v, err := s.Get(key)
	if err != nil {
		slog.Warn("failed to read setting, using default", "key", key, "default", defaultVal, "error", err)
		return defaultVal
	}
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid setting value, using default", "key", key, "value", v, "default", defaultVal)
		return defaultVal
	}
	return n
}
