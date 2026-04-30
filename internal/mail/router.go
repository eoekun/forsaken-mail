package mail

import (
	"log/slog"
	"strings"
	"time"

	"forsaken-mail/internal/audit"
	"forsaken-mail/internal/settings"
	"forsaken-mail/internal/ws"
)

// WebhookSender is the interface for sending webhook notifications.
// Implementations live in the webhook package.
type WebhookSender interface {
	Send(from, to, subject, text string)
}

// Router connects SMTP receipt to storage, WebSocket push, and webhook notification.
type Router struct {
	mailStore  *Store
	hub        *ws.Hub
	settings   *settings.Store
	auditStore *audit.Store
	webhook    WebhookSender
}

// NewRouter creates a new Router with the given dependencies.
func NewRouter(mailStore *Store, hub *ws.Hub, settings *settings.Store, auditStore *audit.Store, webhook WebhookSender) *Router {
	return &Router{
		mailStore:  mailStore,
		hub:        hub,
		settings:   settings,
		auditStore: auditStore,
		webhook:    webhook,
	}
}

// Handle processes an incoming mail: saves it, pushes via WebSocket, records an
// audit event, and sends a webhook notification for each valid recipient.
func (r *Router) Handle(from string, toList []string, subject, textBody, htmlBody string, rawSize int64) {
	mailHost, err := r.settings.Get("mail_host")
	if err != nil {
		slog.Error("failed to get mail_host setting", "error", err)
		return
	}
	mailHost = strings.ToLower(strings.TrimSpace(mailHost))

	for _, addr := range toList {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		shortID, domain := extractShortID(addr)
		if shortID == "" || domain == "" {
			slog.Warn("invalid recipient address", "addr", addr)
			continue
		}
		if domain != mailHost {
			slog.Warn("recipient domain does not match mail_host", "addr", addr, "mail_host", mailHost)
			continue
		}

		// Save to database.
		m := &Mail{
			ShortID:  shortID,
			FromAddr: from,
			ToAddr:   addr,
			Subject:  subject,
			TextBody: textBody,
			HTMLBody: htmlBody,
			RawSize:  rawSize,
		}
		if err := r.mailStore.Save(m); err != nil {
			slog.Error("failed to save mail", "short_id", shortID, "error", err)
			continue
		}

		// Push to WebSocket clients watching this shortID.
		r.hub.SendTo(shortID, ws.MailData{
			ID:        m.ID,
			From:      from,
			To:        addr,
			Subject:   subject,
			HTML:      htmlBody,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		})

		// Record audit event.
		if err := r.auditStore.Record("MAIL_RECEIVED", addr, from, ""); err != nil {
			slog.Error("failed to record audit event", "error", err)
		}

		// Send webhook notification asynchronously.
		if r.webhook != nil {
			go r.webhook.Send(from, addr, subject, textBody)
		}
	}
}

// extractShortID parses an email address into the local part (shortID) and
// domain. Returns empty strings if the address is malformed.
func extractShortID(addr string) (shortID, domain string) {
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return "", ""
	}
	shortID = strings.TrimSpace(parts[0])
	domain = strings.ToLower(strings.TrimSpace(parts[1]))
	if shortID == "" || domain == "" {
		return "", ""
	}
	return shortID, domain
}
